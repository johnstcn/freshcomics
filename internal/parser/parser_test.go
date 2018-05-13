package parser

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

var exampleHTML = `<html>
<body>
<a href="/path/to/somewhere?foo=bar">Somewhere</a>
</body>
</html>
`

type badReader struct{}

func (r *badReader) Read([]byte) (int, error) {
	return 0, fmt.Errorf("read error")
}

func Test_NewXPathParser_OK(t *testing.T) {
	_, err := newXPathParser(strings.NewReader(exampleHTML))
	require.NoError(t, err)
}

func Test_NewXPathParser_Err(t *testing.T) {
	_, err := newXPathParser(&badReader{})
	require.EqualError(t, err, "parsing Page: read error")
}

func Test_Apply_OK(t *testing.T) {
	p, err := newXPathParser(strings.NewReader(exampleHTML))
	require.NoError(t, err)
	r := Rule{
		XPath:  "//a/@href",
		Filter: "foo=([^&]+)",
	}
	val, err := p.Apply(r)
	require.NoError(t, err)
	require.EqualValues(t, "bar", val)
}

func Test_Apply_NoGroup(t *testing.T) {
	p, err := newXPathParser(strings.NewReader(exampleHTML))
	require.NoError(t, err)
	r := Rule{
		XPath:  "//a/@href",
		Filter: ".+",
	}
	val, err := p.Apply(r)
	require.NoError(t, err)
	require.EqualValues(t, "/path/to/somewhere?foo=bar", val)
}

func Test_Apply_XPathNoMatch(t *testing.T) {
	p, err := newXPathParser(strings.NewReader(exampleHTML))
	require.NoError(t, err)
	r := Rule{
		XPath:  "//b/@href",
		Filter: "foo=([^&]+)",
	}
	val, err := p.Apply(r)
	require.EqualValues(t, ErrXPathNoMatch, err)
	require.Zero(t, val)
}

func Test_Apply_RegexNoMatch(t *testing.T) {
	p, err := newXPathParser(strings.NewReader(exampleHTML))
	require.NoError(t, err)
	r := Rule{
		XPath:  "//a/@href",
		Filter: "bazzle=([^&]+)",
	}
	val, err := p.Apply(r)
	require.EqualValues(t, ErrRegexpNoMatch, err)
	require.EqualValues(t, "/path/to/somewhere?foo=bar", val)
}

func Test_Apply_InvalidXPath(t *testing.T) {
	p, err := newXPathParser(strings.NewReader(exampleHTML))
	require.NoError(t, err)
	r := Rule{
		XPath:  "",
		Filter: "foo=([^&]+)",
	}
	val, err := p.Apply(r)
	require.EqualValues(t, ErrInvalidXPath, err)
	require.Zero(t, val)
}

func Test_Apply_InvalidRegexp(t *testing.T) {
	p, err := newXPathParser(strings.NewReader(exampleHTML))
	require.NoError(t, err)
	r := Rule{
		XPath:  "//a/@href",
		Filter: "(",
	}
	val, err := p.Apply(r)
	require.EqualValues(t, ErrInvalidRegexp, err)
	require.Zero(t, val)
}
