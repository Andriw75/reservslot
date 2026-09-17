package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"back/domain/outputs"
	"back/infrastructure/auth"
)

type AuthHandler struct {
	deps *Dependencies
}

func NewAuthHandler(deps *Dependencies) *AuthHandler {
	return &AuthHandler{deps: deps}
}

// Token — POST /auth/token (form-urlencoded: username, password)
func (h *AuthHandler) Token(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		writeJSONError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	user, ok := h.deps.UserStore.IsValidPassword(username, password)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := h.deps.JWTManager.Generate(user.ID, user.Username, user.Email, user.Activo, user.Permisos)
	if err != nil {
		slog.Error("failed to generate token", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	h.setTokenCookie(w, token, time.Now().Add(time.Duration(h.deps.Config.JWTExpirationMin)*time.Minute))
	writeJSON(w, http.StatusOK, outputs.ToUserResponse(user))
}

// Logout — POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r.Context())
	if h.deps.Config.JWTStoreEnabled && claims != nil && claims.ID != "" {
		h.deps.JWTStore.Remove(claims.ID)
	}
	h.clearTokenCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// Me — GET /auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r.Context())
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	writeJSON(w, http.StatusOK, outputs.UserResponse{
		ID:       claims.UserID,
		Username: claims.Username,
		Email:    claims.Email,
		Activo:   claims.Activo,
		Permisos: claims.Permisos,
	})
}

// AuthMiddleware exige sesión válida. No valida permisos: cualquier usuario
// autenticado puede operar todos los módulos (fase inicial).
func (h *AuthHandler) AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := h.resolveClaims(w, r)
			if !ok {
				return
			}
			next.ServeHTTP(w, r.WithContext(WithClaims(r.Context(), claims)))
		})
	}
}

// OptionalAuthMiddleware resuelve la sesión si existe, pero no la exige.
func (h *AuthHandler) OptionalAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(h.deps.Config.CookieName)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			claims, err := h.deps.JWTManager.Validate(cookie.Value)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r.WithContext(WithClaims(r.Context(), claims)))
		})
	}
}

func (h *AuthHandler) resolveClaims(w http.ResponseWriter, r *http.Request) (*auth.Claims, bool) {
	cookie, err := r.Cookie(h.deps.Config.CookieName)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "not authenticated")
		return nil, false
	}

	claims, err := h.deps.JWTManager.Validate(cookie.Value)
	if err != nil {
		h.clearTokenCookie(w)
		writeJSONError(w, http.StatusUnauthorized, "invalid token")
		return nil, false
	}

	if h.deps.Config.JWTStoreEnabled && claims.ID != "" && !h.deps.JWTStore.Exists(claims.ID) {
		user, uerr := h.deps.UserStore.FindByID(claims.UserID)
		if uerr != nil || !user.Activo {
			h.clearTokenCookie(w)
			writeJSONError(w, http.StatusUnauthorized, "session expired")
			return nil, false
		}
		newToken, terr := h.deps.JWTManager.Generate(user.ID, user.Username, user.Email, user.Activo, user.Permisos)
		if terr != nil {
			h.clearTokenCookie(w)
			writeJSONError(w, http.StatusUnauthorized, "session expired")
			return nil, false
		}
		h.setTokenCookie(w, newToken, time.Now().Add(time.Duration(h.deps.Config.JWTExpirationMin)*time.Minute))
		claims, _ = h.deps.JWTManager.Validate(newToken)
		return claims, true
	}

	if h.deps.JWTManager.ShouldRenew(claims) && h.deps.Config.JWTLoadUserOnRenew {
		if user, uerr := h.deps.UserStore.FindByID(claims.UserID); uerr == nil && user.Activo {
			if newToken, terr := h.deps.JWTManager.Generate(user.ID, user.Username, user.Email, user.Activo, user.Permisos); terr == nil {
				h.setTokenCookie(w, newToken, time.Now().Add(time.Duration(h.deps.Config.JWTExpirationMin)*time.Minute))
				claims, _ = h.deps.JWTManager.Validate(newToken)
			}
		}
	}

	return claims, true
}

func (h *AuthHandler) setTokenCookie(w http.ResponseWriter, token string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.deps.Config.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.deps.Config.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  expires,
	})
}

func (h *AuthHandler) clearTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.deps.Config.CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.deps.Config.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
