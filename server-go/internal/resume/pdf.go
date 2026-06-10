package resume

import (
	"bytes"
	"fmt"
	"io"

	"github.com/ledongthuc/pdf"
)

// extractPDFText pulls plain text from a PDF. Works for text-based PDFs only;
// scanned/image PDFs come back empty (they would need OCR).
func extractPDFText(r io.ReaderAt, size int64) (string, error) {
	reader, err := pdf.NewReader(r, size)
	if err != nil {
		return "", fmt.Errorf("open pdf: %w", err)
	}

	textReader, err := reader.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("extract pdf text: %w", err)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(textReader); err != nil {
		return "", fmt.Errorf("read pdf text: %w", err)
	}
	return buf.String(), nil
}
