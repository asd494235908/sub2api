package handler

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/authidentity"
	dbuser "github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/oauth"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/imroc/req/v3"
	"github.com/tidwall/gjson"
)

const (
	casdoorOAuthCookiePath          = "/api/v1/auth/casdoor"
	casdoorOAuthStateCookieName     = "casdoor_oauth_state"
	casdoorOAuthRedirectCookieName  = "casdoor_oauth_redirect"
	casdoorOAuthBrowserCookieName   = "casdoor_oauth_browser"
	casdoorOAuthAffiliateCookieName = "casdoor_oauth_affiliate"
	casdoorOAuthCookieMaxAgeSec     = 10 * 60
	casdoorOAuthTicketTTL           = 2 * time.Minute
	casdoorOAuthDefaultRedirectTo   = "/dashboard"
	casdoorOAuthSuccessPath         = "/auth/casdoor/success"
	casdoorOAuthErrorPath           = "/auth/casdoor/error"
)

type casdoorTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

type casdoorUserInfo struct {
	Subject string
	Email   string
	Phone   string
	Name    string
}

type casdoorExchangeTicketRequest struct {
	Ticket string `json:"ticket" binding:"required"`
}

func (h *AuthHandler) CasdoorOAuthLogin(c *gin.Context) {
	cfg, err := h.getCasdoorOAuthConfig()
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	state, err := oauth.GenerateState()
	if err != nil {
		response.ErrorFrom(c, infraerrors.InternalServer("CASDOOR_STATE_GEN_FAILED", "failed to generate casdoor state").WithCause(err))
		return
	}
	browserKey, err := oauth.GenerateState()
	if err != nil {
		response.ErrorFrom(c, infraerrors.InternalServer("CASDOOR_BROWSER_SESSION_GEN_FAILED", "failed to generate casdoor browser session").WithCause(err))
		return
	}

	redirectTo := sanitizeFrontendRedirectPath(c.Query("redirect"))
	if redirectTo == "" {
		redirectTo = casdoorOAuthDefaultRedirectTo
	}

	secureCookie := isRequestHTTPS(c)
	setCasdoorCookie(c, casdoorOAuthStateCookieName, encodeCookieValue(state), casdoorOAuthCookieMaxAgeSec, secureCookie)
	setCasdoorCookie(c, casdoorOAuthRedirectCookieName, encodeCookieValue(redirectTo), casdoorOAuthCookieMaxAgeSec, secureCookie)
	setCasdoorCookie(c, casdoorOAuthBrowserCookieName, encodeCookieValue(browserKey), casdoorOAuthCookieMaxAgeSec, secureCookie)
	if affCode := strings.TrimSpace(firstNonEmpty(c.Query("aff_code"), c.Query("aff"))); affCode != "" {
		setCasdoorCookie(c, casdoorOAuthAffiliateCookieName, encodeCookieValue(affCode), casdoorOAuthCookieMaxAgeSec, secureCookie)
	} else {
		clearCasdoorCookie(c, casdoorOAuthAffiliateCookieName, secureCookie)
	}

	authorizeURL, err := buildCasdoorAuthorizeURL(cfg, state)
	if err != nil {
		response.ErrorFrom(c, infraerrors.InternalServer("CASDOOR_BUILD_URL_FAILED", "failed to build casdoor authorization url").WithCause(err))
		return
	}
	c.Redirect(http.StatusFound, authorizeURL)
}

