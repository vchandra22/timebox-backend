package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"timebox-backend/internal/entity"
	authrepo "timebox-backend/internal/repository/auth"
	userrepo "timebox-backend/internal/repository/user"

	"golang.org/x/crypto/bcrypt"
)

const (
	defaultAccessTTL  = 15 * time.Minute
	defaultRefreshTTL = 30 * 24 * time.Hour
	tokenTypeBearer   = "Bearer"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidPassword    = errors.New("password must contain uppercase, lowercase, number, and symbol")
	ErrInvalidTimezone    = errors.New("invalid timezone")
	ErrInvalidToken       = errors.New("invalid token")
	ErrRateLimited        = errors.New("too many login attempts")
)

type AuthService struct {
	authRepo authrepo.Repository
	userRepo userrepo.Repository
	options  AuthOptions
}

type AuthOptions struct {
	Secret            string
	AccessTTLSeconds  int
	RefreshTTLSeconds int
}

type TokenSet struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int64
}

type tokenClaims struct {
	Subject   string `json:"sub"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
	Type      string `json:"typ"`
	ID        string `json:"jti"`
}

func newAuthService(repo authrepo.Repository, userRepo userrepo.Repository, options AuthOptions) *AuthService {
	return &AuthService{
		authRepo: repo,
		userRepo: userRepo,
		options:  options,
	}
}

func (s *AuthService) Register(ctx context.Context, user entity.User, password string) (entity.User, TokenSet, error) {
	if err := validatePassword(password); err != nil {
		return entity.User{}, TokenSet{}, err
	}
	if _, err := time.LoadLocation(user.Timezone); err != nil {
		return entity.User{}, TokenSet{}, ErrInvalidTimezone
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return entity.User{}, TokenSet{}, err
	}

	user.PasswordHash = string(hash)
	created, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return entity.User{}, TokenSet{}, userError(err)
	}

	tokens, err := s.issueTokens(ctx, created.ID)
	return created, tokens, err
}

func (s *AuthService) Login(ctx context.Context, email, password string) (entity.User, TokenSet, error) {
	if err := s.checkLoginRateLimit(ctx, email); err != nil {
		return entity.User{}, TokenSet{}, err
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, userrepo.ErrNotFound) {
			return entity.User{}, TokenSet{}, ErrInvalidCredentials
		}
		return entity.User{}, TokenSet{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return entity.User{}, TokenSet{}, ErrInvalidCredentials
	}

	tokens, err := s.issueTokens(ctx, user.ID)
	return user, tokens, err
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (TokenSet, error) {
	claims, err := s.parseToken(refreshToken, "refresh")
	if err != nil {
		return TokenSet{}, err
	}

	userID, err := s.authRepo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, authrepo.ErrRefreshTokenNotFound) {
			return TokenSet{}, ErrInvalidToken
		}
		return TokenSet{}, err
	}
	if userID != claims.Subject {
		return TokenSet{}, ErrInvalidToken
	}

	if err := s.authRepo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return TokenSet{}, err
	}
	return s.issueTokens(ctx, claims.Subject)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if _, err := s.parseToken(refreshToken, "refresh"); err != nil {
		return err
	}
	return s.authRepo.DeleteRefreshToken(ctx, refreshToken)
}

func (s *AuthService) ValidateAccessToken(accessToken string) (string, error) {
	claims, err := s.parseToken(accessToken, "access")
	if err != nil {
		return "", err
	}
	return claims.Subject, nil
}

func (s *AuthService) issueTokens(ctx context.Context, userID string) (TokenSet, error) {
	accessTTL := s.accessTTL()
	refreshTTL := s.refreshTTL()

	accessToken, err := s.signToken(userID, "access", accessTTL)
	if err != nil {
		return TokenSet{}, err
	}
	refreshToken, err := s.signToken(userID, "refresh", refreshTTL)
	if err != nil {
		return TokenSet{}, err
	}
	if err := s.authRepo.SaveRefreshToken(ctx, refreshToken, userID, refreshTTL); err != nil {
		return TokenSet{}, err
	}

	return TokenSet{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    tokenTypeBearer,
		ExpiresIn:    int64(accessTTL.Seconds()),
	}, nil
}

func (s *AuthService) signToken(userID, kind string, ttl time.Duration) (string, error) {
	now := time.Now()
	jti, err := randomHex(16)
	if err != nil {
		return "", err
	}
	header, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	payload, err := json.Marshal(tokenClaims{
		Subject:   userID,
		ExpiresAt: now.Add(ttl).Unix(),
		IssuedAt:  now.Unix(),
		Type:      kind,
		ID:        jti,
	})
	if err != nil {
		return "", err
	}

	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	return unsigned + "." + s.signature(unsigned), nil
}

func (s *AuthService) parseToken(token, kind string) (tokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return tokenClaims{}, ErrInvalidToken
	}
	unsigned := parts[0] + "." + parts[1]
	if !hmac.Equal([]byte(parts[2]), []byte(s.signature(unsigned))) {
		return tokenClaims{}, ErrInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return tokenClaims{}, ErrInvalidToken
	}
	var claims tokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return tokenClaims{}, ErrInvalidToken
	}
	if claims.Type != kind || claims.Subject == "" || claims.ExpiresAt < time.Now().Unix() {
		return tokenClaims{}, ErrInvalidToken
	}
	return claims, nil
}

func (s *AuthService) signature(unsigned string) string {
	mac := hmac.New(sha256.New, []byte(s.options.Secret))
	mac.Write([]byte(unsigned))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *AuthService) checkLoginRateLimit(ctx context.Context, email string) error {
	count, err := s.authRepo.IncrementLoginAttempt(ctx, email, time.Minute)
	if err != nil {
		return err
	}
	if count > 5 {
		return ErrRateLimited
	}
	return nil
}

func (s *AuthService) accessTTL() time.Duration {
	if s.options.AccessTTLSeconds > 0 {
		return time.Duration(s.options.AccessTTLSeconds) * time.Second
	}
	return defaultAccessTTL
}

func (s *AuthService) refreshTTL() time.Duration {
	if s.options.RefreshTTLSeconds > 0 {
		return time.Duration(s.options.RefreshTTLSeconds) * time.Second
	}
	return defaultRefreshTTL
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return ErrInvalidPassword
	}

	var hasUpper, hasLower, hasDigit, hasSymbol bool
	for _, c := range password {
		switch {
		case 'A' <= c && c <= 'Z':
			hasUpper = true
		case 'a' <= c && c <= 'z':
			hasLower = true
		case '0' <= c && c <= '9':
			hasDigit = true
		default:
			hasSymbol = true
		}
	}
	if hasUpper && hasLower && hasDigit && hasSymbol {
		return nil
	}
	return ErrInvalidPassword
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
