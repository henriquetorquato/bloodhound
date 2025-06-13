package cmd

import (
	"bloodhound/lib/client"
	"bloodhound/lib/evaluator"
	"bloodhound/lib/evaluator/pipeline"
	"bloodhound/lib/rules"
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	inputFile   string
	rulesetFile string

	outputFile     string
	logLevelStr    string
	requestRate    int
	requestHeaders []string
	proxyServer    string

	// TODO: Add "passive" option, so that no request is made to the target, and only resource name is evaluated
	cmd = &cobra.Command{
		Use:   "bloodhound",
		Short: "URL resource evaluator and sorter",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			level, err := parseLogLevel(logLevelStr)

			if err != nil {
				log.Warnf("Invalid log level %q, defaulting to INFO", logLevelStr)
				level = log.InfoLevel
			}

			log.SetFormatter(&log.TextFormatter{
				DisableColors: false,
				FullTimestamp: false,
			})

			log.SetOutput(os.Stdout)
			log.SetLevel(level)
		},
		Run: func(cmd *cobra.Command, args []string) {
			// Validate that input file exists
			targetUrls, err := readInputFile(inputFile)

			if err != nil {
				log.Fatalf("Failed to process input file. Reason: %s", err.Error())
				os.Exit(1)
			}

			log.WithFields(log.Fields{
				"size": len(targetUrls),
			}).Trace("Finished reading input file")

			// Validate that rule file exists
			ruleset, err := rules.NewRuleset(rulesetFile)

			if err != nil {
				log.Fatalf("Failed to process ruleset file. Reason: %s", err.Error())
				os.Exit(1)
			}

			log.WithFields(log.Fields{
				"size": len(ruleset.Rules),
			}).Trace("Finished reading ruleset file")

			// Parse client configurations
			headers, err := parseCustomHeaders(requestHeaders)

			if err != nil {
				log.Fatalf("Failed to parse customer headers. Reason: %s", err.Error())
				os.Exit(1)
			}

			clientConfig := client.ClientConfig{
				Rate:    requestRate,
				Headers: headers,
				Proxy:   proxyServer,
			}

			log.WithFields(log.Fields{
				"config": clientConfig,
			}).Trace("Finished creating HTTP client configurations")

			// Execute command
			results := evaluator.Evaluate(targetUrls, ruleset, clientConfig)

			// Write to output file
			err = writeOutputFile(outputFile, results)

			if err != nil {
				log.Fatalf("Failed to write to output file. Reason: %s", err.Error())
			}
		},
	}
)

func init() {
	// Mandatory fields
	cmd.PersistentFlags().StringVarP(&inputFile, "input", "i", "", "Input file with a list of URLs to process (required)")
	cmd.MarkFlagRequired("input")

	cmd.PersistentFlags().StringVarP(&rulesetFile, "rules", "r", "", "Ruleset file with rules and scores (required)")
	cmd.MarkFlagRequired("rules")

	// Optional fields
	cmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "output.txt", "Output file to write sorted list")
	cmd.PersistentFlags().StringVarP(&logLevelStr, "log-level", "l", "info", "Set log level: trace, debug, info, warn, error, fatal, panic")
	cmd.PersistentFlags().IntVarP(&requestRate, "rate", "R", 100, "Number of HTTP requests allowed during a single second on each thread")
	cmd.PersistentFlags().StringArrayVarP(&requestHeaders, "headers", "H", []string{}, "Customer headers to be used when sending HTTP requests (--header \"User-Agent: Mozilla/5.0\")")
	cmd.PersistentFlags().StringVarP(&proxyServer, "proxy", "P", "", "Proxy server in URL format (http://localhost:8080)")
}

func readInputFile(inputFile string) ([]string, error) {
	file, err := os.Open(inputFile)
	if err != nil {
		return nil, errors.New("unable to open input file")
	}

	defer file.Close()

	var targetUrls []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			score := line
			targetUrls = append(targetUrls, score)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.New("unable to read input file")
	}

	return targetUrls, nil
}

// TODO: Write to /temp if unable to write to configured output
func writeOutputFile(outputFile string, results []pipeline.Context) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return errors.New("unable to create output file")
	}

	defer file.Close()

	writer := bufio.NewWriter(file)

	for _, result := range results {
		_, err = writer.WriteString(result.Url)

		if err != nil {
			return err
		}

		writer.WriteString("\n")
	}

	file.Sync()
	writer.Flush()

	return nil
}

func parseLogLevel(level string) (log.Level, error) {
	switch strings.ToLower(level) {
	case "trace":
		return log.TraceLevel, nil
	case "debug":
		return log.DebugLevel, nil
	case "info":
		return log.InfoLevel, nil
	case "warn", "warning":
		return log.WarnLevel, nil
	case "error":
		return log.ErrorLevel, nil
	case "fatal":
		return log.FatalLevel, nil
	case "panic":
		return log.PanicLevel, nil
	default:
		return log.InfoLevel, fmt.Errorf("unknown log level: %s", level)
	}
}

func parseCustomHeaders(inputHeaders []string) (map[string]string, error) {
	headers := make(map[string]string)

	for _, inputHeader := range inputHeaders {
		parts := strings.SplitN(inputHeader, ":", 2)

		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid header format: %q, expected Key:Value", inputHeader)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return nil, fmt.Errorf("invalid header format: %q, expected Key:Value", inputHeader)
		}

		headers[key] = value
	}

	return headers, nil
}

func Execute() error {
	return cmd.Execute()
}
