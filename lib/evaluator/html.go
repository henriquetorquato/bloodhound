package evaluator

import (
	"bloodhound/lib/rules"
	"bloodhound/lib/utils"
	"slices"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

func EvaluateHTML(document *html.Node, ruleList []rules.Rule) EvaluationResult {
	score := 0

	if document == nil {
		return DefaultEvaluationResult()
	}

	var matchedRules []string

	for node := range document.Descendants() {
		if !shouldEvaluate(node) {
			log.WithFields(log.Fields{
				"type": node.Type,
			}).Trace("Skipping rule evaluation for HTML content node")

			continue
		}

		for _, rule := range ruleList {
			if rule.Level != rules.ContentLevel {
				log.WithFields(log.Fields{
					"rule": rule.Name,
				}).Trace("Unable to evaluate rule: Incompatible rule level")

				continue
			}

			// Same rule should not apply twice on the same document
			if slices.Contains(matchedRules, rule.Name) {
				continue
			}

			matches := nodeMatchesRule(node, &rule)

			if matches {
				if rule.Remove {
					log.WithFields(log.Fields{
						"rule": rule.Name,
					}).Trace("Rule with remove parameter matched, resource will be completely skipped")

					return NewEvaluationResult(0, true)
				}

				score += rule.Value
				matchedRules = append(matchedRules, rule.Name)
			}
		}
	}

	return NewEvaluationResult(score, false)
}

func nodeMatchesRule(node *html.Node, rule *rules.Rule) bool {
	if rule.Content.Element != "" {
		return nodeMatchesElementRule(node, rule)
	} else if len(rule.Content.Matches) != 0 {
		return evaluateNodeMatchRule(node, rule)
	}

	return false
}

// TODO: Add support for "any" matchers
func nodeMatchesElementRule(node *html.Node, rule *rules.Rule) bool {
	// Element rules only apply to element node types
	if node.Type != html.ElementNode {
		return false
	}

	// In this context, `node.Data` is the tag name
	if rule.Content.Element != node.Data {
		return false
	}

	if len(rule.Content.Attr) != 0 {
		match := true
		attrMap := getAttrMap(node.Attr)

		for key, value := range rule.Content.Attr {
			nodeAttrValue, hasKey := attrMap[key]

			if hasKey && nodeAttrValue != value {
				match = false
				break
			}
		}

		return match
	}

	log.WithFields(log.Fields{
		"node": node,
		"rule": rule.Name,
	}).Trace("Content level rule match")

	return true
}

func evaluateNodeMatchRule(node *html.Node, rule *rules.Rule) bool {
	// Text matching only applies to text nodes
	if node.Type != html.TextNode {
		return false
	}

	if utils.ContainsAny(node.Data, rule.Content.Matches) {
		log.WithFields(log.Fields{
			"node": node,
			"rule": rule.Name,
		}).Trace("Content level rule match")

		return true
	} else {
		return false
	}
}

func shouldEvaluate(node *html.Node) bool {
	switch node.Type {
	case html.ErrorNode,
		html.DocumentNode,
		html.DoctypeNode:
		return false

	default:
		return true
	}
}

func getAttrMap(attrs []html.Attribute) map[string]string {
	result := make(map[string]string)

	for _, attr := range attrs {
		result[attr.Key] = attr.Val
	}

	return result
}
