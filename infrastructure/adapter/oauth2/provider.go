package oauth2

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	appoauth "github.com/lpxxn/blink/application/oauth"
	"golang.org/x/oauth2"
)

// Provider implements application/oauth.OAuth2Provider using a generic OAuth2 config + userinfo JSON.
type Provider struct {
	Config      *oauth2.Config
	UserInfoURL string
	HTTPClient  *http.Client
}

func (p *Provider) client() *http.Client {
	if p.HTTPClient != nil {
		return p.HTTPClient
	}
	return http.DefaultClient
}

func (p *Provider) AuthCodeURL(state string) string {
	return p.Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (p *Provider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.Config.Exchange(ctx, code)
}

func (p *Provider) UserInfo(ctx context.Context, token *oauth2.Token) (appoauth.UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.UserInfoURL, nil)
	if err != nil {
		return appoauth.UserInfo{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	resp, err := p.client().Do(req)
	if err != nil {
		return appoauth.UserInfo{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return appoauth.UserInfo{}, fmt.Errorf("userinfo: %s: %s", resp.Status, string(b))
	}
	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return appoauth.UserInfo{}, err
	}
	info := appoauth.UserInfo{}
	if s, ok := raw["sub"].(string); ok {
		info.Subject = strings.TrimSpace(s)
	}
	if info.Subject == "" {
		info.Subject = subjectFromIDField(raw["id"])
	}
	if s, ok := raw["email"].(string); ok {
		info.Email = strings.TrimSpace(s)
	}
	if s, ok := raw["name"].(string); ok {
		info.Name = strings.TrimSpace(s)
	}
	if s, ok := raw["login"].(string); ok && info.Name == "" {
		info.Name = strings.TrimSpace(s)
	}
	if info.Subject == "" {
		return info, fmt.Errorf("userinfo: missing subject")
	}
	return info, nil
}

func subjectFromIDField(v any) string {
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case float64:
		return strconv.FormatInt(int64(x), 10)
	case json.Number:
		return string(x)
	default:
		return ""
	}
}
