package checker

import (
	"encoding/json"
	"fmt"
	"os"
)

// Kubernetes retains at most 4096 bytes from one container termination
// message. Failing loudly prevents a silently truncated, invalid JSON result.
const MaxTerminationMessageBytes = 4096

func MarshalResult(result *Result) ([]byte, error) {
	if err := result.Validate(); err != nil {
		return nil, err
	}
	payload, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("marshal checker result: %w", err)
	}
	if len(payload) > MaxTerminationMessageBytes {
		return nil, fmt.Errorf("checker result is %d bytes; Kubernetes termination message limit is %d", len(payload), MaxTerminationMessageBytes)
	}
	return payload, nil
}

func WriteResult(path string, result *Result) error {
	payload, err := MarshalResult(result)
	if err != nil {
		return err
	}
	if path == "-" {
		_, err = os.Stdout.Write(append(payload, '\n'))
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("open checker result %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()
	if _, err = file.Write(payload); err != nil {
		return fmt.Errorf("write checker result %s: %w", path, err)
	}
	return nil
}