func (h *AuthHandler) CasdoorOAuthCallback(c *gin.Context) {
	cfg, cfgErr := h.getCasdoorOAuthConfig()
	if cfgErr != nil {
		response.ErrorFrom(c, cfgErr)
		return
	}

	if providerErr := strings.TrimSpace(c.Query("error")); providerErr != "" {
		redirectCasdoorError(c, "provider_error", providerErr, c.Query("error_description"))
		return
	}

	code := strings.TrimSpace(c.Query("code"))
	state := strings.TrimSpace(c.Query("state"))
	if code == "" || state == "" {
		redirectCasdoorError(c, "missing_params", "missing code/state", "")
		return
	}

	secureCookie := isRequestHTTPS(c)
	defer func() {
		clearCasdoorCookie(c, casdoorOAuthStateCookieName, secureCookie)
		clearCasdoorCookie(c, casdoorOAuthRedirectCookieName, secureCookie)
		clearCasdoorCookie(c, casdoorOAuthAffiliateCookieName, secureCookie)
	}()

	expectedState, err := readCookieDecoded(c, casdoorOAuthStateCookieName)
	if err != nil || expectedState == "" || state != expectedState {
		redirectCasdoorError(c, "invalid_state", "invalid casdoor state", "")
		return
	}

	redirectTo, _ := readCookieDecoded(c, casdoorOAuthRedirectCookieName)
	redirectTo = sanitizeFrontendRedirectPath(redirectTo)
	if redirectTo == "" {
		redirectTo = casdoorOAuthDefaultRedirectTo
	}

	browserKey, _ := readCookieDecoded(c, casdoorOAuthBrowserCookieName)
	if strings.TrimSpace(browserKey) == "" {
		redirectCasdoorError(c, "missing_browser_session", "missing casdoor browser session", "")
		return
	}
	affiliateCode, _ := readCookieDecoded(c, casdoorOAuthAffiliateCookieName)

	tokenResp, err := casdoorExchangeCode(c.Request.Context(), cfg, code)
	if err != nil {
		slog.Warn("casdoor token exchange failed", "error", err)
		redirectCasdoorError(c, "token_exchange_failed", "failed to exchange casdoor code", singleLine(err.Error()))
		return
	}
	userInfo, err := casdoorFetchUserInfo(c.Request.Context(), cfg, tokenResp)
	if err != nil {
		slog.Warn("casdoor userinfo failed", "error", err)
		redirectCasdoorError(c, "userinfo_failed", "failed to fetch casdoor userinfo", "")
		return
	}

	session, err := h.resolveCasdoorLoginSession(c.Request.Context(), cfg, userInfo, redirectTo, browserKey, affiliateCode)
	if err != nil {
		redirectCasdoorError(c, casdoorErrorCode(err), infraerrors.Message(err), "")
		return
	}

	pendingSvc, err := h.pendingIdentityService()
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	ticket, err := pendingSvc.IssueCompletionCode(c.Request.Context(), service.IssuePendingAuthCompletionCodeInput{
		PendingAuthSessionID: session.ID,
		BrowserSessionKey:    browserKey,
		TTL:                  casdoorOAuthTicketTTL,
	})
	if err != nil {
		response.ErrorFrom(c, infraerrors.InternalServer("CASDOOR_TICKET_ISSUE_FAILED", "failed to issue casdoor ticket").WithCause(err))
		return
	}

	values := url.Values{}
	values.Set("ticket", ticket.Code)
	values.Set("redirect", redirectTo)
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	c.Redirect(http.StatusFound, casdoorOAuthSuccessPath+"?"+values.Encode())
}

func (h *AuthHandler) CasdoorExchangeTicket(c *gin.Context) {
	var reqBody casdoorExchangeTicketRequest
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	browserKey, err := readCookieDecoded(c, casdoorOAuthBrowserCookieName)
	if err != nil || strings.TrimSpace(browserKey) == "" {
		response.ErrorFrom(c, infraerrors.Unauthorized("CASDOOR_BROWSER_SESSION_MISSING", "casdoor browser session is missing"))
		return
	}

	pendingSvc, err := h.pendingIdentityService()
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	session, err := pendingSvc.ConsumeCompletionCode(c.Request.Context(), reqBody.Ticket, browserKey)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if session.TargetUserID == nil || *session.TargetUserID <= 0 {
		response.ErrorFrom(c, infraerrors.Unauthorized("CASDOOR_TARGET_USER_MISSING", "casdoor ticket target user is missing"))
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), *session.TargetUserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if err := ensureLoginUserActive(user); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if err := h.ensureBackendModeAllowsUser(c.Request.Context(), user); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	h.authService.RecordSuccessfulLogin(c.Request.Context(), user.ID)
	clearCasdoorCookie(c, casdoorOAuthBrowserCookieName, isRequestHTTPS(c))
	tokenPair, err := h.authService.GenerateTokenPair(c.Request.Context(), user, "")
	if err != nil {
		response.ErrorFrom(c, infraerrors.InternalServer("CASDOOR_TOKEN_PAIR_FAILED", "failed to generate token pair").WithCause(err))
		return
	}

	response.Success(c, AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
		User:         dto.UserFromService(user),
	})
}

