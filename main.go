package main

import (
	"bloodhound/lib/evaluator"
	rules "bloodhound/lib/rule"
	"bufio"
	"errors"
	"net/http"
	"os"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"

	"golang.org/x/net/html"
)

type UrlScore struct {
	url   string
	score int
}

func NewUrlScore(url string) UrlScore {
	return UrlScore{
		url:   url,
		score: 0,
	}
}

func (url *UrlScore) Add(score int) {
	url.score += score
}

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: false,
	})

	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func main() {
	// Prepare ruleset
	ruleset, err := rules.NewRuleset("rules.yml")

	if err != nil {
		log.WithField("reason", err).Fatal("Failed to process ruleset")
		return
	}

	log.WithField("size", len(ruleset.Scores)).Info("Initializing process with loaded ruleset")

	// Prepare client
	client := http.Client{
		// CheckRedirect: true,
	}

	urls, err := readInputFile()

	if err != nil {
		log.WithField("err", err).Fatal("Failed to process input file")
	}

	for index, urlScore := range urls {
		resourceScore := evaluator.EvaluateUrl(urlScore.url, ruleset)
		urls[index].Add(resourceScore)

		log.WithFields(log.Fields{
			"url":   urlScore.url,
			"score": resourceScore,
		}).Info("Finished evaluating URL")

		response, err := client.Get(urlScore.url)

		if err != nil {
			log.WithFields(log.Fields{
				"url": urlScore.url,
				"err": err,
			}).Error("Failed to retrieve page content")

			return
		}

		if response.StatusCode == http.StatusOK {
			document, _ := html.Parse(response.Body)
			contentScore := evaluator.EvaluateHTML(document, ruleset)

			log.WithFields(log.Fields{
				"url":   urlScore.url,
				"score": contentScore,
			}).Info("Finished evaluating content")

			urls[index].Add(contentScore)
		}

		log.WithFields(log.Fields{
			"url":   urlScore.url,
			"score": urlScore.score,
		}).Info("Finished scoring URL")

		response.Body.Close()
	}

	// Rank URLs by score
	sort.Slice(urls, func(i, j int) bool {
		return urls[i].score > urls[j].score
	})

	for _, url := range urls {
		log.Println(url.url)
	}
}

// TODO: Use lib that profile arg utils
func readInputFile() ([]UrlScore, error) {
	args := os.Args[1:]
	filePath := args[0]

	file, err := os.Open(filePath)
	if err != nil {
		return nil, errors.New("unable to open input file")
	}

	defer file.Close()

	var urls []UrlScore
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			score := NewUrlScore(line)
			urls = append(urls, score)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.New("unable to read input file")
	}

	return urls, nil
}
