package health

import "context"

type Repository interface {
	Get(ctx context.Context) (string, error)
}
