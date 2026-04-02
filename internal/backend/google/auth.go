package google

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/oauth2"
	google "golang.org/x/oauth2/google"
)

type AuthManager struct {
	creds_path string
	token_path string
	scopes     []string
}

func NewAuthManager(credsPath, tokenPath string, scopes []string) *AuthManager {
	return &AuthManager{
		creds_path: credsPath,
		token_path: tokenPath,
		scopes:     scopes,
	}
}

func (a *AuthManager) Login(ctx context.Context) error {
	//first check current login state
	cfg, err := a.loadConfig()
	if err != nil {
		return err
	}
	if tok, err := a.loadToken(); err == nil && tok.Valid() {
		//already logged in
		return nil
	}

	//user is not already logged in so initiate login process
	listener, err := net.Listen("tcp", "127.0.0.1:8085")
	if err != nil {
		return fmt.Errorf("error making listener for auth %w", err)
	}
	defer listener.Close()

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return fmt.Errorf("failed to generate oauth state: %w", err)
	}

	state := base64.RawURLEncoding.EncodeToString(stateBytes)

	mux := http.NewServeMux() //create object to create callback endpoint to return after login
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// finding error keyword in url to know there was a failure
		if e := r.URL.Query().Get("error"); e != "" {
			http.Error(w, "OAuth error: "+e, http.StatusBadRequest)
			return
		}

		returnedState := r.URL.Query().Get("state")
		if returnedState == "" || returnedState != state {
			http.Error(w, "invalid state parameter", http.StatusBadRequest)
			select {
			case errCh <- fmt.Errorf("invalid oauth state: got %q expected %q", returnedState, state):
			default:
			}
			return
		}
		// happy path
		auth_code := strings.TrimSpace(r.URL.Query().Get("code"))
		if auth_code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			select {
			case errCh <- errors.New("missing authorization code"):
			default:
			}
			return
		}
		fmt.Fprintln(w, "Authorization successful. You can close this tab.")
		select {
		case codeCh <- auth_code:
		default:
		}
	})

	//build server that runs on port 8085 with login hander (mux)
	auth_server := &http.Server{
		Handler: mux,
	}

	go func() {
		if serve_err := auth_server.Serve(listener); serve_err != nil && !errors.Is(serve_err, http.ErrServerClosed) {
			select {
			case errCh <- fmt.Errorf("callback server error: %w", serve_err):
			default:
			}
		}
	}()

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = auth_server.Shutdown(shutdownCtx)
	}()

	cfg.RedirectURL = "http://127.0.0.1:8085/callback"

	//url that initiates consent for google login
	authURL := cfg.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
	)

	// Attempt Chrome incognito first to avoid stale OAuth session issues.
	// Fallback to default browser if Chrome launch fails.
	if err := exec.Command("open", "-na", "Google Chrome", "--args", "--incognito", authURL).Start(); err != nil {
		if err := exec.Command("open", authURL).Start(); err != nil {
			fmt.Println("Could not open Browser automatically")
		}
	}
	fmt.Println("Complete login in your Browser")
	fmt.Println(authURL)

	var auth_code string
	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-errCh:
		return e
	case auth_code = <-codeCh:
	}

	//convert temporary login to to resusable token
	tok, err := cfg.Exchange(ctx, auth_code)
	if err != nil {
		return fmt.Errorf("Error creating token from login code %w", err)
	}

	// Persist token to disk for future API calls.
	if err := a.saveToken(tok); err != nil {
		return err
	}

	return nil

}

func (a *AuthManager) loadConfig() (*oauth2.Config, error) {
	creds, err := os.ReadFile(a.creds_path)
	if err != nil {
		return nil, fmt.Errorf("error reading credentials file: %w, ", err)
	}

	cfg, err := google.ConfigFromJSON(creds, a.scopes...)
	if err != nil {
		return nil, fmt.Errorf("error parsing credentials %w", err)
	}

	return cfg, nil

}

func (a *AuthManager) loadToken() (*oauth2.Token, error) {
	b, err := os.ReadFile(a.token_path)
	if err != nil {
		return nil, fmt.Errorf("Read token file error: %w", err)
	}

	var tok oauth2.Token
	if err := json.Unmarshal(b, &tok); err != nil {
		return nil, fmt.Errorf("Parse token json: %w", err)
	}

	return &tok, nil
}

// saves token to file for persistant usage
func (a *AuthManager) saveToken(tok *oauth2.Token) error {
	if tok == nil {
		return errors.New("token is nil")
	}

	if err := os.MkdirAll(filepath.Dir(a.token_path), 0o700); err != nil {
		return fmt.Errorf("create token directory: %w", err)
	}
	//convert go token struct to JSON bytes for http requests
	tokenJSON, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal oauth token to json: %w", err)
	}

	if err := os.WriteFile(a.token_path, tokenJSON, 0o600); err != nil {
		return fmt.Errorf("write token file %q: %w", a.token_path, err)
	}

	// Success: token is now persisted.
	return nil

}
