package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/authidentity"
	dbuser "github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCasdoorOAuthLoginRedirectsAndSetsCookies(t *testing.T) {
	handler, _ := newCasdoorOAuthTestHandler(t, casdoorOAuthTestOptions{
		cfg: testCasdoorConfig("https://login.example.com"),
	})

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/casdoor/login?redirect=/billing", nil)

	handler.CasdoorOAuthLogin(c)

	require.Equal(t, http.StatusFound, recorder.Code)
	location := recorder.Header().Get("Location")
	parsed, err := url.Parse(location)
	require.NoError(t, err)
	require.Equal(t, "https", parsed.Scheme)
	require.Equal(t, "login.example.com", parsed.Host)
	require.Equal(t, "/login/oauth/authorize", parsed.Path)
	query := parsed.Query()
	require.Equal(t, "casdoor-client", query.Get("client_id"))
	require.Equal(t, "code", query.Get("response_type"))
	require.Equal(t, "https://token.example.com/api/v1/auth/casdoor/callback", query.Get("redirect_uri"))
	require.Equal(t, "openid profile email phone", query.Get("scope"))
	require.NotEmpty(t, query.Get("state"))

	cookies := recorder.Result().Cookies()
	require.NotNil(t, findCookie(cookies, casdoorOAuthStateCookieName))
	require.NotNil(t, findCookie(cookies, casdoorOAuthRedirectCookieName))
	require.NotNil(t, findCookie(cookies, casdoorOAuthBrowserCookieName))
	require.Equal(t, "/billing", decodeCookieValueForTest(t, findCookie(cookies, casdoorOAuthRedirectCookieName).Value))
}

func TestCasdoorOAuthLoginAddsPhoneScopeWhenConfiguredScopeOmitsIt(t *testing.T) {
	cfg := testCasdoorConfig("https://login.example.com")
	cfg.Scope = "openid profile email"
	handler, _ := newCasdoorOAuthTestHandler(t, casdoorOAuthTestOptions{cfg: cfg})

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/casdoor/login", nil)

	handler.CasdoorOAuthLogin(c)

	require.Equal(t, http.StatusFound, recorder.Code)
	location := recorder.Header().Get("Location")
	parsed, err := url.Parse(location)
	require.NoError(t, err)
	require.Equal(t, "openid profile email phone", parsed.Query().Get("scope"))
}

func TestCasdoorOAuthCallbackIssuesTicketForExistingIdentity(t *testing.T) {
	cfg := testCasdoorConfig("")
	provider, closeProvider := newCasdoorTestProvider(t, casdoorProviderFixture{
		Subject: "casdoor-sub-1",
		Email:   "alice@example.com",
		Name:    "Alice",
	})
	defer closeProvider()
	cfg.Issuer = provider.issuer
	cfg.AuthorizeURL = provider.authorizeURL
	cfg.TokenURL = provider.tokenURL
	cfg.UserInfoURL = provider.userInfoURL

	handler, client := newCasdoorOAuthTestHandler(t, casdoorOAuthTestOptions{cfg: cfg})
	user := createCasdoorTestUser(t, client, "alice@example.com", "", "Alice Local")
	_, err := client.AuthIdentity.Create().
		SetUserID(user.ID).
		SetProviderType("oidc").
		SetProviderKey(cfg.Issuer).
		SetProviderSubject("casdoor-sub-1").
		Save(context.Background())
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/casdoor/callback?code=provider-code&state=state-ok", nil)
	req.AddCookie(&http.Cookie{Name: casdoorOAuthStateCookieName, Value: encodeCookieValue("state-ok")})
	req.AddCookie(&http.Cookie{Name: casdoorOAuthRedirectCookieName, Value: encodeCookieValue("/dashboard")})
	req.AddCookie(&http.Cookie{Name: casdoorOAuthBrowserCookieName, Value: encodeCookieValue("browser-key")})
	c.Request = req

	handler.CasdoorOAuthCallback(c)

	require.Equal(t, http.StatusFound, recorder.Code)
	location := recorder.Header().Get("Location")
	parsed, err := url.Parse(location)
	require.NoError(t, err)
	require.Equal(t, "/auth/casdoor/success", parsed.Path)
	require.Equal(t, "/dashboard", parsed.Query().Get("redirect"))
	require.NotEmpty(t, parsed.Query().Get("ticket"))

	session, err := service.NewAuthPendingIdentityService(client).ConsumeCompletionCode(
		context.Background(),
		parsed.Query().Get("ticket"),
		"browser-key",
	)
	require.NoError(t, err)
	require.Equal(t, user.ID, *session.TargetUserID)
}

