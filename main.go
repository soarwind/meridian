package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// Configuration for Sources
	GFWURL        = "https://raw.githubusercontent.com/Loyalsoldier/surge-rules/release/proxy.txt"
	GFWLiteURL    = "https://raw.githubusercontent.com/Loyalsoldier/surge-rules/release/gfw.txt"
	TelegramIPURL = "https://raw.githubusercontent.com/Loyalsoldier/surge-rules/release/telegramcidr.txt"

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

	// 4. Fetch Telegram IPs
	fmt.Printf("\n--- Processing Telegram IPs ---\n")
	fmt.Printf("Fetching Telegram IP list from %s...\n", TelegramIPURL)
	telegramRemoteLines, err := FetchList(TelegramIPURL)
	if err != nil {
		fmt.Printf("Error fetching Telegram IP list: %v\n", err)
		os.Exit(1)
	}
	parsedTelegramIPs := ParseList(telegramRemoteLines)
	fmt.Printf("Fetched %d remote Telegram IP lines.\n", len(parsedTelegramIPs))

	// Merge Telegram IPs + local IPs, deduplicate
	var mergedTelegramIPs []string
	mergedTelegramIPs = append(mergedTelegramIPs, parsedTelegramIPs...)
	mergedTelegramIPs = append(mergedTelegramIPs, parsedLocalIPs...)
	mergedTelegramIPs = Deduplicate(mergedTelegramIPs)
	fmt.Printf("Created merged Telegram IPs: %d lines.\n", len(mergedTelegramIPs))

	// 5. Generate telegramip.yaml (IP list in payload format)
	fmt.Println("\nGenerating telegramip.yaml...")
	telegramIPLines := GenerateIPList(mergedTelegramIPs)
	telegramIPYAML := GenerateYAMLPayload(telegramIPLines)
	telegramIPPath := filepath.Join(OutputDir, "telegramip.yaml")
	if err := os.WriteFile(telegramIPPath, []byte(telegramIPYAML), 0644); err != nil {
		fmt.Printf("Error saving telegramip.yaml: %v\n", err)
	} else {
		fmt.Println("Saved telegramip.yaml")
	}

	var generatedFiles []string

	// 6. Generate GFW outputs
	fmt.Println("\n--- Generating GFW outputs ---")

	// 6.1 gfw.yaml (domains with +. prefix in payload format)
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

	// 6.2 gfw.mrs (binary format, reuse gfwYAML)
	fmt.Println("Generating gfw.mrs...")
	gfwMRSPath := filepath.Join(OutputDir, "gfw.mrs")
	if err := SaveMetaRuleSet([]byte(gfwYAML), "domain", "yaml", gfwMRSPath); err != nil {
		fmt.Printf("Error saving gfw.mrs: %v\n", err)
	} else {
		fmt.Println("Saved gfw.mrs")
		generatedFiles = append(generatedFiles, "gfw.mrs")
	}

	// 6.3 gfw.srs (merged_gfw + telegramip.txt)
	fmt.Println("Generating gfw.srs...")
	var gfwSRSCombined []string
	gfwSRSCombined = append(gfwSRSCombined, mergedGFW...)
	gfwSRSCombined = append(gfwSRSCombined, mergedTelegramIPs...)
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

	// 7. Generate GFW-Lite outputs
	fmt.Println("\n--- Generating GFW-Lite outputs ---")

	// 7.1 gfw-lite.yaml (domains with +. prefix in payload format)
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

	// 7.2 gfw-lite.mrs (binary format, reuse gfwLiteYAML)
	fmt.Println("Generating gfw-lite.mrs...")
	gfwLiteMRSPath := filepath.Join(OutputDir, "gfw-lite.mrs")
	if err := SaveMetaRuleSet([]byte(gfwLiteYAML), "domain", "yaml", gfwLiteMRSPath); err != nil {
		fmt.Printf("Error saving gfw-lite.mrs: %v\n", err)
	} else {
		fmt.Println("Saved gfw-lite.mrs")
		generatedFiles = append(generatedFiles, "gfw-lite.mrs")
	}

	// 7.3 gfw-lite.srs (merged_gfw-lite + telegramip.txt)
	fmt.Println("Generating gfw-lite.srs...")
	var gfwLiteSRSCombined []string
	gfwLiteSRSCombined = append(gfwLiteSRSCombined, mergedGFWLite...)
	gfwLiteSRSCombined = append(gfwLiteSRSCombined, mergedTelegramIPs...)
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

	// 8. Add telegramip.yaml to file list
	generatedFiles = append(generatedFiles, "telegramip.yaml")

	// 9. Generate File List
	listPath := filepath.Join(OutputDir, "list.txt")
	listContent := strings.Join(generatedFiles, "\n")
	if err := os.WriteFile(listPath, []byte(listContent), 0644); err != nil {
		fmt.Printf("Error saving list.txt: %v\n", err)
	} else {
		fmt.Println("\nSaved release/list.txt.")
	}

	fmt.Println("\nAll tasks completed.")
}
