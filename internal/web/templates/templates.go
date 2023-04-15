package templates

import (
	"embed"
)

//go:embed style.css
var CSS string

//go:embed *.gohtml
var FS embed.FS
