package views

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/flosch/pongo2"
	"github.com/go-chi/chi/v5"
	"github.com/yzzyx/faktura-pdf/models"
)

// Errors used to show specific error pages
// e.g. if ErrForbidden is returned from a view,
// the page 'error/forbidden.html' might be shown
var (
	ErrInternalServerError = errors.New("internal server error")
	ErrBadRequest          = errors.New("error bad request")
	ErrForbidden           = errors.New("error forbidden")
	ErrNotFound            = errors.New("error not found")
)

// ViewContext contains fields that are passed to a view on execution.
type ViewContext struct {
	Request  *http.Request
	Response http.ResponseWriter
	Context  context.Context
	Builder  *ViewBuilder
}

// Viewer is an interface that needs to be fulfilled by a View.
// Usually, this is done by extending the 'View'-struct, and replacing
// Handlers with a new implementation.
//
// Example:
//   type MyView struct {
//      View
//      extraField string
//   }
//
//   func (v *MyView) HandleGet() error {
//       v.SetData("extra", v.extraField)
//       return v.Render("templates/myview.html");
//   }
//
//   myview := &MyView{extraField: "blah"}
//   chi.Get(viewbuilder.Wrap(myview))
type Viewer interface {
	HandleGet() error
	HandlePost() error
	HandlePut() error
	HandleDelete() error
	HandleMethod(method string) error
	HandleError(err error)

	GetCurrentURL() (*url.URL, error)
	SetData(key string, val interface{})
	GetData(key string) interface{}
	Data() pongo2.Context
	URL(viewName string, parameters ...string) (*url.URL, error)

	SetContext(vc ViewContext)
	SetSession(s models.Session)
}

// View defines a view
type View struct {
	w       http.ResponseWriter
	r       *http.Request
	data    pongo2.Context
	Ctx     context.Context
	Session models.Session

	builder *ViewBuilder
}

// HandleGet is the default view GET method
func (v *View) HandleGet() error {
	return ErrBadRequest
}

// HandlePost is the default view POST method
func (v *View) HandlePost() error {
	return ErrBadRequest
}

// HandlePut is the default view PUT method
func (v *View) HandlePut() error {
	return ErrBadRequest
}

// HandleDelete is the default view DELETE method
func (v *View) HandleDelete() error {
	return ErrBadRequest
}

// HandleMethod is the default view function for all methods except GET/POST/PUT/DELETE
func (v *View) HandleMethod(method string) error {
	return ErrBadRequest
}

// HandleError is called when an error occurs. The default is to render the error template with the 'error'-data set
func (v *View) HandleError(err error) {
	v.w.WriteHeader(http.StatusInternalServerError)
	if v.builder.errorTemplate != "" {
		v.SetData("error", err.Error())
		err = v.Render(v.builder.errorTemplate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not render error template: %s\n", err)
		}
	}
}

// SetData exposes data to the view
func (v *View) SetData(key string, val interface{}) {
	v.data[key] = val
}

// GetData returns a previously set data field
func (v *View) GetData(key string) interface{} {
	val := v.data[key]
	return val
}

// Data returns the current dataset
func (v *View) Data() pongo2.Context {
	return v.data
}

func (v *View) SetContext(vc ViewContext) {
	v.r = vc.Request
	v.w = vc.Response
	v.Ctx = vc.Context
	v.builder = vc.Builder

	v.data = make(pongo2.Context)
}

func (v *View) SetSession(s models.Session) {
	v.Session = s
}

// RenderBytes renders raw data to response
func (v *View) RenderBytes(data []byte) error {
	_, err := v.w.Write(data)
	return err
}

// Render renders a template
func (v *View) Render(templateName string) error {
	var tmp *pongo2.Template
	var err error

	// For each route-mapping we can potentially have a separate template set
	// This allows us to switch templates depending on what choices a person makes
	mapping, _ := v.GetData("_template_set").(string)
	templateSet, ok := v.builder.TemplateSets[mapping]
	if !ok {
		templateSet = v.builder.TemplateSets["base"]
	}

	tmp, err = templateSet.FromFile(templateName)
	if err != nil {
		return err
	}
	err = tmp.ExecuteWriter(v.Data(), v.w)
	if err != nil {
		return err
	}
	return nil
}

// ResponseHeaders returns the response headers for the view
func (v *View) ResponseHeaders() http.Header {
	return v.w.Header()
}

// RequestHeaders returns the request headers for the view
func (v *View) RequestHeaders() http.Header {
	return v.r.Header
}

// Redirect redirects to "url"
func (v *View) Redirect(url string) {
	if !strings.HasPrefix(url, "http") {
		url = path.Join(v.builder.BaseURL, url)
	}
	http.Redirect(v.w, v.r, url, http.StatusFound)
}

