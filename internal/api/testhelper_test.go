package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	testZoneId  = "11111111-1111-1111-1111-111111111111"
	testOrgId   = "22222222-2222-2222-2222-222222222222"
	testPrefix  = "/api"
	testToken   = "test-token"
	testCreated = "2026-11-16T00:00:00Z"
)

type recordedRequest struct {
	Method string
	Path   string
	Auth   string
	Body   map[string]interface{}
}

type fakeServer struct {
	t        *testing.T
	requests []recordedRequest
	status   int
	response string
}

// newFakeClient returns a client whose non-organization calls get status/response.
func newFakeClient(t *testing.T, status int, response string) (*APIClient, *fakeServer) {
	t.Helper()
	fs := &fakeServer{t: t, status: status, response: response}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == testPrefix+"/user/organization" {
			_, _ = io.WriteString(w, `{"id":"`+testOrgId+`","created":"`+testCreated+
				`","name":"org","ident":"org","allowed_ips":[]}`)
			return
		}

		rec := recordedRequest{
			Method: r.Method, Path: r.URL.Path, Auth: r.Header.Get("Authorization"),
		}
		raw, _ := io.ReadAll(r.Body)
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &rec.Body); err != nil {
				t.Errorf("request body is not a JSON object: %s", raw)
			}
		}
		fs.requests = append(fs.requests, rec)

		w.WriteHeader(fs.status)
		_, _ = io.WriteString(w, fs.response)
	}))
	t.Cleanup(srv.Close)

	client, err := NewAPIClient(testToken, srv.URL, testPrefix, testZoneId, false)
	if err != nil {
		t.Fatalf("NewAPIClient: %v", err)
	}
	return client, fs
}

func (fs *fakeServer) only() recordedRequest {
	fs.t.Helper()
	if len(fs.requests) != 1 {
		fs.t.Fatalf("expected 1 request, got %d", len(fs.requests))
	}
	return fs.requests[0]
}

func assertRequest(t *testing.T, got recordedRequest, method, path string) {
	t.Helper()
	if got.Method != method || got.Path != testPrefix+path {
		t.Errorf("request = %s %s, want %s %s", got.Method, got.Path, method, testPrefix+path)
	}
	if got.Auth != "Bearer "+testToken {
		t.Errorf("Authorization = %q", got.Auth)
	}
}

// assertJSONEqual compares v against want after a JSON round trip of both.
func assertJSONEqual(t *testing.T, field string, got interface{}, want string) {
	t.Helper()
	var wantV interface{}
	if err := json.Unmarshal([]byte(want), &wantV); err != nil {
		t.Fatalf("bad want JSON for %s: %v", field, err)
	}
	gotRaw, _ := json.Marshal(got)
	wantRaw, _ := json.Marshal(wantV)
	if string(gotRaw) != string(wantRaw) {
		t.Errorf("%s =\n  %s\nwant\n  %s", field, gotRaw, wantRaw)
	}
}