func TestCasdoorOAuthCallbackBindsExistingEmailUser(t *testing.T) {
	cfg := testCasdoorConfig("")
	provider, closeProvider := newCasdoorTestProvider(t, casdoorProviderFixture{
		Subject: "casdoor-sub-email",
		Email:   "bob@example.com",
		Name:    "Bob",
	})
	defer closeProvider()
	cfg.Issuer = provider.issuer
	cfg.TokenURL = provider.tokenURL
	cfg.UserInfoURL = provider.userInfoURL

	handler, client := newCasdoorOAuthTestHandler(t, casdoorOAuthTestOptions{cfg: cfg})
	user := createCasdoorTestUser(t, client, "bob@example.com", "", "Bob Local")

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/casdoor/callback?code=provider-code&state=state-ok", nil)
	req.AddCookie(&http.Cookie{Name: casdoorOAuthStateCookieName, Value: encodeCookieValue("state-ok")})
	req.AddCookie(&http.Cookie{Name: casdoorOAuthBrowserCookieName, Value: encodeCookieValue("browser-key")})
	c.Request = req

	handler.CasdoorOAuthCallback(c)

	require.Equal(t, http.StatusFound, recorder.Code)
	identity, err := client.AuthIdentity.Query().
		Where(
			authidentity.ProviderTypeEQ("oidc"),
			authidentity.ProviderKeyEQ(cfg.Issuer),
			authidentity.ProviderSubjectEQ("casdoor-sub-email"),
		).
		Only(context.Background())
	require.NoError(t, err)
	require.Equal(t, user.ID, identity.UserID)
}

func TestCasdoorOAuthCallbackSyncsPhoneForExistingEmailUser(t *testing.T) {
	cfg := testCasdoorConfig("")
	provider, closeProvider := newCasdoorTestProvider(t, casdoorProviderFixture{
		Subject: "casdoor-sub-phone-sync",
		Email:   "phone-sync@example.com",
		Phone:   "13800138000",
		Name:    "Phone Sync",
	})
	defer closeProvider()
	cfg.Issuer = provider.issuer
	cfg.TokenURL = provider.tokenURL
	cfg.UserInfoURL = provider.userInfoURL

	handler, client := newCasdoorOAuthTestHandler(t, casdoorOAuthTestOptions{cfg: cfg})
	user := createCasdoorTestUser(t, client, "phone-sync@example.com", "", "Phone Sync Local")

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/casdoor/callback?code=provider-code&state=state-ok", nil)
	req.AddCookie(&http.Cookie{Name: casdoorOAuthStateCookieName, Value: encodeCookieValue("state-ok")})
	req.AddCookie(&http.Cookie{Name: casdoorOAuthBrowserCookieName, Value: encodeCookieValue("browser-key")})
	c.Request = req

	handler.CasdoorOAuthCallback(c)

	require.Equal(t, http.StatusFound, recorder.Code)
	updated, err := client.User.Get(context.Background(), user.ID)
	require.NoError(t, err)
	require.Equal(t, "+8613800138000", updated.PhoneNumber)
}

func TestCasdoorOAuthCallbackRejectsExistingIdentityWhenPhoneBelongsToAnotherUser(t *testing.T) {
	cfg := testCasdoorConfig("")
	provider, closeProvider := newCasdoorTestProvider(t, casdoorProviderFixture{
		Subject: "casdoor-sub-phone-conflict",
		Email:   "identity-owner@example.com",
		Phone:   "+8613711112222",
		Name:    "Phone Conflict",
	})
	defer closeProvider()
	cfg.Issuer = provider.issuer
	cfg.TokenURL = provider.tokenURL
	cfg.UserInfoURL = provider.userInfoURL

	handler, client := newCasdoorOAuthTestHandler(t, casdoorOAuthTestOptions{cfg: cfg})
	identityOwner := createCasdoorTestUser(t, client, "identity-owner@example.com", "", "Identity Owner")
	createCasdoorTestUser(t, client, "phone-owner@example.com", "+8613711112222", "Phone Owner")
	_, err := client.AuthIdentity.Create().
		SetUserID(identityOwner.ID).
		SetProviderType("oidc").
		SetProviderKey(cfg.Issuer).
		SetProviderSubject("casdoor-sub-phone-conflict").
		Save(context.Background())
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/casdoor/callback?code=provider-code&state=state-ok", nil)
	req.AddCookie(&http.Cookie{Name: casdoorOAuthStateCookieName, Value: encodeCookieValue("state-ok")})
	req.AddCookie(&http.Cookie{Name: casdoorOAuthBrowserCookieName, Value: encodeCookieValue("browser-key")})
	c.Request = req

	handler.CasdoorOAuthCallback(c)

	require.Equal(t, http.StatusFound, recorder.Code)
	location := recorder.Header().Get("Location")
	parsed, err := url.Parse(location)
	require.NoError(t, err)
	require.Equal(t, "/auth/casdoor/error", parsed.Path)
	require.Equal(t, "account_conflict", parsed.Query().Get("error"))
}