// RedirectRoute redirects to the route with the given name
func (v *View) RedirectRoute(viewName string, parameters ...string) error {
	u, err := v.URL(viewName, parameters...)
	if err != nil {
		return err
	}

	urlString := u.String()
	if !strings.HasPrefix(urlString, "http") {
		urlString = path.Join(v.builder.BaseURL, urlString)
	}
	http.Redirect(v.w, v.r, urlString, http.StatusFound)
	return nil
}

// FormFiles parses a multipart form with a certain key
func (v *View) FormFiles(key string) []*multipart.FileHeader {
	err := v.r.ParseMultipartForm(v.builder.maxFileSizeUploadLimit)
	if err != nil {
		return nil
	}

	return v.r.MultipartForm.File[key]
}

// FormValueStringSlice returns a form value list of strings
func (v *View) FormValueStringSlice(key string) []string {
	v.r.ParseForm()
	return v.r.Form[key]
}

// FormKeys returns a list of all keys in the form structure
func (v *View) FormKeys() []string {
	v.r.ParseForm()
	keys := make([]string, 0, len(v.r.Form))
	for key := range v.r.Form {
		keys = append(keys, key)
	}
	return keys
}

// FormValueString returns a form value string
func (v *View) FormValueExists(key string) bool {
	if len(v.r.Form) == 0 {
		v.r.ParseForm()
	}
	_, ok := v.r.Form[key]
	return ok
}

// FormValueString returns a form value string
func (v *View) FormValueString(key string) string {
	return v.r.FormValue(key)
}

// FormValueInt converts a form value to an int and returns it
func (v *View) FormValueInt(key string) int {
	value, _ := strconv.Atoi(v.r.FormValue(key))
	return value
}

// FormValueFloat converts a form value to an float and returns it
func (v *View) FormValueFloat(key string) float64 {
	value, _ := strconv.ParseFloat(v.r.FormValue(key), 64)
	return value
}

// FormValueBool converts a form value to a bool and returns it
func (v *View) FormValueBool(key string) bool {
	value, _ := strconv.ParseBool(v.r.FormValue(key))
	return value
}

// URLParamString returns a string url param
func (v *View) URLParamString(key string) string {
	return chi.URLParam(v.r, key)
}

// URLParamInt converts a url param to an int and returns it
func (v *View) URLParamInt(key string) int {
	value, _ := strconv.Atoi(chi.URLParam(v.r, key))
	return value
}

// SetCookie adds a cookie to the http response
func (v *View) SetCookie(cookie *http.Cookie) {
	http.SetCookie(v.w, cookie)
}

// GetCookie returns a cookie from the http request
func (v *View) GetCookie(name string) (*http.Cookie, error) {
	return v.r.Cookie(name)
}

// GetCookies returns all cookies from the http request
func (v *View) GetCookies() []*http.Cookie {
	return v.r.Cookies()
}

// URL returns the url for a given view
// The optional parameters can be used to fill in parameters in the url.
// This is done by passing in a multiple of two arguments, where the first
// of each pair is the key, and the second is the value.
//
// Example:
//  For a given mapping from view "request-view" to "/request/{id}",
//  URL("request-view", id, "123") returns "/request/123"
func (v *View) URL(viewName string, parameters ...string) (*url.URL, error) {

	pattern := v.builder.routes[viewName]
	if len(parameters)%2 != 0 {
		return nil, errors.New("view.URL: number of parameters must be divisible by 2")
	}

	for k := 0; k < len(parameters); k += 2 {
		pattern = strings.Replace(pattern, "{"+parameters[k]+"}", parameters[k+1], -1)
	}

	// Is DirectLink set to true for this route?
	pattern = strings.TrimPrefix(pattern, "::")

	mappedPath, _ := v.GetData("route_mapping_path").(string)
	if strings.HasPrefix(pattern, "/api") {
		mappedPath = ""
	}
	p := path.Join(v.builder.BaseURL, mappedPath, pattern)

	if rewrite, ok := v.GetData("_url_rewrite_base").(string); ok {
		p = path.Join(rewrite, p)
	}

	u, err := url.Parse(p)
	if err != nil {
		return nil, err
	}

	return u, nil
}

// GetCurrentURL returns the currently accessed URL
// Note that this function also considers the 'X-Forwarded-Host' HTTP header
// for servers behind a reverse proxy.
func (v *View) GetCurrentURL() (*url.URL, error) {
	requestURL := v.r.URL

	// If X-Forwarded-Host is set, use that, since we want the client to use the correct URL
	forwardedHost := v.r.Header.Get("X-Forwarded-Host")
	if forwardedHost != "" {
		forwardedURL, err := url.Parse(forwardedHost)
		if err != nil {
			return nil, err
		}
		requestURL.Scheme = forwardedURL.Scheme
		requestURL.Host = forwardedURL.Host
	}

	if !requestURL.IsAbs() {
		requestURL.Host = v.r.Host
		if v.r.TLS != nil {
			requestURL.Scheme = "https"
		} else {
			requestURL.Scheme = "http"
		}
	}

	return requestURL, nil
}
