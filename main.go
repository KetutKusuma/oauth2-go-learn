package main

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func main() {
	godotenv.Load(".env")
	l := fmt.Sprintf("papa : %s, %s, %s", os.Getenv("CLIENTID"), os.Getenv("EXAMPLE_CLIENT_ID"), os.Getenv("EXAMPLE_CLIENT_SECRET"))
	fmt.Println("laa : ", l)

	oauthCfg := oauth2.Config{
		ClientID:     os.Getenv("EXAMPLE_CLIENT_ID"),
		ClientSecret: os.Getenv("EXAMPLE_CLIENT_SECRET"),
		Endpoint: oauth2.Endpoint{
			AuthURL:       "https://github.com/login/oauth/authorize",
			DeviceAuthURL: "https://github.com/login/device/code",
			TokenURL:      "https://github.com/login/oauth/access_token",
		},
		RedirectURL: "http://localhost:3000/oauth2/callback",
		Scopes:      []string{"user"},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	tmpl := template.Must(template.New("index.html").Funcs(
		template.FuncMap{
			"join": strings.Join,
		},
	).ParseFiles("index.html"))

	app := App{
		OAuthConfig: &oauthCfg,
		Logger:      logger,
		Template:    tmpl,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", LoggerMiddleware(logger, app.Root))
	mux.HandleFunc("GET /oauth2/callback", LoggerMiddleware(logger, app.OauthCallback))

	server := http.Server{
		Addr:    ":3000",
		Handler: mux,
	}

	logger.Info("start http", "address", server.Addr)

	if err := server.ListenAndServe(); err != nil {
		logger.Error("failed serving http", "error", err)
		os.Exit(1)
	}

}
