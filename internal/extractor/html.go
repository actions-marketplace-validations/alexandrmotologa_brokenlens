package extractor

import (
	"bytes"
	"io"
	"net/url"
	"strings"

	"github.com/alexandrmotologa/brokenlens/internal/model"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// DiscoveredTarget represents an extracted reference inside an HTML document.
type DiscoveredTarget struct {
	RawURL    string
	AssetType model.AssetType
	Text      string
	Fragment  string
	Snippet   string
}

// ExtractionResult holds all links, assets, and DOM fragments extracted from a page.
type ExtractionResult struct {
	Targets      []DiscoveredTarget
	DOMFragments map[string]bool
	PageTitle    string
	CanonicalURL string
	RobotsMeta   string
}

// ExtractHTML parses an HTML document stream and extracts all references, snippets, and DOM IDs.
func ExtractHTML(r io.Reader, baseURL string) (*ExtractionResult, error) {
	tokenizer := html.NewTokenizer(r)

	result := &ExtractionResult{
		DOMFragments: make(map[string]bool),
	}

	var inTitle bool
	var titleBuf strings.Builder
	var inStyle bool
	var styleBuf strings.Builder

	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				break
			}
			return nil, err
		}

		tok := tokenizer.Token()
		rawTagSnippet := tok.String()

		switch tt {
		case html.StartTagToken, html.SelfClosingTagToken:
			// Check for DOM IDs on any element
			for _, attr := range tok.Attr {
				if attr.Key == "id" && attr.Val != "" {
					result.DOMFragments[attr.Val] = true
				}
				// Also extract inline CSS urls: style="background: url(...)"
				if attr.Key == "style" && attr.Val != "" {
					cssURLs := ExtractCSSURLs(attr.Val)
					for _, cu := range cssURLs {
						result.Targets = append(result.Targets, DiscoveredTarget{
							RawURL:    cu,
							AssetType: model.AssetCSS,
							Text:      "inline style",
							Snippet:   rawTagSnippet,
						})
					}
				}
			}

			// Tag-specific inspections
			switch tok.DataAtom {
			case atom.Title:
				inTitle = true
				titleBuf.Reset()

			case atom.Style:
				inStyle = true
				styleBuf.Reset()

			case atom.Meta:
				var name, property, content string
				for _, attr := range tok.Attr {
					switch strings.ToLower(attr.Key) {
					case "name":
						name = strings.ToLower(strings.TrimSpace(attr.Val))
					case "property":
						property = strings.ToLower(strings.TrimSpace(attr.Val))
					case "content":
						content = strings.TrimSpace(attr.Val)
					}
				}

				// Meta Robots
				if (name == "robots" || name == "googlebot") && content != "" {
					result.RobotsMeta = content
				}

				// OpenGraph Image
				if (property == "og:image" || property == "og:image:url") && content != "" {
					result.Targets = append(result.Targets, DiscoveredTarget{
						RawURL:    content,
						AssetType: model.AssetSocialImage,
						Text:      "og:image",
						Snippet:   rawTagSnippet,
					})
				}

				// Twitter Card Image
				if (name == "twitter:image" || property == "twitter:image") && content != "" {
					result.Targets = append(result.Targets, DiscoveredTarget{
						RawURL:    content,
						AssetType: model.AssetSocialImage,
						Text:      "twitter:image",
						Snippet:   rawTagSnippet,
					})
				}

			case atom.A, atom.Area:
				var href, text, nameAttr string
				for _, attr := range tok.Attr {
					switch attr.Key {
					case "href":
						href = strings.TrimSpace(attr.Val)
					case "name":
						nameAttr = strings.TrimSpace(attr.Val)
					case "alt":
						text = strings.TrimSpace(attr.Val)
					}
				}
				if nameAttr != "" {
					result.DOMFragments[nameAttr] = true
				}
				if href != "" {
					frag := extractFragment(href)
					result.Targets = append(result.Targets, DiscoveredTarget{
						RawURL:    href,
						AssetType: model.AssetLink,
						Text:      text,
						Fragment:  frag,
						Snippet:   rawTagSnippet,
					})
				}

			case atom.Img:
				var src, srcset, alt string
				for _, attr := range tok.Attr {
					switch attr.Key {
					case "src":
						src = strings.TrimSpace(attr.Val)
					case "srcset":
						srcset = strings.TrimSpace(attr.Val)
					case "alt":
						alt = strings.TrimSpace(attr.Val)
					}
				}
				if src != "" {
					result.Targets = append(result.Targets, DiscoveredTarget{
						RawURL:    src,
						AssetType: model.AssetImage,
						Text:      alt,
						Snippet:   rawTagSnippet,
					})
				}
				if srcset != "" {
					for _, su := range parseSrcset(srcset) {
						result.Targets = append(result.Targets, DiscoveredTarget{
							RawURL:    su,
							AssetType: model.AssetImage,
							Text:      alt,
							Snippet:   rawTagSnippet,
						})
					}
				}

			case atom.Source:
				var src, srcset string
				for _, attr := range tok.Attr {
					switch attr.Key {
					case "src":
						src = strings.TrimSpace(attr.Val)
					case "srcset":
						srcset = strings.TrimSpace(attr.Val)
					}
				}
				if src != "" {
					result.Targets = append(result.Targets, DiscoveredTarget{
						RawURL:    src,
						AssetType: model.AssetImage,
						Snippet:   rawTagSnippet,
					})
				}
				if srcset != "" {
					for _, su := range parseSrcset(srcset) {
						result.Targets = append(result.Targets, DiscoveredTarget{
							RawURL:    su,
							AssetType: model.AssetImage,
							Snippet:   rawTagSnippet,
						})
					}
				}

			case atom.Script:
				for _, attr := range tok.Attr {
					if attr.Key == "src" && attr.Val != "" {
						result.Targets = append(result.Targets, DiscoveredTarget{
							RawURL:    strings.TrimSpace(attr.Val),
							AssetType: model.AssetScript,
							Snippet:   rawTagSnippet,
						})
					}
				}

			case atom.Link:
				var rel, href string
				for _, attr := range tok.Attr {
					switch attr.Key {
					case "rel":
						rel = strings.ToLower(strings.TrimSpace(attr.Val))
					case "href":
						href = strings.TrimSpace(attr.Val)
					}
				}
				if href != "" {
					assetType := model.AssetLink
					switch {
					case strings.Contains(rel, "stylesheet"):
						assetType = model.AssetStylesheet
					case strings.Contains(rel, "preload") || strings.Contains(rel, "modulepreload"):
						assetType = model.AssetPreload
					case strings.Contains(rel, "icon"):
						assetType = model.AssetFavicon
					case strings.Contains(rel, "canonical"):
						assetType = model.AssetCanonical
						result.CanonicalURL = href
					}
					result.Targets = append(result.Targets, DiscoveredTarget{
						RawURL:    href,
						AssetType: assetType,
						Snippet:   rawTagSnippet,
					})
				}

			case atom.Video, atom.Audio:
				for _, attr := range tok.Attr {
					if (attr.Key == "src" || attr.Key == "poster") && attr.Val != "" {
						result.Targets = append(result.Targets, DiscoveredTarget{
							RawURL:    strings.TrimSpace(attr.Val),
							AssetType: model.AssetMedia,
							Snippet:   rawTagSnippet,
						})
					}
				}

			case atom.Track:
				for _, attr := range tok.Attr {
					if attr.Key == "src" && attr.Val != "" {
						result.Targets = append(result.Targets, DiscoveredTarget{
							RawURL:    strings.TrimSpace(attr.Val),
							AssetType: model.AssetMedia,
							Snippet:   rawTagSnippet,
						})
					}
				}

			case atom.Iframe, atom.Embed:
				for _, attr := range tok.Attr {
					if attr.Key == "src" && attr.Val != "" {
						result.Targets = append(result.Targets, DiscoveredTarget{
							RawURL:    strings.TrimSpace(attr.Val),
							AssetType: model.AssetLink,
							Snippet:   rawTagSnippet,
						})
					}
				}
			}

		case html.TextToken:
			if inTitle {
				titleBuf.WriteString(tok.Data)
			}
			if inStyle {
				styleBuf.WriteString(tok.Data)
			}

		case html.EndTagToken:
			if tok.DataAtom == atom.Title {
				inTitle = false
				result.PageTitle = strings.TrimSpace(titleBuf.String())
			}
			if tok.DataAtom == atom.Style {
				inStyle = false
				cssContent := styleBuf.String()
				cssURLs := ExtractCSSURLs(cssContent)
				for _, cu := range cssURLs {
					result.Targets = append(result.Targets, DiscoveredTarget{
						RawURL:    cu,
						AssetType: model.AssetCSS,
						Text:      "stylesheet block",
						Snippet:   "/* style block */",
					})
				}
			}
		}
	}

	return result, nil
}

// ExtractHTMLFromBytes parses HTML from a byte slice.
func ExtractHTMLFromBytes(body []byte, baseURL string) (*ExtractionResult, error) {
	return ExtractHTML(bytes.NewReader(body), baseURL)
}

func extractFragment(rawURL string) string {
	idx := strings.IndexByte(rawURL, '#')
	if idx >= 0 && idx < len(rawURL)-1 {
		frag := rawURL[idx+1:]
		if unescaped, err := url.PathUnescape(frag); err == nil {
			return unescaped
		}
		return frag
	}
	return ""
}

// parseSrcset parses HTML5 srcset strings into individual URL strings.
func parseSrcset(srcset string) []string {
	var urls []string
	parts := strings.Split(srcset, ",")
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) > 0 {
			urls = append(urls, fields[0])
		}
	}
	return urls
}
