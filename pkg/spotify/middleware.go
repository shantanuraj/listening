package spotify

import (
	"log"
	"net/http"
)

func (c *Client) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if !c.IsAuthenticated() {
			if c.IsTokenExpired() {
				if err := c.RefreshToken(ctx); err != nil {
					log.Printf("auth: failed to refresh token: %v", err)
					http.Error(w, "failed to refresh token", http.StatusInternalServerError)
					return
				}
			} else {
				http.Error(w, "not authenticated", http.StatusUnauthorized)
				return
			}
		}

		next(w, r)
	}
}
