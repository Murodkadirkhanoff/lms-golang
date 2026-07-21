package infrastructure

import (
	"bytes"
	"fmt"

	"github.com/chashma/lms/internal/modules/enrollment/application"
	"github.com/chashma/lms/internal/modules/enrollment/domain"
	"github.com/go-pdf/fpdf"
)

// PDFRenderer renders certificates as A4-landscape PDFs (a double border with
// the student name, course title and issue details).
type PDFRenderer struct{}

// NewPDFRenderer builds a PDFRenderer.
func NewPDFRenderer() *PDFRenderer { return &PDFRenderer{} }

var _ application.CertificateRenderer = (*PDFRenderer)(nil)

// Render produces the certificate PDF bytes.
func (PDFRenderer) Render(studentName string, cert domain.Certificate) ([]byte, error) {
	pdf := fpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	w, h := pdf.GetPageSize()
	pdf.SetDrawColor(79, 70, 229)
	pdf.SetLineWidth(1.2)
	pdf.Rect(8, 8, w-16, h-16, "D")
	pdf.SetLineWidth(0.4)
	pdf.Rect(11, 11, w-22, h-22, "D")

	pdf.SetY(38)
	centered(pdf, "LearnHub", "B", 20, 79, 70, 229, 6)
	centered(pdf, "Certificate of Completion", "B", 30, 0, 0, 0, 14)
	centered(pdf, "This certificate is proudly presented to", "", 13, 100, 116, 139, 10)
	centered(pdf, studentName, "B", 26, 79, 70, 229, 12)
	centered(pdf, "for successfully completing the course", "", 13, 100, 116, 139, 10)
	centered(pdf, cert.CourseTitle, "B", 20, 0, 0, 0, 18)
	issued := cert.IssuedAt.UTC().Format("January 2, 2006")
	centered(pdf, fmt.Sprintf("Issued on %s    -    Certificate ID: LH-%06d", issued, cert.ID), "", 10, 100, 116, 139, 0)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func centered(pdf *fpdf.Fpdf, text, style string, size float64, r, g, b int, spacingAfter float64) {
	pdf.SetFont("Helvetica", style, size)
	pdf.SetTextColor(r, g, b)
	pdf.CellFormat(0, size*0.5+3, text, "", 1, "C", false, 0, "")
	if spacingAfter > 0 {
		pdf.Ln(spacingAfter)
	}
}
