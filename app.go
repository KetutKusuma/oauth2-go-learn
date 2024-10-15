package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

type UserInfo struct {
	ID    int    `json:"Id"`
	Login string `json:"login"`
}

type App struct {
	OAuthConfig *oauth2.Config
	Logger      *slog.Logger
	Template    *template.Template

	AccessToken  string
	RefreshToken string
	UserInfo     *UserInfo
}

func (a *App) Root(w http.ResponseWriter, r *http.Request) {
	if a.AccessToken == "" {
		w.WriteHeader(http.StatusOK)
		if err := a.Template.Execute(w, a); err != nil {
			a.Logger.Error("failed executing template", "err : ", err)
		}
		return
	}

	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		a.Logger.Error("failed creating request", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.AccessToken))
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		a.Logger.Error("failed retriving user details", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	var userInfo UserInfo

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		a.Logger.Error("failed decoding user details", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	a.UserInfo = &userInfo

	w.WriteHeader(http.StatusOK)
	if err := a.Template.Execute(w, a); err != nil {
		a.Logger.Error("failed to execting the template", "err : ", err)
	}
}

func (a *App) OauthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	token, err := a.OAuthConfig.Exchange(r.Context(), code)
	if err != nil {
		a.Logger.Error("failed oauth exchange", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	a.Logger.Info("completed oauth exchange",
		"token_type", token.Type(),
		"experiation", token.Expiry)

	a.AccessToken = token.AccessToken
	a.RefreshToken = token.RefreshToken

	fmt.Println("tokenn : ", token)
	fmt.Println("access token : ", a.AccessToken)
	fmt.Println("refresh token : ", a.RefreshToken)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)

}

func LoggerMiddleware(logger *slog.Logger, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()

		logger.Info(
			"request received",
			"method", r.Method,
			"path", r.URL.Path,
		)

		handler(w, r)

		logger.Info("request complete",
			"method", r.Method,
			"path", r.URL.Path,
			"duration_ms", time.Since(now).Milliseconds(),
		)

	}
}
