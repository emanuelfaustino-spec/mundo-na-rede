package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type driveTransport func(*http.Request) (*http.Response, error)

func (f driveTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestDriveUploadAndRetry(t *testing.T) {
	name := "MundoNaRede-test.zip"
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte("backup payload"), 0600); err != nil {
		t.Fatal(err)
	}
	requests := 0
	client := &http.Client{Transport: driveTransport(func(r *http.Request) (*http.Response, error) {
		requests++
		res := &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("{}"))}
		switch requests {
		case 1, 4:
			if r.Method != "GET" || !strings.Contains(r.URL.Query().Get("q"), name) {
				t.Fatalf("unexpected lookup %v", r.URL)
			}
			if requests == 4 {
				res.Body = io.NopCloser(strings.NewReader(`{"files":[{"id":"already-uploaded"}]}`))
			}
		case 2:
			if r.Method != "POST" || r.URL.Query().Get("uploadType") != "resumable" {
				t.Fatalf("unexpected upload setup %v", r.URL)
			}
			res.Header.Set("Location", "https://www.googleapis.com/upload/drive/v3/files?upload_id=test")
		case 3:
			payload, err := io.ReadAll(r.Body)
			if err != nil || r.Method != "PUT" || string(payload) != "backup payload" {
				t.Fatalf("invalid upload %q %v", payload, err)
			}
		default:
			t.Fatalf("unexpected request %v", r.URL)
		}
		return res, nil
	})}
	for i := 0; i < 2; i++ {
		if err := sendDriveSnapshot(context.Background(), client, path, name); err != nil {
			t.Fatal(err)
		}
	}
	if requests != 4 {
		t.Fatalf("got %d requests, wanted 4", requests)
	}
}
