package textParser

import (
	"io"

	"github.com/sch8ill/gscrawler/types"
)

type TextParser struct {
	site       *types.Site
	bodyStream io.ReadCloser
}

func New(site *types.Site, bodyStream io.ReadCloser) *TextParser {
	return &TextParser{
		site:       site,
		bodyStream: bodyStream,
	}
}

func (p *TextParser) Parse() error {
	binary, err := io.ReadAll(p.bodyStream)
	if err != nil {
		return err
	}

	p.site.Text = []string{string(binary)}
	return nil
}
