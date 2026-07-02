package jsonview

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
)

var (
	keyStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Bold(true)
	stringStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("113"))
	numberStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	literalStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("176"))
)

// Highlight color-codes already-indented JSON: a "..." string immediately
// followed by ':' (ignoring whitespace) is treated as a key, otherwise a
// value. This single lookahead avoids needing a full JSON parser.
func Highlight(indented string) string {
	runes := []rune(indented)
	n := len(runes)
	var out strings.Builder

	i := 0
	for i < n {
		c := runes[i]

		if c == '"' {
			j := i + 1
			for j < n {
				if runes[j] == '\\' {
					j += 2
					continue
				}
				if runes[j] == '"' {
					break
				}
				j++
			}
			end := j
			if end < n {
				end++ // include closing quote
			}
			str := string(runes[i:end])

			k := end
			for k < n && unicode.IsSpace(runes[k]) {
				k++
			}
			if k < n && runes[k] == ':' {
				out.WriteString(keyStyle.Render(str))
			} else {
				out.WriteString(stringStyle.Render(str))
			}
			i = end
			continue
		}

		if c == '-' || unicode.IsDigit(c) {
			j := i
			for j < n && isNumberRune(runes[j]) {
				j++
			}
			if j > i {
				out.WriteString(numberStyle.Render(string(runes[i:j])))
				i = j
				continue
			}
		}

		if lit, ok := literalAt(runes, i); ok {
			out.WriteString(literalStyle.Render(lit))
			i += len(lit)
			continue
		}

		out.WriteRune(c)
		i++
	}

	return out.String()
}

func isNumberRune(r rune) bool {
	return unicode.IsDigit(r) || r == '.' || r == 'e' || r == 'E' || r == '+' || r == '-'
}

func literalAt(runes []rune, i int) (string, bool) {
	for _, lit := range []string{"true", "false", "null"} {
		l := len(lit)
		if i+l > len(runes) {
			continue
		}
		if string(runes[i:i+l]) != lit {
			continue
		}
		// word boundary check: next rune (if any) must not be alnum
		if i+l < len(runes) && (unicode.IsLetter(runes[i+l]) || unicode.IsDigit(runes[i+l])) {
			continue
		}
		return lit, true
	}
	return "", false
}
