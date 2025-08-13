package monitoring

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AuthConfig represents authentication configuration
type AuthConfig struct {
	Enabled      bool         `yaml:"enabled"`
	APIKeyAuth   APIKeyConfig `yaml:"api_key"`
	JWTAuth      JWTConfig    `yaml:"jwt"`
	OAuthConfig  OAuthConfig  `yaml:"oauth"`
	RequiredRole string       `yaml:"required_role"` // admin, monitor, readonly
}

// APIKeyConfig represents API key authentication configuration
type APIKeyConfig struct {
	Enabled    bool              `yaml:"enabled"`
	Keys       map[string]string `yaml:"keys"`        // key -> role mapping
	HeaderName string            `yaml:"header_name"` // default: X-API-Key
}

// JWTConfig represents JWT authentication configuration
type JWTConfig struct {
	Enabled          bool          `yaml:"enabled"`
	Secret           string        `yaml:"secret"`
	TokenExpiry      time.Duration `yaml:"token_expiry"`
	RefreshExpiry    time.Duration `yaml:"refresh_expiry"`
	Issuer           string        `yaml:"issuer"`
	RequireSignedJWT bool          `yaml:"require_signed_jwt"`
}

// OAuthConfig represents OAuth2 authentication configuration
type OAuthConfig struct {
	Enabled      bool     `yaml:"enabled"`
	Provider     string   `yaml:"provider"` // google, github, custom
	ClientID     string   `yaml:"client_id"`
	ClientSecret string   `yaml:"client_secret"`
	RedirectURL  string   `yaml:"redirect_url"`
	Scopes       []string `yaml:"scopes"`
}

// AuthManager handles authentication and authorization
type AuthManager struct {
	config    AuthConfig
	jwtSecret []byte
}

// UserContext represents an authenticated user
type UserContext struct {
	UserID   string
	Role     string
	APIKey   string
	JWTToken string
	Claims   jwt.MapClaims
}

// Role constants
const (
	RoleAdmin    = "admin"
	RoleMonitor  = "monitor"
	RoleReadOnly = "readonly"
)

// NewAuthManager creates a new authentication manager
func NewAuthManager(config AuthConfig) (*AuthManager, error) {
	am := &AuthManager{
		config: config,
	}

	if config.JWTAuth.Enabled {
		if config.JWTAuth.Secret == "" {
			// Generate a random secret if none provided
			secret := make([]byte, 32)
			if _, err := rand.Read(secret); err != nil {
				return nil, fmt.Errorf("failed to generate JWT secret: %w", err)
			}
			am.jwtSecret = secret
		} else {
			am.jwtSecret = []byte(config.JWTAuth.Secret)
		}
	}

	return am, nil
}

// AuthenticateRequest authenticates an HTTP request
func (am *AuthManager) AuthenticateRequest(r *http.Request) (*UserContext, error) {
	if !am.config.Enabled {
		// Authentication disabled, allow all requests
		return &UserContext{
			UserID: "anonymous",
			Role:   RoleAdmin, // Grant admin role when auth is disabled
		}, nil
	}

	// Try API key authentication first
	if am.config.APIKeyAuth.Enabled {
		if userCtx, err := am.authenticateAPIKey(r); err == nil {
			return userCtx, nil
		}
	}

	// Try JWT authentication
	if am.config.JWTAuth.Enabled {
		if userCtx, err := am.authenticateJWT(r); err == nil {
			return userCtx, nil
		}
	}

	return nil, fmt.Errorf("authentication required")
}

// authenticateAPIKey authenticates using API key
func (am *AuthManager) authenticateAPIKey(r *http.Request) (*UserContext, error) {
	headerName := am.config.APIKeyAuth.HeaderName
	if headerName == "" {
		headerName = "X-API-Key"
	}

	apiKey := r.Header.Get(headerName)
	if apiKey == "" {
		return nil, fmt.Errorf("API key not provided")
	}

	// Check if API key exists and get role
	role, exists := am.config.APIKeyAuth.Keys[apiKey]
	if !exists {
		return nil, fmt.Errorf("invalid API key")
	}

	return &UserContext{
		UserID: fmt.Sprintf("apikey-%s", hashAPIKey(apiKey)),
		Role:   role,
		APIKey: apiKey,
	}, nil
}

