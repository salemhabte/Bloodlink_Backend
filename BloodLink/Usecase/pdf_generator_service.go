package Usecase

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jung-kurt/gofpdf"
	_ "golang.org/x/image/webp"
)

type IPDFGeneratorService interface {
	GenerateDraftContract(contractID, renderedText string) (string, error)
	GenerateFinalContract(contractID, renderedText, hospitalSigPath, adminSigPath string) (string, error)
}

type pdfGeneratorService struct {
	uploadsDir string
}

func NewPDFGeneratorService(uploadsDir string) IPDFGeneratorService {
	os.MkdirAll(uploadsDir, 0755)
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
		localPath, imgType := s.ensureLocalFile(hospitalSigPath)
		if localPath != "" {
			opts := gofpdf.ImageOptions{ImageType: imgType}
			pdf.ImageOptions(localPath, 10, pdf.GetY(), 50, 0, false, opts, 0, "")
			if localPath != hospitalSigPath {
				defer os.Remove(localPath)
			}
		}
	}
	if adminSigPath != "" {
		localPath, imgType := s.ensureLocalFile(adminSigPath)
		if localPath != "" {
			opts := gofpdf.ImageOptions{ImageType: imgType}
			pdf.ImageOptions(localPath, 100, pdf.GetY(), 50, 0, false, opts, 0, "")
			if localPath != adminSigPath {
				defer os.Remove(localPath)
			}
		}
	}

	fileName := fmt.Sprintf("contract_%s_final.pdf", contractID)
	fullPath := filepath.Join(s.uploadsDir, fileName)

	err := pdf.OutputFileAndClose(fullPath)
	if err != nil {
		return "", err
	}
	return fullPath, nil
}

func (s *pdfGeneratorService) ensureLocalFile(pathOrURL string) (string, string) {
	var data []byte
	var err error

	if strings.HasPrefix(pathOrURL, "http") {
		resp, err := http.Get(pathOrURL)
		if err != nil || resp.StatusCode != http.StatusOK {
			return "", ""
		}
		defer resp.Body.Close()
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return "", ""
		}
	} else {
		data, err = os.ReadFile(pathOrURL)
		if err != nil {
			return "", ""
		}
	}

	// Decode the image (supports PNG, JPG, WEBP)
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", ""
	}

	// Always convert to JPEG for maximum compatibility with gofpdf
	ext := ".jpg"
	imgType := "JPG"

	tmpFile, err := os.CreateTemp("", "sig_*"+ext)
	if err != nil {
		return "", ""
	}
	defer tmpFile.Close()

	// Handle transparency: Draw the image onto a white background
	// This prevents transparent backgrounds from turning black when converted to JPEG
	bounds := img.Bounds()
	newImg := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(newImg, newImg.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	draw.Draw(newImg, newImg.Bounds(), img, bounds.Min, draw.Over)

	// Encode as JPEG with 95 quality
	err = jpeg.Encode(tmpFile, newImg, &jpeg.Options{Quality: 95})
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", ""
	}

	return tmpFile.Name(), imgType
}
