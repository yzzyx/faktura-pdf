package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/yzzyx/zerr"
)

type Customer struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Address1  string `json:"address1"`
	Address2  string `json:"address2"`
	Postcode  string `json:"postcode"`
	City      string `json:"city"`
	PNR       string `json:"pnr"`
	Telephone string `json:"telephone"`

	CompanyID int `json:"company_id"`
}

type CustomerFilter struct {
	ID        int
	CompanyID int
	Search    string // Search for customers that match this string (name)

	Limit     int
	Offset    int
	OrderBy   string
	Direction string
}

func CustomerList(ctx context.Context, filter CustomerFilter) ([]Customer, error) {
	query := `
SELECT 
    id,
    name,
    email,
    address1,
    address2,
    postcode,
    pnr,
    telephone
FROM customer
`
	filterstrings := []string{}

	if filter.CompanyID > 0 {
		filterstrings = append(filterstrings, "company_id = :company_id")
	}

	if filter.ID > 0 {
		filterstrings = append(filterstrings, "id = :id")
	}

	if filter.Search != "" {
		filterstrings = append(filterstrings, "name ILIKE '%'||:search||'%'")
	}

	if len(filterstrings) > 0 {
		query += " WHERE " + strings.Join(filterstrings, " AND ")
	}

	orderMap := map[string]string{
		"name":  "customer.name",
		"email": "customer.email",
	}

	orderBy, ok := orderMap[filter.OrderBy]
	if !ok {
		orderBy = "customer.name"
	}

	if strings.ToUpper(filter.Direction) != "DESC" {
		filter.Direction = "ASC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", orderBy, filter.Direction)

	if filter.Limit > 0 {
		query += "LIMIT :limit"
	}

	if filter.Offset > 0 {
		query += "OFFSET :limit"
	}

	tx := getContextTx(ctx)
	rows, err := tx.NamedQuery(ctx, query, filter)
	if err != nil {
		return nil, zerr.Wrap(err).WithString("query", query).WithAny("filter", filter)
	}
	defer rows.Close()

	var result []Customer
	for rows.Next() {
		var c Customer

		err = rows.StructScan(&c)

		if err != nil {
			return nil, zerr.Wrap(err).WithString("query", query).WithAny("filter", filter)
		}
		result = append(result, c)
	}

	return result, nil
}

func CustomerSave(ctx context.Context, customer Customer) (int, error) {
	tx := getContextTx(ctx)

	if customer.ID > 0 {
		query := `UPDATE customer SET 
name = $2,
email = $3,
address1 = $4,
address2 = $5,
postcode = $6,
city = $7,
pnr = $8,
telephone = $9
WHERE id = $1`
		_, err := tx.Exec(ctx, query, customer.ID,
			customer.Name,
			customer.Email,
			customer.Address1,
			customer.Address2,
			customer.Postcode,
			customer.City,
			customer.PNR,
			customer.Telephone)
		if err != nil {
			return 0, zerr.Wrap(err).WithString("query", query).WithAny("customer", customer)
		}
		return customer.ID, err
	}

	query := `INSERT INTO customer 
(name, email, address1, address2, postcode, city, pnr, telephone, company_id)
VALUES
($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id`

	err := tx.QueryRow(ctx, query,
		customer.Name,
		customer.Email,
		customer.Address1,
		customer.Address2,
		customer.Postcode,
		customer.City,
		customer.PNR,
		customer.Telephone,
		customer.CompanyID).Scan(&customer.ID)
	if err != nil {
		return 0, zerr.Wrap(err).WithString("query", query).WithAny("customer", customer)
	}

	return customer.ID, nil
}
