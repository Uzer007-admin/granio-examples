package lang

import (
	"github.com/jeandeaual/go-locale"
	"golang.org/x/text/language"

	"github.com/Uzer007-admin/granio/v2"
)

// SystemLocale returns the primary locale on the current system.
// This may refer to a language that Fyne does not have translations for.
func SystemLocale() granio.Locale {
	loc, err := locale.GetLocale()
	if err != nil {
		granio.LogError("Failed to look up user locale", err)
	}
	if len(loc) < 2 {
		loc = "en"
	}

	tag, err := language.Parse(loc)
	if err != nil {
		granio.LogError("Error parsing user locale "+loc, err)
	}
	return localeFromTag(tag)
}

func closestSupportedLocale(locs []string) granio.Locale {
	matcher := language.NewMatcher(translated)

	tags := make([]language.Tag, len(locs))
	for i, loc := range locs {
		tag, err := language.Parse(loc)
		if err != nil && loc != "C" {
			granio.LogError("Error parsing user locale "+loc, err)
		}
		tags[i] = tag
	}
	best, _, conf := matcher.Match(tags...)
	// When confidence is No the matcher may pick a language that shares only a script
	// (for example Serbian Cyrillic resolving to Russian). Prefer the default fallback
	// in that case rather than presenting an unrelated language.
	if conf == language.No && len(translated) > 0 {
		best = translated[0]
	}
	return localeFromTag(best)
}

func localeFromTag(in language.Tag) granio.Locale {
	b, s, r := in.Raw()
	ret := b.String()

	if r.String() != "ZZ" {
		ret += "-" + r.String()

		if s.String() != "Zzzz" {
			ret += "-" + s.String()
		}
	}

	return granio.Locale(ret)
}
