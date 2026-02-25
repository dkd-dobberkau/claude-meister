package squirrel

import "encoding/json"

// ParseStatus parses the JSON output of `squirrel status --json`.
func ParseStatus(data []byte) (*Status, error) {
	var s Status
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}
