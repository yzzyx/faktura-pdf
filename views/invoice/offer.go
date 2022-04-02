package invoice

import (
	"fmt"
	"strings"
	"time"

	"github.com/yzzyx/faktura-pdf/models"
	"github.com/yzzyx/faktura-pdf/views"
)

// OfferPDF is the view-handler for creating and returning a PDF-offer
type OfferPDF struct {
	views.View
}

// NewOfferPDF creates a new handler for creating a PDF-offer
func NewOfferPDF() *OfferPDF {
	return &OfferPDF{}
}

// HandleGet renders a offer PDF
func (v *OfferPDF) HandleGet() error {
	id := v.URLParamInt("id")

	invoice, err := models.InvoiceGet(v.Ctx, id)
	if err != nil {
		return err
	}

	data, err := generatePDF(v.Ctx, invoice, "offer.tex")
	if err != nil {
		return err
	}

	now := time.Now()
	name := fmt.Sprintf("offert-%d-%s-%s.pdf", invoice.Number, invoice.Name, now.Format("2006-01-02"))
	name = strings.ReplaceAll(name, " ", "_")

	header := v.ResponseHeaders()
	header.Set("Content-Type", "application/pdf")
	header.Set("Content-Disposition", "attachment; filename="+name)
	return v.RenderBytes(data)
}