// authenticateJWT authenticates using JWT token
func (am *AuthManager) authenticateJWT(r *http.Request) (*UserContext, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("authorization header not provided")
	}

	// Extract Bearer token
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	tokenString := parts[1]

	// Parse and validate JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return am.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid JWT token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid JWT token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid JWT claims")
	}

	// Extract user information from claims
	userID, _ := claims["sub"].(string)
	role, _ := claims["role"].(string)

	if userID == "" {
		return nil, fmt.Errorf("user ID not found in JWT claims")
	}

	if role == "" {
		role = RoleReadOnly // Default role
	}

	return &UserContext{
		UserID:   userID,
		Role:     role,
		JWTToken: tokenString,
		Claims:   claims,
	}, nil
}

// Authorize checks if user has required permissions
func (am *AuthManager) Authorize(userCtx *UserContext, requiredRole string) error {
	if !am.config.Enabled {
		return nil // Authorization disabled
	}

	// Check if user has required role
	if !am.hasRole(userCtx.Role, requiredRole) {
		return fmt.Errorf("insufficient permissions: required role %s, user has %s", requiredRole, userCtx.Role)
	}

	return nil
}

// hasRole checks if user role satisfies required role
func (am *AuthManager) hasRole(userRole, requiredRole string) bool {
	roleHierarchy := map[string]int{
		RoleReadOnly: 1,
		RoleMonitor:  2,
		RoleAdmin:    3,
	}

	userLevel, userExists := roleHierarchy[userRole]
	requiredLevel, requiredExists := roleHierarchy[requiredRole]

	if !userExists || !requiredExists {
		return false
	}

	return userLevel >= requiredLevel
}

// GenerateJWT generates a JWT token for a user
func (am *AuthManager) GenerateJWT(userID, role string) (string, error) {
	if !am.config.JWTAuth.Enabled {
		return "", fmt.Errorf("JWT authentication not enabled")
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"iat":  now.Unix(),
		"exp":  now.Add(am.config.JWTAuth.TokenExpiry).Unix(),
		"iss":  am.config.JWTAuth.Issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(am.jwtSecret)
}

// GenerateAPIKey generates a new API key
func (am *AuthManager) GenerateAPIKey() (string, error) {
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", fmt.Errorf("failed to generate API key: %w", err)
	}
	return base64.URLEncoding.EncodeToString(keyBytes), nil
}

// AuthMiddleware returns an HTTP middleware for authentication
func (am *AuthManager) AuthMiddleware(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for health check
			if r.URL.Path == "/api/v1/health" {
				next.ServeHTTP(w, r)
				return
			}

			userCtx, err := am.AuthenticateRequest(r)
			if err != nil {
				http.Error(w, fmt.Sprintf("Authentication failed: %v", err), http.StatusUnauthorized)
				return
			}

			if err := am.Authorize(userCtx, requiredRole); err != nil {
				http.Error(w, fmt.Sprintf("Authorization failed: %v", err), http.StatusForbidden)
				return
			}

			// Add user context to request
			ctx := context.WithValue(r.Context(), "user", userCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext extracts user context from request context
func GetUserFromContext(ctx context.Context) (*UserContext, bool) {
	user, ok := ctx.Value("user").(*UserContext)
	return user, ok
}

// hashAPIKey creates a hash of the API key for logging purposes
func hashAPIKey(apiKey string) string {
	if len(apiKey) < 8 {
		return "short"
	}
	return apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
}

// ValidateRole checks if a role is valid
func ValidateRole(role string) bool {
	switch role {
	case RoleAdmin, RoleMonitor, RoleReadOnly:
		return true
	default:
		return false
	}
}

// CompareAPIKeys performs constant-time comparison of API keys
func CompareAPIKeys(provided, stored string) bool {
	return subtle.ConstantTimeCompare([]byte(provided), []byte(stored)) == 1
}
