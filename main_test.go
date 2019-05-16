package main_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	main "github.com/cs3238-tsuzu/coding_challenge_01"
)

func parseURL(t *testing.T, ul string) *url.URL {
	t.Helper()
	u, err := url.Parse(ul)

	if err != nil {
		t.Fatal("url parse error", err)
	}

	return u
}

func TestInitHandler200(t *testing.T) {
	handler := main.InitHandler()
	server := httptest.NewServer(handler)
	defer server.Close()

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	const expectedCode = http.StatusOK

	req, err := http.NewRequest("GET", server.URL, nil)

	if err != nil {
		t.Fatal("new request error", err)
	}

	resp, err := client.Do(req)

	if err != nil {
		t.Fatal("http request error", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedCode {
		t.Fatalf("expected status: %d, but got %d", expectedCode, resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Fatal("content-typ error", ct)
	}

	type responseBody struct {
		Message string `json:"message"`
	}

	var body responseBody

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal("json parse error", err)
	}

	if body.Message != "Hello World!!" {
		t.Fatal("incorrect response body message", body.Message)
	}
}

func TestInitHandler404(t *testing.T) {
	handler := main.InitHandler()
	server := httptest.NewServer(handler)
	defer server.Close()

	type param struct {
		name, path, method string
	}

	params := []*param{
		{name: "/ POST", path: "/", method: "POST"},
		{name: "/endpoint GET", path: "/endpoint", method: "GET"},
		{name: "/endpoint POST", path: "/endpoint", method: "POST"},
	}

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	const expectedCode = http.StatusNotFound

	for i := range params {
		p := params[i]
		t.Run(p.name, func(t *testing.T) {
			u := parseURL(t, server.URL)
			u.Path = p.path

			req, err := http.NewRequest(p.method, u.String(), nil)

			if err != nil {
				t.Fatal("new request error", err)
			}

			resp, err := client.Do(req)

			if err != nil {
				t.Fatal("http request error", err)
			}
			resp.Body.Close()

			if resp.StatusCode != expectedCode {
				t.Fatalf("expected status: %d, but got %d", expectedCode, resp.StatusCode)
			}
		})
	}
}
