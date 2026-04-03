package workosclient

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/workos/workos-go/v6/pkg/organizations"
	"github.com/workos/workos-go/v6/pkg/usermanagement"
)

type AuthorizationURLOpts struct {
	RedirectURI string
	State       string
	ScreenHint  string
}

type AuthenticateWithCodeOpts struct {
	Code      string
	IPAddress string
	UserAgent string
}

type AuthenticateWithRefreshTokenOpts struct {
	RefreshToken   string
	OrganizationID string
	IPAddress      string
	UserAgent      string
}

type AuthenticateResult struct {
	UserID         string
	Email          string
	DisplayName    string
	OrganizationID string
	AccessToken    string
	RefreshToken   string
}

type RefreshResult struct {
	AccessToken  string
	RefreshToken string
}

type Organization struct {
	ID   string
	Name string
}

type Client interface {
	Configured() bool
	ClientID() string
	AuthorizationURL(opts AuthorizationURLOpts) (string, error)
	AuthenticateWithCode(ctx context.Context, opts AuthenticateWithCodeOpts) (AuthenticateResult, error)
	AuthenticateWithRefreshToken(ctx context.Context, opts AuthenticateWithRefreshTokenOpts) (RefreshResult, error)
	GetOrganization(ctx context.Context, organizationID string) (Organization, error)
	LogoutURL(sessionID string, returnTo string) (string, error)
}

type SDKClient struct {
	apiKey          string
	clientID        string
	userManagement  *usermanagement.Client
	organizationAPI *organizations.Client
}

