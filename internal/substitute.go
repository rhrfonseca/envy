package envy

import (
	"regexp"
	"strings"
)

var substitutionRe = regexp.MustCompile(`\$\$\{[^}]*\}|\$\{([A-Za-z_][A-Za-z0-9_]*)(?::-([^}]*))?\}`)

func Substitute(content string, envMap map[string]string) (string, []string) {
	var missingVars []string
	seenMissing := make(map[string]bool)
	var sb strings.Builder
	lastIdx := 0

	for _, loc := range substitutionRe.FindAllStringSubmatchIndex(content, -1) {
		sb.WriteString(content[lastIdx:loc[0]])

		match := content[loc[0]:loc[1]]

		if strings.HasPrefix(match, "$$") {
			sb.WriteString(match[1:])
		} else {
			varName := content[loc[2]:loc[3]]
			hasDefault := loc[4] != -1
			var defaultVal string
			if hasDefault {
				defaultVal = content[loc[4]:loc[5]]
			}

			val, found := envMap[varName]
			if !found {
				if hasDefault {
					val = defaultVal
				} else if !seenMissing[varName] {
					missingVars = append(missingVars, varName)
					seenMissing[varName] = true
				}
			}
			if found || hasDefault {
				sb.WriteString(val)
			} else {
				sb.WriteString(match)
			}
		}

		lastIdx = loc[1]
	}

	sb.WriteString(content[lastIdx:])
	return sb.String(), missingVars
}
