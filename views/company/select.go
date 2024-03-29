package company

import (
	"errors"

	"github.com/yzzyx/faktura-pdf/models"
	"github.com/yzzyx/faktura-pdf/views"
)

// Select is the view-handler for selecting a company
type Select struct {
	views.View
}

// NewSelect creates a new handler for selecting a company
func NewSelect() *Select {
	return &Select{}
}

// HandleGet selects a company
func (v *Select) HandleGet() error {
	var err error
	var company models.Company
	id := v.URLParamInt("id")
	company, err = models.CompanyGet(v.Ctx, models.CompanyFilter{ID: id, UserID: v.Session.User.ID})
	if err != nil {
		return err
	}

	if company.ID == 0 {
		return errors.New("Inget sådant företag existerar")
	}

	v.Session.Company = company
	_, err = models.SessionSave(v.Ctx, v.Session)
	if err != nil {
		return err
	}

	redirect := v.FormValueString("r")
	if redirect != "" {
		v.Redirect(redirect)
		return nil
	}
	return v.RedirectRoute("start")
}
