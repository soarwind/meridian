package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// Configuration for Sources
	GFWURL     = "https://raw.githubusercontent.com/Loyalsoldier/surge-rules/release/proxy.txt"
	GFWLiteURL = "https://raw.githubusercontent.com/Loyalsoldier/surge-rules/release/gfw.txt"
	IPCIDRURL  = "https://raw.githubusercontent.com/Loyalsoldier/surge-rules/release/telegramcidr.txt"

	// Local Data to be merged
	LocalDomainFile = "domain.txt"
	LocalIPFile     = "ipcidr.txt"

	OutputDir = "release"
)

func main() {
	// 0. Prepare Output Directory
	if err := os.MkdirAll(OutputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// 1. Fetch Local Data
	fmt.Printf("Fetching Local Domain list from %s...\n", LocalDomainFile)
	localDomainLines, err := FetchList(LocalDomainFile)
	if err != nil {
		fmt.Printf("Warning: Error fetching local domain list: %v\n", err)
	}
	parsedLocalDomains := ParseList(localDomainLines)
	fmt.Printf("Loaded %d local domain lines.\n", len(parsedLocalDomains))

	fmt.Printf("Fetching Local IP list from %s...\n", LocalIPFile)
	localIPLines, err := FetchList(LocalIPFile)
	if err != nil {
		fmt.Printf("Warning: Error fetching local IP list: %v\n", err)
	}
	parsedLocalIPs := ParseList(localIPLines)
	fmt.Printf("Loaded %d local IP lines.\n", len(parsedLocalIPs))

	// 2. Fetch GFW domains
	fmt.Printf("\n--- Processing GFW domains ---\n")
	fmt.Printf("Fetching remote GFW list from %s...\n", GFWURL)
	gfwRemoteLines, err := FetchList(GFWURL)
	if err != nil {
		fmt.Printf("Error fetching GFW list: %v\n", err)
		os.Exit(1)
	}
	parsedGFWRemote := ParseList(gfwRemoteLines)
	fmt.Printf("Fetched %d remote GFW lines.\n", len(parsedGFWRemote))

	// Merge GFW + local domains, deduplicate → merged_gfw cache
	var mergedGFW []string
	mergedGFW = append(mergedGFW, parsedGFWRemote...)
	mergedGFW = append(mergedGFW, parsedLocalDomains...)
	mergedGFW = Deduplicate(mergedGFW)
	fmt.Printf("Created merged_gfw cache: %d lines.\n", len(mergedGFW))

	// 3. Fetch GFW-Lite domains
	fmt.Printf("\n--- Processing GFW-Lite domains ---\n")
	fmt.Printf("Fetching remote GFW-Lite list from %s...\n", GFWLiteURL)
	gfwLiteRemoteLines, err := FetchList(GFWLiteURL)
	if err != nil {
		fmt.Printf("Error fetching GFW-Lite list: %v\n", err)
		os.Exit(1)
	}
	parsedGFWLiteRemote := ParseList(gfwLiteRemoteLines)
	fmt.Printf("Fetched %d remote GFW-Lite lines.\n", len(parsedGFWLiteRemote))

	// Merge GFW-Lite + local domains, deduplicate → merged_gfw-lite cache
	var mergedGFWLite []string
	mergedGFWLite = append(mergedGFWLite, parsedGFWLiteRemote...)
	mergedGFWLite = append(mergedGFWLite, parsedLocalDomains...)
	mergedGFWLite = Deduplicate(mergedGFWLite)
	fmt.Printf("Created merged_gfw-lite cache: %d lines.\n", len(mergedGFWLite))

	// 4. Fetch Remote IPs
	fmt.Printf("\n--- Processing Remote IPs ---\n")
	fmt.Printf("Fetching Remote IP list from %s...\n", IPCIDRURL)
	remoteIPLines, err := FetchList(IPCIDRURL)
	if err != nil {
		fmt.Printf("Error fetching Remote IP list: %v\n", err)
		os.Exit(1)
	}
	parsedRemoteIPs := ParseList(remoteIPLines)
	fmt.Printf("Fetched %d remote IP lines.\n", len(parsedRemoteIPs))

	// Merge Remote IPs + local IPs, deduplicate
	var mergedIPs []string
	mergedIPs = append(mergedIPs, parsedRemoteIPs...)
	mergedIPs = append(mergedIPs, parsedLocalIPs...)
	mergedIPs = Deduplicate(mergedIPs)
	fmt.Printf("Created merged IPs: %d lines.\n", len(mergedIPs))

	// 5. Generate plain text files from merged data
	fmt.Println("\n--- Generating plain text files ---")

	// 5.1 gfw.txt
	fmt.Println("Generating gfw.txt...")
	gfwTxtPath := filepath.Join(OutputDir, "gfw.txt")
	gfwTxtContent := strings.Join(mergedGFW, "\n")
	if err := os.WriteFile(gfwTxtPath, []byte(gfwTxtContent), 0644); err != nil {
		fmt.Printf("Error saving gfw.txt: %v\n", err)
	} else {
		fmt.Printf("Saved gfw.txt (%d lines)\n", len(mergedGFW))
	}

	// 5.2 gfw-lite.txt
	fmt.Println("Generating gfw-lite.txt...")
	gfwLiteTxtPath := filepath.Join(OutputDir, "gfw-lite.txt")
	gfwLiteTxtContent := strings.Join(mergedGFWLite, "\n")
	if err := os.WriteFile(gfwLiteTxtPath, []byte(gfwLiteTxtContent), 0644); err != nil {
		fmt.Printf("Error saving gfw-lite.txt: %v\n", err)
	} else {
		fmt.Printf("Saved gfw-lite.txt (%d lines)\n", len(mergedGFWLite))
	}

	// 5.3 ipcidr.txt
	fmt.Println("Generating ipcidr.txt...")
	ipcidrTxtPath := filepath.Join(OutputDir, "ipcidr.txt")
	ipcidrTxtContent := strings.Join(mergedIPs, "\n")
	if err := os.WriteFile(ipcidrTxtPath, []byte(ipcidrTxtContent), 0644); err != nil {
		fmt.Printf("Error saving ipcidr.txt: %v\n", err)
	} else {
		fmt.Printf("Saved ipcidr.txt (%d lines)\n", len(mergedIPs))
	}

	// 6. Generate ipcidr.yaml (IP list in payload format)
	fmt.Println("\nGenerating ipcidr.yaml...")
	ipLines := GenerateIPList(mergedIPs)
	ipYAML := GenerateYAMLPayload(ipLines)
	ipPath := filepath.Join(OutputDir, "ipcidr.yaml")
	if err := os.WriteFile(ipPath, []byte(ipYAML), 0644); err != nil {
		fmt.Printf("Error saving ipcidr.yaml: %v\n", err)
	} else {
		fmt.Println("Saved ipcidr.yaml")
	}

	var generatedFiles []string
	// Add plain text files to the list
	generatedFiles = append(generatedFiles, "gfw.txt", "gfw-lite.txt", "ipcidr.txt")

	// 7. Generate GFW outputs
	fmt.Println("\n--- Generating GFW outputs ---")

	// 7.1 gfw.yaml (domains with +. prefix in payload format)
	fmt.Println("Generating gfw.yaml...")
	gfwMihomoLines := GenerateMihomoText(mergedGFW)
	gfwYAML := GenerateYAMLPayload(gfwMihomoLines)
	gfwYAMLPath := filepath.Join(OutputDir, "gfw.yaml")
	if err := os.WriteFile(gfwYAMLPath, []byte(gfwYAML), 0644); err != nil {
		fmt.Printf("Error saving gfw.yaml: %v\n", err)
	} else {
		fmt.Println("Saved gfw.yaml")
		generatedFiles = append(generatedFiles, "gfw.yaml")
	}

	// 7.2 gfw.mrs (binary format, reuse gfwYAML)
	fmt.Println("Generating gfw.mrs...")
	gfwMRSPath := filepath.Join(OutputDir, "gfw.mrs")
	if err := SaveMetaRuleSet([]byte(gfwYAML), "domain", "yaml", gfwMRSPath); err != nil {
		fmt.Printf("Error saving gfw.mrs: %v\n", err)
	} else {
		fmt.Println("Saved gfw.mrs")
		generatedFiles = append(generatedFiles, "gfw.mrs")
	}

	// 7.3 gfw.srs (merged_gfw + merged IPs)
	fmt.Println("Generating gfw.srs...")
	var gfwSRSCombined []string
	gfwSRSCombined = append(gfwSRSCombined, mergedGFW...)
	gfwSRSCombined = append(gfwSRSCombined, mergedIPs...)
	gfwSRSCombined = Deduplicate(gfwSRSCombined)

	gfwSRSRules := GenerateSingBoxRules(gfwSRSCombined)
	gfwSRSIPRules := GenerateSingBoxIPRules(gfwSRSCombined)
	if len(gfwSRSIPRules) > 0 {
		gfwSRSRules = append(gfwSRSRules, gfwSRSIPRules...)
	}

	gfwSRSPath := filepath.Join(OutputDir, "gfw")
	if err := SaveSingRuleSet(gfwSRSRules, gfwSRSPath); err != nil {
		fmt.Printf("Error saving gfw.srs: %v\n", err)
	} else {
		fmt.Println("Saved gfw.srs")
		generatedFiles = append(generatedFiles, "gfw.srs")
	}

	// 7.4 gfw-domains.srs (domains from the merged GFW rule set only)
	fmt.Println("Generating gfw-domains.srs...")
	gfwDomainsSRSRules := GenerateSingBoxRules(mergedGFW)
	gfwDomainsSRSPath := filepath.Join(OutputDir, "gfw-domains")
	if err := SaveSingRuleSet(gfwDomainsSRSRules, gfwDomainsSRSPath); err != nil {
		fmt.Printf("Error saving gfw-domains.srs: %v\n", err)
	} else {
		fmt.Println("Saved gfw-domains.srs")
		generatedFiles = append(generatedFiles, "gfw-domains.srs")
	}

	// 8. Generate GFW-Lite outputs
	fmt.Println("\n--- Generating GFW-Lite outputs ---")

	// 8.1 gfw-lite.yaml (domains with +. prefix in payload format)
	fmt.Println("Generating gfw-lite.yaml...")
	gfwLiteMihomoLines := GenerateMihomoText(mergedGFWLite)
	gfwLiteYAML := GenerateYAMLPayload(gfwLiteMihomoLines)
	gfwLiteYAMLPath := filepath.Join(OutputDir, "gfw-lite.yaml")
	if err := os.WriteFile(gfwLiteYAMLPath, []byte(gfwLiteYAML), 0644); err != nil {
		fmt.Printf("Error saving gfw-lite.yaml: %v\n", err)
	} else {
		fmt.Println("Saved gfw-lite.yaml")
		generatedFiles = append(generatedFiles, "gfw-lite.yaml")
	}

	// 8.2 gfw-lite.mrs (binary format, reuse gfwLiteYAML)
	fmt.Println("Generating gfw-lite.mrs...")
	gfwLiteMRSPath := filepath.Join(OutputDir, "gfw-lite.mrs")
	if err := SaveMetaRuleSet([]byte(gfwLiteYAML), "domain", "yaml", gfwLiteMRSPath); err != nil {
		fmt.Printf("Error saving gfw-lite.mrs: %v\n", err)
	} else {
		fmt.Println("Saved gfw-lite.mrs")
		generatedFiles = append(generatedFiles, "gfw-lite.mrs")
	}

	// 8.3 gfw-lite.srs (merged_gfw-lite + merged IPs)
	fmt.Println("Generating gfw-lite.srs...")
	var gfwLiteSRSCombined []string
	gfwLiteSRSCombined = append(gfwLiteSRSCombined, mergedGFWLite...)
	gfwLiteSRSCombined = append(gfwLiteSRSCombined, mergedIPs...)
	gfwLiteSRSCombined = Deduplicate(gfwLiteSRSCombined)

	gfwLiteSRSRules := GenerateSingBoxRules(gfwLiteSRSCombined)
	gfwLiteSRSIPRules := GenerateSingBoxIPRules(gfwLiteSRSCombined)
	if len(gfwLiteSRSIPRules) > 0 {
		gfwLiteSRSRules = append(gfwLiteSRSRules, gfwLiteSRSIPRules...)
	}

	gfwLiteSRSPath := filepath.Join(OutputDir, "gfw-lite")
	if err := SaveSingRuleSet(gfwLiteSRSRules, gfwLiteSRSPath); err != nil {
		fmt.Printf("Error saving gfw-lite.srs: %v\n", err)
	} else {
		fmt.Println("Saved gfw-lite.srs")
		generatedFiles = append(generatedFiles, "gfw-lite.srs")
	}

	// 9. Add ipcidr.yaml to file list
	generatedFiles = append(generatedFiles, "ipcidr.yaml")

	// 10. Generate File List
	listPath := filepath.Join(OutputDir, "list.txt")
	listContent := strings.Join(generatedFiles, "\n")
	if err := os.WriteFile(listPath, []byte(listContent), 0644); err != nil {
		fmt.Printf("Error saving list.txt: %v\n", err)
	} else {
		fmt.Println("\nSaved release/list.txt.")
	}

	fmt.Println("\nAll tasks completed.")
}
