package document

import (
	"github.com/yzzyx/faktura-pdf/views"
)

// View is the view-handler for viewing a ROT/RUT request
type View struct {
	views.View
}

// NewView creates a new handler for viewing a ROT/RUT request
func NewView() *View {
	return &View{}
}

// HandleGet displays a ROT/RUT request
func (v *View) HandleGet() error {
	return v.Render("document/view.html")
}
