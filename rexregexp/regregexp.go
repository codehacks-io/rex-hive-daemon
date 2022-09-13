package rexregexp

import "regexp"

func MatchNamedCapturingGroups(text *string, r *regexp.Regexp) map[string]string {
	match := r.FindStringSubmatch(*text)

	result := make(map[string]string)
	for ii, name := range r.SubexpNames() {
		if ii != 0 && name != "" && len(match) > ii {
			result[name] = match[ii]
		}
	}
	return result
}
