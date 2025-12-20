package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/sagernet/sing-box/option"
)

// FetchList retrieves the content of a list from a URL or local file.
func FetchList(source string) ([]string, error) {
	var rc io.ReadCloser
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		resp, err := http.Get(source)
		if err != nil {
			return nil, err
		}
		rc = resp.Body
	} else {
		f, err := os.Open(source)
		if err != nil {
			return nil, err
		}
		rc = f
	}
	defer rc.Close()

	var lines []string
	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// ParseList cleans the raw lines by trimming whitespace and removing comments/empty lines.
func ParseList(lines []string) []string {
	var result []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Skip comments and headers
		if strings.HasPrefix(line, "#") ||
			strings.HasPrefix(line, "//") ||
			strings.HasPrefix(line, "!") ||
			strings.HasPrefix(line, "[") {
			continue
		}
		result = append(result, line)
	}
	return result
}

// ParseRuleType identifies the rule type and value from a line.
// Returns empty strings if invalid or unknown.
func ParseRuleType(line string) (string, string) {
	parts := strings.Split(line, ",")

	// Case 1: Surge/Clash format (TYPE,VALUE,...)
	if len(parts) >= 2 {
		rType := strings.ToUpper(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])

		// Whitelist check
		switch rType {
		case "DOMAIN-SUFFIX", "DOMAIN", "DOMAIN-KEYWORD", "IP-CIDR", "IP-CIDR6", "URL-REGEX", "DOMAIN-REGEX", "PROCESS-NAME":
			return rType, val
		}
		// If unknown type, treat as plain?? No, likely invalid or unsupported.
		return "", ""
	}

	// Case 2: Plain format
	// Check for implicit CIDR
	if strings.Contains(line, "/") {
		// Basic validation for CIDR chars?
		if strings.Contains(line, ":") {
			return "IP-CIDR6", line
		}
		return "IP-CIDR", line
	}

	// Check domain behaviors
	if strings.HasPrefix(line, ".") {
		return "DOMAIN-SUFFIX", strings.TrimPrefix(line, ".")
	}

	// Default to DOMAIN-SUFFIX for plain domains in these lists
	// Ensure it looks like a domain (no spaces)
	if strings.Contains(line, " ") {
		return "", ""
	}

	// Strict validation: Reject lines with characters invalid for standard domain syntax options.
	// This helps filter out regexes, base64 garbage, or unparsed headers that might crash strict parsers.
	if strings.ContainsAny(line, `\[]*(){}^$|\`) {
		return "", ""
	}

	// Adblock syntax mitigation (||example.com) - strict handling
	if strings.HasPrefix(line, "||") {
		return "DOMAIN-SUFFIX", strings.TrimPrefix(line, "||")
	}

	return "DOMAIN-SUFFIX", line
}

// GenerateSingBoxRules converts lines to sing-box DefaultHeadlessRule.
// Consolidates all domains/keywords into a single rule object.
func GenerateSingBoxRules(lines []string) []option.DefaultHeadlessRule {
	var (
		suffixes []string
		domains  []string
		keywords []string
		regex    []string
	)

	for _, line := range lines {
		rType, val := ParseRuleType(line)
		if rType == "" || val == "" {
			continue
		}
		switch rType {
		case "DOMAIN-SUFFIX":
			suffixes = append(suffixes, val)
		case "DOMAIN":
			domains = append(domains, val)
		case "DOMAIN-KEYWORD":
			keywords = append(keywords, val)
		case "URL-REGEX", "DOMAIN-REGEX":
			regex = append(regex, val)
		}
	}

	if len(suffixes) == 0 && len(domains) == 0 && len(keywords) == 0 && len(regex) == 0 {
		return nil
	}

	// Return a single consolidated rule
	return []option.DefaultHeadlessRule{{
		DomainSuffix:  suffixes,
		Domain:        domains,
		DomainKeyword: keywords,
		DomainRegex:   regex,
	}}
}

// GenerateMihomoText transforms lines for Mihomo text format (Domain Set).
// Specifically: DOMAIN-SUFFIX,example.com -> +.example.com
// DOMAIN,example.com -> example.com
func GenerateMihomoText(lines []string) []string {
	var result []string
	for _, line := range lines {
		rType, val := ParseRuleType(line)
		if rType == "" || val == "" {
			continue
		}
		switch rType {
		case "DOMAIN-SUFFIX":
			// User requested format: '+.docker.com'
			// Ensure we don't double dot if val already has one
			if strings.HasPrefix(val, ".") {
				result = append(result, "+"+val)
			} else {
				result = append(result, "+."+val)
			}
		case "DOMAIN":
			// User requested format: 'api.my-ip.io' (no prefix)
			result = append(result, val)
		}
	}
	return result
}

// GenerateMihomoClassical transforms lines for Mihomo classical format (Mixed).
// Reconstructs TYPE,VALUE lines.
func GenerateMihomoClassical(lines []string) []string {
	var result []string
	for _, line := range lines {
		rType, val := ParseRuleType(line)
		if rType == "" || val == "" {
			continue
		}
		// Filter out unknown or unsupported types if needed?
		// But mostly just pass through.
		// Construct Surge/Mihomo compatible line.
		result = append(result, fmt.Sprintf("%s,%s", rType, val))
	}
	return result
}

// GenerateSingBoxIPRules parses implicit or explicit IP-CIDR lines.
// Consolidates all CIDRs into a single rule.
func GenerateSingBoxIPRules(lines []string) []option.DefaultHeadlessRule {
	var cidrs []string
	for _, line := range lines {
		rType, val := ParseRuleType(line)
		if rType == "" || val == "" {
			continue
		}
		switch rType {
		case "IP-CIDR", "IP-CIDR6":
			cidrs = append(cidrs, val)
		}
	}

	if len(cidrs) == 0 {
		return nil
	}
	return []option.DefaultHeadlessRule{{
		IPCIDR: cidrs,
	}}
}

// Deduplicate removes duplicate lines while preserving order.
func Deduplicate(lines []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, line := range lines {
		if !seen[line] {
			seen[line] = true
			result = append(result, line)
		}
	}
	return result
}

// GenerateIPList extracts IP-CIDR/IP-CIDR6 values.
// Returns just the value (e.g. 1.2.3.4/24).
func GenerateIPList(lines []string) []string {
	var result []string
	for _, line := range lines {
		rType, val := ParseRuleType(line)
		if rType == "" || val == "" {
			continue
		}
		switch rType {
		case "IP-CIDR", "IP-CIDR6":
			result = append(result, val)
		}
	}
	return result
}

// GenerateYAMLPayload generates a YAML payload format string from a list of lines.
// Format:
//
//	payload:
//	  - 'line1'
//	  - 'line2'
func GenerateYAMLPayload(lines []string) string {
	var builder strings.Builder
	builder.WriteString("payload:\n")
	for _, line := range lines {
		builder.WriteString(fmt.Sprintf("  - '%s'\n", line))
	}
	return builder.String()
}
