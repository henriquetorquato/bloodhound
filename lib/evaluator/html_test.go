package evaluator

import (
	"bloodhound/lib/rules"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestEvaluateHTML(t *testing.T) {

	ruleList := []rules.Rule{
		rules.NewContentRule("Has form", 1, false, rules.NewElementRuleContent("form", nil)),
		rules.NewContentRule("Has hidden input", 2, false, rules.NewElementRuleContent("input", map[string]string{"type": "hidden", "hidden": "true"})),
	}

	assert := func(t *testing.T, expected EvaluationResult, actual EvaluationResult) {
		if expected != actual {
			t.Errorf("EvaluateHTML; want %+v; got %+v", expected, actual)
		}
	}

	t.Run("page is not available", func(t *testing.T) {
		evaluation := EvaluateHTML(nil, ruleList)
		assert(t, DefaultEvaluationResult(), evaluation)
	})

	t.Run("login page", func(t *testing.T) {
		page := `<!DOCTYPE html>
			<head>
				<title>Login Page</title>
			</head>
			<body>
				<form action="/login" method="post">
					<input type="email" name="login" id="login">
					<input type="password" name="password" id="password">
					<input type="hidden" name="_token">
					<input type="hidden" name="_source">
					<input type="button" value="send">
				</form>
			</body>
			</html>`

		document := getHTMLDocument(page)

		t.Run("element found on page", func(t *testing.T) {
			evaluation := EvaluateHTML(document, ruleList)
			assert(t, NewEvaluationResult(3, false), evaluation)
		})
	})

	t.Run("about page", func(t *testing.T) {
		page := `<!DOCTYPE html>
			<head>
				<title>About Page</title>
			</head>
			<body>
				<p>Lorem ipsum dolor sit amet, consectetur adipiscing elit.</p>
				<p>Nulla sed tellus a dolor tristique congue. Nullam id velit in massa pulvinar ullamcorper.</p>
			</body>
			</html>`

		document := getHTMLDocument(page)

		t.Run("element found on page", func(t *testing.T) {
			evaluation := EvaluateHTML(document, ruleList)
			assert(t, NewEvaluationResult(0, false), evaluation)
		})
	})
}

func getHTMLDocument(content string) *html.Node {
	reader := strings.NewReader(content)
	document, _ := html.Parse(reader)

	return document
}
