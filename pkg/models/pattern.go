package models

import "regexp"

// MatchPattern reports whether value satisfies an access rule's pattern. The pattern is a regular
// expression that must match value in full, so the rule "staging" selects the device named
// "staging" and neither "notstaging" nor "staging-db"; an empty pattern imposes no restriction and
// matches any value. It returns an error when the pattern does not compile.
//
// It is the shared matcher for every administrator-written selector that decides access — device
// filters, public-key username restrictions and firewall rules — so that all of them read the way
// their author wrote them.
func MatchPattern(pattern, value string) (bool, error) {
	if pattern == "" {
		return true, nil
	}

	return regexp.MatchString(`\A(?:`+pattern+`)\z`, value)
}
