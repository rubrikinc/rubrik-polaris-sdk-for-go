package cdm

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// errorMessage returns an error message from the specified response and HTTP
// status code.
func errorMessage(res []byte, code int) string {
	msg := fmt.Sprintf("%s (%d)", http.StatusText(code), code)

	var cdmErr struct {
		Type    string `json:"errorType"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(res, &cdmErr); err == nil {
		if cdmErr.Type != "" && cdmErr.Message != "" {
			return fmt.Sprintf("%s: %s: %s", msg, cdmErr.Type, cdmErr.Message)
		}
	}

	if res := string(res); res != "" {
		msg = fmt.Sprintf("%s: %s", msg, res)
	}

	return msg
}
