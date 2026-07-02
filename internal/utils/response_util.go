package utils

import (
	"errors"
	"strconv"
)

const (
	defaultPage  = 1
	defaultLimit = 10
	maxLimit     = 100
)

type PaginationFilter struct {
	Page  int
	Limit int
}

func NewPaginationFilter(pageValue, limitValue string) (PaginationFilter, error) {
	page, err := queryInt(pageValue, "page", defaultPage)
	if err != nil {
		return PaginationFilter{}, err
	}
	limit, err := queryInt(limitValue, "limit", defaultLimit)
	if err != nil {
		return PaginationFilter{}, err
	}
	if page < 1 {
		return PaginationFilter{}, errors.New("page must be greater than 0")
	}
	if limit < 1 {
		return PaginationFilter{}, errors.New("limit must be greater than 0")
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	return PaginationFilter{Page: page, Limit: limit}, nil
}

func (f PaginationFilter) Offset() int {
	return (f.Page - 1) * f.Limit
}

func queryInt(value, key string, fallback int) (int, error) {
	if value == "" {
		return fallback, nil
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.New(key + " must be a number")
	}

	return result, nil
}
