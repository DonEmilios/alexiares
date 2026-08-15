// Package html extracts structural artifacts from already-collected
// HTML: forms, input fields, hidden fields, metadata, comments, and
// external resource references. It performs no network I/O — it is a
// pure parser over the bytes the collector already fetched.
package html

import (
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/alexiares/alexiares/internal/artifact"
)

// Extract parses rawHTML into artifact.HTMLArtifacts. Malformed HTML
// is tolerated (as golang.org/x/net/html does for any real browser
// target) rather than treated as an error — a phishing page's HTML is
// rarely well-formed.
func Extract(rawHTML string) artifact.HTMLArtifacts {
	out := artifact.HTMLArtifacts{
		Metadata: map[string]string{},
	}

	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return out
	}

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		switch n.Type {
		case html.ElementNode:
			switch n.DataAtom {
			case atom.Form:
				out.Forms = append(out.Forms, extractForm(n))
			case atom.Meta:
				if name, content := metaPair(n); name != "" {
					out.Metadata[name] = content
				}
			case atom.Title:
				out.Title = textContent(n)
			case atom.Script:
				if src := attr(n, "src"); src != "" {
					out.ExternalResources = append(out.ExternalResources, src)
				}
			case atom.Link:
				if href := attr(n, "href"); href != "" {
					out.ExternalResources = append(out.ExternalResources, href)
				}
			case atom.Img:
				if src := attr(n, "src"); src != "" {
					out.ExternalResources = append(out.ExternalResources, src)
				}
			case atom.Iframe:
				if src := attr(n, "src"); src != "" {
					out.ExternalResources = append(out.ExternalResources, src)
				}
			}
		case html.CommentNode:
			if c := strings.TrimSpace(n.Data); c != "" {
				out.Comments = append(out.Comments, c)
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)

	return out
}

// extractForm reads a <form>'s action, method, and field names,
// separating hidden inputs from visible ones.
func extractForm(form *html.Node) artifact.Form {
	f := artifact.Form{
		Action: attr(form, "action"),
		Method: strings.ToUpper(attr(form, "method")),
	}
	if f.Method == "" {
		f.Method = "GET"
	}

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.DataAtom {
			case atom.Input, atom.Select, atom.Textarea:
				name := attr(n, "name")
				if name == "" {
					name = attr(n, "id")
				}
				if name == "" {
					return
				}
				if strings.EqualFold(attr(n, "type"), "hidden") {
					f.HiddenFields = append(f.HiddenFields, name)
				} else {
					f.Fields = append(f.Fields, name)
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			// Do not descend into a nested <form>; it is walked separately.
			if child.Type == html.ElementNode && child.DataAtom == atom.Form {
				continue
			}
			walk(child)
		}
	}
	walk(form)

	return f
}

// metaPair reads a <meta> tag's name/content (or property/content, for
// Open Graph-style tags) as a key-value pair.
func metaPair(n *html.Node) (name, content string) {
	name = attr(n, "name")
	if name == "" {
		name = attr(n, "property")
	}
	return name, attr(n, "content")
}

func textContent(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(n)
	return strings.TrimSpace(b.String())
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return a.Val
		}
	}
	return ""
}
