package evaluator

import (
	rules "bloodhound/lib/rule"
	utils "bloodhound/lib/utils"

	log "github.com/sirupsen/logrus"
)

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

		if utils.ContainsAny(url, rule.Content.Matches) {
			score += rule.Value
		}
	}

	return score
}
