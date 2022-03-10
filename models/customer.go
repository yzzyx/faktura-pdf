package models

import (
	"context"
	"fmt"
)

type Customer struct {
	ID       int
	Name     string
	Email    string
	Address1 string
	Address2 string
	Postcode string
	City     string
	PNR      string
}

func CustomerSave(ctx context.Context, customer Customer) (int, error) {

	if customer.ID > 0 {
		_, err := dbpool.Exec(ctx, `UPDATE customer SET 
name = $2,
email = $3,
address1 = $4,
address2 = $5,
postcode = $6,
city = $7,
pnr = $8
WHERE id = $1`, customer.ID,
			customer.Name,
			customer.Email,
			customer.Address1,
			customer.Address2,
			customer.Postcode,
			customer.City,
			customer.PNR)
		return customer.ID, err
	}

	query := `INSERT INTO customer 
(name, email, address1, address2, postcode, city)
VALUES
($1, $2, $3, $4, $5, $6, $7)
RETURNING id`

	err := dbpool.QueryRow(ctx, query,
		customer.Name,
		customer.Email,
		customer.Address1,
		customer.Address2,
		customer.Postcode,
		customer.City,
		customer.PNR).Scan(&customer.ID)
	if err != nil {
		fmt.Println("cannot create customer:", err)
		fmt.Println("query:", query)
		return 0, err
	}

	return customer.ID, err
}
