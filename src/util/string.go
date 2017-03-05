package util

import (
	"regexp"
	"strings"
)

type RegExp struct {
	Pattern string
	Replace string
}

func Replace(str string, rs []RegExp) string {
	for _, r := range rs {
		patt := regexp.MustCompile(r.Pattern)
		repl := r.Replace
		str = patt.ReplaceAllString(str, repl)
	}

	return str
}

func XPathToSelector(xpath string) string {
	var parts []string

	rs := []RegExp{}

	rs = append(rs, RegExp{`\[@id=['"]([^'"]*?)['"]\]`, `#$1`})
	rs = append(rs, RegExp{`\[@class=['"]([^'"]*?)['"]\]`, `.$1`})

	arr := strings.Split(xpath, "/")
	for _, v := range arr {
		if v != "" {
			parts = append(parts, Replace(v, rs))
		}
	}

	return strings.Join(parts, " ")
}

func Substring(str string, start, length int) string {
	if start < 0 || length <= 0 {
		return str
	}

	r := []rune(str)
	if start + length - 1 > len(r) {
		return string(r[start:])
	} else {
		return string(r[start:(start + length - 1)])
	}
}
