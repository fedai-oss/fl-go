package monitoring

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthManager_APIKeyAuthentication(t *testing.T) {
	config := AuthConfig{
		Enabled: true,
		APIKeyAuth: APIKeyConfig{
			Enabled:    true,
			HeaderName: "X-API-Key",
			Keys: map[string]string{
				"admin-key":    RoleAdmin,
				"monitor-key":  RoleMonitor,
				"readonly-key": RoleReadOnly,
			},
		},
	}

	authManager, err := NewAuthManager(config)
	if err != nil {
		t.Fatalf("Failed to create auth manager: %v", err)
	}

	tests := []struct {
		name      string
		apiKey    string
		wantRole  string
		wantError bool
	}{
		{
			name:      "valid admin key",
			apiKey:    "admin-key",
			wantRole:  RoleAdmin,
			wantError: false,
		},
		{
			name:      "valid monitor key",
			apiKey:    "monitor-key",
			wantRole:  RoleMonitor,
			wantError: false,
		},
		{
			name:      "valid readonly key",
			apiKey:    "readonly-key",
			wantRole:  RoleReadOnly,
			wantError: false,
		},
		{
			name:      "invalid key",
			apiKey:    "invalid-key",
			wantRole:  "",
			wantError: true,
		},
		{
			name:      "empty key",
			apiKey:    "",
			wantRole:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/test", nil)
			if tt.apiKey != "" {
				req.Header.Set("X-API-Key", tt.apiKey)
			}

			userCtx, err := authManager.AuthenticateRequest(req)
			if (err != nil) != tt.wantError {
				t.Errorf("AuthenticateRequest() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && userCtx.Role != tt.wantRole {
				t.Errorf("AuthenticateRequest() role = %v, want %v", userCtx.Role, tt.wantRole)
			}
		})
	}
}

func TestAuthManager_JWTAuthentication(t *testing.T) {
	config := AuthConfig{
		Enabled: true,
		JWTAuth: JWTConfig{
			Enabled:     true,
			Secret:      "test-secret",
			TokenExpiry: time.Hour,
			Issuer:      "test-issuer",
		},
	}

	authManager, err := NewAuthManager(config)
	if err != nil {
		t.Fatalf("Failed to create auth manager: %v", err)
	}

	// Generate a valid JWT token
	token, err := authManager.GenerateJWT("test-user", RoleMonitor)
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	tests := []struct {
		name       string
		authHeader string
		wantRole   string
		wantError  bool
	}{
		{
			name:       "valid JWT token",
			authHeader: "Bearer " + token,
			wantRole:   RoleMonitor,
			wantError:  false,
		},
		{
			name:       "invalid token format",
			authHeader: "InvalidFormat",
			wantRole:   "",
			wantError:  true,
		},
		{
			name:       "missing bearer prefix",
			authHeader: token,
			wantRole:   "",
			wantError:  true,
		},
		{
			name:       "empty header",
			authHeader: "",
			wantRole:   "",
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			userCtx, err := authManager.AuthenticateRequest(req)
			if (err != nil) != tt.wantError {
				t.Errorf("AuthenticateRequest() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && userCtx.Role != tt.wantRole {
				t.Errorf("AuthenticateRequest() role = %v, want %v", userCtx.Role, tt.wantRole)
			}
		})
	}
}

