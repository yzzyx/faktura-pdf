package invoice

import (
	"encoding/json"
	"image/png"
	"io"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/shopspring/decimal"
)

type PaymentQRInfo struct {
	Version int    `json:"uqr"`
	Type    int    `json:"tp"`
	Name    string `json:"nme"`

	CompanyID        string          `json:"cid"`
	InvoiceReference string          `json:"iref"`
	InvoiceDate      string          `json:"idt,omitempty"`
	DueDate          string          `json:"ddt,omitempty"`
	DueAmount        decimal.Decimal `json:"due"`
	PaymentType      string          `json:"pt"` // one of IBAN, BBAN, BG, PG
	Account          string          `json:"acc"`

	// Not used for domestic invoices
	Currency string `json:"cur,omitempty"`

	// Only used for IBAN
	CountryCode string `json:"cc,omitempty"`
}

func GenerateQR(info PaymentQRInfo, w io.Writer) error {
	content, _ := json.Marshal(info)

	qrcode, err := qr.Encode(string(content), qr.M, qr.Unicode)
	if err != nil {
		return err
	}

	qrcode, err = barcode.Scale(qrcode, 256, 256)
	if err != nil {
		return err
	}

	err = png.Encode(w, qrcode)
	return err
}
