package script

import (
	"errors"
	"strings"

	"gopkg.in/yaml.v3"
)

//nolint:gosec // markers are not credentials
const (
	// tokensStartMarker marks the start of the token block in the output of a script.
	tokensStartMarker = "--START TOKENS--"
	// tokensEndMarker marks the end of the token block in the output of an initialization script.
	tokensEndMarker = "--END TOKENS--"
)

// ParseTokens parses the yaml-formatted token block, enclosed in start and end markers, of a script output.
// The last block in the output is used.
// stdout: The output of the script.
func ParseTokens(stdout string) (map[string]any, error) {
	startBlock := strings.LastIndex(stdout, tokensStartMarker)
	endBlock := strings.LastIndex(stdout, tokensEndMarker)
	if startBlock < 0 || endBlock < startBlock {
		return nil, errors.New("could not find the token block in the output")
	}

	var out map[string]any
	if err := yaml.Unmarshal([]byte(stdout[startBlock+len(tokensStartMarker):endBlock]), &out); err != nil {
		return nil, err
	}

	return out, nil
}
