package hosts

import "strings"

const (
	StartMarker  = "# === LocalHosts START ==="
	EndMarker    = "# === LocalHosts END ==="
	HeaderMarker = "# Managed by LocalHosts. Do not edit within this block."
)

func DetectEOL(raw string) string {
	if strings.Contains(raw, "\r\n") {
		return "\r\n"
	}
	return "\n"
}

func SplitDocument(raw []byte) (HostDocument, []string, bool, bool) {
	rawStr := string(raw)
	eol := DetectEOL(rawStr)
	endsWithEOL := strings.HasSuffix(rawStr, eol)

	lines := strings.Split(rawStr, eol)
	if endsWithEOL && len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	startIndex := -1
	endIndex := -1

	for i, line := range lines {
		if line == StartMarker && startIndex == -1 {
			startIndex = i
			continue
		}
		if line == EndMarker && startIndex != -1 {
			endIndex = i
			break
		}
	}

	doc := HostDocument{
		EOL:         eol,
		EndsWithEOL: endsWithEOL,
	}

	if startIndex == -1 && endIndex == -1 {
		doc.PrefixLines = lines
		doc.SuffixLines = []string{}
		return doc, []string{}, false, false
	}

	if startIndex == -1 || endIndex == -1 || endIndex <= startIndex {
		doc.PrefixLines = lines
		doc.SuffixLines = []string{}
		return doc, []string{}, false, true
	}

	doc.PrefixLines = lines[:startIndex]
	blockLines := lines[startIndex : endIndex+1]
	doc.SuffixLines = lines[endIndex+1:]
	return doc, blockLines, true, false
}