func (h *AuthHandler) getCasdoorOAuthConfig() (config.CasdoorConfig, error) {
	if h == nil || h.cfg == nil {
		return config.CasdoorConfig{}, infraerrors.ServiceUnavailable("CONFIG_NOT_READY", "config not loaded")
	}
	cfg := h.cfg.Casdoor
	if !cfg.Enabled {
		return config.CasdoorConfig{}, infraerrors.NotFound("CASDOOR_DISABLED", "casdoor login is disabled")
	}
	return cfg, nil
}

func buildCasdoorAuthorizeURL(cfg config.CasdoorConfig, state string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(cfg.AuthorizeURL))
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", strings.TrimSpace(cfg.ClientID))
	q.Set("redirect_uri", strings.TrimSpace(cfg.RedirectURI))
	if scope := normalizeCasdoorAuthorizeScope(cfg.Scope); scope != "" {
		q.Set("scope", scope)
	}
	q.Set("state", strings.TrimSpace(state))
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func normalizeCasdoorAuthorizeScope(scope string) string {
	fields := strings.Fields(scope)
	seen := make(map[string]bool, len(fields)+2)
	normalized := make([]string, 0, len(fields)+2)
	for _, field := range fields {
		if seen[field] {
			continue
		}
		seen[field] = true
		normalized = append(normalized, field)
	}
	for _, required := range []string{"openid", "phone"} {
		if !seen[required] {
			seen[required] = true
			normalized = append(normalized, required)
		}
	}
	return strings.Join(normalized, " ")
}

func casdoorExchangeCode(ctx context.Context, cfg config.CasdoorConfig, code string) (*casdoorTokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", strings.TrimSpace(cfg.ClientID))
	form.Set("client_secret", strings.TrimSpace(cfg.ClientSecret))
	form.Set("code", strings.TrimSpace(code))
	form.Set("redirect_uri", strings.TrimSpace(cfg.RedirectURI))

	resp, err := req.C().SetTimeout(30*time.Second).R().
		SetContext(ctx).
		SetHeader("Accept", "application/json").
		SetFormDataFromValues(form).
		Post(cfg.TokenURL)
	if err != nil {
		return nil, fmt.Errorf("request token: %w", err)
	}
	body := strings.TrimSpace(resp.String())
	if !resp.IsSuccessState() {
		providerErr, providerDesc := parseOAuthProviderError(body)
		return nil, fmt.Errorf("token exchange status=%d error=%s description=%s", resp.StatusCode, providerErr, providerDesc)
	}

	tokenResp, ok := parseCasdoorTokenResponse(body)
	if !ok || strings.TrimSpace(tokenResp.AccessToken) == "" {
		return nil, fmt.Errorf("token response missing access_token")
	}
	if strings.TrimSpace(tokenResp.TokenType) == "" {
		tokenResp.TokenType = "Bearer"
	}
	return tokenResp, nil
}

func parseCasdoorTokenResponse(body string) (*casdoorTokenResponse, bool) {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, false
	}

	accessToken := strings.TrimSpace(getGJSON(body, "access_token"))
	if accessToken != "" {
		return &casdoorTokenResponse{
			AccessToken:  accessToken,
			TokenType:    strings.TrimSpace(getGJSON(body, "token_type")),
			ExpiresIn:    gjson.Get(body, "expires_in").Int(),
			RefreshToken: strings.TrimSpace(getGJSON(body, "refresh_token")),
			Scope:        strings.TrimSpace(getGJSON(body, "scope")),
		}, true
	}

	values, err := url.ParseQuery(body)
	if err != nil {
		return nil, false
	}
	accessToken = strings.TrimSpace(values.Get("access_token"))
	if accessToken == "" {
		return nil, false
	}
	return &casdoorTokenResponse{
		AccessToken:  accessToken,
		TokenType:    strings.TrimSpace(values.Get("token_type")),
		RefreshToken: strings.TrimSpace(values.Get("refresh_token")),
		Scope:        strings.TrimSpace(values.Get("scope")),
	}, true
}

