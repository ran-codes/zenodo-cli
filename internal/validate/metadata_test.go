package validate

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ran-codes/zenodo-cli/internal/model"
)

func validMetadata() model.Metadata {
	return model.Metadata{
		Title:           "Test Record",
		Description:     "A test description",
		UploadType:      "dataset",
		PublicationDate: "2024-01-01",
		AccessRight:     "open",
		License:         json.RawMessage(`"cc-by-4.0"`),
		Creators:        []model.Creator{{Name: "Test Author"}},
	}
}

func TestValidMetadata(t *testing.T) {
	errs := Metadata(validMetadata())
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestMissingTitle(t *testing.T) {
	m := validMetadata()
	m.Title = ""
	errs := Metadata(m)
	if !containsError(errs, "title is required") {
		t.Errorf("expected title error, got: %v", errs)
	}
}

func TestMissingDescription(t *testing.T) {
	m := validMetadata()
	m.Description = ""
	errs := Metadata(m)
	if !containsError(errs, "description is required") {
		t.Errorf("expected description error, got: %v", errs)
	}
}

func TestInvalidUploadType(t *testing.T) {
	m := validMetadata()
	m.UploadType = "invalid_type"
	errs := Metadata(m)
	if !containsError(errs, "upload_type") {
		t.Errorf("expected upload_type error, got: %v", errs)
	}
}

func TestMissingCreators(t *testing.T) {
	m := validMetadata()
	m.Creators = nil
	errs := Metadata(m)
	if !containsError(errs, "creator") {
		t.Errorf("expected creator error, got: %v", errs)
	}
}

func TestCreatorMissingName(t *testing.T) {
	m := validMetadata()
	m.Creators = []model.Creator{{Name: ""}}
	errs := Metadata(m)
	if !containsError(errs, "creators[0].name") {
		t.Errorf("expected creator name error, got: %v", errs)
	}
}

func TestMissingLicenseForOpen(t *testing.T) {
	m := validMetadata()
	m.AccessRight = "open"
	m.License = nil
	errs := Metadata(m)
	if !containsError(errs, "license is required") {
		t.Errorf("expected license error, got: %v", errs)
	}
}

func TestMissingEmbargoDate(t *testing.T) {
	m := validMetadata()
	m.AccessRight = "embargoed"
	m.License = json.RawMessage(`"cc-by-4.0"`)
	m.EmbargoDate = ""
	errs := Metadata(m)
	if !containsError(errs, "embargo_date") {
		t.Errorf("expected embargo_date error, got: %v", errs)
	}
}

func TestMissingAccessConditions(t *testing.T) {
	m := validMetadata()
	m.AccessRight = "restricted"
	m.License = nil
	m.AccessConditions = ""
	errs := Metadata(m)
	if !containsError(errs, "access_conditions") {
		t.Errorf("expected access_conditions error, got: %v", errs)
	}
}

func TestClosedAccessNoLicenseRequired(t *testing.T) {
	m := validMetadata()
	m.AccessRight = "closed"
	m.License = nil
	errs := Metadata(m)
	if containsError(errs, "license") {
		t.Errorf("license should not be required for closed access, got: %v", errs)
	}
}

func TestAllFieldsMissing(t *testing.T) {
	errs := Metadata(model.Metadata{})
	if len(errs) < 5 {
		t.Errorf("expected at least 5 errors for empty metadata, got %d: %v", len(errs), errs)
	}
}

func containsError(errs []string, substr string) bool {
	for _, e := range errs {
		if strings.Contains(e, substr) {
			return true
		}
	}
	return false
}
