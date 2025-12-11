package utils

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("secret")

func SetJWTKey(key string) {
	jwtKey = []byte(key)
}

type ContextJWTKey struct{}

func GetClaimsFromContext(ctx context.Context) jwt.MapClaims {
	if ctx == nil {
		return nil
	}
	v := ctx.Value(ContextJWTKey{})
	if v == nil {
		return nil
	}
	claims, ok := v.(jwt.MapClaims)
	if !ok {
		return nil
	}
	return claims
}

func GetUserIDFromClaims(claims jwt.MapClaims) int64 {
	if claims == nil {
		return 0
	}

	if sub, ok := claims["sub"].(float64); ok {
		return int64(sub)
	}

	if s, ok := claims["sub"].(string); ok {
		var id int64
		_, err := fmt.Sscan(s, &id)
		if err == nil {
			return id
		}
	}
	return 0
}

func GenerateToken(userID int64, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtKey)
}

func ParseToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	return claims, nil
}

func ParseJWTFromRequest(r *http.Request) (jwt.MapClaims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("missing Authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, errors.New("invalid Authorization header")
	}

	return ParseToken(parts[1])
}
