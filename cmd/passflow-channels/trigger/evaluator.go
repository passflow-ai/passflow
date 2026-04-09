package trigger

import (
	"bytes"
	"regexp"
	"strings"
	"text/template"

	"github.com/jaak-ai/passflow-channels/domain"
)

// Matches returns true if the event satisfies the rule's condition.
func Matches(rule domain.TriggerRule, event domain.Event) bool {
	if !rule.Enabled {
		return false
	}

	cond := rule.Condition

	// "always" operator → matches everything (e.g. cron triggers)
	if cond.Operator == "always" || cond.Field == "" {
		return true
	}

	value, ok := event.Fields[cond.Field]
	if !ok {
		return false
	}

	switch cond.Operator {
	case "contains":
		return strings.Contains(strings.ToLower(value), strings.ToLower(cond.Value))
	case "equals":
		return strings.EqualFold(value, cond.Value)
	case "starts_with":
		return strings.HasPrefix(strings.ToLower(value), strings.ToLower(cond.Value))
	case "ends_with":
		return strings.HasSuffix(strings.ToLower(value), strings.ToLower(cond.Value))
	case "matches_regex":
		re, err := regexp.Compile(cond.Value)
		if err != nil {
			return false
		}
		return re.MatchString(value)
	case "not_empty":
		return value != ""
	default:
		return false
	}
}

// RenderInput renders the action's InputTemplate with the event fields as context.
func RenderInput(action domain.Action, event domain.Event) (string, error) {
	if action.InputTemplate == "" {
		// Default: use the raw "text" or "body" field, or a JSON dump
		if text, ok := event.Fields["text"]; ok {
			return text, nil
		}
		if body, ok := event.Fields["body"]; ok {
			return body, nil
		}
		return "Process this event.", nil
	}

	tmpl, err := template.New("input").Parse(action.InputTemplate)
	if err != nil {
		return "", err
	}

	// Make fields accessible as .FieldName in the template
	data := make(map[string]string, len(event.Fields))
	for k, v := range event.Fields {
		// Capitalize first letter for template access ({{.Text}}, {{.Subject}}, etc.)
		key := strings.ToUpper(k[:1]) + k[1:]
		data[key] = v
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
