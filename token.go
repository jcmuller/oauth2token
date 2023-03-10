package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/99designs/keyring"
	"github.com/adrg/xdg"
	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	progName   = "oauth2token"
	secretKey  = "token"
	configFile = "config.json"
	scopesFile = "scopes.json"
	stateLen   = 32

	reauthScope        = "https://www.googleapis.com/auth/accounts.reauth"
	cloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"
)

var (
	keyringConfig = keyring.Config{
		FileDir:          filepath.Join(xdg.DataHome, progName),
		FilePasswordFunc: func(s string) (string, error) { return progName, nil },
		AllowedBackends:  []keyring.BackendType{keyring.FileBackend},
	}
)

// Retrieves a token from secret store
func retrieveToken(ctx context.Context) (*oauth2.Token, error) {
	slog.Log(slog.LevelDebug, "retrieveToken")
	ring, err := keyring.Open(keyringConfig)
	if err != nil {
		return nil, fmt.Errorf("error opening keyring: %w", err)
	}

	item, err := ring.Get(secretKey)
	if err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return mintNewToken(ctx)
		}

		return nil, fmt.Errorf("error retrieving item from secret store: %w", err)
	}

	token := new(oauth2.Token)
	err = json.NewDecoder(bytes.NewReader(item.Data)).Decode(token)
	if err != nil {
		return nil, fmt.Errorf("error decoding token: %w", err)
	}

	if !token.Valid() {
		slog.Log(slog.LevelDebug, "token is not valid. Refreshing")

		token, err = refreshToken(ctx, token)
		if err != nil {
			return nil, fmt.Errorf("error refreshing token: %w", err)
		}
	}

	return token, nil
}

// Request a token from the web, then returns the retrieved token.
func mintNewToken(ctx context.Context) (*oauth2.Token, error) {
	slog.Log(slog.LevelDebug, "mintNewToken")
	config, err := getConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting oauth2 config: %w", err)
	}

	code, err := getCode(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("error getting code: %w", err)
	}

	slog.Log(slog.LevelDebug, "mintNewToken", slog.String("code", code))

	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %w", err)
	}

	if err := saveToken(token); err != nil {
		return nil, fmt.Errorf("error saving minted token: %w", err)
	}

	return token, nil
}

func refreshToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	slog.Log(slog.LevelDebug, "refreshToken")

	config, err := getConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting oauth2 config: %w", err)
	}

	token, err = config.TokenSource(ctx, token).Token()
	if err != nil {
		return nil, fmt.Errorf("error retrieving token from token source: %w", err)
	}

	if err := saveToken(token); err != nil {
		return nil, fmt.Errorf("error saving refreshed token: %w", err)
	}

	return token, nil
}

func getConfig() (*oauth2.Config, error) {
	slog.Log(slog.LevelDebug, "getConfig")
	configFile, err := readConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	scopesFile, err := readConfig(scopesFile)
	if err != nil {
		return nil, fmt.Errorf("error reading scopes file: %w", err)
	}

	var scopes []string
	if err := json.Unmarshal(scopesFile, &scopes); err != nil {
		return nil, fmt.Errorf("error unmarshaling scopes: %w", err)
	}
	scopes = append(scopes, reauthScope, cloudPlatformScope)

	config, err := google.ConfigFromJSON(configFile, scopes...)
	if err != nil {
		return nil, fmt.Errorf("error parsing creds: %w", err)
	}

	return config, nil
}

func getCode(ctx context.Context, config *oauth2.Config) (string, error) {
	slog.Log(slog.LevelDebug, "getCode")
	state, err := randString(stateLen)
	if err != nil {
		return "", fmt.Errorf("error creating random state string: %w", err)
	}

	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Fprintln(os.Stderr, "Please open this URL to authorize access:", authURL)
	if err := exec.Command("xdg-open", authURL).Start(); err != nil {
		return "", fmt.Errorf("error opening auth URL: %w", err)
	}

	code, err := callbackServer(ctx, state, config.RedirectURL)
	if err != nil {
		return "", fmt.Errorf("error starting callback server: %w", err)
	}

	return code, nil
}

func readConfig(fileName string) ([]byte, error) {
	slog.Log(slog.LevelDebug, "readConfig")
	configPath, err := xdg.SearchConfigFile(filepath.Join(progName, fileName))
	if err != nil {
		return nil, fmt.Errorf("error searching config file: %w", err)
	}
	return os.ReadFile(configPath)
}

// Saves a token to a file path.
func saveToken(token *oauth2.Token) error {
	slog.Log(slog.LevelDebug, "saveToken")
	ring, err := keyring.Open(keyringConfig)
	if err != nil {
		return fmt.Errorf("error opening keyring: %w", err)
	}

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(token); err != nil {
		return fmt.Errorf("error encoding token: %w", err)
	}

	if err := ring.Set(keyring.Item{
		Key:  secretKey,
		Data: buf.Bytes(),
	}); err != nil {
		return fmt.Errorf("error setting secret: %w", err)
	}

	return nil
}

func randString(length int) (string, error) {
	slog.Log(slog.LevelDebug, "randString")
	data := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(data), nil
}
