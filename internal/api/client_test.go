package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ran-codes/zenodo-cli/internal/model"
)

func TestGet_Success(t *testing.T) {
	expected := map[string]string{"title": "Test Record"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("expected Accept: application/json, got %s", r.Header.Get("Accept"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-token")
	var result map[string]string
	err := client.Get("/test", nil, &result)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if result["title"] != "Test Record" {
		t.Errorf("got title %q, want %q", result["title"], "Test Record")
	}
}

func TestGet_AuthHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer my-secret" {
			t.Errorf("expected 'Bearer my-secret', got %q", auth)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "my-secret")
	err := client.Get("/auth-test", nil, nil)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
}

func TestGet_NoAuthWhenEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "" {
			t.Errorf("expected no auth header, got %q", auth)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "")
	err := client.Get("/no-auth", nil, nil)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
}

func TestGet_QueryParams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") != "test" {
			t.Errorf("expected q=test, got %q", r.URL.Query().Get("q"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("expected page=2, got %q", r.URL.Query().Get("page"))
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "")
	q := make(map[string][]string)
	q["q"] = []string{"test"}
	q["page"] = []string{"2"}
	err := client.Get("/search", q, nil)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
}

func TestPost_SendsJSON(t *testing.T) {
	type reqBody struct {
		Title string `json:"title"`
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", ct)
		}
		var body reqBody
		json.NewDecoder(r.Body).Decode(&body)
		if body.Title != "New Title" {
			t.Errorf("got title %q, want %q", body.Title, "New Title")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": "123"})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "tok")
	var result map[string]string
	err := client.Post("/create", reqBody{Title: "New Title"}, &result)
	if err != nil {
		t.Fatalf("Post() error: %v", err)
	}
	if result["id"] != "123" {
		t.Errorf("got id %q, want %q", result["id"], "123")
	}
}

func TestPut_SendsJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "tok")
	err := client.Put("/update", map[string]string{"title": "Updated"}, nil)
	if err != nil {
		t.Fatalf("Put() error: %v", err)
	}
}

func TestDelete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "tok")
	err := client.Delete("/remove", nil)
	if err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
}

func TestAPIError_Structured(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(model.APIError{
			Status:  400,
			Message: "Validation error",
			Errors: []model.Detail{
				{Field: "title", Message: "required"},
			},
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "tok")
	err := client.Get("/bad", nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*model.APIError)
	if !ok {
		t.Fatalf("expected *model.APIError, got %T", err)
	}
	if apiErr.Status != 400 {
		t.Errorf("status = %d, want 400", apiErr.Status)
	}
	if apiErr.Message != "Validation error" {
		t.Errorf("message = %q, want %q", apiErr.Message, "Validation error")
	}
	if len(apiErr.Errors) != 1 || apiErr.Errors[0].Field != "title" {
		t.Errorf("expected field error on title, got %+v", apiErr.Errors)
	}
}

func TestAPIError_401Hint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"status":401,"message":"Unauthorized"}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "bad-token")
	err := client.Get("/secret", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr := err.(*model.APIError)
	if apiErr.Status != 401 {
		t.Errorf("status = %d, want 401", apiErr.Status)
	}
	if hint := apiErr.Hint(); hint == "" {
		t.Error("expected hint for 401, got empty")
	}
}

func TestAPIError_NonJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "tok")
	err := client.Get("/crash", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr := err.(*model.APIError)
	if apiErr.Status != 500 {
		t.Errorf("status = %d, want 500", apiErr.Status)
	}
}

func TestGetRaw(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept")
		if accept != "application/x-bibtex" {
			t.Errorf("expected Accept application/x-bibtex, got %q", accept)
		}
		w.Write([]byte("@article{test, title={Test}}"))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "tok")
	data, err := client.GetRaw("/record/123", "application/x-bibtex")
	if err != nil {
		t.Fatalf("GetRaw() error: %v", err)
	}
	if string(data) != "@article{test, title={Test}}" {
		t.Errorf("got %q", string(data))
	}
}
