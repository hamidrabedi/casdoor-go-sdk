// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package casdoorsdk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OAuthLoginResult includes OAuth tokens plus the IdP session id returned at login.
type OAuthLoginResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	SessionId    string `json:"session_id"`
}

// ExchangeOAuthCode exchanges an authorization code and returns tokens with the IdP session id.
func (c *Client) ExchangeOAuthCode(code string) (*OAuthLoginResult, error) {
	contentType, body, err := createForm(map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     c.ClientId,
		"client_secret": c.ClientSecret,
		"code":          code,
	})
	if err != nil {
		return nil, err
	}

	respBytes, err := c.DoPostBytesRaw(c.GetUrl("login/oauth/access_token", nil), contentType, body)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		OAuthLoginResult
		Error     string `json:"error"`
		ErrorDesc string `json:"error_description"`
	}
	if err := json.Unmarshal(respBytes, &wrapper); err != nil {
		return nil, err
	}
	if wrapper.Error != "" {
		if wrapper.ErrorDesc != "" {
			return nil, fmt.Errorf("%s: %s", wrapper.Error, wrapper.ErrorDesc)
		}
		return nil, fmt.Errorf("%s", wrapper.Error)
	}

	return &wrapper.OAuthLoginResult, nil
}

// DeleteSessionById removes one IdP session id for a user/application.
func (c *Client) DeleteSessionById(session *Session, sessionId string) (bool, error) {
	if sessionId == "" {
		return false, nil
	}

	queryMap := map[string]string{
		"id":        fmt.Sprintf("%s/%s", session.Owner, session.Name),
		"sessionId": sessionId,
	}

	session.Owner = c.OrganizationName
	postBytes, err := json.Marshal(session)
	if err != nil {
		return false, err
	}

	resp, err := c.DoPost("delete-session", queryMap, postBytes, false, false)
	if err != nil {
		return false, err
	}

	return resp.Data == "Affected", nil
}

// DeleteOAuthTokenByAccessToken removes the OAuth token record for this login.
// The access token is sent in the POST body so IdP does not auto-sign-in via query param.
func (c *Client) DeleteOAuthTokenByAccessToken(accessToken string) (bool, error) {
	if accessToken == "" {
		return false, nil
	}

	postBytes, err := json.Marshal(map[string]string{"accessToken": accessToken})
	if err != nil {
		return false, err
	}

	resp, err := c.DoPost("delete-token", nil, postBytes, false, false)
	if err != nil {
		return false, err
	}

	return resp.Data == "Affected", nil
}

// SsoLogout calls IdP /api/sso-logout to propagate logout to SSO apps (OIDC back-channel,
// notification providers, etc.). The access token authenticates the request via AutoSigninFilter.
func (c *Client) SsoLogout(accessToken string, logoutAll bool) error {
	if accessToken == "" {
		return fmt.Errorf("access token is required for SSO logout")
	}

	logoutParam := "false"
	if logoutAll {
		logoutParam = "true"
	}
	logoutURL := fmt.Sprintf("%s/api/sso-logout?logoutAll=%s", c.Endpoint, logoutParam)

	req, err := http.NewRequest(http.MethodPost, logoutURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", string(body))
	}

	var wrapper Response
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return err
	}
	if wrapper.Status != "ok" {
		if wrapper.Msg != "" {
			return fmt.Errorf("%s", wrapper.Msg)
		}
		return fmt.Errorf("SSO logout failed")
	}
	return nil
}

// LogoutUserCompletely deletes all IdP sessions (every application) and OAuth tokens for a user.
func (c *Client) LogoutUserCompletely(owner, username string) error {
	postBytes, err := json.Marshal(map[string]string{
		"owner": owner,
		"name":  username,
	})
	if err != nil {
		return err
	}

	_, err = c.DoPost("logout-user-completely", nil, postBytes, false, false)
	return err
}
