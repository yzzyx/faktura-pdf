package models

type PaymentType int

const (
	PaymentTypeBG PaymentType = iota + 1
	PaymentTypePG
)

var paymentTypeStrings = map[PaymentType]string{
	PaymentTypeBG: "BG",
	PaymentTypePG: "PG",
}

func (t PaymentType) String() string {
	return paymentTypeStrings[t]
}

type Company struct {
	ID        int
	Name      string
	Email     string
	Address1  string
	Address2  string
	Postcode  string
	City      string
	Telephone string

	CompanyID      string
	PaymentAccount string
	PaymentType    PaymentType
	VATNumber      string

	InvoiceNumber    int
	InvoiceDueDays   int
	InvoiceReference string
	InvoiceText      string
	InvoiceTemplate  string
	OfferTemplate    string
	OfferText        string
}
