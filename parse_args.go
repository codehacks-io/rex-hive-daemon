package main

import (
	"fmt"
	"regexp"
	"rex-daemon/rexregexp"
	"strconv"
)

var getUniqueInSequenceRegex = regexp.MustCompile(`{unique-in-sequence:(?P<from>\d+)-(?P<to>\d+)}`)

func getDynamicArgsOrPanic(originalArgs []string, used *map[int]bool) []string {

	replacedArgs := make([]string, len(originalArgs))
	copy(replacedArgs, originalArgs)

	for i, arg := range originalArgs {
		matches := rexregexp.MatchNamedCapturingGroups(&arg, getUniqueInSequenceRegex)
		if len(matches["from"]) > 0 && len(matches["to"]) > 0 {
			from, _ := strconv.Atoi(matches["from"])
			to, _ := strconv.Atoi(matches["to"])

			// Swap is from is greater than to
			if from > to {
				oldTo := to
				to = from
				from = oldTo
			}

			didAssign := false
			for seq := from; seq <= to; seq++ {
				if !(*used)[seq] {
					replacedArgs[i] = strconv.Itoa(seq)
					(*used)[seq] = true
					didAssign = true
					break
				}
			}
			if !didAssign {
				panic(fmt.Sprintf("dynamic argument %s cannot be allocated a value, all values in the sequense have been reserved", arg))
			}
		}
	}

	return replacedArgs
}
