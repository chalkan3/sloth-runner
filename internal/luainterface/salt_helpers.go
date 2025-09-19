package luainterface

import (
	"encoding/json"
	"fmt"
)

// SaltKeyAction performs a salt-key action.
func SaltKeyAction(args ...string) (string, error) {
	stdout, stderr, err := RunCommand("salt-key", args...)
	if err != nil {
		return "", fmt.Errorf("command failed: %w, stderr: %s", err, stderr)
	}
	if stderr != "" {
		// salt-key often prints non-error info to stderr, but we return it anyway
		return stdout, fmt.Errorf(stderr)
	}
	return stdout, nil
}

// SaltRunAction performs a salt-run action.
func SaltRunAction(args ...string) (string, error) {
	stdout, stderr, err := RunCommand("salt-run", args...)
	if err != nil {
		return "", fmt.Errorf("command failed: %w, stderr: %s", err, stderr)
	}
	if stderr != "" {
		return stdout, fmt.Errorf(stderr)
	}
	return stdout, nil
}

// SaltPing pings a minion.
func SaltPing(minionID string) (bool, error) {
	output, stderr, err := RunCommand("salt", minionID, "test.ping", "--out=json")
	if err != nil {
		return false, fmt.Errorf("command failed: %w, stderr: %s", err, stderr)
	}
	if stderr != "" {
		return false, fmt.Errorf(stderr)
	}

	var result map[string]bool
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return false, fmt.Errorf("failed to unmarshal ping result: %w", err)
	}

	pingSuccess, ok := result[minionID]
	if !ok {
		return false, fmt.Errorf("minion ID '%s' not found in ping response", minionID)
	}

	return pingSuccess, nil
}
