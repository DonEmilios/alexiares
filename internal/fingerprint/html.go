package fingerprint

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// structuralHash returns a SHA256 hash of the DOM's tag nesting shape:
// a bracket-nested serialization of element tag names in document
// order, ignoring text content, attributes, and comments. Two pages
// built from the same template produce the same hash even if their
// wallet address, copy, or images differ; changing the template's
// structure at all changes the hash.
func structuralHash(rawHTML string) string {
	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return ""
	}

	var b strings.Builder
	writeShape(&b, doc)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(b.String())))
}

func writeShape(b *strings.Builder, n *html.Node) {
	if n.Type == html.ElementNode {
		b.WriteByte('(')
		b.WriteString(n.Data)
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		writeShape(b, child)
	}
	if n.Type == html.ElementNode {
		b.WriteByte(')')
	}
}

// tagSequence returns the flat, preorder sequence of element tag
// names in rawHTML, used as the shingling input for the similarity
// fingerprint.
func tagSequence(rawHTML string) []string {
	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return nil
	}

	var out []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			out = append(out, n.Data)
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	return out
}
