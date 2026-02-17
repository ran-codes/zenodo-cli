package main

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/ran-codes/zenodo-cli/internal/cli"
	"github.com/ran-codes/zenodo-cli/internal/model"
)

func main() {
	if err := cli.Execute(); err != nil {
		code := exitCode(err)

		// If output is json mode, emit structured error to stderr.
		if isJSONOutput() {
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

func isJSONOutput() bool {
	for _, arg := range os.Args {
		if arg == "--output" || arg == "-o" {
			return false // Next arg would be the value, check below
		}
		if arg == "--output=json" || arg == "-o=json" {
			return true
		}
	}
	// Check for -o json or --output json pattern.
	for i, arg := range os.Args {
		if (arg == "--output" || arg == "-o") && i+1 < len(os.Args) && os.Args[i+1] == "json" {
			return true
		}
	}
	return false
}
