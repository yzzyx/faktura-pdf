package main

import (
	"net/http"
	"path"

	"github.com/flosch/pongo2"
	"github.com/go-chi/chi/v5"
	"github.com/yzzyx/faktura-pdf/models"
	"github.com/yzzyx/faktura-pdf/session"
	"github.com/yzzyx/faktura-pdf/tags/static"
	tagurl "github.com/yzzyx/faktura-pdf/tags/url"
	"github.com/yzzyx/faktura-pdf/views"
	"github.com/yzzyx/faktura-pdf/views/company"
	"github.com/yzzyx/faktura-pdf/views/invoice"
	"github.com/yzzyx/faktura-pdf/views/login"
	"github.com/yzzyx/faktura-pdf/views/register"
	"github.com/yzzyx/faktura-pdf/views/rut"
	"github.com/yzzyx/faktura-pdf/views/start"
)

const (
	MethodALL = iota
	MethodGET
	MethodPOST
)

type routeInfo struct {
	URL          string
	Path         string
	View         views.Viewer
	Methods      int
	RequireLogin bool
}

var routes = []routeInfo{
	{URL: "start", Path: "/", View: start.New(), Methods: MethodGET, RequireLogin: false},
	{URL: "register", Path: "/register", View: register.New(), RequireLogin: false},
	{URL: "login", Path: "/login", View: login.New(), RequireLogin: false},
	{URL: "company-view", Path: "/company/{id}", View: company.NewView(), RequireLogin: true},
	{URL: "rut-list", Path: "/rut", View: rut.NewList(), Methods: MethodGET, RequireLogin: true},
	{URL: "rut-view", Path: "/rut/{id}", View: rut.NewView(), RequireLogin: true},
	{URL: "rut-export", Path: "/rut/{id}/export", View: rut.NewExport(), RequireLogin: true},
	{URL: "invoice-list", Path: "/invoice", View: invoice.NewList(), Methods: MethodGET, RequireLogin: true},
	{URL: "invoice-view", Path: "/invoice/{id}", View: invoice.NewView(), RequireLogin: true},
	{URL: "invoice-view-offer", Path: "/invoice/{id}/offer", View: invoice.NewOfferPDF(), Methods: MethodGET, RequireLogin: true},
	{URL: "invoice-view-invoice", Path: "/invoice/{id}/invoice", View: invoice.NewInvoicePDF(), Methods: MethodGET, RequireLogin: true},
	{URL: "invoice-set-flag", Path: "/invoice/{id}/flag", View: invoice.NewFlag(), RequireLogin: true},
}

func RegisterViews(baseURL string, r chi.Router) error {
	urlMap := map[string]string{}

	for _, r := range routes {
		urlMap[r.URL] = r.Path
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
		r.Use(models.TransactionMiddleware)

		for _, route := range routes {
			if route.Methods == MethodGET {
				r.Get(route.Path, viewBuilder.Wrap(route.View))
			} else if route.Methods == MethodPOST {
				r.Post(route.Path, viewBuilder.Wrap(route.View))
			} else if route.Methods == MethodALL {
				r.HandleFunc(route.Path, viewBuilder.Wrap(route.View))
			}
		}
	})

	return nil
}

func viewPreRender(v views.Viewer, r *http.Request, w http.ResponseWriter) error {
	hasSession := false

	c, err := r.Cookie("_fp_login")
	if err == nil && c != nil {
		s, ok := session.Validate(c.Value)
		if ok {
			v.SetSession(s)
			v.SetData("session", s)
			v.SetData("logged_in", true)
			hasSession = true
		} else {
			// Clear session cookie
			http.SetCookie(w, &http.Cookie{Name: "_fp_login", MaxAge: -1})
		}
	}

	// Make sure that user can access page
	for _, route := range routes {
		if route.URL == v.GetData("currentPage").(string) {
			if route.RequireLogin && !hasSession {
				u, err := v.URL("login")
				if err != nil {
					return err
				}

				q := u.Query()
				q.Add("r", v.GetData("currentURL").(string))
				u.RawQuery = q.Encode()
				http.Redirect(w, r, u.String(), http.StatusFound)
				return views.ErrViewRedirect
			}
			break
		}
	}

	return nil
}
