package update

import (
	"fmt"
	"strconv"
	"strings"
)

type SemVer struct {
	Major int
	Minor int
	Patch int
}

func NormalizeSemverString(raw string) (string, error) {
	v, err := ParseSemver(raw)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch), nil
}

func ParseSemver(raw string) (SemVer, error) {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(s, "v")
	if i := strings.IndexAny(s, "-+"); i >= 0 {
		s = s[:i]
	}
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return SemVer{}, fmt.Errorf("invalid semver: %q", raw)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return SemVer{}, fmt.Errorf("invalid semver: %q", raw)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return SemVer{}, fmt.Errorf("invalid semver: %q", raw)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return SemVer{}, fmt.Errorf("invalid semver: %q", raw)
	}
	return SemVer{Major: major, Minor: minor, Patch: patch}, nil
}

func CompareSemver(a SemVer, b SemVer) int {
	if a.Major != b.Major {
		return cmp(a.Major, b.Major)
	}
	if a.Minor != b.Minor {
		return cmp(a.Minor, b.Minor)
	}
	return cmp(a.Patch, b.Patch)
}

func cmp(a int, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
