package evaluator

import (
	"bloodhound/lib/rules"
	"testing"
)

// TODO: Add tests for subdomain matching
func TestEvaluateUrl(t *testing.T) {
	ruleList := []rules.Rule{
		rules.NewResourceRule("Match login page", 1, false, rules.NewMatchRuleContent([]string{"login", "auth"})),
		rules.NewResourceRule("Remove about page", 2, true, rules.NewMatchRuleContent([]string{"about"})),
		rules.NewResourceRule("Rule with no match words", 4, false, rules.NewMatchRuleContent([]string{})),
		rules.NewContentRule("Rule with miss-matching level", 8, false, rules.NewMatchRuleContent([]string{"login", "auth"})),
		rules.NewResourceRule("Match interesting file extensions", 16, false, rules.NewMatchRuleContent([]string{".bak", ".log"})),
		rules.NewResourceRule("Possible SSRF", 32, false, rules.NewMatchRuleContent([]string{"url=", "callback="})),
	}

	assert := func(t *testing.T, expected EvaluationResult, actual EvaluationResult) {
		if expected != actual {
			t.Errorf("EvaluateUrl; want %+v; got %+v", expected, actual)
		}
	}

	t.Run("simple pattern matching", func(t *testing.T) {
		t.Run("name not in list", func(t *testing.T) {
			url := "http://localhost/potato"
			evaluation := EvaluateUrl(&url, &ruleList)
			assert(t, NewEvaluationResult(0, false), evaluation)
		})

		t.Run("name in list", func(t *testing.T) {
			url := "http://localhost/login"
			evaluation := EvaluateUrl(&url, &ruleList)
			assert(t, NewEvaluationResult(1, false), evaluation)
		})

		t.Run("name twice in list", func(t *testing.T) {
			url := "http://localhost/auth/login"
			evaluation := EvaluateUrl(&url, &ruleList)
			assert(t, NewEvaluationResult(1, false), evaluation)
		})

		t.Run("name in list with remove filter", func(t *testing.T) {
			url := "http://localhost/about"
			evaluation := EvaluateUrl(&url, &ruleList)
			assert(t, NewEvaluationResult(0, true), evaluation)
		})
	})

	t.Run("complex pattern matching", func(t *testing.T) {
		t.Run("with extension matching", func(t *testing.T) {
			url := "http://localhost/backup.bak"
			evaluation := EvaluateUrl(&url, &ruleList)
			assert(t, NewEvaluationResult(16, false), evaluation)
		})

		t.Run("with parameter names", func(t *testing.T) {
			url := "http://localhost/request?url=http%3A%2F%2Flocalhost%2Fsensitive"
			evaluation := EvaluateUrl(&url, &ruleList)
			assert(t, NewEvaluationResult(32, false), evaluation)
		})
	})
}
