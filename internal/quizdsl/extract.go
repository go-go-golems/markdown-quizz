package quizdsl

import (
	stderrors "errors"
	"regexp"
	"strings"

	pkgerrors "github.com/pkg/errors"
)

type ExtractedForm struct {
	FormID     string
	Definition any
}

var (
	formOpenTagIDRegexp = regexp.MustCompile(`(?i)\bid\s*=\s*(?:"([^"]+)"|'([^']+)'|([^\s>]+))`)
)

func ExtractFormsFromContent(content string) ([]ExtractedForm, error) {
	var forms []ExtractedForm
	var errs []error

	lower := strings.ToLower(content)
	i := 0
	for {
		start := strings.Index(lower[i:], "<form")
		if start == -1 {
			break
		}
		start += i

		openEnd := strings.IndexByte(content[start:], '>')
		if openEnd == -1 {
			break
		}
		openEnd += start

		openTag := content[start : openEnd+1]
		formID := parseFormID(openTag)
		if formID == "" {
			i = openEnd + 1
			continue
		}

		closeStart := strings.Index(lower[openEnd+1:], "</form>")
		if closeStart == -1 {
			i = openEnd + 1
			continue
		}
		closeStart += openEnd + 1

		yamlContent := strings.TrimSpace(content[openEnd+1 : closeStart])
		definition, err := ParseYAMLDefinition(yamlContent)
		if err != nil {
			errs = append(errs, pkgerrors.Wrapf(err, "parse yaml for form %s", formID))
			i = closeStart + len("</form>")
			continue
		}

		forms = append(forms, ExtractedForm{
			FormID:     formID,
			Definition: definition,
		})

		i = closeStart + len("</form>")
	}

	if len(errs) > 0 {
		return forms, pkgerrors.Wrap(stderrors.Join(errs...), "extract forms")
	}
	return forms, nil
}

func parseFormID(openTag string) string {
	m := formOpenTagIDRegexp.FindStringSubmatch(openTag)
	if len(m) == 0 {
		return ""
	}
	for i := 1; i < len(m); i++ {
		if m[i] != "" {
			return m[i]
		}
	}
	return ""
}
