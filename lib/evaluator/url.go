package evaluator

import (
	"bloodhound/lib/rules"
	"bloodhound/lib/utils"
)

func EvaluateUrl(url *string, ruleList *[]rules.Rule) EvaluationResult {
	score := 0

	for _, rule := range *ruleList {
		if rule.Level != rules.ResourceLevel {
			continue
		}

		if len(rule.Content.Matches) == 0 {
			continue
		}

		if utils.ContainsAny(*url, rule.Content.Matches) {
			if rule.Remove {
				return NewEvaluationResult(0, rule.Remove)
			}

			score += rule.Value
		}
	}

	return NewEvaluationResult(score, false)
}
