package invoice

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

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

	totals := models.InvoiceTotals{}

	if startRow > -1 && endRow > -1 {
		rowStr := template[startRow+5 : endRow]
		rowData := ""

		for _, row := range invoice.Rows {
			rowTotals := row.Totals(true, true)
			totals = totals.Add(rowTotals)

			s := strings.ReplaceAll(rowStr, "<description>", latexEscape(row.Description))
			s = strings.ReplaceAll(s, "<price>", rowTotals.PPU.StringFixedBank(2))
			s = strings.ReplaceAll(s, "<count>", row.Count.Truncate(2).String())
			s = strings.ReplaceAll(s, "<unit>", latexEscape(row.Unit.String()))
			s = strings.ReplaceAll(s, "<vat>", latexEscape(row.VAT.String()))
			s = strings.ReplaceAll(s, "<rowtotal>", rowTotals.Total.StringFixedBank(2))

			rotRut := ""
			if row.IsRotRut {
				rotRut = "ja"
			}
			s = strings.ReplaceAll(s, "<isRotRut>", rotRut)
			rowData += s
		}
		template = template[0:startRow] + rowData + template[endRow+6:]
	}

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

	qrImagePath := path.Join(tmpdir, "qrimage.png")
	qrImage, err := os.OpenFile(qrImagePath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}
	defer qrImage.Close()

	qrInfo := PaymentQRInfo{
		Version:          2,
		Type:             1,
		Name:             invoice.Company.Name,
		CompanyID:        invoice.Company.CompanyID,
		InvoiceReference: strconv.Itoa(invoice.Number),
		InvoiceDate:      invoicedate.Format("20060102"),
		DueDate:          dueDate.Format("20060102"),
		DueAmount:        totals.Incl.Sub(totals.ROTRUT),
		PaymentType:      invoice.Company.PaymentType.String(),
		Account:          invoice.Company.PaymentAccount,
	}

	err = GenerateQR(qrInfo, qrImage)
	if err != nil {
		return nil, err
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
		"total":            totals.Total.StringFixedBank(2),
		"totalexcl":        totals.Excl.StringFixedBank(2),
		"totalinclrut":     totals.Incl.Sub(totals.ROTRUT).StringFixedBank(2),
		"totalvat25":       totals.VAT25.StringFixedBank(2),
		"totalvat12":       totals.VAT12.StringFixedBank(2),
		"totalvat6":        totals.VAT6.StringFixedBank(2),
		"totalrut":         totals.ROTRUT.StringFixedBank(2),
		"totalrot":         totals.ROTRUT.StringFixedBank(2),
		"additionalinfo":   invoice.AdditionalInfo,
		"qrimage":          qrImagePath,

		"companyname":           invoice.Company.Name,
		"companyemail":          invoice.Company.Email,
		"companyaddress1":       invoice.Company.Address1,
		"companyaddress2":       invoice.Company.Address2,
		"companypostcode":       invoice.Company.Postcode,
		"companycity":           invoice.Company.City,
		"companytelephone":      invoice.Company.Telephone,
		"companyid":             invoice.Company.CompanyID,
		"companypaymentaccount": invoice.Company.PaymentAccount,
		"companypaymenttype":    invoice.Company.PaymentType.String(),
		"companyvatnumber":      invoice.Company.VATNumber,
		"companyreference":      invoice.Company.InvoiceReference,
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

	filename := fmt.Sprintf("invoice-%s-%d", time.Now().Format("2006-01-02"), invoice.Number)
	cmd := exec.CommandContext(ctx, "xelatex",
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
