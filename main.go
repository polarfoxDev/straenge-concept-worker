package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"straenge-concept-worker/m/ai"
	"straenge-concept-worker/m/models"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var (
	ctx    = context.Background()
	client *redis.Client
)

const (
	threshold     = 15
	checkInterval = 5 * time.Second
)

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func enqueueConcept(concept models.RiddleConcept) error {
	conceptJson, err := json.Marshal(concept)
	if err != nil {
		return fmt.Errorf("marshal concept: %w", err)
	}
	job := models.Job{
		Type:    "new",
		Payload: string(conceptJson),
	}
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal job: %w", err)
	}
	if err := client.LPush(ctx, "generate-riddle", data).Err(); err != nil {
		return fmt.Errorf("insert job: %w", err)
	}
	return nil
}

func init() {
	// read dotenv file
	err := godotenv.Load()
	if err != nil {
		logrus.Warn("No .env file found")
	}
	// setup logging
	lvl, ok := os.LookupEnv("LOG_LEVEL")
	// LOG_LEVEL not set, let's default to info
	if !ok {
		lvl = "info"
	}
	// parse string, this is built-in feature of logrus
	ll, err := logrus.ParseLevel(lvl)
	if err != nil {
		ll = logrus.InfoLevel
	}
	// set global log level
	logrus.SetLevel(ll)
	logrus.Info("Logging initialized with level: ", lvl)
}

func main() {
	redisUrl, success := os.LookupEnv("REDIS_URL")
	if !success {
		logrus.Fatal("REDIS_URL not set")
		return
	}
	lang, success := os.LookupEnv("LANGUAGE")
	if !success {
		logrus.Fatal("LANGUAGE not set")
		return
	}
	lang = strings.TrimSpace(lang)
	supportedLangs := []string{"de", "sv"}
	if !contains(supportedLangs, lang) {
		logrus.Fatalf("Language '%s' is not supported", lang)
		return
	}
	apiKey, success := os.LookupEnv("OPENAI_API_KEY")
	if !success {
		logrus.Fatal("OPENAI_API_KEY not set")
		return
	}
	client = redis.NewClient(&redis.Options{
		Addr: redisUrl,
	})
	predefinedSuperSolutions := make([]string, 0)
	predefinedSuperSolutionsRaw, success := os.LookupEnv("PREDEFINED_SUPER_SOLUTIONS")
	if success {
		// PREDEFINED_SUPER_SOLUTIONS is a comma-separated list, so we split it
		for _, solution := range strings.Split(predefinedSuperSolutionsRaw, ",") {
			trimmedSolution := strings.TrimSpace(solution)
			if trimmedSolution != "" {
				predefinedSuperSolutions = append(predefinedSuperSolutions, trimmedSolution)
			}
		}
		logrus.Infof("Using predefined super solutions: %v", predefinedSuperSolutions)
	} else {
		logrus.Info("No predefined super solutions found, using empty list")
	}

	generator := ai.IdeaGenerator{}
	generator.Login(apiKey)
	generator.SetLanguage(lang)

	logrus.Info("Started worker...")

	for {
		logrus.Info("Checking queue...")
		len, err := client.LLen(ctx, "generate-riddle").Result()
		if err != nil {
			logrus.Errorf("❌ Redis Error: %v", err)
			time.Sleep(30 * time.Second)
			continue
		}

		if len < threshold {
			logrus.Infof("⚠️ queue only has %d elements – filling...", len)
			count, err := generateConcepts(&generator, predefinedSuperSolutions, enqueueConcept)
			if err != nil {
				logrus.Error("❌ generateConcepts failed: ", err)
				time.Sleep(30 * time.Second)
				continue
			}
			// empty predefinedSuperSolutions after single use
			predefinedSuperSolutions = make([]string, 0)
			if count == 0 {
				logrus.Warn("⚠️ No concepts enqueued during this iteration")
			}
		} else {
			logrus.Infof("✅ queue is filled (%d jobs)", len)
		}

		time.Sleep(checkInterval)
	}
}
