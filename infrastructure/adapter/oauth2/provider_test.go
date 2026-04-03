package oauth2

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/oauth2"
)

func TestProvider_UserInfo_GoogleShape(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sub":"g1","email":"u@example.com","name":"U"}`))
	}))
	defer srv.Close()

	p := &Provider{
		UserInfoURL: srv.URL,
		HTTPClient:  srv.Client(),
	}
	info, err := p.UserInfo(context.Background(), &oauth2.Token{AccessToken: "t"})
	if err != nil {
		t.Fatal(err)
	}
	if info.Subject != "g1" || info.Email != "u@example.com" || info.Name != "U" {
		t.Fatalf("info: %+v", info)
	}
}

func TestProvider_UserInfo_GitHubShape(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":12345,"login":"octo","name":"Octo Cat"}`))
	}))
	defer srv.Close()

	p := &Provider{
		UserInfoURL: srv.URL,
		HTTPClient:  srv.Client(),
	}
	info, err := p.UserInfo(context.Background(), &oauth2.Token{AccessToken: "t"})
	if err != nil {
		t.Fatal(err)
	}
	if info.Subject != "12345" || info.Name != "Octo Cat" {
		t.Fatalf("info: %+v", info)
	}
}
