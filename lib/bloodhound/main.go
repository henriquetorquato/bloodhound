package bloodhound

import (
	"bloodhound/lib/evaluator"
	rules "bloodhound/lib/rule"
	"net/http"
	"sort"

	log "github.com/sirupsen/logrus"

	"golang.org/x/net/html"
)

type Context struct {
	Url   string
	Score int
}

func NewContext(url string) Context {
	return Context{
		Url:   url,
		Score: 0,
	}
}

func (context *Context) AddScore(score int) {
	context.Score += score
}

func Execute(targetUrls []string, ruleset rules.Ruleset) []Context {
	log.WithField("size", len(ruleset.Scores)).Info("Initializing process with loaded ruleset")

	// Prepare client
	client := http.Client{
		// CheckRedirect: true,
	}

	var contexts []Context

	for _, url := range targetUrls {
		context := NewContext(url)

		resourceScore := evaluator.EvaluateUrl(url, &ruleset)
		context.AddScore(resourceScore)

		log.WithFields(log.Fields{
			"url":   url,
			"score": resourceScore,
		}).Debug("Finished evaluating URL")

		response, err := client.Get(url)

		if err != nil {
			log.WithFields(log.Fields{
				"url": url,
				"err": err,
			}).Error("Failed to retrieve page content. Skipping url")

			continue
		}

		if response.StatusCode == http.StatusOK {
			document, _ := html.Parse(response.Body)
			contentScore := evaluator.EvaluateHTML(document, &ruleset)

			log.WithFields(log.Fields{
				"url":   url,
				"score": contentScore,
			}).Debug("Finished evaluating content")

			context.AddScore(contentScore)
		}

		log.WithFields(log.Fields{
			"url":   url,
			"score": context.Score,
		}).Debug("Finished scoring URL")

		response.Body.Close()
		contexts = append(contexts, context)
	}

	// Rank URLs by score
	sort.Slice(contexts, func(i, j int) bool {
		return contexts[i].Score > contexts[j].Score
	})

	return contexts
}