func casdoorFetchUserInfo(ctx context.Context, cfg config.CasdoorConfig, token *casdoorTokenResponse) (casdoorUserInfo, error) {
	authorization, err := buildBearerAuthorization(token.TokenType, token.AccessToken)
	if err != nil {
		return casdoorUserInfo{}, err
	}

	resp, err := req.C().SetTimeout(30*time.Second).R().
		SetContext(ctx).
		SetHeader("Accept", "application/json").
		SetHeader("Authorization", authorization).
		Get(cfg.UserInfoURL)
	if err != nil {
		return casdoorUserInfo{}, fmt.Errorf("request userinfo: %w", err)
	}
	if !resp.IsSuccessState() {
		return casdoorUserInfo{}, fmt.Errorf("userinfo status=%d", resp.StatusCode)
	}
	info := parseCasdoorUserInfo(resp.String())
	if strings.TrimSpace(info.Subject) == "" {
		return casdoorUserInfo{}, fmt.Errorf("userinfo missing sub")
	}
	return info, nil
}

func parseCasdoorUserInfo(body string) casdoorUserInfo {
	return casdoorUserInfo{
		Subject: firstNonEmpty(getGJSON(body, "sub"), getGJSON(body, "id"), getGJSON(body, "name")),
		Email:   strings.ToLower(firstNonEmpty(getGJSON(body, "email"), getGJSON(body, "data.email"))),
		Phone: firstNonEmpty(
			getGJSON(body, "phone"),
			getGJSON(body, "phone_number"),
			getGJSON(body, "phoneNumber"),
			getGJSON(body, "data.phone"),
			getGJSON(body, "data.phone_number"),
			getGJSON(body, "data.phoneNumber"),
			getGJSON(body, "mobile"),
			getGJSON(body, "data.mobile"),
		),
		Name: firstNonEmpty(getGJSON(body, "displayName"), getGJSON(body, "display_name"), getGJSON(body, "name")),
	}
}

func (h *AuthHandler) resolveCasdoorLoginSession(
	ctx context.Context,
	cfg config.CasdoorConfig,
	info casdoorUserInfo,
	redirectTo string,
	browserKey string,
	affiliateCode string,
) (*dbent.PendingAuthSession, error) {
	client := h.entClient()
	if client == nil {
		return nil, infraerrors.ServiceUnavailable("CASDOOR_NOT_READY", "casdoor auth is not ready")
	}
	subject := strings.TrimSpace(info.Subject)
	if subject == "" {
		return nil, infraerrors.BadRequest("CASDOOR_SUB_MISSING", "casdoor subject is missing")
	}

	identityKey := service.PendingAuthIdentityKey{
		ProviderType:    "oidc",
		ProviderKey:     strings.TrimSpace(cfg.Issuer),
		ProviderSubject: subject,
	}
	upstreamClaims := casdoorUpstreamClaims(cfg, info)

	userEntity, err := h.findOAuthIdentityUser(ctx, identityKey)
	if err != nil {
		return nil, err
	}
	createdUser := false
	if userEntity == nil {
		userEntity, err = resolveCasdoorUserByEmailPhone(ctx, client, info)
		if err != nil {
			return nil, err
		}
		if userEntity == nil {
			userEntity, err = h.createCasdoorUser(ctx, client, info)
			if err != nil {
				return nil, err
			}
			createdUser = true
		}
	}
	userEntity, err = syncCasdoorUserPhone(ctx, client, userEntity, info)
	if err != nil {
		return nil, err
	}
	if err := ensureCasdoorIdentityForUser(ctx, client, identityKey, userEntity.ID, upstreamClaims); err != nil {
		return nil, err
	}
	if createdUser && h.authService != nil {
		h.authService.BindOAuthAffiliate(ctx, userEntity.ID, affiliateCode)
	}

	pendingSvc, err := h.pendingIdentityService()
	if err != nil {
		return nil, err
	}
	return pendingSvc.CreatePendingSession(ctx, service.CreatePendingAuthSessionInput{
		Intent:                 oauthIntentLogin,
		Identity:               identityKey,
		TargetUserID:           &userEntity.ID,
		RedirectTo:             redirectTo,
		ResolvedEmail:          userEntity.Email,
		BrowserSessionKey:      browserKey,
		UpstreamIdentityClaims: upstreamClaims,
		ExpiresAt:              time.Now().UTC().Add(casdoorOAuthTicketTTL),
	})
}

