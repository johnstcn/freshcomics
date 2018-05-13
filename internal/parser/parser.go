package parser

import (
	"regexp"
	"strings"

	"io"

	"github.com/pkg/errors"
	"gopkg.in/xmlpath.v2"
)

//go:generate mockery -interface Parser -package parsertest

var (
	ErrInvalidRegexp = errors.New("invalid regexp")
	ErrInvalidXPath  = errors.New("invalid xpath")
	ErrRegexpNoMatch = errors.New("no match for regexp")
	ErrXPathNoMatch  = errors.New("no match for xpath")
)

// Rule represents a targeted element on a Page
type Rule struct {
	XPath  string
	Filter string
}

type xPathCompiler func(path string) (*xmlpath.Path, error)
type regexpCompiler func(expr string) (*regexp.Regexp, error)

// Parser applies a Rule to its parsed Page
type Parser interface {
	Apply(r Rule) (string, error)
}

func NewParser(r io.Reader) (Parser, error) {
	return newXPathParser(r)
}

func newXPathParser(r io.Reader) (Parser, error) {
	page, err := xmlpath.ParseHTML(r)
	if err != nil {
		return &xPathParser{}, errors.Wrap(err, "parsing Page")
	}
	return &xPathParser{
		Page:          page,
		CompileXPath:  xmlpath.Compile,
		CompileRegexp: regexp.Compile,
	}, nil
}

// xPathParser implements Parser
type xPathParser struct {
	Page          *xmlpath.Node
	CompileXPath  xPathCompiler
	CompileRegexp regexpCompiler
}

var _ Parser = (*xPathParser)(nil)

func (p *xPathParser) Apply(r Rule) (string, error) {
	rawValue, err := p.applyXPath(r.XPath)
	if err != nil {
		return "", err
	}

	return p.applyFilter(r.Filter, rawValue)
}

func (p *xPathParser) applyXPath(path string) (string, error) {
	xp, err := p.CompileXPath(path)
	if err != nil {
		return "", ErrInvalidXPath
	}

	val, ok := xp.String(p.Page)
	if !ok {
		return "", ErrXPathNoMatch
	}

	return val, nil
}

func (p *xPathParser) applyFilter(expr, text string) (string, error) {
	r, err := p.CompileRegexp(expr)
	if err != nil {
		return "", ErrInvalidRegexp
	}

	match := r.FindStringSubmatch(text)
	if match == nil || len(match) == 0 {
		return text, ErrRegexpNoMatch
	}

	if len(match) == 1 {
		return match[0], nil
	}

	return strings.TrimSpace(match[1]), nil
}
