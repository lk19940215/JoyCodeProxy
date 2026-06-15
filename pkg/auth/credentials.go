package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
)

// Credentials holds JoyCode authentication data.
type Credentials struct {
	PtKey  string
	UserID string
}

type stateData struct {
	JoyCoderUser struct {
		PtKey  string `json:"ptKey"`
		UserID string `json:"userId"`
	} `json:"joyCoderUser"`
}

const (
	stateDBEnv       = "JOYCODE_STATE_DB"
	containerStateDB = "/root/.joycode-ide/state.vscdb"
)

// JoyCodeStateDBPath returns the platform-specific path to JoyCode's state.vscdb.
func JoyCodeStateDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support",
			"JoyCode", "User", "globalStorage", "state.vscdb"), nil
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "JoyCode", "User", "globalStorage", "state.vscdb"), nil
	case "linux":
		return filepath.Join(home, ".config", "JoyCode", "User", "globalStorage", "state.vscdb"), nil
	default:
		return "", fmt.Errorf("unsupported OS %s for auto credential detection; set %s manually", runtime.GOOS, stateDBEnv)
	}
}

// LoadFromSystem reads ptKey from local JoyCode state database.
func LoadFromSystem() (*Credentials, error) {
	if dbPath := os.Getenv(stateDBEnv); dbPath != "" {
		return loadFromStateDB(dbPath)
	}
	if _, err := os.Stat(containerStateDB); err == nil {
		return loadFromStateDB(containerStateDB)
	}

	dbPath, err := JoyCodeStateDBPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("JoyCode state database not found at %s\n  Please install and log in to JoyCode IDE first", dbPath)
	}

	return loadFromStateDB(dbPath)
}

func loadFromStateDB(dbPath string) (*Credentials, error) {
	if _, err := os.Stat(dbPath); err != nil {
		return nil, fmt.Errorf("JoyCode state database not found at %s: %w", dbPath, err)
	}

	db, err := sql.Open("sqlite3", dbPath+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("cannot open JoyCode database: %w", err)
	}
	defer db.Close()

	var value string
	if err := db.QueryRow(
		"SELECT value FROM ItemTable WHERE key='JoyCoder.IDE'",
	).Scan(&value); err != nil {
		return nil, fmt.Errorf("login info not found in database\n  Please log in to JoyCode IDE first")
	}

	var data stateData
	if err := json.Unmarshal([]byte(value), &data); err != nil {
		return nil, fmt.Errorf("cannot parse login data from database: %w", err)
	}
	if data.JoyCoderUser.PtKey == "" {
		return nil, fmt.Errorf("ptKey is empty in stored credentials\n  Please re-login to JoyCode IDE")
	}
	if data.JoyCoderUser.UserID == "" {
		return nil, fmt.Errorf("userId is empty in stored credentials\n  Please re-login to JoyCode IDE")
	}
	return &Credentials{
		PtKey:  data.JoyCoderUser.PtKey,
		UserID: data.JoyCoderUser.UserID,
	}, nil
}
