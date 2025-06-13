package evaluator

import (
	"bloodhound/lib/client"
	"bloodhound/lib/evaluator/pipeline"
	"bloodhound/lib/rules"
	"sort"

	log "github.com/sirupsen/logrus"
)

type EvaluationResult struct {
	Score  int
	Remove bool
}

func NewEvaluationResult(score int, remove bool) EvaluationResult {
	return EvaluationResult{
		Score:  score,
		Remove: remove,
	}
}

func DefaultEvaluationResult() EvaluationResult {
	return NewEvaluationResult(0, false)
}

// TODO: Add stopwatch
func Evaluate(targetUrls []string, ruleset *rules.Ruleset, clientConfig client.ClientConfig) []pipeline.Context {
	log.WithFields(log.Fields{
		"targetsSize": len(targetUrls),
		"rulesetSize": len(ruleset.Rules),
	}).Trace("Initializing evaluation pipeline")

	maxChannelSize := len(targetUrls)

	// Put context into pipeline
	inputChannel := make(chan pipeline.Context, maxChannelSize)
	go pipelineInput(targetUrls, inputChannel)

	// Apply resource name rules
	resourceLevelResultChannel := make(chan pipeline.Context, maxChannelSize)
	go applyResourceNameRules(ruleset, inputChannel, resourceLevelResultChannel)

	// Retrieve resource
	requestResultChannel := make(chan pipeline.Context, maxChannelSize)
	go pipeline.RetrieveResource(clientConfig, resourceLevelResultChannel, requestResultChannel)

	// Apply content level evaluation
	contentLevelResultChannel := make(chan pipeline.Context, maxChannelSize)
	go applyContentRules(ruleset, requestResultChannel, contentLevelResultChannel)

	log.WithFields(log.Fields{
		"targetsSize": len(targetUrls),
		"rulesetSize": len(ruleset.Rules),
	}).Info("Initialized evaluation pipeline")

	return pipelineOutput(contentLevelResultChannel)
}

func pipelineInput(targetUrls []string, out chan<- pipeline.Context) {
	defer close(out)

	for _, targetUrl := range targetUrls {
		log.WithFields(log.Fields{
			"target": targetUrl,
		}).Trace("Created new context")

		out <- pipeline.NewContext(targetUrl)
	}
}

func pipelineOutput(in <-chan pipeline.Context) []pipeline.Context {
	var contexts []pipeline.Context
	for context := range in {
		log.WithFields(log.Fields{
			"target": context.Url,
			"score":  context.Score,
		}).Info("Finished processing target")

		contexts = append(contexts, context)
	}

	// Rank URLs by score
	sort.Slice(contexts, func(i, j int) bool {
		return contexts[i].Score > contexts[j].Score
	})

	return contexts
}

func applyResourceNameRules(ruleset *rules.Ruleset, in <-chan pipeline.Context, out chan<- pipeline.Context) {
	defer close(out)
	resourceRules := ruleset.GetRules(rules.ResourceLevel)

	for context := range in {
		log.WithFields(log.Fields{
			"target": context.Url,
		}).Trace("Started resource level rule evaluation")

		evaluation := EvaluateUrl(&context.Url, &resourceRules)

		log.WithFields(log.Fields{
			"target":     context.Url,
			"evaluation": evaluation,
		}).Trace("Finished resource level rule evaluation")

		if !evaluation.Remove {
			context.AddScore(evaluation.Score)
			out <- context
		}
	}
}

func applyContentRules(ruleset *rules.Ruleset, in <-chan pipeline.Context, out chan<- pipeline.Context) {
	defer close(out)
	contentRules := ruleset.GetRules(rules.ContentLevel)

	for context := range in {
		log.WithFields(log.Fields{
			"target": context.Url,
		}).Trace("Started content level evaluation")

		// TODO: Add content type check and call popper content type evaluator
		evaluation := EvaluateHTML(context.Content, contentRules)

		log.WithFields(log.Fields{
			"target":     context.Url,
			"evaluation": evaluation,
		}).Trace("Finished content level evaluation")

		if !evaluation.Remove {
			context.AddScore(evaluation.Score)
			out <- context
		}
	}
}
