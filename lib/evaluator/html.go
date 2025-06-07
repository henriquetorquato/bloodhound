package evaluator

import (
	utils "bloodhound/lib"
	rules "bloodhound/lib/rule"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

func EvaluateHTML(document *html.Node, ruleset *rules.Ruleset) int {
	score := 0

	for node := range document.Descendants() {
		if !shouldEvaluate(node) {
			log.WithFields(log.Fields{
				"type": node.Type,
			}).Debug("Skipping rule evaluation for HTML content node")

			continue
		}

		score += evaluateNode(node, ruleset)
	}

	return score
}

func evaluateNode(node *html.Node, ruleset *rules.Ruleset) int {
	score := 0

	for _, rule := range ruleset.Scores {
		// Skip rules that don't apply
		if rule.Level != rules.ContentLevel {
			log.WithFields(log.Fields{
				"node": node.Data,
				"rule": rule.Name,
			}).Trace("Skipping rule while evaluating Node")

			continue
		}

		if rule.Content.Element != "" {
			// Element rules only apply to element node types
			if node.Type != html.ElementNode {
				continue
			}

			// In this context, `node.Data` is the tag name
			if rule.Content.Element != node.Data {
				continue
			}

			// element type rule can apply with or without attr matching
			if len(rule.Content.Attr) != 0 {
				attrMap := getAttrMap(node.Attr)

				for key, value := range rule.Content.Attr {
					nodeAttrValue := attrMap[key]

					if nodeAttrValue == value {
						log.WithFields(log.Fields{
							"node": node,
							"rule": rule.Name,
						}).Debug("Rule match")

						score += 1
					}
				}

			} else {
				log.WithFields(log.Fields{
					"node": node,
					"rule": rule.Name,
				}).Debug("Rule match")

				score += 1
			}
		} else if len(rule.Content.Matches) != 0 {
			// Text matching only applies to text nodes
			if node.Type != html.TextNode {
				continue
			}

			if utils.ContainsAny(node.Data, rule.Content.Matches) {
				log.WithFields(log.Fields{
					"node": node,
					"rule": rule.Name,
				}).Debug("Rule match")

				score += 1
			}
		}
	}

	return score
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