func TestAuthManager_Authorization(t *testing.T) {
	config := AuthConfig{
		Enabled: true,
	}

	authManager, err := NewAuthManager(config)
	if err != nil {
		t.Fatalf("Failed to create auth manager: %v", err)
	}

	tests := []struct {
		name         string
		userRole     string
		requiredRole string
		wantError    bool
	}{
		{
			name:         "admin accessing admin endpoint",
			userRole:     RoleAdmin,
			requiredRole: RoleAdmin,
			wantError:    false,
		},
		{
			name:         "admin accessing monitor endpoint",
			userRole:     RoleAdmin,
			requiredRole: RoleMonitor,
			wantError:    false,
		},
		{
			name:         "admin accessing readonly endpoint",
			userRole:     RoleAdmin,
			requiredRole: RoleReadOnly,
			wantError:    false,
		},
		{
			name:         "monitor accessing monitor endpoint",
			userRole:     RoleMonitor,
			requiredRole: RoleMonitor,
			wantError:    false,
		},
		{
			name:         "monitor accessing readonly endpoint",
			userRole:     RoleMonitor,
			requiredRole: RoleReadOnly,
			wantError:    false,
		},
		{
			name:         "monitor accessing admin endpoint",
			userRole:     RoleMonitor,
			requiredRole: RoleAdmin,
			wantError:    true,
		},
		{
			name:         "readonly accessing readonly endpoint",
			userRole:     RoleReadOnly,
			requiredRole: RoleReadOnly,
			wantError:    false,
		},
		{
			name:         "readonly accessing monitor endpoint",
			userRole:     RoleReadOnly,
			requiredRole: RoleMonitor,
			wantError:    true,
		},
		{
			name:         "readonly accessing admin endpoint",
			userRole:     RoleReadOnly,
			requiredRole: RoleAdmin,
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userCtx := &UserContext{
				UserID: "test-user",
				Role:   tt.userRole,
			}

			err := authManager.Authorize(userCtx, tt.requiredRole)
			if (err != nil) != tt.wantError {
				t.Errorf("Authorize() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestAuthManager_DisabledAuthentication(t *testing.T) {
	config := AuthConfig{
		Enabled: false,
	}

	authManager, err := NewAuthManager(config)
	if err != nil {
		t.Fatalf("Failed to create auth manager: %v", err)
	}

	// Test that authentication is bypassed when disabled
	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	userCtx, err := authManager.AuthenticateRequest(req)
	if err != nil {
		t.Errorf("AuthenticateRequest() should not fail when auth is disabled: %v", err)
	}

	if userCtx.Role != RoleAdmin {
		t.Errorf("AuthenticateRequest() should grant admin role when auth is disabled, got: %v", userCtx.Role)
	}

	// Test that authorization is bypassed when disabled
	err = authManager.Authorize(userCtx, RoleAdmin)
	if err != nil {
		t.Errorf("Authorize() should not fail when auth is disabled: %v", err)
	}
}

func TestAuthMiddleware(t *testing.T) {
	config := AuthConfig{
		Enabled: true,
		APIKeyAuth: APIKeyConfig{
			Enabled:    true,
			HeaderName: "X-API-Key",
			Keys: map[string]string{
				"admin-key":    RoleAdmin,
				"readonly-key": RoleReadOnly,
			},
		},
	}

	authManager, err := NewAuthManager(config)
	if err != nil {
		t.Fatalf("Failed to create auth manager: %v", err)
	}

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Health endpoint doesn't require user context
		if r.URL.Path == "/api/v1/health" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
			return
		}

		user, ok := GetUserFromContext(r.Context())
		if !ok {
			http.Error(w, "User context not found", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello " + user.UserID))
	})

	// Apply auth middleware
	authMiddleware := authManager.AuthMiddleware(RoleReadOnly)
	protectedHandler := authMiddleware(testHandler)

	tests := []struct {
		name           string
		path           string
		apiKey         string
		expectedStatus int
	}{
		{
			name:           "health endpoint without auth",
			path:           "/api/v1/health",
			apiKey:         "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "protected endpoint with valid key",
			path:           "/api/v1/test",
			apiKey:         "admin-key",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "protected endpoint with valid readonly key",
			path:           "/api/v1/test",
			apiKey:         "readonly-key",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "protected endpoint without auth",
			path:           "/api/v1/test",
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "protected endpoint with invalid key",
			path:           "/api/v1/test",
			apiKey:         "invalid-key",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			if tt.apiKey != "" {
				req.Header.Set("X-API-Key", tt.apiKey)
			}

			rr := httptest.NewRecorder()
			protectedHandler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("AuthMiddleware() status = %v, want %v", rr.Code, tt.expectedStatus)
			}
		})
	}
}

func TestValidateRole(t *testing.T) {
	tests := []struct {
		role string
		want bool
	}{
		{RoleAdmin, true},
		{RoleMonitor, true},
		{RoleReadOnly, true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			if got := ValidateRole(tt.role); got != tt.want {
				t.Errorf("ValidateRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateAPIKey(t *testing.T) {
	config := AuthConfig{}
	authManager, err := NewAuthManager(config)
	if err != nil {
		t.Fatalf("Failed to create auth manager: %v", err)
	}

	// Test API key generation
	key1, err := authManager.GenerateAPIKey()
	if err != nil {
		t.Errorf("GenerateAPIKey() error = %v", err)
	}

	key2, err := authManager.GenerateAPIKey()
	if err != nil {
		t.Errorf("GenerateAPIKey() error = %v", err)
	}

	// Keys should be different
	if key1 == key2 {
		t.Error("GenerateAPIKey() should generate unique keys")
	}

	// Keys should not be empty
	if key1 == "" || key2 == "" {
		t.Error("GenerateAPIKey() should not generate empty keys")
	}
}
