package workosclient

import (
	"fmt"
	"net/url"
	"strings"
)

type AuthorizationURLOpts struct {
	RedirectURI string
	State       string
}

type Client interface {
	Configured() bool
	ClientID() string
	AuthorizationURL(opts AuthorizationURLOpts) (string, error)
}

type SDKClient struct {
	apiKey   string
	clientID string
}

var authorizationBaseURL = "https://api.workos.com/user_management/authorize"

func New(apiKey, clientID string) *SDKClient {
	return &SDKClient{
		apiKey:   strings.TrimSpace(apiKey),
		clientID: strings.TrimSpace(clientID),
	}
}

func (c *SDKClient) Configured() bool {
	return c.apiKey != "" && c.clientID != ""
}

func (c *SDKClient) ClientID() string {
	return c.clientID
}

func (c *SDKClient) AuthorizationURL(opts AuthorizationURLOpts) (string, error) {
	if !c.Configured() {
		return "", fmt.Errorf("workos client is not configured")
	}

	u, err := url.Parse(authorizationBaseURL)
	if err != nil {
		return "", err
	}

	query := u.Query()
	query.Set("provider", "authkit")
	query.Set("client_id", c.clientID)
	query.Set("redirect_uri", strings.TrimSpace(opts.RedirectURI))
	query.Set("state", strings.TrimSpace(opts.State))
	u.RawQuery = query.Encode()

	return u.String(), nil
}

type FakeClient struct {
	ConfiguredValue       bool
	ClientIDValue         string
	AuthorizationURLValue string
	AuthorizationURLError error
}

func (f *FakeClient) Configured() bool {
	return f.ConfiguredValue
}

func (f *FakeClient) ClientID() string {
	if strings.TrimSpace(f.ClientIDValue) != "" {
		return f.ClientIDValue
	}
	return "fake-client-id"
}

func (f *FakeClient) AuthorizationURL(opts AuthorizationURLOpts) (string, error) {
	if f.AuthorizationURLError != nil {
		return "", f.AuthorizationURLError
	}
	if strings.TrimSpace(f.AuthorizationURLValue) != "" {
		return f.AuthorizationURLValue, nil
	}
	return (&SDKClient{apiKey: "fake-api-key", clientID: f.ClientID()}).AuthorizationURL(opts)
}
