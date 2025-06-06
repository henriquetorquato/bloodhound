package evaluator

import (
	rules "bloodhound/lib/rule"
	"strings"

	log "github.com/sirupsen/logrus"
)

type UrlEvaluation struct {
	url      string
	absolute bool
	score    int
}

func EvaluateUrl(url string, ruleset *rules.Ruleset) int {
	score := 0

	for _, rule := range ruleset.Scores {
		if rule.Level != rules.ResourceLevel {
			log.WithFields(log.Fields{
				"url":  url,
				"rule": rule.Name,
			}).Trace("Skipping rule while evaluating URL")

			continue
		}

		if len(rule.Content.Matches) == 0 {
			log.WithFields(log.Fields{
				"url":  url,
				"rule": rule.Name,
			}).Trace("Rule does not apply: Match list is empty")

			continue
		}

		for _, word := range rule.Content.Matches {
			if !strings.Contains(url, word) {
				continue
			}

			score += rule.Value
			break
		}
	}

	return score
}
