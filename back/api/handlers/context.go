package handlers

import (
	"context"

	"back/infrastructure/auth"
)

type contextKey string

const claimsKey contextKey = "claims"

func WithClaims(ctx context.Context, claims *auth.Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

func GetClaims(ctx context.Context) *auth.Claims {
	claims, _ := ctx.Value(claimsKey).(*auth.Claims)
	return claims
}
