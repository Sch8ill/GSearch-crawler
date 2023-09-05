package htmlParser

import (
	"io"
	"strings"

	"golang.org/x/net/html"

	"github.com/sch8ill/gscrawler/crawler/parser/parseUtils"
	"github.com/sch8ill/gscrawler/types"
)

type HtmlParser struct {
	site       *types.Site
	bodyStream io.ReadCloser
	tokenizer  *html.Tokenizer
}

// list of html text tokens
var textTokenStartTags []string = []string{
	"p",
	"h",
	"h1",
	"h2",
	"h3",
	"h4",
	"h5",
	"h6",
	"strong",
	"b",
	"em",
	"i",
	"u",
	"s",
	"sub",
	"sup",
	"blockquote",
	"cite",
	"code",
	"pre",
	"title",

	"span",
}

var metaContentTags []string = []string{"description", "keywords", "author"}

func New(site *types.Site, bodyStream io.ReadCloser) *HtmlParser {
	return &HtmlParser{
		site:       site,
		bodyStream: bodyStream,
	}
}

func (p *HtmlParser) Parse() {
	p.tokenizer = html.NewTokenizer(p.bodyStream)
	for {
		tokenType := p.tokenizer.Next()

		if tokenType == html.ErrorToken {
			break
		}

		token := p.tokenizer.Token()

		if tokenType == html.StartTagToken {
			p.parseStartTagToken(token)

		} else if parseUtils.Contains(textTokenStartTags, token.Data) {
			p.parseText()
		}
	}
}

func (p *HtmlParser) parseText() {
	for {
		tokenType := p.tokenizer.Next()

		if tokenType == html.TextToken {
			content := strings.TrimSpace(p.tokenizer.Token().Data)
			if content != "" {
				p.site.Text = append(p.site.Text, content)
			}
			break

		} else if tokenType == html.StartTagToken {
			token := p.tokenizer.Token()
			p.parseStartTagToken(token)

		}
	}
}

// parses a html.StartTag and returns usable data found in it
func (p *HtmlParser) parseStartTagToken(token html.Token) {
	// check if the tag is an anchor tag
	if token.Data == "a" {
		url := parseHref(token)
		if url != "" {
			p.site.Links = append(p.site.Links, url)
		}

		// check if the tag is a meta tag
	} else if token.Data == "meta" {
		metadata := parseMeta(token)
		if metadata != "" {
			p.site.Text = append(p.site.Text, metadata)
		}

	} else if parseUtils.Contains(textTokenStartTags, token.Data) {
		p.parseText()
	}
}

// parses the url out of an anchor tag
func parseHref(token html.Token) string {
	for _, attr := range token.Attr {
		if attr.Key == "href" {
			// check if the href is not a mail link
			if strings.Contains(attr.Val, "mailto") {
				return ""
			}
			return parseUtils.RemoveTagsfromUrl(attr.Val)
		}
	}
	return ""
}

// parses SE relevant data out of a meta tag
func parseMeta(token html.Token) string {
	usable := false
	for _, attr := range token.Attr {
		if parseUtils.Contains(metaContentTags, attr.Key) {
			usable = true
		} else if attr.Key == "content" {
			if usable {
				return attr.Val
			}
		}
	}
	return ""
}
