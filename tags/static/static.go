package static

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/flosch/pongo2"
)

// Config defines the configuration parameters available for the static tag
type Config struct {
	// If set to an absolute URL (e.g. http://www.mytest.com/), all static links will be generated with this
	// as the base. E.g. the template {% static '/css/test.css' %} will become "http://www.mytest.com/css/test.css",
	// and no additional query parameters will be added to the URL.
	URL string

	// If set, static will check if the requested asset exists in this folder, and add a query parameter to the
	// generated URL with the last modification date, for cache-breaking purposes.
	Path string
}

type staticTag struct {
	argument pongo2.IEvaluator
	name     string
	config   Config
}

// Execute joins the static path to another path
func (pt *staticTag) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	var query string

	if pt.argument != nil {
		ival, err := pt.argument.Evaluate(ctx)
		if err != nil {
			return err
		}
		pt.name = ival.String()
	}

	staticPath := path.Join(pt.config.URL, pt.name)

	// Only check modified timestamp on local files
	if pt.config.Path != "" && (strings.HasPrefix(pt.config.URL, "/") || pt.config.URL == "") {
		var filename string

		// If the user is requesting the URL of a custom asset, we need to modify the filename to point to our custom asset folder
		filename = filepath.Join(pt.config.Path, pt.name)
		st, err := os.Stat(filename)
		if err == nil {
			query = fmt.Sprintf("?m=%d", st.ModTime().Unix())
		}

		if rewrite, ok := ctx.Public["_url_rewrite_base"].(string); ok {
			staticPath = path.Join(rewrite, staticPath)
		}
	}

	_, _ = writer.WriteString(staticPath + query)
	return nil
}

// tagStatic wraps a pongo2.TagParser with an absolute basepath to create a "static"-tag
func tagStatic(conf Config) pongo2.TagParser {
	return func(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
		tag := &staticTag{
			config: conf,
		}

		if nameToken := arguments.MatchType(pongo2.TokenString); nameToken != nil {
			tag.name = pongo2.AsValue(nameToken.Val).String()
		} else if arguments.Remaining() == 1 {
			valueExpr, err := arguments.ParseExpression()
			if err != nil {
				return nil, err
			}
			tag.argument = valueExpr
		}

		return tag, nil
	}
}

// RegisterTag registers the 'static'-tag in pango2
func RegisterTag(conf Config) error {
	return pongo2.RegisterTag("static", tagStatic(conf))
}
