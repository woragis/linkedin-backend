package slug

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

var invalidChars = regexp.MustCompile(`[^a-z0-9]+`)

func FromName(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	if s == "" {
		return "user"
	}
	s = invalidChars.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "user"
	}
	return s
}

func EnsureUnique(ctx context.Context, db *gorm.DB, base string) (string, error) {
	candidate := base
	for i := 0; i < 100; i++ {
		var count int64
		if err := db.WithContext(ctx).Table("profiles").Where("slug = ?", candidate).Count(&count).Error; err != nil {
			return "", fmt.Errorf("slug check: %w", err)
		}
		if count == 0 {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", base, i+2)
	}
	return "", fmt.Errorf("could not allocate unique slug for %q", base)
}
