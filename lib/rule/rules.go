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
	ElementName string
	Attr        map[string]string
	Matches     []string
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
		"filepath":   filepath,
		"scores_len": len(ruleset.Scores),
	}).Debug("Finished parsing rulesets from filepath")

	return &ruleset, nil
}
