package main

import (
	"os"

	"github.com/sagernet/sing-box/common/srs"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
)

type DefaultHeadlessRule = option.DefaultHeadlessRule

const RuleSetVersion = C.RuleSetVersion2

func SaveSingRuleSet(rules []DefaultHeadlessRule, outputPath string) error {
	plainRuleSet := option.PlainRuleSetCompat{
		Version: RuleSetVersion,
		Options: option.PlainRuleSet{
			Rules: common.Map(rules, func(it option.DefaultHeadlessRule) option.HeadlessRule {
				return option.HeadlessRule{
					Type:           C.RuleTypeDefault,
					DefaultOptions: it,
				}
			}),
		},
	}
	// Only generate .srs binary file
	if err := saveSingBinaryRuleSet(&plainRuleSet, outputPath+".srs"); err != nil {
		return err
	}
	return nil
}

func saveSingBinaryRuleSet(ruleset *option.PlainRuleSetCompat, outputPath string) error {
	ruleSet, err := ruleset.Upgrade()
	if err != nil {
		return err
	}
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	err = srs.Write(outputFile, ruleSet, RuleSetVersion)
	if err != nil {
		outputFile.Close()
		os.Remove(outputPath)
		return err
	}
	outputFile.Close()
	return nil
}
