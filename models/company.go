package models

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/yzzyx/zerr"
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
	Homepage  string

	CompanyID      string
	PaymentAccount string
	PaymentType    PaymentType
	VATNumber      string `db:"vat_number"`

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

func (c *Company) AddUser(ctx context.Context, u User) error {
	if c.ID == 0 {
		return errors.New("cannot add user to company before company is created")
	}

	if u.ID == 0 {
		return errors.New("cannot add user to company before user is created")
	}
	tx := getContextTx(ctx)

	query := `INSERT INTO company_user (user_id, company_id) VALUES ($1, $2) ON CONFLICT ON CONSTRAINT company_user_unique DO NOTHING`
	_, err := tx.Exec(ctx, query, u.ID, c.ID)
	if err != nil {
		return zerr.Wrap(err).WithString("query", query).WithInt("company-id", c.ID).WithInt("user-id", u.ID)
	}
	return nil
}

func (c *Company) RemoveUser(ctx context.Context, u User) error {
	return errors.New("not implemented")
}

type CompanyFilter struct {
	ID     int
	UserID int
}

func CompanyList(ctx context.Context, filter CompanyFilter) ([]Company, error) {
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
	homepage,

    company.company_id,
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
FROM company
`
	filterstrings := []string{}
	joinstrings := []string{}

	if filter.ID > 0 {
		filterstrings = append(filterstrings, "id = :id")
	}

	if filter.UserID > 0 {
		joinstrings = append(joinstrings, "INNER JOIN company_user cu ON company.id = cu.company_id AND cu.user_id = :user_id")
	}

	if len(joinstrings) > 0 {
		query += strings.Join(joinstrings, "\n")
	}

	if len(filterstrings) > 0 {
		query += " WHERE " + strings.Join(filterstrings, " AND ")
	}

	tx := getContextTx(ctx)
	rows, err := tx.NamedQuery(ctx, query, filter)
	if err != nil {
		return nil, zerr.Wrap(err).WithString("query", query).WithAny("filter", filter)
	}
	defer rows.Close()

	var result []Company
	for rows.Next() {
		var c Company

		err = rows.StructScan(&c)

		if err != nil {
			return nil, zerr.Wrap(err).WithString("query", query).WithAny("filter", filter)
		}
		result = append(result, c)
	}

	return result, nil
}

func CompanyGet(ctx context.Context, f CompanyFilter) (Company, error) {
	var c Company
	lst, err := CompanyList(ctx, f)
	if err != nil {
		return Company{}, err
	}

	if len(lst) == 0 {
		return c, sql.ErrNoRows
	}

	if len(lst) > 1 {
		return c, zerr.Wrap(errTooManyRows).WithAny("filter", f)
	}
	c = lst[0]

	return c, err
}

func CompanySave(ctx context.Context, c Company) (int, error) {
	tx := getContextTx(ctx)

	if c.ID > 0 {
		query := `UPDATE "company" SET 
    name = :name,
    email = :email,
    address1 = :address1,
    address2 = :address2,
    postcode = :postcode,
    city = :city,
    telephone = :telephone,
	homepage = :homepage,
    company_id = :company_id,
    payment_account = :payment_account,
    payment_type = :payment_type,
    vat_number = :vat_number,

    invoice_number = :invoice_number,
    invoice_due_days = :invoice_due_days,
    invoice_reference = :invoice_reference,
    invoice_text = :invoice_text,
    invoice_template = :invoice_template,

    offer_text= :offer_text,
    offer_template = :offer_template 
WHERE id = :id`

		_, err := tx.NamedExec(ctx, query, c)
		if err != nil {
			return 0, zerr.Wrap(err).WithString("query", query).WithAny("company", c)
		}
		return c.ID, nil
	}

	query := `INSERT INTO "company"
    (name,
    email,
    address1,
    address2,
    postcode,
    city,
    telephone,
	homepage,

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
(:name,
:email,
:address1,
:address2,
:postcode,
:city,
:telephone,
:homepage,
:company_id,
:payment_account,
:payment_type,
:vat_number,
:invoice_number,
:invoice_due_days,
:invoice_reference,
:invoice_text,
:invoice_template,
:offer_text,
:offer_template)
RETURNING id`

	rows, err := tx.NamedQuery(ctx, query, c)
	if err != nil {
		return 0, zerr.Wrap(err).WithString("query", query).WithAny("company", c)
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&c.ID)
		if err != nil {
			return 0, zerr.Wrap(err).WithString("query", query).WithAny("company", c)
		}
	}
	return c.ID, nil
}
