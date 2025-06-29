package coreutils

import (
	"strconv"
	"strings"

	"github.com/forPelevin/gomoji"
	"github.com/jfrog/jfrog-client-go/utils/log"

	"github.com/gookit/color"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

// Removes the provided flag and value from the command arguments
func RemoveFlagFromCommand(args *[]string, flagIndex, flagValueIndex int) {
	argsCopy := *args
	// Remove flag from Command if required.
	if flagIndex != -1 {
		argsCopy = append(argsCopy[:flagIndex], argsCopy[flagValueIndex+1:]...)
	}
	*args = argsCopy
}

// Find value of required CLI flag in Command.
// If flag does not exist, the returned index is -1 and nil is returned as the error.
// Return values:
// err - error if flag exists but failed to extract its value.
// flagIndex - index of flagName in Command.
// flagValueIndex - index in Command in which the value of the flag exists.
// flagValue - value of flagName.
func FindFlag(flagName string, args []string) (flagIndex, flagValueIndex int, flagValue string, err error) {
	flagIndex = -1
	flagValueIndex = -1
	for index, arg := range args {
		// Check current argument.
		if !strings.HasPrefix(arg, flagName) {
			continue
		}

		// Get flag value.
		flagValue, flagValueIndex, err = getFlagValueAndValueIndex(flagName, args, index)
		if err != nil {
			return
		}

		// If was not the correct flag, continue looking.
		if flagValueIndex == -1 {
			continue
		}

		// Return value.
		flagIndex = index
		return
	}

	// Flag not found.
	return
}

// Get the provided flag's value, and the index of the value.
// Value-index can either be same as flag's index, or the next one.
// Return error if flag is found, but couldn't extract value.
// If the provided index doesn't contain the searched flag, return flagIndex = -1.
func getFlagValueAndValueIndex(flagName string, args []string, flagIndex int) (flagValue string, flagValueIndex int, err error) {
	indexValue := args[flagIndex]

	// Check if flag is in form '--key=value'
	indexValue = strings.TrimPrefix(indexValue, flagName)
	if strings.HasPrefix(indexValue, "=") {
		if len(indexValue) > 1 {
			return indexValue[1:], flagIndex, nil
		}
		return "", -1, errorutils.CheckErrorf("Flag %s is provided with empty value.", flagName)
	}

	// Check if it is a different flag with same prefix, e.g --server-id-another
	if len(indexValue) > 0 {
		return "", -1, nil
	}

	// If reached here, expect the flag value in next argument.
	if len(args) < flagIndex+2 {
		// Flag value does not exist.
		return "", -1, errorutils.CheckErrorf("Failed extracting value of provided flag: %s.", flagName)
	}

	nextIndexValue := args[flagIndex+1]
	// Don't allow next value to be a flag.
	if strings.HasPrefix(nextIndexValue, "-") {
		// Flag value does not exist.
		return "", -1, errorutils.CheckErrorf("Failed extracting value of provided flag: %s.", flagName)
	}

	return nextIndexValue, flagIndex + 1, nil
}

// Boolean flag can be provided in one of the following forms:
// 1. --flag=value, where value can be true/false
// 2. --flag, here the value is true
// Return values:
// flagIndex - index of flagName in args.
// flagValue - value of flagName.
// err - error if flag exists, but we failed to extract its value.
// If flag does not exist flagIndex = -1 with false value and nil error.
func FindBooleanFlag(flagName string, args []string) (flagIndex int, flagValue bool, err error) {
	var arg string
	for flagIndex, arg = range args {
		if strings.HasPrefix(arg, flagName) {
			value := strings.TrimPrefix(arg, flagName)
			switch {
			case len(value) == 0:
				flagValue = true
			case strings.HasPrefix(value, "="):
				flagValue, err = strconv.ParseBool(value[1:])
			default:
				continue
			}
			return
		}
	}
	return -1, false, nil
}

// Find the first match of the provided flags in args.
// Return same values as FindFlag.
func FindFlagFirstMatch(flags, args []string) (flagIndex, flagValueIndex int, flagValue string, err error) {
	// Look for provided flags.
	for _, flag := range flags {
		flagIndex, flagValueIndex, flagValue, err = FindFlag(flag, args)
		if err != nil {
			return
		}
		if flagIndex != -1 {
			// Found value for flag.
			return
		}
	}
	return
}

func ExtractServerIdFromCommand(args []string) (cleanArgs []string, serverId string, err error) {
	return ExtractStringOptionFromArgs(args, "server-id")
}

func ExtractThreadsFromArgs(args []string, defaultValue int) (cleanArgs []string, threads int, err error) {
	cleanArgs = append([]string(nil), args...)
	threads = defaultValue
	// Extract threads information from the args.
	flagIndex, valueIndex, numOfThreads, err := FindFlag("--threads", cleanArgs)
	if err != nil {
		return
	}

	RemoveFlagFromCommand(&cleanArgs, flagIndex, valueIndex)
	if numOfThreads != "" {
		threads, err = strconv.Atoi(numOfThreads)
		if err != nil {
			err = errorutils.CheckError(err)
		}
	}
	return
}

func ExtractInsecureTlsFromArgs(args []string) (cleanArgs []string, insecureTls bool, err error) {
	return extractBoolOptionFromArgs(args, "insecure-tls")
}

// Used by docker
func ExtractSkipLoginFromArgs(args []string) (cleanArgs []string, skipLogin bool, err error) {
	return extractBoolOptionFromArgs(args, "skip-login")
}

// Used by docker
func ExtractBoolFlagFromArgs(args []string, flagName string) (cleanArgs []string, flagValue bool, err error) {
	return extractBoolOptionFromArgs(args, flagName)
}

// Used by docker
func ExtractFailFromArgs(args []string) (cleanArgs []string, fail bool, err error) {
	return extractBoolOptionFromArgs(args, "fail")
}

// Used by docker  scan (Xray)
func ExtractLicensesFromArgs(args []string) (cleanArgs []string, licenses bool, err error) {
	return extractBoolOptionFromArgs(args, "licenses")
}

// Used by docker scan (Xray)
func ExtractRepoPathFromArgs(args []string) (cleanArgs []string, repoPath string, err error) {
	return ExtractStringOptionFromArgs(args, "repo-path")
}

// Used by docker scan (Xray)
func ExtractWatchesFromArgs(args []string) (cleanArgs []string, watches string, err error) {
	return ExtractStringOptionFromArgs(args, "watches")
}

func ExtractDetailedSummaryFromArgs(args []string) (cleanArgs []string, detailedSummary bool, err error) {
	return extractBoolOptionFromArgs(args, "detailed-summary")
}

func ExtractXrayScanFromArgs(args []string) (cleanArgs []string, xrayScan bool, err error) {
	return extractBoolOptionFromArgs(args, "scan")
}

func ExtractXrayOutputFormatFromArgs(args []string) (cleanArgs []string, format string, err error) {
	return ExtractStringOptionFromArgs(args, "format")
}

func ExtractTagFromArgs(args []string) (cleanArgs []string, tag string, err error) {
	return ExtractStringOptionFromArgs(args, "tag")
}

func ExtractStringOptionFromArgs(args []string, optionName string) (cleanArgs []string, value string, err error) {
	cleanArgs = append([]string(nil), args...)

	flagIndex, valIndex, value, err := FindFlag("--"+optionName, cleanArgs)
	if err != nil {
		return
	}
	RemoveFlagFromCommand(&cleanArgs, flagIndex, valIndex)
	return
}

func ExtractUseNativeFromArgs(args []string) (cleanArgs []string, useNpmrc bool, err error) {
	return extractBoolOptionFromArgs(args, "run-native")
}

func extractBoolOptionFromArgs(args []string, optionName string) (cleanArgs []string, value bool, err error) {
	cleanArgs = append([]string(nil), args...)

	flagIndex, value, err := FindBooleanFlag("--"+optionName, cleanArgs)
	if err != nil {
		return
	}
	// Since boolean flag might appear as --flag or --flag=value, the value index is the same as the flag index.
	RemoveFlagFromCommand(&cleanArgs, flagIndex, flagIndex)
	return
}

// Add green color style to the string if possible.
func PrintTitle(str string) string {
	return PrintGreen(str)
}

// Add cyan color style to the string if possible.
func PrintLink(str string) string {
	return colorStr(str, color.Cyan)
}

// Add bold style to the string if possible.
func PrintBold(str string) string {
	return colorStr(str, color.Bold)
}

// Add bold and green style to the string if possible.
func PrintBoldTitle(str string) string {
	return PrintBold(PrintTitle(str))
}

// Add gray color style to the string if possible.
func PrintComment(str string) string {
	return colorStr(str, color.Gray)
}

// Add red color style to the string if possible.
func PrintRed(str string) string {
	return colorStr(str, color.Red)
}

// Add yellow color style to the string if possible.
func PrintYellow(str string) string {
	return colorStr(str, color.Yellow)
}

// Add green color style to the string if possible.
func PrintGreen(str string) string {
	return colorStr(str, color.Green)
}

// Add the requested style to the string if possible.
func colorStr(str string, c color.Color) string {
	// Add styles only on supported terminals
	if log.IsStdOutTerminal() && log.IsColorsSupported() {
		return c.Render(str)
	}
	// Remove emojis from non-supported terminals
	if gomoji.ContainsEmoji(str) {
		str = gomoji.RemoveEmojis(str)
	}
	return str
}

// Remove emojis from non-supported terminals
func RemoveEmojisIfNonSupportedTerminal(msg string) string {
	if !(log.IsStdOutTerminal() && log.IsColorsSupported()) {
		if gomoji.ContainsEmoji(msg) {
			msg = gomoji.RemoveEmojis(msg)
		}
	}
	return msg
}
