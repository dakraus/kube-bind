/*
Copyright 2022 The Kube Bind Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package http

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	echo2 "github.com/labstack/echo/v4"

	"github.com/kube-bind/kube-bind/contrib/example-backend/kubernetes"
	"github.com/kube-bind/kube-bind/contrib/example-backend/kubernetes/resources"
	"github.com/kube-bind/kube-bind/pkg/apis/kubebind/v1alpha1"
)

type handler struct {
	oidc *oidcServiceProvider

	backendCallbackURL string
	providerPrettyName string

	client *http.Client

	kubeManager *kubernetes.Manager
}

func NewHandler(provider *oidcServiceProvider, backendCallbackURL, providerPrettyName string, mgr *kubernetes.Manager) (*handler, error) {
	return &handler{
		oidc:               provider,
		backendCallbackURL: backendCallbackURL,
		providerPrettyName: providerPrettyName,
		client:             http.DefaultClient,
		kubeManager:        mgr,
	}, nil
}

func (h *handler) handleServiceExport(c echo2.Context) error {
	serviceProvider := &v1alpha1.APIServiceProvider{
		Spec: v1alpha1.APIServiceProviderSpec{
			AuthenticatedClientURL: h.backendCallbackURL,
			ProviderPrettyName:     h.providerPrettyName,
		},
	}

	bs, err := json.Marshal(serviceProvider)
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	return c.Blob(http.StatusOK, "application/json", bs)
}

func (h *handler) handleAuthorize(c echo2.Context) error {
	scopes := []string{"openid", "profile", "email", "offline_access"}
	code := &resources.AuthCode{
		RedirectURL: c.QueryParam("redirect_url"),
		SessionID:   c.QueryParam("session_id"),
	}
	if code.RedirectURL == "" || code.SessionID == "" {
		http.Error(c.Response(), "missing redirect_url or session_id", http.StatusBadRequest)
		return nil
	}

	dataCode, err := json.Marshal(code)
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(dataCode)
	authURL := h.oidc.OIDCProviderConfig(scopes).AuthCodeURL(encoded)
	return c.Redirect(http.StatusSeeOther, authURL)
}

func parseJWT(p string) ([]byte, error) {
	parts := strings.Split(p, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("oidc: malformed jwt, expected 3 parts got %d", len(parts))
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("oidc: malformed jwt payload: %v", err)
	}
	return payload, nil
}

// handleCallback handle the authorization redirect callback from OAuth2 auth flow.
func (h *handler) handleCallback(c echo2.Context) error {
	if errMsg := c.FormValue("error"); errMsg != "" {
		http.Error(c.Response(), errMsg+": "+c.FormValue("error_description"), http.StatusBadRequest)
		return errors.New(errMsg)
	}
	code := c.FormValue("code")
	if code == "" {
		http.Error(c.Response(), fmt.Sprintf("no code in request: %q", c.Request().Form), http.StatusBadRequest)
		return nil
	}

	state := c.FormValue("state")
	decode, err := base64.StdEncoding.DecodeString(state)
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	authCode := &resources.AuthCode{}
	if err := json.Unmarshal(decode, authCode); err != nil {
		c.Logger().Error(err)
		return err
	}

	// TODO: sign state and verify that it is not faked by the oauth provider

	token, err := h.oidc.OIDCProviderConfig(nil).Exchange(c.Request().Context(), code)
	if err != nil {
		http.Error(c.Response(), fmt.Sprintf("failed to get token: %v", err), http.StatusInternalServerError)
		c.Logger().Error(err)
		return err
	}
	jwtStr, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(c.Response(), fmt.Sprintf("failed to get ID token: %v", err), http.StatusInternalServerError)
		err := fmt.Errorf("invalid id token: %v", token.Extra("id_token"))
		c.Logger().Error(err)
		return err
	}

	jwt, err := parseJWT(jwtStr)
	if !ok {
		http.Error(c.Response(), fmt.Sprintf("failed to parse JWT: %v", err), http.StatusInternalServerError)
		err := fmt.Errorf("invalid id token: %v", token.Extra("id_token"))
		c.Logger().Error(err)
		return err
	}

	var idToken struct {
		Subject string `json:"sub"`
	}

	if err := json.Unmarshal(jwt, &idToken); err != nil {
		http.Error(c.Response(), fmt.Sprintf("failed to parse ID token: %v", err), http.StatusInternalServerError)
		c.Logger().Error(fmt.Errorf("invalid id token: %v", jwt))
		return err
	}

	kfg, err := h.kubeManager.HandleResources(c.Request().Context(), idToken.Subject)
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	authResponse := resources.AuthResponse{
		SessionID:  authCode.SessionID,
		Kubeconfig: kfg,
	}

	payload, err := json.Marshal(authResponse)
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(payload)

	parsedAuthURL, err := url.Parse(authCode.RedirectURL)
	if err != nil {
		return fmt.Errorf("failed to parse auth url: %v", err)
	}

	values := parsedAuthURL.Query()
	values.Add("auth_response", encoded)

	parsedAuthURL.RawQuery = values.Encode()

	if err := c.Redirect(301, parsedAuthURL.String()); err != nil {
		c.Logger().Error(err)
		return err
	}

	return nil
}
