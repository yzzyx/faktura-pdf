package models

import (
	"context"

	"github.com/jackc/pgx/v4"
)

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

func (c *Company) ListUsers() ([]User, error) {

	return []User{}, nil
}

func (c *Company) AddUser(User) error {

	return nil
}

func (c *Company) RemoveUser(User) error {

	return nil
}

func CompanyGet(ctx context.Context, id int) (Company, error) {
	query := `
SELECT 
    id,
    name,
    email,
    address1,
    address2,
    postcode,
    city,
    telephone,
    company_id,
    payment_account,
    payment_type,
    vat_number,

    invoice_number,
    invoice_due_days,
    invoice_reference,
    invoice_text,
    invoice_template,

    offer_text,
    offer_template
WHERE id = $!
`

	var c Company
	err := dbpool.QueryRow(ctx, query, id).Scan(
		&c.ID,
		&c.Name,
		&c.Email,
		&c.Address1,
		&c.Address2,
		&c.Postcode,
		&c.City,
		&c.Telephone,
		&c.CompanyID,
		&c.PaymentAccount,
		&c.PaymentType,
		&c.VATNumber,
		&c.InvoiceNumber,
		&c.InvoiceDueDays,
		&c.InvoiceText,
		&c.InvoiceTemplate,
		&c.OfferText,
		&c.OfferTemplate)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c, nil
		}
	}
	return c, err
}

func CompanySave(ctx context.Context, c Company) (int, error) {
	if c.ID > 0 {
		_, err := dbpool.Exec(ctx, `UPDATE "company" SET 
    name = $2,
    email = $3,
    address1 = $4,
    address2 = $5,
    postcode = $6,
    city = $7,
    telephone = $8,
    company_id = $9,
    payment_account = $10,
    payment_type = $11,
    vat_number = $12,

    invoice_number = $13,
    invoice_due_days = $14,
    invoice_reference = $15,
    invoice_text = $16,
    invoice_template = $17,

    offer_text= $18,
    offer_template = $19
WHERE id = $1`, c.ID,
			c.Name,
			c.Email,
			c.Address1,
			c.Address2,
			c.Postcode,
			c.City,
			c.Telephone,
			c.CompanyID,
			c.PaymentAccount,
			c.PaymentType,
			c.VATNumber,
			c.InvoiceNumber,
			c.InvoiceDueDays,
			c.InvoiceReference,
			c.InvoiceText,
			c.InvoiceTemplate,
			c.OfferText,
			c.OfferTemplate)
		return c.ID, err
	}

	query := `INSERT INTO "company"
    (name,
    email,
    address1,
    address2,
    postcode,
    city,
    telephone,
    company_id,
    payment_account,
    payment_type,
    vat_number,

    invoice_number,
    invoice_due_days,
    invoice_reference,
    invoice_text,
    invoice_template,

    offer_text,
    offer_template)
VALUES
($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
RETURNING id`

	err := dbpool.QueryRow(ctx, query,
		c.Name,
		c.Email,
		c.Address1,
		c.Address2,
		c.Postcode,
		c.City,
		c.Telephone,
		c.CompanyID,
		c.PaymentAccount,
		c.PaymentType,
		c.VATNumber,
		c.InvoiceNumber,
		c.InvoiceDueDays,
		c.InvoiceReference,
		c.InvoiceText,
		c.InvoiceTemplate,
		c.OfferText,
		c.OfferTemplate).Scan(&c.ID)

	if err != nil {
		return 0, err
	}

	return c.ID, nil
}
