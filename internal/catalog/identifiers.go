package catalog

import (
	"strings"
	"unicode"
)

func CompactISBN(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	return strings.Map(func(c rune) rune {
		if unicode.IsSpace(c) || c == '-' || c == '‐' || c == '‑' || c == '–' {
			return -1
		}
		return c
	}, value)
}
func ValidISBN10(value string) bool {
	if len(value) != 10 {
		return false
	}
	sum := 0
	for i, c := range value {
		n := int(c - '0')
		if i == 9 && c == 'X' {
			n = 10
		} else if c < '0' || c > '9' {
			return false
		}
		sum += (10 - i) * n
	}
	return sum%11 == 0
}
func ValidISBN13(value string) bool {
	if len(value) != 13 || (!strings.HasPrefix(value, "978") && !strings.HasPrefix(value, "979")) {
		return false
	}
	sum := 0
	for i, c := range value {
		if c < '0' || c > '9' {
			return false
		}
		weight := 1
		if i%2 == 1 {
			weight = 3
		}
		sum += int(c-'0') * weight
	}
	return sum%10 == 0
}
func ISBN13(value string) string {
	value = CompactISBN(value)
	if ValidISBN13(value) {
		return value
	}
	if !ValidISBN10(value) {
		return ""
	}
	base := "978" + value[:9]
	sum := 0
	for i, c := range base {
		w := 1
		if i%2 == 1 {
			w = 3
		}
		sum += int(c-'0') * w
	}
	return base + string(rune('0'+(10-sum%10)%10))
}
func ISBN10(value string) string {
	value = CompactISBN(value)
	if ValidISBN10(value) {
		return value
	}
	if !ValidISBN13(value) || !strings.HasPrefix(value, "978") {
		return ""
	}
	base := value[3:12]
	sum := 0
	for i, c := range base {
		sum += (10 - i) * int(c-'0')
	}
	n := (11 - sum%11) % 11
	check := string(rune('0' + n))
	if n == 10 {
		check = "X"
	}
	return base + check
}
func ValidIdentifier(v Identifier) bool {
	if v.Value == "" {
		return false
	}
	switch v.Scheme {
	case "isbn10":
		return ValidISBN10(v.Value)
	case "isbn13":
		return ValidISBN13(v.Value)
	case "asin":
		return len(v.Value) == 10 && strings.IndexFunc(v.Value, func(c rune) bool { return !(c >= 'A' && c <= 'Z' || c >= '0' && c <= '9') }) < 0
	case "other":
		return true
	}
	return false
}
func NormalizeIdentifier(raw, source string) []Identifier {
	compact := CompactISBN(raw)
	if value := ISBN13(compact); value != "" {
		result := []Identifier{{Scheme: "isbn13", Value: value, Source: source, Derived: len(compact) == 10}}
		if short := ISBN10(value); short != "" {
			result = append(result, Identifier{Scheme: "isbn10", Value: short, Source: source, Derived: len(compact) == 13})
		}
		return result
	}
	if len(compact) == 10 && strings.HasPrefix(compact, "B") && ValidIdentifier(Identifier{Scheme: "asin", Value: compact}) {
		return []Identifier{{Scheme: "asin", Value: compact, Source: source}}
	}
	if strings.TrimSpace(raw) != "" {
		return []Identifier{{Scheme: "other", Value: strings.TrimSpace(raw), Source: source}}
	}
	return []Identifier{}
}
func MergeIdentifiers(values ...[]Identifier) []Identifier {
	result := make([]Identifier, 0)
	seen := map[string]bool{}
	for _, list := range values {
		for _, v := range list {
			key := v.Scheme + ":" + v.Value
			if !seen[key] && ValidIdentifier(v) {
				seen[key] = true
				result = append(result, v)
			}
		}
	}
	return result
}
func PreferredISBN(values []Identifier) string {
	for _, v := range values {
		if v.Scheme == "isbn13" && ValidISBN13(v.Value) {
			return v.Value
		}
	}
	return ""
}

// A conflicting ISBN-10 is evidence, not a second canonical alias of the ISBN-13.
func AcceptDeclared(values []Identifier) ([]Identifier, []string) {
	var ten, thirteen string
	warnings := []string{}
	extra := []Identifier{}
	for _, v := range values {
		value := CompactISBN(v.Value)
		if v.Scheme == "isbn13" && ValidISBN13(value) {
			thirteen = value
		}
		if v.Scheme == "isbn10" && ValidISBN10(value) {
			ten = value
		}
		if v.Scheme == "asin" {
			v.Value = value
			if ValidIdentifier(v) {
				extra = append(extra, v)
			}
		}
		if (v.Scheme == "isbn10" && !ValidISBN10(value)) || (v.Scheme == "isbn13" && !ValidISBN13(value)) {
			warnings = append(warnings, "El proveedor declaró un ISBN inválido; no se usó para deduplicar.")
		}
	}
	if thirteen != "" {
		if ten != "" && ISBN13(ten) != thirteen {
			warnings = append(warnings, "Los ISBN-10/13 declarados no son equivalentes; se conservó el ISBN-13 y su equivalencia calculada.")
		}
		return MergeIdentifiers(NormalizeIdentifier(thirteen, Hardcover), extra), warnings
	}
	if ten != "" {
		return MergeIdentifiers(NormalizeIdentifier(ten, Hardcover), extra), warnings
	}
	return extra, warnings
}
