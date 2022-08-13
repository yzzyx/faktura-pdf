package rut

import (
	"fmt"
	"strconv"

	"github.com/shopspring/decimal"
	"github.com/yzzyx/faktura-pdf/models"
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
	f := models.RUTFilter{
		ID:        v.URLParamInt("id"),
		CompanyID: v.Session.Company.ID,
	}

	if f.ID <= 0 {
		return views.ErrBadRequest
	}

	rutRequest, err := models.RUTGet(v.Ctx, f)
	if err != nil {
		return err
	}

	multiplier := decimal.NewFromFloat(0.5)
	if rutRequest.Type == models.RUTTypeROT {
		multiplier = decimal.NewFromFloat(0.3)
	}

	maxAmount := decimal.NewFromInt(0)
	filteredRows := []models.InvoiceRow{}
	for _, r := range rutRequest.Invoice.Rows {
		if !r.IsRotRut || r.RotRutServiceType == nil {
			continue
		}

		if (rutRequest.Type == models.RUTTypeRUT && r.RotRutServiceType.IsRUT()) ||
			(rutRequest.Type == models.RUTTypeROT && r.RotRutServiceType.IsROT()) {
			maxAmount = maxAmount.Add(r.Total.Mul(multiplier))
			filteredRows = append(filteredRows, r)
		}
	}

	canExport := len(filteredRows) > 0 && rutRequest.Invoice.Customer.PNR != "" && rutRequest.RequestedSum != nil && *rutRequest.RequestedSum != 0

	v.SetData("rut", rutRequest)
	v.SetData("maxAmount", maxAmount)
	v.SetData("filteredRows", filteredRows)
	v.SetData("canExport", canExport)
	v.SetData("hasRequestedSum", rutRequest.RequestedSum != nil)

	return v.Render("rut/view.html")
}

// HandlePost updates a ROT/RUT request
func (v *View) HandlePost() error {
	f := models.RUTFilter{
		ID:        v.URLParamInt("id"),
		CompanyID: v.Session.Company.ID,
	}

	if f.ID <= 0 {
		return views.ErrBadRequest
	}

	rutRequest, err := models.RUTGet(v.Ctx, f)
	if err != nil {
		return err
	}

	sum := v.FormValueInt("request-sum")
	rutRequest.RequestedSum = &sum

	for _, formKey := range v.FormKeys() {
		var rowID int
		n, err := fmt.Sscanf(formKey, "hours[%d]", &rowID)
		if n != 1 || err != nil {
			continue
		}

		if v.FormValueString(formKey) == "" {
			continue
		}

		hours := v.FormValueInt(formKey)
		for k := range rutRequest.Invoice.Rows {
			if rutRequest.Invoice.Rows[k].ID != rowID {
				continue
			}
			rutRequest.Invoice.Rows[k].RotRutHours = &hours
		}
	}

	_, err = models.RUTSave(v.Ctx, rutRequest)
	if err != nil {
		return err
	}

	return v.RedirectRoute("rut-view", "id", strconv.Itoa(rutRequest.ID))
}
