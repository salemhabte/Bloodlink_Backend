package Usecase

import (
	"fmt"
	"path/filepath"

	"github.com/jung-kurt/gofpdf"
)

type IPDFGeneratorService interface {
	GenerateDraftContract(contractID, renderedText string) (string, error)
	GenerateFinalContract(contractID, renderedText, hospitalSigPath, adminSigPath string) (string, error)
}

type pdfGeneratorService struct {
	uploadsDir string
}

func NewPDFGeneratorService(uploadsDir string) IPDFGeneratorService {
	return &pdfGeneratorService{uploadsDir: uploadsDir}
}

func (s *pdfGeneratorService) GenerateDraftContract(contractID, renderedText string) (string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Blood Bank - Hospital Contract")

	pdf.Ln(20)
	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 10, renderedText, "", "L", false)

	pdf.Ln(40)
	pdf.SetFont("Arial", "I", 10)
	pdf.Cell(0, 10, "Awaiting Signatures...")

	fileName := fmt.Sprintf("contract_%s.pdf", contractID)
	fullPath := filepath.Join(s.uploadsDir, fileName)

	err := pdf.OutputFileAndClose(fullPath)
	if err != nil {
		return "", err
	}
	return fullPath, nil
}

func (s *pdfGeneratorService) GenerateFinalContract(contractID, renderedText, hospitalSigPath, adminSigPath string) (string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Blood Bank - Hospital Contract (FINAL)")

	pdf.Ln(20)
	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 10, renderedText, "", "L", false)

	pdf.Ln(40)
	pdf.Cell(90, 10, "Hospital Administrator Signature:")
	pdf.Cell(90, 10, "Blood Bank Administrator Signature:")

	pdf.Ln(10)
	if hospitalSigPath != "" {
		pdf.ImageOptions(hospitalSigPath, 10, pdf.GetY(), 50, 0, false, gofpdf.ImageOptions{}, 0, "")
	}
	if adminSigPath != "" {
		pdf.ImageOptions(adminSigPath, 100, pdf.GetY(), 50, 0, false, gofpdf.ImageOptions{}, 0, "")
	}

	fileName := fmt.Sprintf("contract_%s_final.pdf", contractID)
	fullPath := filepath.Join(s.uploadsDir, fileName)

	err := pdf.OutputFileAndClose(fullPath)
	if err != nil {
		return "", err
	}
	return fullPath, nil
}