func New(apiKey, clientID string) *SDKClient {
	apiKey = strings.TrimSpace(apiKey)
	clientID = strings.TrimSpace(clientID)

	return &SDKClient{
		apiKey:   apiKey,
		clientID: clientID,
		userManagement: func() *usermanagement.Client {
			client := usermanagement.NewClient(apiKey)
			client.APIKey = apiKey
			return client
		}(),
		organizationAPI: &organizations.Client{
			APIKey:   apiKey,
			Endpoint: "https://api.workos.com",
		},
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

	screenHint := usermanagement.ScreenHint("")
	switch strings.TrimSpace(opts.ScreenHint) {
	case "", "sign-in":
		screenHint = usermanagement.SignIn
	case "sign-up":
		screenHint = usermanagement.SignUp
	default:
		return "", fmt.Errorf("invalid screen hint")
	}

	u, err := c.userManagement.GetAuthorizationURL(usermanagement.GetAuthorizationURLOpts{
		ClientID:    c.clientID,
		RedirectURI: strings.TrimSpace(opts.RedirectURI),
		Provider:    "authkit",
		State:       strings.TrimSpace(opts.State),
		ScreenHint:  screenHint,
	})
	if err != nil {
		return "", err
	}

	return u.String(), nil
}

func (c *SDKClient) AuthenticateWithCode(ctx context.Context, opts AuthenticateWithCodeOpts) (AuthenticateResult, error) {
	response, err := c.userManagement.AuthenticateWithCode(ctx, usermanagement.AuthenticateWithCodeOpts{
		ClientID:  c.clientID,
		Code:      strings.TrimSpace(opts.Code),
		IPAddress: strings.TrimSpace(opts.IPAddress),
		UserAgent: strings.TrimSpace(opts.UserAgent),
	})
	if err != nil {
		return AuthenticateResult{}, err
	}

	return AuthenticateResult{
		UserID:         response.User.ID,
		Email:          response.User.Email,
		DisplayName:    displayName(response.User.FirstName, response.User.LastName, response.User.Email),
		OrganizationID: response.OrganizationID,
		AccessToken:    response.AccessToken,
		RefreshToken:   response.RefreshToken,
	}, nil
}

func (c *SDKClient) AuthenticateWithRefreshToken(ctx context.Context, opts AuthenticateWithRefreshTokenOpts) (RefreshResult, error) {
	response, err := c.userManagement.AuthenticateWithRefreshToken(ctx, usermanagement.AuthenticateWithRefreshTokenOpts{
		ClientID:       c.clientID,
		RefreshToken:   strings.TrimSpace(opts.RefreshToken),
		OrganizationID: strings.TrimSpace(opts.OrganizationID),
		IPAddress:      strings.TrimSpace(opts.IPAddress),
		UserAgent:      strings.TrimSpace(opts.UserAgent),
	})
	if err != nil {
		return RefreshResult{}, err
	}

	return RefreshResult{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
	}, nil
}

func (c *SDKClient) GetOrganization(ctx context.Context, organizationID string) (Organization, error) {
	response, err := c.organizationAPI.GetOrganization(ctx, organizations.GetOrganizationOpts{
		Organization: strings.TrimSpace(organizationID),
	})
	if err != nil {
		return Organization{}, err
	}

	return Organization{
		ID:   response.ID,
		Name: response.Name,
	}, nil
}

func (c *SDKClient) LogoutURL(sessionID string, returnTo string) (string, error) {
	u, err := c.userManagement.GetLogoutURL(usermanagement.GetLogoutURLOpts{
		SessionID: strings.TrimSpace(sessionID),
		ReturnTo:  strings.TrimSpace(returnTo),
	})
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

type FakeClient struct {
	ConfiguredValue          bool
	ClientIDValue            string
	AuthorizationURLValue    string
	AuthorizationURLError    error
	AuthenticateResultValue  AuthenticateResult
	AuthenticateResultError  error
	RefreshResultValue       RefreshResult
	RefreshResultError       error
	OrganizationValue        Organization
	OrganizationError        error
	LogoutURLValue           string
	LogoutURLError           error
	LastAuthenticateCodeOpts AuthenticateWithCodeOpts
	LastRefreshOpts          AuthenticateWithRefreshTokenOpts
	LastOrganizationID       string
	LastLogoutSessionID      string
	LastLogoutReturnTo       string
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

	u, err := url.Parse("https://api.workos.com/user_management/authorize")
	if err != nil {
		return "", err
	}
	query := u.Query()
	query.Set("provider", "authkit")
	query.Set("client_id", f.ClientID())
	query.Set("redirect_uri", opts.RedirectURI)
	query.Set("state", opts.State)
	if opts.ScreenHint != "" {
		query.Set("screen_hint", opts.ScreenHint)
	}
	u.RawQuery = query.Encode()
	return u.String(), nil
}

func (f *FakeClient) AuthenticateWithCode(_ context.Context, opts AuthenticateWithCodeOpts) (AuthenticateResult, error) {
	f.LastAuthenticateCodeOpts = opts
	if f.AuthenticateResultError != nil {
		return AuthenticateResult{}, f.AuthenticateResultError
	}
	if f.AuthenticateResultValue.UserID != "" {
		return f.AuthenticateResultValue, nil
	}
	return AuthenticateResult{
		UserID:         "user_123",
		Email:          "user@example.com",
		DisplayName:    "Test User",
		OrganizationID: "org_123",
		AccessToken:    "header.payload.signature",
		RefreshToken:   "refresh_token",
	}, nil
}

func (f *FakeClient) AuthenticateWithRefreshToken(_ context.Context, opts AuthenticateWithRefreshTokenOpts) (RefreshResult, error) {
	f.LastRefreshOpts = opts
	if f.RefreshResultError != nil {
		return RefreshResult{}, f.RefreshResultError
	}
	if f.RefreshResultValue.AccessToken != "" {
		return f.RefreshResultValue, nil
	}
	return RefreshResult{
		AccessToken:  "header.payload.signature",
		RefreshToken: "new_refresh_token",
	}, nil
}

func (f *FakeClient) GetOrganization(_ context.Context, organizationID string) (Organization, error) {
	f.LastOrganizationID = organizationID
	if f.OrganizationError != nil {
		return Organization{}, f.OrganizationError
	}
	if f.OrganizationValue.ID != "" || f.OrganizationValue.Name != "" {
		return f.OrganizationValue, nil
	}
	return Organization{
		ID:   organizationID,
		Name: "Current workspace",
	}, nil
}

func (f *FakeClient) LogoutURL(sessionID string, returnTo string) (string, error) {
	f.LastLogoutSessionID = sessionID
	f.LastLogoutReturnTo = returnTo
	if f.LogoutURLError != nil {
		return "", f.LogoutURLError
	}
	if strings.TrimSpace(f.LogoutURLValue) != "" {
		return f.LogoutURLValue, nil
	}
	return "https://api.workos.com/user_management/sessions/logout?session_id=" + url.QueryEscape(sessionID) + "&return_to=" + url.QueryEscape(returnTo), nil
}

func displayName(firstName string, lastName string, email string) string {
	fullName := strings.TrimSpace(strings.TrimSpace(firstName) + " " + strings.TrimSpace(lastName))
	if fullName != "" {
		return fullName
	}
	if email != "" {
		localPart := strings.TrimSpace(strings.Split(email, "@")[0])
		if localPart != "" {
			return titleWords(strings.NewReplacer(".", " ", "_", " ", "-", " ").Replace(localPart))
		}
	}
	return "WorkOS User"
}

func titleWords(value string) string {
	parts := strings.Fields(strings.TrimSpace(value))
	for index, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(strings.ToLower(part))
		runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
		parts[index] = string(runes)
	}
	return strings.Join(parts, " ")
}
