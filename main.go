package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"unicode"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/yzzyx/faktura-pdf/config"
	"github.com/yzzyx/faktura-pdf/models"
	"github.com/yzzyx/faktura-pdf/sqlx"
	"github.com/yzzyx/zerr"
	"gopkg.in/yaml.v2"
)

func main() {
	ctx := context.Background()

	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	cfg := config.Config{}
	configFile, err := os.Open("config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot open config file: %+v\n", err)
		os.Exit(1)
	}

	err = yaml.NewDecoder(configFile).Decode(&cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot parse config file: %+v\n", err)
		os.Exit(1)
	}

	lg, err := setupLogger(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot setup logger: %+v\n", err)
		os.Exit(1)
	}

	err = models.Setup(ctx, cfg.Database.URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database URL %s: %v\n", cfg.Database.URL, err)
		os.Exit(1)
	}
	defer models.Shutdown()

	// Map from go CamelCase to sql snake_case
	sqlx.NameMapper = func(s string) string {
		result := ""
		prevUpper := true
		for _, v := range s {
			if unicode.IsUpper(v) {
				if !prevUpper {
					result += "_"
				}
				prevUpper = true
			} else {
				prevUpper = false
			}
			result += string(unicode.ToLower(v))
		}
		return result
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot get current directory: %v\n", err)
	}

	err = FileServer(r, "/static/", http.Dir("static"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %+v\n", err)
		return
	}

	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(currentDir, "static/img/favicon.ico"))
	})

	err = RegisterViews("", r, lg)
	if err != nil {
		zerr.Wrap(err).LogError(lg)
		os.Exit(1)
		return
	}

	tlsConfig := &tls.Config{}

	protocol := "http"
	if cfg.Server.EnableTLS {
		protocol = "https"

		tlsConfig.ClientAuth = tls.VerifyClientCertIfGiven

		// Use user defined CA
		if cfg.Server.CACertFile != "" {
			caCert, err := ioutil.ReadFile(cfg.Server.CACertFile)
			if err != nil {
				zerr.Wrap(err).LogError(lg)
				os.Exit(1)
			}

			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)

			tlsConfig.ClientCAs = caCertPool
		}
	}

	fmt.Printf("Starting webserver on %s://%s\n", protocol, cfg.Server.Address)

	server := http.Server{
		Addr:      cfg.Server.Address,
		TLSConfig: tlsConfig,
		Handler:   r,
	}

	errorCh := make(chan error)
	go func() {
		var err error

		if !cfg.Server.EnableTLS {
			// start server
			err = server.ListenAndServe()
		} else {
			// start TLS server
			err = server.ListenAndServeTLS(cfg.Server.CertFile, cfg.Server.KeyFile)
		}

		// ListenAndServer always returns a non-nil error -
		// either it's been shut down (via server.Shutdown()),
		// or an error occurred
		if err != http.ErrServerClosed {
			errorCh <- zerr.Wrap(err).WithString("address", server.Addr)
		}
	}()

	select {
	case err := <-errorCh:
		zerr.Wrap(err).LogError(lg)
	case <-ctx.Done():
	}
}
