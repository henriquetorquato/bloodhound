package rules

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"

	log "github.com/sirupsen/logrus"
)

type Level string

const (
	UnknownLevel  Level = ""
	ResourceLevel Level = "resource"
	ContentLevel  Level = "content"
)

type RuleContent struct {
	Element string
	Attr    map[string]string
	Matches []string
}

type Rule struct {
	Name    string
	Value   int
	Level   Level
	Content RuleContent
}

type Ruleset struct {
	Name   string
	Scores []Rule
}

func NewRuleset(filepath string) (*Ruleset, error) {
	data, err := os.ReadFile(filepath)

	if err != nil {
		return nil, errors.New("Unable to open ruleset file. " + err.Error())
	}

	ruleset := Ruleset{}
	err = yaml.Unmarshal(data, &ruleset)

	if err != nil {
		return nil, errors.New("Unable to parse ruleset file. " + err.Error())
	}

	log.WithFields(log.Fields{
		"filepath": filepath,
		"data":     ruleset,
	}).Trace("Finished reading ruleset from filepath")

	for _, rule := range ruleset.Scores {
		if !rule.isValid() {
			return nil, errors.New("Unable to parse rule with name '" + rule.Name + "': Invalid rule configurations.")
		}
	}

	log.WithFields(log.Fields{
		"filepath":   filepath,
		"scores_len": len(ruleset.Scores),
	}).Debug("Finished parsing rulesets from filepath")

	return &ruleset, nil
}

func (rule *Rule) isValid() bool {
	switch rule.Level {
	case ResourceLevel:
		return rule.isResourceRuleValid()
	case ContentLevel:
		return rule.isContentRuleValid()
	}

	return false
}

func (rule *Rule) isResourceRuleValid() bool {
	return len(rule.Content.Matches) != 0
}

func (rule *Rule) isContentRuleValid() bool {
	return len(rule.Content.Matches) != 0 || rule.Content.Element != ""
}
