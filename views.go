package main

import (
	"net/http"
	"path"

	"github.com/flosch/pongo2"
	"github.com/go-chi/chi/v5"
	"github.com/yzzyx/faktura-pdf/session"
	"github.com/yzzyx/faktura-pdf/tags/static"
	tagurl "github.com/yzzyx/faktura-pdf/tags/url"
	"github.com/yzzyx/faktura-pdf/views"
	"github.com/yzzyx/faktura-pdf/views/invoice"
	"github.com/yzzyx/faktura-pdf/views/login"
	"github.com/yzzyx/faktura-pdf/views/rut"
	"github.com/yzzyx/faktura-pdf/views/start"
)

func RegisterViews(baseURL string, r chi.Router) error {
	urlMap := map[string]string{
		"start":                "/",
		"login":                "/login",
		"invoice-list":         "/invoice",
		"invoice-view":         "/invoice/{id}",
		"invoice-set-flag":     "/invoice/{id}/flag",
		"invoice-view-offer":   "/invoice/{id}/offer",
		"invoice-view-invoice": "/invoice/{id}/invoice",
		"rut-list":             "/rut",
		"rut-view":             "/rut/{id}",
		"rut-export":           "/rut/{id}/export",
	}

	// Add base url to all routes
	for k := range urlMap {
		urlMap[k] = path.Join(baseURL, urlMap[k])
	}

	err := pongo2.ReplaceFilter("date", Date)
	if err != nil {
		return err
	}
	err = pongo2.RegisterFilter("money", Money)
	if err != nil {
		return err
	}
	err = pongo2.RegisterFilter("json", JSON)
	if err != nil {
		return err
	}

	err = static.RegisterTag(static.Config{URL: path.Join(baseURL, "/static"), Path: "static"})
	if err != nil {
		return err
	}

	err = tagurl.RegisterTag("", urlMap)
	if err != nil {
		return err
	}

	viewBuilder, err := views.NewBuilder(views.BuilderConfig{
		BaseURL:   "",
		PreRender: viewPreRender,
		//OnError:                viewErrorHandler,
		ErrorTemplate:          "error.html",
		MaxFileSizeUploadLimit: 0,
	})
	if err != nil {
		return err
	}

	ts, err := pongo2.NewLocalFileSystemLoader("templates")
	if err != nil {
		return err
	}

	viewBuilder.AddTemplateSet("base", pongo2.NewSet("base", ts))
	err = viewBuilder.RegisterRoutes(urlMap)
	if err != nil {
		return err
	}

	r.Route("/", func(r chi.Router) {
		//r.Use(TransactionMiddleware)
		r.Get("/", viewBuilder.Wrap(start.New()))
		r.HandleFunc("/login", viewBuilder.Wrap(login.New()))
		r.Get("/rut", viewBuilder.Wrap(rut.NewList()))
		r.HandleFunc("/rut/{id}", viewBuilder.Wrap(rut.NewView()))
		r.HandleFunc("/rut/{id}/export", viewBuilder.Wrap(rut.NewExport()))
		r.Get("/invoice", viewBuilder.Wrap(invoice.NewList()))
		r.HandleFunc("/invoice/{id}", viewBuilder.Wrap(invoice.NewView()))
		r.Get("/invoice/{id}/offer", viewBuilder.Wrap(invoice.NewOfferPDF()))
		r.Get("/invoice/{id}/invoice", viewBuilder.Wrap(invoice.NewInvoicePDF()))
		r.HandleFunc("/invoice/{id}/flag", viewBuilder.Wrap(invoice.NewFlag()))
	})

	return nil
}

func viewPreRender(v views.Viewer, r *http.Request) error {
	c, err := r.Cookie("_fp_login")
	if err == nil && c != nil {
		s, ok := session.Validate(c.Value)
		if ok {
			v.SetData("session", s)
			v.SetData("logged_in", true)
		}
	}

	return nil
}
