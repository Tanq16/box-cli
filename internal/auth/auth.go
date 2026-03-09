package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/tanq16/box/internal/types"
	"golang.org/x/oauth2"
)

const (
	redirectURI = "http://localhost:8080"
	authURL     = "https://account.box.com/api/oauth2/authorize"
	tokenURL    = "https://api.box.com/oauth2/token"
)

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	dir := filepath.Join(home, ".box")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}
	return dir, nil
}

func credentialsPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "credentials.json"), nil
}

func tokenPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "token.json"), nil
}

func loadCredentials() (*types.BoxCredentials, error) {
	p, err := credentialsPath()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file at %s: %v\nRun 'box login' first or create ~/.box/credentials.json with client_id and client_secret", p, err)
	}
	var creds types.BoxCredentials
	if err := json.Unmarshal(b, &creds); err != nil {
		return nil, fmt.Errorf("unable to parse credentials file: %v", err)
	}
	if creds.ClientID == "" || creds.ClientSecret == "" {
		return nil, fmt.Errorf("credentials file must contain client_id and client_secret")
	}
	return &creds, nil
}

func oauthConfig(creds *types.BoxCredentials) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
		RedirectURL: redirectURI,
		Scopes:      []string{"root_readwrite"},
	}
}

func loadToken() (*oauth2.Token, error) {
	p, err := tokenPath()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

func saveToken(token *oauth2.Token) error {
	p, err := tokenPath()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %v", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

func Login() error {
	creds, err := loadCredentials()
	if err != nil {
		return err
	}
	config := oauthConfig(creds)
	state := fmt.Sprintf("st%d", os.Getpid())
	authorizationURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	fmt.Println("Open the following URL in your browser to authorize:")
	fmt.Println()
	fmt.Println(authorizationURL)
	fmt.Println()
	fmt.Print("After authorizing, paste the full redirect URL here: ")

	var redirectURLStr string
	fmt.Scanln(&redirectURLStr)

	parsedURL, err := url.Parse(redirectURLStr)
	if err != nil {
		return fmt.Errorf("could not parse the pasted URL: %v", err)
	}
	code := parsedURL.Query().Get("code")
	returnedState := parsedURL.Query().Get("state")
	if code == "" {
		return fmt.Errorf("pasted URL did not contain an authorization 'code'")
	}
	if returnedState != state {
		return fmt.Errorf("CSRF state mismatch: expected '%s' but got '%s'", state, returnedState)
	}

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		return fmt.Errorf("unable to exchange auth code for token: %v", err)
	}
	if err := saveToken(token); err != nil {
		return fmt.Errorf("unable to save token: %v", err)
	}
	fmt.Println("Login successful! Token saved.")
	return nil
}

func GetClient() (*http.Client, error) {
	creds, err := loadCredentials()
	if err != nil {
		return nil, err
	}
	config := oauthConfig(creds)

	token, err := loadToken()
	if err != nil {
		return nil, fmt.Errorf("no saved token found — run 'box login' first: %v", err)
	}

	ctx := context.Background()
	tokenSource := config.TokenSource(ctx, token)

	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("unable to refresh token — run 'box login' again: %v", err)
	}

	if newToken.AccessToken != token.AccessToken {
		saveToken(newToken)
	}

	return oauth2.NewClient(ctx, tokenSource), nil
}

func ConfigDir() (string, error) {
	return configDir()
}
