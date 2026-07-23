package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	u "github.com/tanq16/box/utils"
	"golang.org/x/oauth2"
)

const (
	authURL     = "https://account.box.com/api/oauth2/authorize"
	tokenURL    = "https://api.box.com/oauth2/token"
	redirectURI = "http://localhost:8080"
)

func ConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		u.PrintFatal("cannot determine home directory", err)
	}
	dir := filepath.Join(home, ".config", "box")
	if err := os.MkdirAll(dir, 0700); err != nil {
		u.PrintFatal("cannot create config directory", err)
	}
	return dir
}

func LoadCredentials() (*oauth2.Config, error) {
	credPath := filepath.Join(ConfigDir(), "credentials.json")
	data, err := os.ReadFile(credPath)
	if err != nil {
		return nil, fmt.Errorf("create %s with your OAuth client credentials", credPath)
	}

	var creds struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	}
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("unable to parse credentials file: %w", err)
	}
	if creds.ClientID == "" || creds.ClientSecret == "" {
		return nil, fmt.Errorf("credentials file must contain client_id and client_secret")
	}

	return &oauth2.Config{
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
		RedirectURL: redirectURI,
		Scopes:      []string{"root_readwrite"},
	}, nil
}

func Login(config *oauth2.Config, mode string) (*oauth2.Token, error) {
	state, err := generateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	switch mode {
	case "manual":
		return loginWithManual(config, state)
	default:
		return loginWithCallback(config, state)
	}
}

func loginWithCallback(config *oauth2.Config, state string) (*oauth2.Token, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		return nil, fmt.Errorf("port 8080 unavailable — use 'box login --manual' instead: %w", err)
	}

	authorizationURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			errCh <- fmt.Errorf("state mismatch — possible CSRF attack")
			http.Error(w, "State mismatch", http.StatusBadRequest)
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no auth code in callback")
			http.Error(w, "Missing code", http.StatusBadRequest)
			return
		}
		fmt.Fprint(w, "<html><body><h2>Authentication successful!</h2><p>You can close this tab.</p></body></html>")
		codeCh <- code
	})

	srv := &http.Server{Handler: mux}
	go func() {
		if err := srv.Serve(listener); err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	u.PrintInfo("Opening browser for authentication...")
	if err := openBrowser(authorizationURL); err != nil {
		srv.Close()
		return nil, fmt.Errorf("cannot open browser — use 'box login --manual' instead")
	}
	u.PrintInfo("Waiting for authorization in browser...")

	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		srv.Close()
		return nil, err
	case <-time.After(5 * time.Minute):
		srv.Close()
		return nil, fmt.Errorf("authentication timed out")
	}

	srv.Close()

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	if err := SaveToken(token); err != nil {
		return nil, err
	}
	return token, nil
}

func loginWithManual(config *oauth2.Config, state string) (*oauth2.Token, error) {
	authorizationURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	u.PrintInfo("Visit this URL to authenticate:")
	u.PrintGeneric(authorizationURL)
	u.PrintGeneric("")
	u.PrintInfo("After authorizing, paste the full redirect URL or authorization code.")

	input, err := u.PromptInput("Authorization:", "http://localhost:8080?code=...")
	if err != nil {
		return nil, fmt.Errorf("input error: %w", err)
	}
	if input == "" {
		return nil, fmt.Errorf("no code provided")
	}

	code := extractCode(input)

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	if err := SaveToken(token); err != nil {
		return nil, err
	}
	return token, nil
}

func extractCode(input string) string {
	if !strings.Contains(input, "code=") {
		return input
	}
	parts := strings.SplitN(input, "?", 2)
	if len(parts) < 2 {
		return input
	}
	for _, param := range strings.Split(parts[1], "&") {
		kv := strings.SplitN(param, "=", 2)
		if len(kv) == 2 && kv[0] == "code" {
			return kv[1]
		}
	}
	return input
}

func LoadToken() (*oauth2.Token, error) {
	tokenPath := filepath.Join(ConfigDir(), "token.json")
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("run 'box login' first")
	}
	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("corrupt token file — run 'box login' again")
	}
	return &token, nil
}

func SaveToken(token *oauth2.Token) error {
	tokenPath := filepath.Join(ConfigDir(), "token.json")
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}
	if err := os.WriteFile(tokenPath, data, 0600); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}
	return nil
}

func GetHTTPClient() (*http.Client, error) {
	config, err := LoadCredentials()
	if err != nil {
		return nil, err
	}

	token, err := LoadToken()
	if err != nil {
		return nil, err
	}

	tokenSource := config.TokenSource(context.Background(), token)

	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("token refresh failed — run 'box login' again")
	}
	if newToken.AccessToken != token.AccessToken {
		if err := SaveToken(newToken); err != nil {
			return nil, err
		}
	}

	return oauth2.NewClient(context.Background(), tokenSource), nil
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Run()
}