func resolveCasdoorUserByEmailPhone(ctx context.Context, client *dbent.Client, info casdoorUserInfo) (*dbent.User, error) {
	var emailUser *dbent.User
	var phoneUser *dbent.User
	var err error

	if email := strings.ToLower(strings.TrimSpace(info.Email)); email != "" {
		emailUser, err = findUserByNormalizedEmail(ctx, client, email)
		if err != nil && !errors.Is(err, service.ErrUserNotFound) {
			return nil, err
		}
	}

	if phone := service.NormalizePhoneNumber(info.Phone, "86"); phone != "" {
		phoneUser, err = findUserByNormalizedPhone(ctx, client, phone)
		if err != nil && !errors.Is(err, service.ErrUserNotFound) {
			return nil, err
		}
	}

	if emailUser != nil && phoneUser != nil && emailUser.ID != phoneUser.ID {
		return nil, infraerrors.Conflict("CASDOOR_ACCOUNT_CONFLICT", "casdoor email and phone match different users")
	}
	if emailUser != nil {
		return emailUser, nil
	}
	return phoneUser, nil
}

func syncCasdoorUserPhone(ctx context.Context, client *dbent.Client, userEntity *dbent.User, info casdoorUserInfo) (*dbent.User, error) {
	if userEntity == nil {
		return nil, infraerrors.ServiceUnavailable("CASDOOR_TARGET_USER_MISSING", "casdoor target user is missing")
	}
	normalized := service.NormalizePhoneNumber(info.Phone, "86")
	if normalized == "" {
		return userEntity, nil
	}
	existing, err := findUserByNormalizedPhone(ctx, client, normalized)
	if err != nil && !errors.Is(err, service.ErrUserNotFound) {
		return nil, err
	}
	if existing != nil && existing.ID != userEntity.ID {
		return nil, infraerrors.Conflict("CASDOOR_ACCOUNT_CONFLICT", "casdoor phone matches another user")
	}
	if strings.TrimSpace(userEntity.PhoneNumber) == normalized {
		return userEntity, nil
	}
	updated, err := client.User.UpdateOneID(userEntity.ID).
		SetPhoneNumber(normalized).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func findUserByNormalizedPhone(ctx context.Context, client *dbent.Client, phone string) (*dbent.User, error) {
	normalized := service.NormalizePhoneNumber(phone, "86")
	if normalized == "" {
		return nil, service.ErrUserNotFound
	}
	matches, err := client.User.Query().
		Where(dbuser.PhoneNumberEQ(normalized)).
		Order(dbent.Asc(dbuser.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, service.ErrUserNotFound
	}
	if len(matches) > 1 {
		return nil, infraerrors.Conflict("USER_PHONE_CONFLICT", "phone number matched multiple users")
	}
	return matches[0], nil
}

func (h *AuthHandler) createCasdoorUser(ctx context.Context, client *dbent.Client, info casdoorUserInfo) (*dbent.User, error) {
	if h.authService == nil || !h.authService.IsRegistrationEnabled(ctx) {
		return nil, service.ErrRegDisabled
	}

	email := strings.ToLower(strings.TrimSpace(info.Email))
	if email == "" {
		email = casdoorSyntheticEmail(info.Subject)
	}
	randomPassword, err := casdoorRandomHex(32)
	if err != nil {
		return nil, infraerrors.InternalServer("CASDOOR_PASSWORD_GEN_FAILED", "failed to generate casdoor user password").WithCause(err)
	}
	passwordHash, err := h.authService.HashPassword(randomPassword)
	if err != nil {
		return nil, err
	}

	create := client.User.Create().
		SetEmail(email).
		SetPhoneNumber(service.NormalizePhoneNumber(info.Phone, "86")).
		SetUsername(casdoorTruncateUsername(info.Name)).
		SetPasswordHash(passwordHash).
		SetRole(service.RoleUser).
		SetStatus(service.StatusActive).
		SetSignupSource("oidc").
		SetConcurrency(casdoorDefaultUserConcurrency(h)).
		SetBalance(casdoorDefaultUserBalance(h)).
		SetRpmLimit(casdoorDefaultUserRPMLimit(ctx, h))

	userEntity, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}
	return userEntity, nil
}

func ensureCasdoorIdentityForUser(
	ctx context.Context,
	client *dbent.Client,
	identity service.PendingAuthIdentityKey,
	userID int64,
	metadata map[string]any,
) error {
	record, err := client.AuthIdentity.Query().
		Where(
			authidentity.ProviderTypeEQ(identity.ProviderType),
			authidentity.ProviderKeyEQ(identity.ProviderKey),
			authidentity.ProviderSubjectEQ(identity.ProviderSubject),
		).
		Only(ctx)
	if err != nil && !dbent.IsNotFound(err) {
		return err
	}
	if record != nil {
		if record.UserID != userID {
			activeOwner, err := findActiveUserByID(ctx, client, record.UserID)
			if err != nil {
				return err
			}
			if activeOwner != nil {
				return infraerrors.Conflict("AUTH_IDENTITY_OWNERSHIP_CONFLICT", "auth identity already belongs to another user")
			}
			_, err = client.AuthIdentity.UpdateOneID(record.ID).
				SetUserID(userID).
				SetIssuer(identity.ProviderKey).
				SetMetadata(mergeOAuthMetadata(record.Metadata, metadata)).
				Save(ctx)
			return err
		}
		_, err = client.AuthIdentity.UpdateOneID(record.ID).
			SetIssuer(identity.ProviderKey).
			SetMetadata(mergeOAuthMetadata(record.Metadata, metadata)).
			Save(ctx)
		return err
	}

	_, err = client.AuthIdentity.Create().
		SetUserID(userID).
		SetProviderType(identity.ProviderType).
		SetProviderKey(identity.ProviderKey).
		SetProviderSubject(identity.ProviderSubject).
		SetIssuer(identity.ProviderKey).
		SetMetadata(metadata).
		Save(ctx)
	return err
}

func casdoorUpstreamClaims(cfg config.CasdoorConfig, info casdoorUserInfo) map[string]any {
	return map[string]any{
		"issuer":   strings.TrimSpace(cfg.Issuer),
		"sub":      strings.TrimSpace(info.Subject),
		"email":    strings.TrimSpace(info.Email),
		"phone":    strings.TrimSpace(info.Phone),
		"name":     strings.TrimSpace(info.Name),
		"provider": "casdoor",
	}
}

func casdoorSyntheticEmail(subject string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(subject)))
	return "casdoor-" + hex.EncodeToString(sum[:12]) + "@oidc-connect.invalid"
}

