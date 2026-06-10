package compiler

import (
	"archive/zip"
	"context"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// EPUBExporter generates EPUB 3 archives from campaign markdown.
type EPUBExporter struct{}

// NewEPUBExporter creates a new EPUB exporter.
func NewEPUBExporter() *EPUBExporter {
	return &EPUBExporter{}
}

// Format returns the exporter format identifier.
func (e *EPUBExporter) Format() string {
	return "epub"
}

// Export compiles the campaign directory into an EPUB archive.
func (e *EPUBExporter) Export(ctx context.Context, campaignDir, title string) (string, error) {
	if title == "" {
		title = filepath.Base(campaignDir)
	}

	// Generate HTML using existing compiler logic
	comp := New(campaignDir, "")
	htmlParts, err := comp.generateHTML(title)
	if err != nil {
		return "", fmt.Errorf("failed to generate HTML: %w", err)
	}
	fullHTML := strings.Join(htmlParts, "\n")

	// Split HTML into chapters at h2/h3 boundaries
	chapters := splitHTMLIntoChapters(fullHTML, title)

	// Write EPUB
	epubPath := filepath.Join(campaignDir, "campaign.epub")
	if err := e.writeEPUB(epubPath, title, chapters); err != nil {
		return "", fmt.Errorf("failed to write EPUB: %w", err)
	}

	return epubPath, nil
}

type epubChapter struct {
	ID      string
	Title   string
	Content string
}

var headingTagRe = regexp.MustCompile(`(?i)<h[23][^>]*>([^<]+)</h[23]>`)

func splitHTMLIntoChapters(fullHTML, title string) []epubChapter {
	var chapters []epubChapter

	// Find all h2/h3 splits
	matches := headingTagRe.FindAllStringIndex(fullHTML, -1)
	if len(matches) == 0 {
		// No headings found: treat entire HTML as one chapter
		return []epubChapter{{
			ID:      "chapter-1",
			Title:   title,
			Content: fullHTML,
		}}
	}

	// Extract cover/intro before first heading as chapter 1 if substantial
	firstMatch := matches[0]
	intro := strings.TrimSpace(fullHTML[:firstMatch[0]])
	if len(intro) > 200 {
		chapters = append(chapters, epubChapter{
			ID:      "chapter-1",
			Title:   title,
			Content: intro,
		})
	}

	for i, match := range matches {
		start := match[0]
		end := len(fullHTML)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}

		segment := fullHTML[start:end]
		titleText := extractHeadingText(segment)
		if titleText == "" {
			titleText = fmt.Sprintf("Chapter %d", i+1)
		}

		chapters = append(chapters, epubChapter{
			ID:      fmt.Sprintf("chapter-%d", len(chapters)+1),
			Title:   titleText,
			Content: segment,
		})
	}

	return chapters
}

func extractHeadingText(htmlSegment string) string {
	m := headingTagRe.FindStringSubmatch(htmlSegment)
	if len(m) > 1 {
		return html.UnescapeString(strings.TrimSpace(m[1]))
	}
	return ""
}

func (e *EPUBExporter) writeEPUB(path, title string, chapters []epubChapter) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	// mimetype must be first and uncompressed
	h, err := zw.CreateHeader(&zip.FileHeader{
		Name:   "mimetype",
		Method: zip.Store,
	})
	if err != nil {
		return err
	}
	if _, err := h.Write([]byte("application/epub+zip")); err != nil {
		return err
	}

	// META-INF/container.xml
	containerXML := `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`
	if err := writeZipEntry(zw, "META-INF/container.xml", containerXML); err != nil {
		return err
	}

	// Build OPF manifest and spine
	var manifestItems, spineItems, navPoints strings.Builder
	for _, ch := range chapters {
		filename := ch.ID + ".xhtml"
		manifestItems.WriteString(fmt.Sprintf(
			`    <item id="%s" href="%s" media-type="application/xhtml+xml"/>%s`,
			ch.ID, filename, "\n"))
		spineItems.WriteString(fmt.Sprintf(
			`    <itemref idref="%s"/>%s`,
			ch.ID, "\n"))
		navPoints.WriteString(fmt.Sprintf(
			`    <navPoint id="%s" playOrder="%d"><navLabel><text>%s</text></navLabel><content src="%s"/></navPoint>%s`,
			ch.ID, len(chapters), html.EscapeString(ch.Title), filename, "\n"))

		xhtml := wrapXHTML(ch.Title, ch.Content)
		if err := writeZipEntry(zw, "OEBPS/"+filename, xhtml); err != nil {
			return err
		}
	}

	// OEBPS/content.opf
	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	opf := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<package version="3.0" xmlns="http://www.idpf.org/2007/opf">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>%s</dc:title>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">%s</meta>
  </metadata>
  <manifest>
%s    <item id="toc" href="toc.ncx" media-type="application/x-dtbncx+xml"/>
  </manifest>
  <spine toc="toc">
%s  </spine>
</package>`,
		html.EscapeString(title), now, manifestItems.String(), spineItems.String())
	if err := writeZipEntry(zw, "OEBPS/content.opf", opf); err != nil {
		return err
	}

	// OEBPS/toc.ncx
	ncx := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ncx version="2005-1" xmlns="http://www.daisy.org/z3986/2005/ncx/">
  <head>
    <meta name="dtb:uid" content="grimorio-%d"/>
    <meta name="dtb:depth" content="1"/>
    <meta name="dtb:totalPageCount" content="0"/>
    <meta name="dtb:maxPageNumber" content="0"/>
  </head>
  <docTitle><text>%s</text></docTitle>
  <navMap>
%s  </navMap>
</ncx>`, time.Now().Unix(), html.EscapeString(title), navPoints.String())
	if err := writeZipEntry(zw, "OEBPS/toc.ncx", ncx); err != nil {
		return err
	}

	return zw.Close()
}

func wrapXHTML(title, body string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.1//EN" "http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
  <title>%s</title>
  <meta http-equiv="Content-Type" content="application/xhtml+xml; charset=utf-8"/>
  <style type="text/css">
    body { font-family: serif; margin: 1em; }
    h1, h2, h3 { page-break-after: avoid; }
    p { margin: 0.5em 0; }
  </style>
</head>
<body>
%s
</body>
</html>`, html.EscapeString(title), body)
}

func writeZipEntry(zw *zip.Writer, name, content string) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(content))
	return err
}