func TestCasdoorExchangeTicketConsumesOnceAndReturnsProductJWT(t *testing.T) {
	handler, client := newCasdoorOAuthTestHandler(t, casdoorOAuthTestOptions{
		cfg: testCasdoorConfig("https://login.example.com"),
	})
	user := createCasdoorTestUser(t, client, "carol@example.com", "", "Carol")
	pendingSvc := service.NewAuthPendingIdentityService(client)
	session, err := pendingSvc.CreatePendingSession(context.Background(), service.CreatePendingAuthSessionInput{
		Intent: "login",
		Identity: service.PendingAuthIdentityKey{
			ProviderType:    "oidc",
			ProviderKey:     "https://login.example.com",
			ProviderSubject: "casdoor-sub-carol",
		},
		TargetUserID:           &user.ID,
		BrowserSessionKey:      "browser-key",
		ResolvedEmail:          user.Email,
		UpstreamIdentityClaims: map[string]any{"email": user.Email},
	})
	require.NoError(t, err)
	ticket, err := pendingSvc.IssueCompletionCode(context.Background(), service.IssuePendingAuthCompletionCodeInput{
		PendingAuthSessionID: session.ID,
		BrowserSessionKey:    "browser-key",
	})
	require.NoError(t, err)

	body := bytes.NewBufferString(`{"ticket":"` + ticket.Code + `"}`)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/casdoor/exchange-ticket", body)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: casdoorOAuthBrowserCookieName, Value: encodeCookieValue("browser-key")})
	c.Request = req

	handler.CasdoorExchangeTicket(c)

	require.Equal(t, http.StatusOK, recorder.Code)
	var envelope struct {
		Code int `json:"code"`
		Data struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			ExpiresIn    int    `json:"expires_in"`
			TokenType    string `json:"token_type"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
	require.Equal(t, 0, envelope.Code)
	require.NotEmpty(t, envelope.Data.AccessToken)
	require.NotEmpty(t, envelope.Data.RefreshToken)
	require.Equal(t, "Bearer", envelope.Data.TokenType)

	secondRecorder := httptest.NewRecorder()
	secondCtx, _ := gin.CreateTestContext(secondRecorder)
	secondReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/casdoor/exchange-ticket", bytes.NewBufferString(`{"ticket":"`+ticket.Code+`"}`))
	secondReq.Header.Set("Content-Type", "application/json")
	secondReq.AddCookie(&http.Cookie{Name: casdoorOAuthBrowserCookieName, Value: encodeCookieValue("browser-key")})
	secondCtx.Request = secondReq

	handler.CasdoorExchangeTicket(secondCtx)

	require.Equal(t, http.StatusUnauthorized, secondRecorder.Code)
}

func TestCasdoorOAuthCallbackRejectsMismatchedEmailAndPhoneUsers(t *testing.T) {
	cfg := testCasdoorConfig("")
	provider, closeProvider := newCasdoorTestProvider(t, casdoorProviderFixture{
		Subject: "casdoor-sub-conflict",
		Email:   "email-owner@example.com",
		Phone:   "+8613711112222",
		Name:    "Conflict",
	})
	defer closeProvider()
	cfg.Issuer = provider.issuer
	cfg.TokenURL = provider.tokenURL
	cfg.UserInfoURL = provider.userInfoURL

	handler, client := newCasdoorOAuthTestHandler(t, casdoorOAuthTestOptions{cfg: cfg})
	createCasdoorTestUser(t, client, "email-owner@example.com", "", "Email Owner")
	createCasdoorTestUser(t, client, "phone-owner@example.com", "+8613711112222", "Phone Owner")

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/casdoor/callback?code=provider-code&state=state-ok", nil)
	req.AddCookie(&http.Cookie{Name: casdoorOAuthStateCookieName, Value: encodeCookieValue("state-ok")})
	req.AddCookie(&http.Cookie{Name: casdoorOAuthBrowserCookieName, Value: encodeCookieValue("browser-key")})
	c.Request = req

	handler.CasdoorOAuthCallback(c)

	require.Equal(t, http.StatusFound, recorder.Code)
	location := recorder.Header().Get("Location")
	parsed, err := url.Parse(location)
	require.NoError(t, err)
	require.Equal(t, "/auth/casdoor/error", parsed.Path)
	require.Equal(t, "account_conflict", parsed.Query().Get("error"))
}

func TestParseCasdoorUserInfoReadsPhoneAliases(t *testing.T) {
	require.Equal(t, "13800138000", parseCasdoorUserInfo(`{"sub":"sub-1","phone_number":"13800138000"}`).Phone)
	require.Equal(t, "13800138001", parseCasdoorUserInfo(`{"sub":"sub-1","data":{"phone_number":"13800138001"}}`).Phone)
	require.Equal(t, "13800138002", parseCasdoorUserInfo(`{"sub":"sub-1","data":{"phoneNumber":"13800138002"}}`).Phone)
	require.Equal(t, "13800138003", parseCasdoorUserInfo(`{"sub":"sub-1","mobile":"13800138003"}`).Phone)
	require.Equal(t, "13800138004", parseCasdoorUserInfo(`{"sub":"sub-1","data":{"mobile":"13800138004"}}`).Phone)
}

type casdoorOAuthTestOptions struct {
	cfg                 config.CasdoorConfig
	registrationEnabled bool
}

func newCasdoorOAuthTestHandler(t *testing.T, options casdoorOAuthTestOptions) (*AuthHandler, *dbent.Client) {
	t.Helper()
	if !options.registrationEnabled {
		options.registrationEnabled = true
	}
	handler, client := newOAuthPendingFlowTestHandlerWithDependencies(t, oauthPendingFlowTestHandlerOptions{
		settingValues: map[string]string{
			service.SettingKeyRegistrationEnabled: boolSettingValue(options.registrationEnabled),
		},
	})
	if handler.cfg == nil {
		handler.cfg = &config.Config{}
	}
	handler.cfg.Casdoor = options.cfg
	return handler, client
}

func testCasdoorConfig(issuer string) config.CasdoorConfig {
	if issuer == "" {
		issuer = "https://login.example.com"
	}
	return config.CasdoorConfig{
		Enabled:      true,
		Issuer:       issuer,
		ClientID:     "casdoor-client",
		ClientSecret: "casdoor-secret",
		RedirectURI:  "https://token.example.com/api/v1/auth/casdoor/callback",
		Scope:        "openid profile email phone",
		AuthorizeURL: issuer + "/login/oauth/authorize",
		TokenURL:     issuer + "/api/login/oauth/access_token",
		UserInfoURL:  issuer + "/api/userinfo",
	}
}

type casdoorProviderFixture struct {
	Subject string
	Email   string
	Phone   string
	Name    string
}

type casdoorTestProvider struct {
	issuer       string
	authorizeURL string
	tokenURL     string
	userInfoURL  string
}

func newCasdoorTestProvider(t *testing.T, fixture casdoorProviderFixture) (casdoorTestProvider, func()) {
	t.Helper()
	var issuer string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/login/oauth/access_token":
			require.Equal(t, "authorization_code", r.FormValue("grant_type"))
			require.Equal(t, "casdoor-client", r.FormValue("client_id"))
			require.Equal(t, "casdoor-secret", r.FormValue("client_secret"))
			require.Equal(t, "provider-code", r.FormValue("code"))
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "casdoor-access",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
		case "/api/userinfo":
			require.Equal(t, "Bearer casdoor-access", r.Header.Get("Authorization"))
			_ = json.NewEncoder(w).Encode(map[string]any{
				"sub":   fixture.Subject,
				"email": fixture.Email,
				"phone": fixture.Phone,
				"name":  fixture.Name,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	issuer = server.URL
	return casdoorTestProvider{
		issuer:       issuer,
		authorizeURL: issuer + "/login/oauth/authorize",
		tokenURL:     issuer + "/api/login/oauth/access_token",
		userInfoURL:  issuer + "/api/userinfo",
	}, server.Close
}

func createCasdoorTestUser(t *testing.T, client *dbent.Client, email, phone, username string) *dbent.User {
	t.Helper()
	user, err := client.User.Create().
		SetEmail(email).
		SetPhoneNumber(phone).
		SetUsername(username).
		SetPasswordHash("hashed-password").
		SetRole(service.RoleUser).
		SetStatus(service.StatusActive).
		SetSignupSource("email").
		SetConcurrency(1).
		Save(context.Background())
	require.NoError(t, err)

	loaded, err := client.User.Query().Where(dbuser.IDEQ(user.ID)).Only(context.Background())
	require.NoError(t, err)
	return loaded
}
