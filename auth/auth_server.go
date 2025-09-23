package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type TokenInfo struct {
	User        string   `yaml:"user"`
	Permissions []string `yaml:"permissions"`
	Expires     string   `yaml:"expires"`
}

type TokensConfig struct {
	Tokens map[string]TokenInfo `yaml:"tokens"`
}

func loadTokens() (*TokensConfig, error) {
	data, err := os.ReadFile("tokens.yaml")
	if err != nil {
		return nil, err
	}

	var config TokensConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func validateToken(w http.ResponseWriter, r *http.Request) {
	// Extract token from URL path using native Go 1.22+ path values
	token := r.PathValue("token")

	config, err := loadTokens()
	if err != nil {
		http.Error(w, "Configuration error", http.StatusInternalServerError)
		return
	}

	tokenInfo, exists := config.Tokens[token]
	if !exists {
		http.Error(w, "Invalid token", http.StatusForbidden)
		return
	}

	// Check expiration
	expires, err := time.Parse("2006-01-02", tokenInfo.Expires)
	if err != nil {
		http.Error(w, "Invalid expiration date", http.StatusInternalServerError)
		return
	}

	if time.Now().After(expires) {
		http.Error(w, "Token expired", http.StatusForbidden)
		return
	}

	// Set response headers for Caddy
	w.Header().Set("X-Validated-User", tokenInfo.User)
	w.Header().Set("Content-Type", "application/json")

	// Return success response
	response := map[string]interface{}{
		"user":        tokenInfo.User,
		"permissions": tokenInfo.Permissions,
	}

	json.NewEncoder(w).Encode(response)
}

func main() {
	// Use native Go 1.22+ HTTP routing with path patterns
	http.HandleFunc("GET /validate/{token}", validateToken)

	fmt.Println("Token validation server starting on :9000 (using native Go HTTP router)")
	log.Fatal(http.ListenAndServe(":9000", nil))
}