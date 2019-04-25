package cli

import (
	"errors"
	"flag"
	"log"
	"strings"
)

type Flag struct {
	Key         string
	Default     string
	Description string
	Validator   func(string) bool
	Value       *string
}

func ParseFlags(cliFlags []*Flag) map[string]string {
	parsedFlags, err := doParse(cliFlags)
	return HandleParseFlags(parsedFlags, err)
}

func doParse(cliFlags []*Flag) (map[string]string, error) {
	parsedValues := make(map[string]string)
	for _, cliFlag := range cliFlags {
		cliFlag.Value = flag.String(cliFlag.Key, cliFlag.Default, cliFlag.Description)
	}
	flag.Parse()
	errorFound := false
	for _, cliFlag := range cliFlags {
		if *cliFlag.Value == "" {
			errorFound = true
			log.Printf("Missing flag \033[32m-%s\033[97m, must be \033[33m%s\033[97m", cliFlag.Key, cliFlag.Description)
		} else if !cliFlag.Validator(*cliFlag.Value) {
			errorFound = true
			log.Printf("Invalid \033[32m-%s\033[97m: \033[31m%s\033[97m, must be \033[33m%s\033[97m", cliFlag.Key, *cliFlag.Value, cliFlag.Description)
		} else {
			parsedValues[cliFlag.Key] = strings.TrimRight(*cliFlag.Value, "/")
		}
	}

	if errorFound {
		return nil, errors.New("Error parsing CLI flags, see log.")
	} else {
		return parsedValues, nil
	}
}

func HandleParseFlags(flags map[string]string, err error) map[string]string {
	if err != nil {
		panic(err.Error())
	} else if flags == nil || len(flags) == 0 {
		panic("Something's gone horribly wrong, parsed CLI flags are empty!")
	}
	return flags
}
