package infrastructure

import (
	"path/filepath"
	"regexp"
	"strings"
)

// keep letters/numbers/._-@, replace the rest with _
var reUnsafe = regexp.MustCompile(`[^a-zA-Z0-9._\-@]+`)

func SafeFilePart(s string) string {
	s = strings.TrimSpace(s)
	// remove any path separators from the current OS just in case
	s = strings.ReplaceAll(s, string(filepath.Separator), "_")
	// and the second separator (relevant for Windows, where both may occur)
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")

	s = reUnsafe.ReplaceAllString(s, "_")
	if s == "" {
		s = "unknown"
	}
	return s
}
