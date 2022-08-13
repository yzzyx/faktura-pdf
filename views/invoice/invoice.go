package invoice

import (
	"fmt"
	"strings"
	"time"

	"github.com/yzzyx/faktura-pdf/models"
	"github.com/yzzyx/faktura-pdf/views"
)

// InvoicePDF is the view-handler for creating and returning a PDF-invoice
type InvoicePDF struct {
	views.View
}

// NewInvoicePDF creates a new handler for creating a PDF-invoice
func NewInvoicePDF() *InvoicePDF {
	return &InvoicePDF{}
}

// HandleGet renders a PDF invoice
func (v *InvoicePDF) HandleGet() error {
	f := models.InvoiceFilter{
		ID:             v.URLParamInt("id"),
		CompanyID:      v.Session.Company.ID,
		IncludeCompany: true,
	}

	if f.ID <= 0 {
		return views.ErrBadRequest
	}

	invoice, err := models.InvoiceGet(v.Ctx, f)
	if err != nil {
		return err
	}

	data, err := generatePDF(v.Ctx, invoice, "invoice.tex")
	if err != nil {
		return err
	}

	now := time.Now()
	name := fmt.Sprintf("faktura-%d-%s-%s.pdf", invoice.Number, invoice.Name, now.Format("2006-01-02"))
	name = strings.ReplaceAll(name, " ", "_")

	headers := v.ResponseHeaders()
	headers.Set("Content-Type", "application/pdf")
	headers.Set("Content-Disposition", "attachment; filename="+name)
	return v.RenderBytes(data)
}
