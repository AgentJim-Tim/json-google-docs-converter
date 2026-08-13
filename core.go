package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type JSONStats struct {
	Objects int
	Arrays  int
	Values  int
	Depth   int
}

type DocumentPreview struct {
	Title       string
	PlainText   string
	HTML        string
	Stats       JSONStats
	ApproxPages int
}

var acronym = map[string]string{
	"api": "API", "url": "URL", "id": "ID", "json": "JSON", "docx": "DOCX",
	"gpu": "GPU", "ram": "RAM", "mfa": "MFA", "sso": "SSO", "vpn": "VPN",
	"ip": "IP", "cpu": "CPU", "ui": "UI", "ux": "UX", "http": "HTTP", "https": "HTTPS",
}

var preferredRootOrder = []string{
	"title", "subtitle", "document_title", "document_date", "date", "status", "version", "confidential",
	"owner", "executive_summary", "summary", "overview", "objectives", "project_metrics", "metrics",
	"milestones", "architecture", "risks", "change_windows", "notes", "approval",
}

func parseJSON(data []byte) (any, error) {
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	var v any
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	if dec.More() {
		return nil, fmt.Errorf("multiple JSON values found")
	}
	return v, nil
}

func buildPreview(data []byte, fallbackTitle string) (DocumentPreview, error) {
	v, err := parseJSON(data)
	if err != nil {
		return DocumentPreview{}, err
	}
	title := deriveTitle(v, fallbackTitle)
	stats := collectStats(v, 1)
	plain := renderPlainDocument(title, v)
	rich := renderHTMLDocument(title, v)
	pages := estimatePages(plain, stats)
	return DocumentPreview{Title: title, PlainText: plain, HTML: rich, Stats: stats, ApproxPages: pages}, nil
}

