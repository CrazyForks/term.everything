package termeverything

import (
	"strings"

	"github.com/mmulet/term.everything/escapecodes"
)

func RenderMarkdownToTerminal(markdown string) string {
	var outLines []string
	for _, line := range strings.Split(markdown, "\n") {
		if strings.HasPrefix(line, "# ") {
			outLines = append(outLines, escapecodes.FgGreen+escapecodes.Underline+renderCode(line[2:])+escapecodes.Reset)
			continue
		}
		if strings.HasPrefix(line, "## ") {
			outLines = append(outLines, escapecodes.FgCyan+escapecodes.Underline+renderCode(line[3:])+escapecodes.Reset)
			continue
		}
		outLines = append(outLines, renderCode(line))
	}
	return strings.Join(outLines, "\n")
}

func renderCode(line string) string {
	var outLine strings.Builder
	inCode := false
	for _, char := range line {
		if char != '`' {
			outLine.WriteRune(char)
			continue
		}
		if inCode {
			outLine.WriteString(escapecodes.Reset)
			inCode = false
			continue
		}
		inCode = true
		outLine.WriteString(escapecodes.FgYellow)
	}
	return outLine.String()
}
