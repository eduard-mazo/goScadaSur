// pkg/api/auth.go
package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("goscadasur_secure_key_2026") // Reemplazar con config persistente

// Claims representa el payload del token JWT
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken genera un nuevo token JWT
func GenerateToken(userID uint, username, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// AuthMiddleware intercepta las peticiones para validar el token
func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Si no hay base de datos conectada, permitir todo (Modo Lite)
		if s.PGClient == nil {
			next.ServeHTTP(w, r)
			return
		}

		// Solo proteger rutas que empiecen con /api/
		// Omitir login/register de la validación
		if !strings.HasPrefix(r.URL.Path, "/api/") || 
			r.URL.Path == "/api/auth/login" || 
			r.URL.Path == "/api/auth/register" {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			s.ErrorResponse(w, "Autorización requerida", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		})

		if err != nil || !token.Valid {
			s.ErrorResponse(w, "Token inválido o expirado", http.StatusUnauthorized)
			return
		}

		// Inyectar el UserID en el contexto para uso posterior
		// Context propagation... (implementación simplificada para el ejemplo)
		next.ServeHTTP(w, r)
	})
}
