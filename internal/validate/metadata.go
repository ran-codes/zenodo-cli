package validate

import (
	"fmt"
	"strings"

	"github.com/ran-codes/zenodo-cli/internal/model"
)

var validUploadTypes = map[string]bool{
	"publication": true, "poster": true, "presentation": true,
	"dataset": true, "image": true, "video": true,
	"software": true, "lesson": true, "physicalobject": true,
	"other": true,
}

var validAccessRights = map[string]bool{
	"open": true, "embargoed": true, "restricted": true, "closed": true,
}

// Metadata validates required fields on a metadata struct.
// Returns a slice of validation errors (empty if valid).
func Metadata(m model.Metadata) []string {
	var errs []string

	if m.Title == "" {
		errs = append(errs, "title is required")
	}

	if m.Description == "" {
		errs = append(errs, "description is required")
	}

	if m.UploadType == "" {
		errs = append(errs, "upload_type is required")
	} else if !validUploadTypes[m.UploadType] {
		errs = append(errs, fmt.Sprintf("upload_type %q is invalid; valid types: %s",
			m.UploadType, joinKeys(validUploadTypes)))
	}

	if len(m.Creators) == 0 {
		errs = append(errs, "at least one creator is required")
	} else {
		for i, c := range m.Creators {
			if c.Name == "" {
				errs = append(errs, fmt.Sprintf("creators[%d].name is required", i))
			}
		}
	}

	if m.PublicationDate == "" {
		errs = append(errs, "publication_date is required")
	}

	if m.AccessRight == "" {
		errs = append(errs, "access_right is required")
	} else {
		if !validAccessRights[m.AccessRight] {
			errs = append(errs, fmt.Sprintf("access_right %q is invalid; valid values: %s",
				m.AccessRight, joinKeys(validAccessRights)))
		}

		// License required for open/embargoed.
		if (m.AccessRight == "open" || m.AccessRight == "embargoed") && m.LicenseString() == "" {
			errs = append(errs, "license is required when access_right is open or embargoed")
		}

		// Embargo date required for embargoed.
		if m.AccessRight == "embargoed" && m.EmbargoDate == "" {
			errs = append(errs, "embargo_date is required when access_right is embargoed")
		}

		// Access conditions required for restricted.
		if m.AccessRight == "restricted" && m.AccessConditions == "" {
			errs = append(errs, "access_conditions is required when access_right is restricted")
		}
	}

	return errs
}

func joinKeys(m map[string]bool) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}
