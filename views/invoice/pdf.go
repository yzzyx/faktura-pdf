package invoice

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/yzzyx/faktura-pdf/models"
)

func generatePDF(ctx context.Context, invoice models.Invoice, templateFile string) (pdfFile []byte, err error) {
	rep := strings.NewReplacer(`\`, `\textbackslash{}`,
		`^`, `\textasciicircum{}`,
		`~`, `\textasciitilde{}`,
		`<`, `\textless{}`,
		`>`, `\textgreater{}`,
		`#`, `\#`,
		`$`, `\$`,
		`%`, `\%`,
		`&`, `\&`,
		`_`, `\_`,
		`{`, `\{`,
		`}`, `\}`)

	latexEscape := func(s string) string {
		if s == "" {
			return "~"
		}

		return rep.Replace(s)
	}

	d, err := ioutil.ReadFile(templateFile)
	if err != nil {
		return nil, err
	}

	template := string(d)

	invoicedate := time.Now()
	dueDate := time.Now().AddDate(0, 1, 0)
	if invoice.DateInvoiced != nil {
		invoicedate = *invoice.DateInvoiced
	}

	if invoice.DateDue != nil {
		dueDate = *invoice.DateDue
	}

	startRow := strings.Index(template, "<row>")
	endRow := strings.Index(template, "</row>")

	totalVAT := map[models.VATType]decimal.Decimal{}
	totalExcl := decimal.Decimal{}
	totalROTRUT := map[models.RUTType]decimal.Decimal{}

	if startRow > -1 && endRow > -1 {
		rowStr := template[startRow+5 : endRow]
		rowData := ""

		for _, row := range invoice.Rows {
			priceExcl := row.Cost.Div(decimal.NewFromInt(1).Add(row.VAT.Amount()))
			vatAmount := row.Cost.Sub(priceExcl)
			totalVAT[row.VAT] = totalVAT[row.VAT].Add(vatAmount)
			totalExcl = totalExcl.Add(priceExcl)

			priceInclRUT := row.Cost
			if row.IsRotRut && row.RotRutServiceType != nil {
				if row.RotRutServiceType.IsROT() {
					priceInclRUT = row.Cost.Mul(decimal.NewFromFloat(0.7))
					totalROTRUT[models.RUTTypeROT] = totalROTRUT[models.RUTTypeROT].Add(row.Cost.Mul(decimal.NewFromFloat(0.3)))
				} else {
					priceInclRUT = row.Cost.Mul(decimal.NewFromFloat(0.5))
					totalROTRUT[models.RUTTypeRUT] = totalROTRUT[models.RUTTypeRUT].Add(priceInclRUT)
				}
			}

			s := strings.ReplaceAll(rowStr, "<description>", latexEscape(row.Description))
			s = strings.ReplaceAll(s, "<price>", row.Cost.StringFixedBank(2))
			s = strings.ReplaceAll(s, "<priceExcl>", priceExcl.StringFixedBank(2))
			s = strings.ReplaceAll(s, "<priceInclRUT>", priceInclRUT.StringFixedBank(2))
			s = strings.ReplaceAll(s, "<count>", row.Count.Truncate(2).String())
			s = strings.ReplaceAll(s, "<unit>", latexEscape(row.Unit.String()))
			s = strings.ReplaceAll(s, "<vat>", latexEscape(row.VAT.String()))
			s = strings.ReplaceAll(s, "<vatAmount>", vatAmount.StringFixedBank(2))
			s = strings.ReplaceAll(s, "<rowtotal>", row.Total.StringFixedBank(2))

			rotRut := ""
			if row.IsRotRut {
				rotRut = "ja"
			}
			s = strings.ReplaceAll(s, "<isRotRut>", rotRut)
			rowData += s
		}
		template = template[0:startRow] + rowData + template[endRow+6:]
	}

	replaceMap := map[string]string{
		"customername":     invoice.Customer.Name,
		"customeremail":    invoice.Customer.Email,
		"customeraddress1": invoice.Customer.Address1,
		"customeraddress2": invoice.Customer.Address2,
		"customerpostcode": invoice.Customer.Postcode,
		"customercity":     invoice.Customer.City,
		"invoicedate":      invoicedate.Format("2006-01-02"),
		"duedate":          dueDate.Format("2006-01-02"),
		"invoicenumber":    fmt.Sprintf("%d", invoice.Number),
		"total":            invoice.TotalSum.StringFixedBank(2),
		"totalexcl":        totalExcl.StringFixedBank(2),
		"totalvat25":       totalVAT[models.VATType(0)].StringFixedBank(2),
		"totalvat12":       totalVAT[models.VATType(1)].StringFixedBank(2),
		"totalvat6":        totalVAT[models.VATType(2)].StringFixedBank(2),
		"totalrut":         totalROTRUT[models.RUTTypeRUT].StringFixedBank(2),
		"totalrot":         totalROTRUT[models.RUTTypeROT].StringFixedBank(2),
		"additionalinfo":   invoice.AdditionalInfo,
	}

	re := regexp.MustCompile("<([^>]*)>")
	matches := re.FindAllStringSubmatchIndex(template, -1)
	updatedTemplate := ""
	prevPos := 0
	for _, m := range matches {
		keyname := strings.ToLower(template[m[2]:m[3]])
		if v, ok := replaceMap[keyname]; ok {
			updatedTemplate += template[prevPos:m[0]] + latexEscape(v)
			prevPos = m[1]
		}
	}
	template = updatedTemplate + template[prevPos:]

	tmpdir, err := ioutil.TempDir("", "faktura-pdf-*")
	if err != nil {
		return nil, err
	}
	defer func() {
		err := os.RemoveAll(tmpdir)
		if err != nil {
			log.Printf("could not remove temp folder: %v", err)
		}
	}()

	filename := fmt.Sprintf("invoice-%s-%d", time.Now().Format("2006-01-02"), invoice.Number)
	cmd := exec.CommandContext(ctx, "pdflatex",
		"-jobname", filename,
		"-output-directory", tmpdir,
		"-no-shell-escape",
		"-8bit",
		"-file-line-error",
		"-interaction=batch")
	cmd.Stdin = bytes.NewBufferString(template)
	stdout := &bytes.Buffer{}
	cmd.Stdout = stdout
	err = cmd.Run()
	if err != nil {
		fmt.Println(template)
		fmt.Println(stdout.String())
		return nil, err
	}

	pdfFile, err = ioutil.ReadFile(filepath.Join(tmpdir, filename+".pdf"))
	if err != nil {
		return nil, err
	}

	return pdfFile, nil
}