func casdoorRandomHex(byteLen int) (string, error) {
	if byteLen <= 0 {
		byteLen = 16
	}
	buf := make([]byte, byteLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func casdoorTruncateUsername(value string) string {
	value = strings.TrimSpace(value)
	if len([]rune(value)) > 100 {
		return string([]rune(value)[:100])
	}
	return value
}

func casdoorDefaultUserConcurrency(h *AuthHandler) int {
	if h != nil && h.cfg != nil && h.cfg.Default.UserConcurrency > 0 {
		return h.cfg.Default.UserConcurrency
	}
	return 5
}

func casdoorDefaultUserBalance(h *AuthHandler) float64 {
	if h != nil && h.cfg != nil {
		return h.cfg.Default.UserBalance
	}
	return 0
}

func casdoorDefaultUserRPMLimit(ctx context.Context, h *AuthHandler) int {
	if h != nil && h.settingSvc != nil {
		return h.settingSvc.GetDefaultUserRPMLimit(ctx)
	}
	return 0
}

func casdoorErrorCode(err error) string {
	switch {
	case errors.Is(err, service.ErrRegDisabled):
		return "registration_disabled"
	case infraerrors.IsConflict(err):
		return "account_conflict"
	case infraerrors.IsForbidden(err):
		return "registration_disabled"
	case infraerrors.IsBadRequest(err):
		return "invalid_userinfo"
	default:
		return "server_error"
	}
}

func redirectCasdoorError(c *gin.Context, code string, message string, description string) {
	values := url.Values{}
	values.Set("error", truncateFragmentValue(code))
	if strings.TrimSpace(message) != "" {
		values.Set("error_message", truncateFragmentValue(message))
	}
	if strings.TrimSpace(description) != "" {
		values.Set("error_description", truncateFragmentValue(description))
	}
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	c.Redirect(http.StatusFound, casdoorOAuthErrorPath+"?"+values.Encode())
}

func setCasdoorCookie(c *gin.Context, name string, value string, maxAgeSec int, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     casdoorOAuthCookiePath,
		MaxAge:   maxAgeSec,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearCasdoorCookie(c *gin.Context, name string, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     casdoorOAuthCookiePath,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}
