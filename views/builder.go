package views

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"runtime"
	"strings"

	"github.com/flosch/pongo2"
)

// ErrViewRedirect is used to signal that redirect was done, and no futher processing is necessary
var ErrViewRedirect = errors.New("view will redirect")

// ViewBuilder defines a view builder
type ViewBuilder struct {
	Basepath     string
	TemplateSets map[string]*pongo2.TemplateSet
	BaseURL      string

	maxFileSizeUploadLimit int64

	// Map view-names with patterns
	routes map[string]string

	// errorTemplate defines what template to use by default when an error occurs.
	errorTemplate string

	// onError is called for all encountered errors.
	onError func(err error)

	// preRender is called for all views before the controller is called
	preRender func(v Viewer, r *http.Request, w http.ResponseWriter) error
}

type BuilderConfig struct {
	// BaseURL is the base URL that all routes starts with.
	// Can either be a path or an absolute URL. Will be prepended to all URLs
	BaseURL string

	// PreRender defines a function that is called before each view
	// This allows for automatically including session information or
	// other additional fields that should always be available in every view
	PreRender func(v Viewer, r *http.Request, w http.ResponseWriter) error

	// ErrorTemplate defines what template to use by default when an error occurs.
	// This can be overriden by the 'HandleError'-method of 'Viewer'
	ErrorTemplate string

	// OnError is called for all encountered errors.
	// This can be used for logging of errors to any backend logger, etc.
	OnError func(err error)

	// MaxFilesizeUploadLimit
	MaxFileSizeUploadLimit int64
}

func NewBuilder(cfg BuilderConfig) (*ViewBuilder, error) {
	builder := &ViewBuilder{
		BaseURL:                cfg.BaseURL,
		preRender:              cfg.PreRender,
		errorTemplate:          cfg.ErrorTemplate,
		onError:                cfg.OnError,
		maxFileSizeUploadLimit: cfg.MaxFileSizeUploadLimit,
	}

	builder.TemplateSets = make(map[string]*pongo2.TemplateSet)

	return builder, nil
}

func (v *ViewBuilder) AddTemplateSet(path string, templateSet *pongo2.TemplateSet) {
	if path == "" {
		path = "base"
	}

	v.TemplateSets[path] = templateSet
}

func matchCurrentPage(routes map[string]string, baseURL string, path string) (string, error) {
	var currentPage string

	re := regexp.MustCompile("{[^}]+}")
	for name, urlPart := range routes {
		urlPart = re.ReplaceAllString(strings.TrimPrefix(urlPart, "/"), "[^/]+")
		// All admin routes are "DirectLinks" which are prefixed with "::/".
		// Remove the prefix so that the regex can match admin routes
		urlPart = strings.TrimPrefix(urlPart, "::")
		urlRe, err := regexp.Compile("^" + baseURL + "/?" + urlPart + "$")
		if err != nil {
			return "", err
		}

		if urlRe.MatchString(path) {
			// Is this an exact match? Return directly
			if path == baseURL+"/"+strings.TrimPrefix(urlPart, "/") {
				return name, nil
			}
			currentPage = name
		}
	}
	return currentPage, nil
}

// Wrap initializes a new view and calls the given template controller
func (v *ViewBuilder) Wrap(view Viewer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		copyObj := reflect.New(reflect.TypeOf(view).Elem())
		oldVal := reflect.ValueOf(view).Elem()
		newVal := copyObj.Elem()
		for i := 0; i < oldVal.NumField(); i++ {
			newValField := newVal.Field(i)
			if newValField.CanSet() {
				newValField.Set(oldVal.Field(i))
			}
		}
		clone := copyObj.Interface().(Viewer)

		clone.SetContext(ViewContext{
			Request:  r,
			Response: w,
			Context:  r.Context(),
			Builder:  v,
		})

		defer func() {
			// If a panic occurs, we want to log it, and show an error to the user
			if err := recover(); err != nil {
				var stackTrace string
				buf := make([]byte, 8192)
				nb := runtime.Stack(buf, false)
				stackTrace = string(buf[0:nb])

				// Combine panic message with stacktrace
				e := fmt.Errorf("%s\n%s", err, stackTrace)
				clone.HandleError(e)
				if v.onError != nil {
					v.onError(e)
				} else {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}
		}()

		currentURL, err := clone.GetCurrentURL()
		if err != nil {
			_, _ = fmt.Fprintf(w, "cannot get current url: %s", err)
		}

		// Set current page
		currentPage, err := matchCurrentPage(v.routes, v.BaseURL, currentURL.Path)
		if err != nil {
			return
		}

		clone.SetData("currentPage", currentPage)
		clone.SetData("currentURL", currentURL.String())

		if v.preRender != nil {
			err = v.preRender(clone, r, w)
			if err != nil {
				if errors.Is(err, ErrViewRedirect) {
					return
				}

				clone.HandleError(err)
				if v.onError != nil {
					v.onError(err)
				}
				return
			}
		}

		// Only call controller if we don't have any errors
		if err == nil {
			switch r.Method {
			case http.MethodGet:
				err = clone.HandleGet()
			case http.MethodPost:
				err = clone.HandlePost()
			case http.MethodPut:
				err = clone.HandlePut()
			case http.MethodDelete:
				err = clone.HandleDelete()
			default:
				err = clone.HandleMethod(r.Method)
			}
			if err == nil {
				return
			}
		}

		clone.HandleError(err)
		if v.onError != nil {
			v.onError(err)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
	}
}
