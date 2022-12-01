package models

import (
	"context"
	"strings"

	"github.com/yzzyx/zerr"
)

type File struct {
	ID        int
	Name      string
	CompanyID int
	MIMEType  string

	// Used for storage
	Backend  *string
	Contents []byte
}

type FileFilter struct {
	ID        int
	CompanyID int
	InvoiceID int

	IncludeContent bool
}

func (f File) IsImage() bool {
	return strings.HasPrefix(f.MIMEType, "image/")
}

func FileAdd(ctx context.Context, f File) (int, error) {
	tx := getContextTx(ctx)

	query := `
INSERT INTO file
(name, company_id, mimetype, backend, contents)
VALUES ($1, $2, $3, $4, $5)
RETURNING id`
	err := tx.QueryRow(ctx, query, f.Name, f.CompanyID, f.MIMEType, f.Backend, f.Contents).Scan(&f.ID)
	if err != nil {
		return 0, zerr.Wrap(err).
			WithString("query", query).
			WithString("name", f.Name).
			WithString("mimetype", f.MIMEType).
			WithInt("company_id", f.CompanyID).
			WithStringp("backend", f.Backend)
	}
	return f.ID, nil
}

func FileRemove(ctx context.Context, f File) error {
	tx := getContextTx(ctx)

	query := `DELETE FROM file
USING file f
LEFT OUTER JOIN invoice_attachments a ON a.file_id = f.id
WHERE file.id = $1 AND
file.id = f.id AND
a.file_id  IS NULL
`
	_, err := tx.Exec(ctx, query, f.ID)
	if err != nil {
		return zerr.Wrap(err).
			WithString("query", query).
			WithInt("id", f.ID)
	}
	return nil
}

// FileList returns a list of files
func FileList(ctx context.Context, f FileFilter) ([]File, error) {
	tx := getContextTx(ctx)

	filterStrings := []string{}
	joinStrings := []string{}

	var additionalCol string
	if f.IncludeContent {
		additionalCol = ", contents"
	}
	query := `SELECT
id,
name,
company_id,
mimetype,
backend` + additionalCol + `
FROM file
`

	if f.ID != 0 {
		filterStrings = append(filterStrings, "file.id = :id")
	}

	if f.CompanyID != 0 {
		filterStrings = append(filterStrings, "company_id = :company_id")
	}

	if f.InvoiceID != 0 {
		joinStrings = append(joinStrings, "INNER JOIN invoice_attachments ia ON ia.invoice_id = :invoice_id AND ia.file_id = file.id")
	}

	if len(joinStrings) > 0 {
		query += strings.Join(joinStrings, "\n")
	}

	if len(filterStrings) > 0 {
		query += "\nWHERE\n" + strings.Join(filterStrings, " AND ")
	}

	rows, err := tx.NamedQuery(ctx, query, f)
	if err != nil {
		return nil, zerr.Wrap(err).WithString("query", query).WithAny("filter", f)
	}

	var result []File
	for rows.Next() {
		var file File
		err = rows.StructScan(&file)
		if err != nil {
			return nil, zerr.Wrap(err).WithString("query", query).WithAny("filter", f)
		}

		result = append(result, file)
	}

	return result, nil
}
