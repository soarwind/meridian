package main

import (
	"fmt"
	"os"

	P "github.com/metacubex/mihomo/constant/provider"
	RP "github.com/metacubex/mihomo/rules/provider"
)

func SaveMetaRuleSet(buf []byte, b string, f string, outputPath string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in SaveMetaRuleSet: %v", r)
		}
	}()

	behavior, err := P.ParseBehavior(b)
	if err != nil {
		return err
	}
	format, err := P.ParseRuleFormat(f)
	if err != nil {
		return err
	}
	targetFile, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer targetFile.Close() // Ensure close even if panic occurs (though recover handles return)

	err = RP.ConvertToMrs(buf, behavior, format, targetFile)
	return err
}
