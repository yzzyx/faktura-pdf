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

	"github.com/flosch/pongo2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/yzzyx/faktura-pdf/models"
	"github.com/yzzyx/zerr"
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

	dbURL := "postgres://faktura:faktura@localhost:2211/faktura?sslmode=disable"
	err := models.Setup(ctx, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer models.Shutdown()

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

	err = pongo2.ReplaceFilter("date", Date)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %+v\n", err)
		return
	}
	err = pongo2.RegisterFilter("money", Money)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %+v\n", err)
		return
	}

	TemplateSetup()

	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(currentDir, "static/img/favicon.ico"))
	})

	r.Get("/", ViewInvoiceList)
	r.Get("/invoice/{id}", ViewInvoice)
	r.Get("/invoice/{id}/offer", ViewInvoiceOffer)
	r.Get("/invoice/{id}/invoice", ViewInvoiceInvoice)
	r.Get("/invoice/{id}/flag", SetInvoiceFlag)
	r.Post("/invoice/{id}", SaveInvoice)

	serveAddress := ":3000"
	fmt.Printf("Listening on %s\n", serveAddress)
	err = http.ListenAndServe(serveAddress, r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %+v\n", err)
	}

	tlsConfig := &tls.Config{}

	enableTLS := false
	caCert := "" //viper.GetString("tls.ca_certificate")
	key := ""    //viper.GetString("tls.key")
	cert := ""   //viper.GetString("tls.certificate")
	protocol := "http"
	if enableTLS {
		protocol = "https"

		tlsConfig.ClientAuth = tls.VerifyClientCertIfGiven

		// Use user defined CA
		if caCert != "" {
			caCert, err := ioutil.ReadFile(caCert)
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, "Failed to load CA certificate:", err)
				os.Exit(1)
			}

			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)

			tlsConfig.ClientCAs = caCertPool
		}
	}

	bindAddr := fmt.Sprintf("%s:%d", "127.0.0.1", "3000")
	fmt.Printf("Starting webserver on %s://%s\n", protocol, bindAddr)

	server := http.Server{
		Addr:      bindAddr,
		TLSConfig: tlsConfig,
		Handler:   r,
	}

	errorCh := make(chan error)
	go func() {
		var err error

		if !enableTLS {
			// start server
			err = server.ListenAndServe()
		} else {
			// start TLS server
			err = server.ListenAndServeTLS(cert, key)
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
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	case <-ctx.Done():
	}
}
