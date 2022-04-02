package url

import (
	"path"
	"strings"

	"github.com/flosch/pongo2"
)

type urlTag struct {
	name      string
	nameEval  pongo2.IEvaluator
	arguments map[string]pongo2.IEvaluator

	baseURL string
	routes  map[string]string
}

// Execute joins the static path to another path
func (pt *urlTag) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	if pt.nameEval != nil {
		val, err := pt.nameEval.Evaluate(ctx)
		if err != nil {
			return err
		}
		pt.name = val.String()
	}

	pattern := pt.routes[pt.name]

	for argkey, evaluator := range pt.arguments {
		ival, err := evaluator.Evaluate(ctx)
		if err != nil {
			return err
		}
		str := ival.String()
		pattern = strings.Replace(pattern, "{"+argkey+"}", str, -1)
	}

	p := path.Join(pt.baseURL, pattern)
	if rewrite, ok := ctx.Public["_url_rewrite_base"].(string); ok {
		p = path.Join(rewrite, p)
	}

	// Check if user has access the page via a mapping.
	_, _ = writer.WriteString(p)
	return nil
}

func createTag(baseURL string, routes map[string]string) func(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	return func(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (n pongo2.INodeTag, err *pongo2.Error) {
		tag := &urlTag{baseURL: baseURL, routes: routes}

		if strToken := arguments.PeekType(pongo2.TokenString); strToken != nil {
			tag.name = strToken.Val
		} else if identifierToken := arguments.PeekType(pongo2.TokenIdentifier); identifierToken != nil {
			tag.nameEval, err = arguments.ParseExpression()
			if err != nil {
				return nil, err
			}
		}
		arguments.Consume()

		tag.arguments = make(map[string]pongo2.IEvaluator)

		for arguments.Remaining() > 0 {
			keyToken := arguments.MatchType(pongo2.TokenIdentifier)
			if keyToken == nil {
				return nil, arguments.Error("Expected an identifier", nil)
			}
			if arguments.Match(pongo2.TokenSymbol, "=") == nil {
				return nil, arguments.Error("Expected '='.", nil)
			}
			valueExpr, err := arguments.ParseExpression()
			if err != nil {
				return nil, err
			}
			tag.arguments[keyToken.Val] = valueExpr
		}

		return tag, nil
	}
}

// Register is used to register the 'url' tag
// Tag usage:
//  {% url <view-name> [urlparam=val] [urlparam=val] ... %}
// Where 'view-name' is the name of the view we want to get the URL to.
// This argument is required.
//
// Additional urlparams can be specified. Any matching urlparams in the URL of the view
// will then be replaced by 'val'.
//
// Example:
//  Given a view 'test' with URL '/my-first-test',
//  {% url 'test' %} returns '/my-first-test'
//
//  Given a view 'test2' with URL '/my-second-test/{id}',
//  {% url 'test2' id=123 %} returns '/my-second-test/123'
func RegisterTag(baseURL string, routes map[string]string) error {
	return pongo2.RegisterTag("url", createTag(baseURL, routes))
}
