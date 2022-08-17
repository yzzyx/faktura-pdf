package start

import (
	"github.com/yzzyx/faktura-pdf/models"
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
	if v.Session.User.ID == 0 {
		return v.Render("start/start.html")
	}

	if v.Session.Company.ID == 0 {
		companyList, err := models.CompanyList(v.Ctx, models.CompanyFilter{UserID: v.Session.User.ID})
		if err != nil {
			return err
		}
		v.SetData("companyList", companyList)
	}
	return v.Render("start/logged-in.html")
}
