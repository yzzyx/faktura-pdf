package start

import (
	"github.com/yzzyx/faktura-pdf/views"
)

// Start is the view-handler for the start-page
type Start struct {
	views.View
}

// New creates a new handler for the start page
func New() *Start {
	return &Start{}
}

// HandleGet displays a list of requests
func (v *Start) HandleGet() error {
	return v.Render("start.html")
}