func deriveTitle(v any, fallback string) string {
	if m, ok := v.(map[string]any); ok {
		for _, k := range []string{"title", "document_title", "name", "subject", "project"} {
			if s, ok := m[k].(string); ok && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
	}
	fallback = strings.TrimSpace(fallback)
	if fallback == "" {
		return "Converted JSON"
	}
	fallback = strings.TrimSuffix(fallback, ".json")
	return labelize(fallback)
}

func collectStats(v any, depth int) JSONStats {
	s := JSONStats{Depth: depth}
	switch x := v.(type) {
	case map[string]any:
		s.Objects = 1
		for _, child := range x {
			c := collectStats(child, depth+1)
			s.Objects += c.Objects
			s.Arrays += c.Arrays
			s.Values += c.Values
			if c.Depth > s.Depth {
				s.Depth = c.Depth
			}
		}
	case []any:
		s.Arrays = 1
		for _, child := range x {
			c := collectStats(child, depth+1)
			s.Objects += c.Objects
			s.Arrays += c.Arrays
			s.Values += c.Values
			if c.Depth > s.Depth {
				s.Depth = c.Depth
			}
		}
	default:
		s.Values = 1
	}
	return s
}

func estimatePages(plain string, s JSONStats) int {
	weighted := float64(len([]rune(plain))) + float64(s.Objects*85+s.Arrays*45)
	n := int(math.Ceil(weighted / 3400.0))
	if n < 1 {
		n = 1
	}
	return n
}

func orderedKeys(m map[string]any) []string {
	seen := make(map[string]bool, len(m))
	out := make([]string, 0, len(m))
	for _, k := range preferredRootOrder {
		if _, ok := m[k]; ok {
			out = append(out, k)
			seen[k] = true
		}
	}
	rest := make([]string, 0, len(m))
	for k := range m {
		if !seen[k] {
			rest = append(rest, k)
		}
	}
	sort.Strings(rest)
	return append(out, rest...)
}

func labelize(s string) string {
	s = strings.TrimSpace(s)
	s = strings.NewReplacer("_", " ", "-", " ").Replace(s)
	fields := strings.Fields(s)
	for i, f := range fields {
		low := strings.ToLower(f)
		if a, ok := acronym[low]; ok {
			fields[i] = a
			continue
		}
		r := []rune(low)
		if len(r) > 0 {
			r[0] = unicode.ToUpper(r[0])
		}
		fields[i] = string(r)
	}
	return strings.Join(fields, " ")
}

func scalarString(v any) string {
	switch x := v.(type) {
	case nil:
		return "—"
	case bool:
		if x {
			return "Yes"
		}
		return "No"
	case json.Number:
		return x.String()
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case string:
		return x
	default:
		return fmt.Sprint(x)
	}
}

func isScalar(v any) bool {
	switch v.(type) {
	case nil, bool, string, json.Number, float64, int, int64:
		return true
	}
	return false
}

func tableCandidate(a []any) ([]string, bool) {
	if len(a) < 2 || len(a) > 100 {
		return nil, false
	}
	keySet := map[string]bool{}
	for _, row := range a {
		m, ok := row.(map[string]any)
		if !ok {
			return nil, false
		}
		for k, v := range m {
			if !isScalar(v) {
				return nil, false
			}
			keySet[k] = true
			if len(keySet) > 8 {
				return nil, false
			}
		}
	}
	keys := make([]string, 0, len(keySet))
	priority := []string{"id", "name", "phase", "status", "owner", "start_date", "end_date", "date", "completion_percent"}
	used := map[string]bool{}
	for _, k := range priority {
		if keySet[k] {
			keys = append(keys, k)
			used[k] = true
		}
	}
	rest := []string{}
	for k := range keySet {
		if !used[k] {
			rest = append(rest, k)
		}
	}
	sort.Strings(rest)
	return append(keys, rest...), true
}

func renderPlainDocument(title string, v any) string {
	var b strings.Builder
	b.WriteString(title)
	b.WriteString("\n\n")
	renderPlain(&b, v, 0, true)
	return strings.TrimSpace(b.String()) + "\n"
}

func renderPlain(b *strings.Builder, v any, depth int, root bool) {
	indent := strings.Repeat("  ", depth)
	switch x := v.(type) {
	case map[string]any:
		keys := orderedKeys(x)
		for _, k := range keys {
			if root && (k == "title" || k == "document_title") {
				continue
			}
			val := x[k]
			if isScalar(val) {
				text := scalarString(val)
				if len([]rune(text)) > 180 || strings.Contains(text, "\n") {
					b.WriteString("\n" + indent + labelize(k) + "\n")
					b.WriteString(indent + text + "\n")
				} else {
					b.WriteString(indent + labelize(k) + ": " + text + "\n")
				}
				continue
			}
			b.WriteString("\n" + indent + labelize(k) + "\n")
			renderPlain(b, val, depth+1, false)
		}
	case []any:
		if keys, ok := tableCandidate(x); ok {
			for _, row := range x {
				m := row.(map[string]any)
				cells := make([]string, 0, len(keys))
				for _, k := range keys {
					cells = append(cells, labelize(k)+": "+scalarString(m[k]))
				}
				b.WriteString(indent + "• " + strings.Join(cells, " | ") + "\n")
			}
		} else {
			for _, item := range x {
				if isScalar(item) {
					b.WriteString(indent + "• " + scalarString(item) + "\n")
				} else {
					renderPlain(b, item, depth+1, false)
				}
			}
		}
	default:
		b.WriteString(indent + scalarString(x) + "\n")
	}
}

func renderHTMLDocument(title string, v any) string {
	var b strings.Builder
	b.WriteString(`<!doctype html><html><head><meta charset="utf-8"><style>`)
	b.WriteString(`body{font-family:Arial,sans-serif;color:#262421;font-size:11pt;line-height:1.45}h1{font-family:Georgia,serif;font-size:24pt;margin:0 0 18px}h2{font-family:Georgia,serif;font-size:16pt;margin:22px 0 8px}h3{font-family:Georgia,serif;font-size:13pt;margin:16px 0 6px}p{margin:4px 0 9px}ul{margin:5px 0 12px 22px}table{border-collapse:collapse;width:100%;margin:8px 0 16px}th,td{border:1px solid #d9d4cc;padding:6px 8px;text-align:left;vertical-align:top}th{background:#f1eee8;font-weight:600}.kv{margin:3px 0}.label{font-weight:600}`)
	b.WriteString(`</style></head><body>`)
	b.WriteString("<h1>" + html.EscapeString(title) + "</h1>")
	renderHTML(&b, v, 0, true)
	b.WriteString("</body></html>")
	return b.String()
}

func renderHTML(b *strings.Builder, v any, depth int, root bool) {
	switch x := v.(type) {
	case map[string]any:
		for _, k := range orderedKeys(x) {
			if root && (k == "title" || k == "document_title") {
				continue
			}
			val := x[k]
			if isScalar(val) {
				text := scalarString(val)
				if len([]rune(text)) > 180 || strings.Contains(text, "\n") {
					b.WriteString("<h2>" + html.EscapeString(labelize(k)) + "</h2><p>" + strings.ReplaceAll(html.EscapeString(text), "\n", "<br>") + "</p>")
				} else {
					b.WriteString(`<p class="kv"><span class="label">` + html.EscapeString(labelize(k)) + `:</span> ` + html.EscapeString(text) + `</p>`)
				}
				continue
			}
			level := 2
			if depth > 0 {
				level = 3
			}
			b.WriteString(fmt.Sprintf("<h%d>%s</h%d>", level, html.EscapeString(labelize(k)), level))
			renderHTML(b, val, depth+1, false)
		}
	case []any:
		if keys, ok := tableCandidate(x); ok {
			b.WriteString("<table><thead><tr>")
			for _, k := range keys {
				b.WriteString("<th>" + html.EscapeString(labelize(k)) + "</th>")
			}
			b.WriteString("</tr></thead><tbody>")
			for _, row := range x {
				m := row.(map[string]any)
				b.WriteString("<tr>")
				for _, k := range keys {
					b.WriteString("<td>" + html.EscapeString(scalarString(m[k])) + "</td>")
				}
				b.WriteString("</tr>")
			}
			b.WriteString("</tbody></table>")
		} else {
			b.WriteString("<ul>")
			for _, item := range x {
				if isScalar(item) {
					b.WriteString("<li>" + html.EscapeString(scalarString(item)) + "</li>")
				} else {
					b.WriteString("<li>")
					renderHTML(b, item, depth+1, false)
					b.WriteString("</li>")
				}
			}
			b.WriteString("</ul>")
		}
	default:
		b.WriteString("<p>" + html.EscapeString(scalarString(x)) + "</p>")
	}
}

var keyValueLine = regexp.MustCompile(`^\s*([^:]{1,80}):\s*(.+?)\s*$`)

func clipboardTextToJSON(text string) map[string]any {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	out := map[string]any{}
	content := []any{}
	titleSet := false
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if !titleSet {
			out["title"] = line
			titleSet = true
			continue
		}
		if m := keyValueLine.FindStringSubmatch(line); m != nil {
			out[toSnake(m[1])] = inferScalar(m[2])
			continue
		}
		if strings.HasPrefix(line, "• ") || strings.HasPrefix(line, "- ") {
			content = append(content, strings.TrimSpace(line[2:]))
			continue
		}
		content = append(content, line)
	}
	if len(content) > 0 {
		out["content"] = content
	}
	if !titleSet {
		out["title"] = "Google Doc Export"
	}
	return out
}

func inferScalar(s string) any {
	t := strings.TrimSpace(s)
	low := strings.ToLower(t)
	switch low {
	case "yes", "true":
		return true
	case "no", "false":
		return false
	case "null", "—", "-":
		return nil
	}
	if i, err := strconv.ParseInt(t, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(t, 64); err == nil && strings.ContainsAny(t, ".eE") {
		return f
	}
	return t
}

func toSnake(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "_")
	return strings.Trim(s, "_")
}
