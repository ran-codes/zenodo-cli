package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/ran-codes/zenodo-cli/internal/cli"
	"github.com/ran-codes/zenodo-cli/internal/model"
)

func main() {
	err, outputFmt := cli.Execute()
	if err != nil {
		code := exitCode(err)

		if outputFmt == "json" {
			// Structured JSON error to stderr.
			errObj := map[string]interface{}{
				"error": err.Error(),
				"code":  code,
			}
			var apiErr *model.APIError
			if errors.As(err, &apiErr) {
				errObj["status"] = apiErr.Status
				errObj["message"] = apiErr.Message
				if hint := apiErr.Hint(); hint != "" {
					errObj["hint"] = hint
				}
			}
			json.NewEncoder(os.Stderr).Encode(errObj)
		} else {
			// Plain text error to stderr.
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		}

		os.Exit(code)
	}
}

func exitCode(err error) int {
	var apiErr *model.APIError
	if errors.As(err, &apiErr) {
		switch {
		case apiErr.Status == 401 || apiErr.Status == 403:
			return 2 // Auth error
		case apiErr.LikelyAuthError():
			return 2 // Auth error (Zenodo returns 400 for bad tokens on some endpoints)
		case apiErr.Status == 429:
			return 4 // Rate limit
		default:
			return 1 // API error
		}
	}

	// Check for validation error message pattern.
	if err.Error() == "metadata validation failed" {
		return 3
	}

	return 1
}
