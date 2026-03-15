package auth

import (
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kidandcat/hogar/internal/db"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret []byte

func init() {
	secret := os.Getenv("HOGAR_JWT_SECRET")
	if secret == "" {
		secret = "hogar-default-secret-change-me"
	}
	jwtSecret = []byte(secret)
}

const (
	cookieName = "hogar_session"
	jwtExpiry  = 30 * 24 * time.Hour // 30 days
)

// EnsureAdminUser creates the admin user if it doesn't exist.
func EnsureAdminUser(database *db.DB) error {
	user, err := database.GetUserByUsername("admin")
	if err != nil {
		return err
	}
	if user != nil {
		return nil // already exists
	}

	password := os.Getenv("HOGAR_PASSWORD")
	if password == "" {
		password = "hogar123"
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return database.CreateUser("admin", string(hashed))
}

// Login validates credentials and sets a session cookie.
func Login(database *db.DB, w http.ResponseWriter, username, password string) bool {
	user, err := database.GetUserByUsername(username)
	if err != nil || user == nil {
		return false
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return false
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.Username,
		"exp": time.Now().Add(jwtExpiry).Unix(),
		"iat": time.Now().Unix(),
	})

	tokenStr, err := token.SignedString(jwtSecret)
	if err != nil {
		return false
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    tokenStr,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(jwtExpiry.Seconds()),
	})
	return true
}

// Logout clears the session cookie.
func Logout(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

// ValidateRequest checks the JWT cookie and returns the username if valid.
func ValidateRequest(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return "", false
	}

	token, err := jwt.Parse(cookie.Value, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid {
		return "", false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", false
	}

	sub, err := claims.GetSubject()
	if err != nil {
		return "", false
	}
	return sub, true
}

// Middleware protects routes that require authentication.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow login page and static assets without auth
		path := r.URL.Path
		if path == "/login" || path == "/static/" || hasPrefix(path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}

		_, ok := ValidateRequest(r)
		if !ok {
			// For API requests, return 401
			if hasPrefix(path, "/api/") {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			// For page requests, redirect to login
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
