package views

// RegisterRoutes adds routes to a ViewBuilder, and registers the 'url'-tag with pongo2
// The routes-argument is a map from 'pattern' to Route
func (vs *ViewBuilder) RegisterRoutes(routes map[string]string) error {
	vs.routes = routes
	return nil
}
