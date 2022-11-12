package sie

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"golang.org/x/text/encoding/charmap"
)

// esc escapes strings used in SIE file
func esc(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}

// Transaction describes a single transaction in a verification
type Transaction struct {
	KontoNr     int
	ObjektLista string
	Belopp      decimal.Decimal
	TransDat    time.Time
	TransText   string
	Kvantitet   string
	Sign        string
}

// Write writes the transaction to w
func (t *Transaction) Write(w io.Writer) error {
	if t.Belopp.IsZero() {
		return nil
	}

	// #TRANS kontonr {objektlista} belopp transdat transtext kvantitet sign
	_, err := fmt.Fprintf(w, "    #TRANS %d {%s} %s\n", t.KontoNr, t.ObjektLista, t.Belopp.StringFixedBank(2))
	return err
}

// Verification describes a single verification in a SIE file
type Verification struct {
	Serie    string
	VerNr    string
	VerDatum time.Time
	VerText  string
	RegDatum time.Time
	Sign     string

	Transactions []Transaction
}

// Write writes the verification to w
func (v *Verification) Write(w io.Writer) error {
	// #VER serie vernr verdatum vertext regdatum sign
	_, err := fmt.Fprintf(w,
		`#VER "%s" "%s" %s "%s"
{
`, esc(v.Serie),
		esc(v.VerNr),
		v.VerDatum.Format("20060102"),
		esc(v.VerText))

	if err != nil {
		return err
	}

	for _, t := range v.Transactions {
		err = t.Write(w)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(w, "}\n")
	return err
}

// SIE describes a SIE file
type SIE struct {
	Flag           int
	Fnamn          string
	Program        string
	ProgramVersion string
	Type           int

	Verifications []Verification
}

// Write writes the SIE to w, converted to CP437
func (s *SIE) Write(w io.Writer) error {
	w = charmap.CodePage437.NewEncoder().Writer(w)

	_, err := fmt.Fprintf(w,
		`#FLAGGA %d
#FNAMN "%s"
#FORMAT PC8
#GEN %s
#PROGRAM "%s" %s
#SIETYP 4

`, s.Flag,
		esc(s.Fnamn),
		time.Now().Format("20060102"),
		esc(s.Program),
		s.ProgramVersion)

	if err != nil {
		return err
	}

	for _, v := range s.Verifications {
		err = v.Write(w)
		if err != nil {
			return err
		}
	}

	return nil
}
