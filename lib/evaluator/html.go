package evaluator

import (
	rules "bloodhound/lib/rule"

	"golang.org/x/net/html"
)

var types = [7]string{
	"ErrorNode",
	"TextNode",
	"DocumentNode",
	"ElementNode",
	"CommentNode",
	"DoctypeNode",
	"RawNode",
}

func EvaluateHTML(document *html.Node, ruleset rules.Ruleset) (int, error) {
	score := 0

	for node := range document.Descendants() {
		if !shouldEvaluate(node) {
			continue
		}

		// attrMap := getAttrMap(&node.Attr)

	}

	return score, nil
}

func evaluateNode() {

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

func getAttrMap(attrs *[]html.Attribute) map[string]string {
	result := make(map[string]string)

	for _, attr := range *attrs {
		result[attr.Key] = attr.Val
	}

	return result
}
