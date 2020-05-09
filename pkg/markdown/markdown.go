package markdown

import (
	md "github.com/MichaelMure/go-term-markdown"
)

func ConvertShellString(source string) []byte {
	result := md.Render(string(source), 80, 6)
	return result
}
