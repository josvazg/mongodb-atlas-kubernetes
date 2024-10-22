package akogen

import (
	"fmt"
	"unicode"
)

var knownShortNames = map[string]string{
	"context.Context": "ctx",
	"error":           "err",
}

func shortenName(name string) string {
	return shorten(".", name)
}

func shorten(base, name string) string {
	if len(name) == 0 {
		return ""
	}
	fullName := fullName(base, name)
	shortName, ok := knownShortNames[fullName]
	if ok {
		return shortName
	}
	filtered := string(name[0])
	for _, char := range name[1:] {
		if unicode.IsUpper(char) {
			filtered += string(unicode.ToLower(char))
		}
	}
	if base != "." {
		r := fmt.Sprintf("%s%s", base, firstToUpper(filtered))
		return r
	}
	r := firstToLower(filtered)
	return r
}

func fullName(base, name string) string {
	if base == "." {
		return name
	}
	return fmt.Sprintf("%s.%s", base, name)
}

// func removeBase(s, pkgPath string) string {
// 	if len(s) == 0 {
// 		return ""
// 	}
// 	base := filepath.Base(pkgPath)
// 	if strings.HasPrefix(s, base) && len(s) > len(base) {
// 		return firstToLower(strings.Replace(s, base, "", 1))
// 	}
// 	return s
// }
