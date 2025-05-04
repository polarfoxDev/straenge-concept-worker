package main

import (
	"straenge-concept-worker/m/ai"
	"straenge-concept-worker/m/models"

	"github.com/sirupsen/logrus"
)

func generateConcepts(generator *ai.IdeaGenerator) *[]models.RiddleConcept {
	concepts := make([]models.RiddleConcept, 0)
	logrus.Info("Generating super solutions...")
	superSolutions, err := generator.GetSuperSolutions()

	if err != nil {
		logrus.Error("Error getting super solutions:", err)
		return nil
	}

	logrus.Infof("Generated %d super solutions", len(superSolutions))

	for _, superSolution := range superSolutions {
		logrus.Info("Generating theme for super solution: " + superSolution)

		theme, err := generator.GetThemeBySuperSolution(superSolution)
		if err != nil {
			logrus.Error("Error getting theme for super solution:", err)
			continue
		}

		logrus.Info("Generating word pool for super solution: " + superSolution)

		wordList, err := generator.GetWordPoolBySuperSolution(superSolution)
		if err != nil {
			logrus.Error("Error getting word pool for super solution:", err)
			continue
		}

		concept := models.RiddleConcept{
			SuperSolution:    superSolution,
			ThemeDescription: theme,
			WordPool:         wordList,
		}
		concepts = append(concepts, concept)
	}

	return &concepts
}
