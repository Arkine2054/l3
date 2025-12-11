package middlewares

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"gitlab.com/arkine/l3/7/internal/repository"
	"gitlab.com/arkine/l3/7/internal/utils"
)

type ctxKey string

const RoleKey ctxKey = "role"

func JWTAuth(repo *repository.Repo) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			claims, err := utils.ParseJWTFromRequest(r)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			uid := utils.GetUserIDFromClaims(claims)

			role := claims["role"].(string)

			ctx := context.WithValue(r.Context(), utils.ContextJWTKey{}, claims)
			ctx = context.WithValue(ctx, RoleKey, role)

			if uid > 0 {
				log.Println("[JWTAuth] Setting app.current_user_id =", uid)

				_, err = repo.DB.ExecContext(
					ctx,
					"SELECT set_config('app.current_user_id', $1, false)",
					strconv.FormatInt(uid, 10),
				)

				if err != nil {
					log.Println("[JWTAuth] ERROR setting session variable:", err)
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool)
	for _, r := range roles {
		allowed[r] = true
	}

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(RoleKey).(string)

			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !allowed[role] {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
