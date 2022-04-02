package rut

import (
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
	id := v.URLParamInt("id")

	rutRequest, err := models.RUTGet(v.Ctx, id)
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

	v.SetData("rut", rutRequest)
	v.SetData("maxAmount", maxAmount)
	v.SetData("filteredRows", filteredRows)

	return v.Render("rut/view.html")
}
