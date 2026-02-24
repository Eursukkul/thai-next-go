package main

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"thaid-backend/config"
)

// OpenIDConfiguration เก็บข้อมูล Well-Known Configuration
type OpenIDConfiguration struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
	IntrospectionEndpoint string `json:"introspection_endpoint"`
}

// TokenResponse เก็บข้อมูล Token ที่ได้รับจาก ThaID
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope"`
}

var (
	cfg          *config.Config
	oidcConfig   *OpenIDConfiguration
	httpClient   = &http.Client{Timeout: 30 * time.Second}
)

func main() {
	cfg = config.LoadConfig()

	// โหลด OpenID Configuration จาก ThaID
	if err := loadOIDCConfig(); err != nil {
		log.Fatalf("Failed to load OIDC config: %v", err)
	}

	r := gin.Default()

	// CORS สำหรับให้ Next.js frontend เรียก API ได้
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendURL},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	// Session middleware
	store := cookie.NewStore([]byte(cfg.SessionSecret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 วัน
		HttpOnly: true,
		Secure:   false, // เปลี่ยนเป็น true ใน production (HTTPS)
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(sessions.Sessions("thaid_session_v2", store))

	// Routes
	r.GET("/api/auth/login", handleLogin)
	r.GET("/api/auth/callback", handleCallback) // For direct callback from ThaID
	r.POST("/api/auth/exchange", handleExchangeCode) // For FE to send code
	r.GET("/api/auth/logout", handleLogout)
	r.GET("/api/auth/me", handleGetMe)
	r.POST("/api/auth/introspect", handleIntrospect)

	log.Printf("Server starting on %s", cfg.BackendURL)
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// loadOIDCConfig โหลด OpenID Configuration จาก ThaID
func loadOIDCConfig() error {
	resp, err := httpClient.Get(cfg.ThaidWellKnownURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	oidcConfig = &OpenIDConfiguration{}
	return json.Unmarshal(body, oidcConfig)
}

// handleLogin สร้าง URL สำหรับ redirect ไปยัง ThaID
func handleLogin(c *gin.Context) {
	state := generateState()
	session := sessions.Default(c)
	session.Set("oauth_state", state)
	session.Save()

	// สร้าง Authorization URL
	authURL, _ := url.Parse(oidcConfig.AuthorizationEndpoint)
	q := authURL.Query()
	q.Set("client_id", cfg.ThaidClientID)
	q.Set("redirect_uri", cfg.FrontendURL+"/auth/callback")
	q.Set("response_type", "code")
	q.Set("state", state)
	q.Set("scope", "openid pid")
	authURL.RawQuery = q.Encode()

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL.String(),
	})
}

// handleCallback รับ authorization code และแลกเปลี่ยนเป็น token
func handleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	session := sessions.Default(c)
	storedState := session.Get("oauth_state")

	// ตรวจสอบ state ป้องกัน CSRF
	if state == "" || state != storedState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code not found"})
		return
	}

	// แลกเปลี่ยน code เป็น token
	token, err := exchangeCodeForToken(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token: " + err.Error()})
		return
	}
	log.Printf("[ThaID] Token Response: %+v", token)

	// Decode ID Token เพื่อดึง userinfo
	userInfo, err := decodeIDToken(token.IDToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode ID token"})
		return
	}
	log.Printf("[ThaID] User Info from ID Token: %+v", userInfo)

	// เก็บข้อมูลใน session
	session.Delete("oauth_state")
	session.Set("user", userInfo)
	session.Set("access_token", token.AccessToken)
	session.Set("id_token", token.IDToken)
	session.Save()

	// Redirect กลับไปยัง frontend
	c.Redirect(http.StatusFound, cfg.FrontendURL+"/dashboard")
}

// handleExchangeCode รับ code จาก FE แล้วแลกเปลี่ยนเป็น token
func handleExchangeCode(c *gin.Context) {
	var req struct {
		Code  string `json:"code" binding:"required"`
		State string `json:"state" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	session := sessions.Default(c)
	storedState := session.Get("oauth_state")

	// ตรวจสอบ state ป้องกัน CSRF
	if req.State == "" || req.State != storedState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}

	// แลกเปลี่ยน code เป็น token
	token, err := exchangeCodeForToken(req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token: " + err.Error()})
		return
	}
	log.Printf("[ThaID] Token Response: %+v", token)

	// Decode ID Token เพื่อดึง userinfo
	userInfo, err := decodeIDToken(token.IDToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode ID token"})
		return
	}
	log.Printf("[ThaID] User Info from ID Token: %+v", userInfo)

	// เก็บข้อมูลใน session
	session.Delete("oauth_state")
	session.Set("user", userInfo)
	session.Set("access_token", token.AccessToken)
	session.Set("id_token", token.IDToken)
	session.Save()

	c.JSON(http.StatusOK, gin.H{
		"message": "Authentication successful",
		"user":    userInfo,
	})
}

// handleLogout ลบ session และ logout
func handleLogout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// handleGetMe ดึงข้อมูลผู้ใช้ปัจจุบัน
func handleGetMe(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")
	accessToken := session.Get("access_token")
	idToken := session.Get("id_token")

	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":         user,
		"access_token": accessToken,
		"id_token":     idToken,
	})
}

// handleIntrospect ตรวจสอบ token กับ ThaID
func handleIntrospect(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header"})
		return
	}

	accessToken := parts[1]

	// สร้าง Basic Auth header
	secret := cfg.ThaidClientID + ":" + cfg.ThaidClientSecret
	bearer := base64.StdEncoding.EncodeToString([]byte(secret))

	req, err := http.NewRequest("POST", oidcConfig.IntrospectionEndpoint, 
		strings.NewReader(url.Values{"token": {accessToken}}.Encode()))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header.Set("Authorization", "Basic "+bearer)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to introspect token"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", body)
}

// exchangeCodeForToken แลกเปลี่ยน authorization code เป็น access token
func exchangeCodeForToken(code string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", cfg.FrontendURL+"/auth/callback")
	data.Set("client_id", cfg.ThaidClientID)
	data.Set("client_secret", cfg.ThaidClientSecret)

	log.Printf("[ThaID] Token Request - redirect_uri: %s", cfg.FrontendURL+"/auth/callback")
	log.Printf("[ThaID] Token Request - client_id: %s... (length: %d)", cfg.ThaidClientID[:10], len(cfg.ThaidClientID))
	log.Printf("[ThaID] Token Request - client_secret length: %d", len(cfg.ThaidClientSecret))

	resp, err := httpClient.Post(oidcConfig.TokenEndpoint, 
		"application/x-www-form-urlencoded", 
		strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Printf("[ThaID] Token Endpoint Response Status: %d", resp.StatusCode)
	log.Printf("[ThaID] Token Endpoint Response Body: %s", string(body))

	var token TokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

// decodeIDToken ถอดรหัส JWT (แบบง่าย - ไม่ตรวจสอบลายเซ็น)
func decodeIDToken(idToken string) (map[string]interface{}, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, nil
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}

	return claims, nil
}

// generateState สร้าง random state string
func generateState() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = byte(65 + (i % 26))
	}
	return base64.URLEncoding.EncodeToString(b)
}
