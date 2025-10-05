package main

import (
	"straenge-concept-worker/m/ai"
	"straenge-concept-worker/m/models"

	"github.com/sirupsen/logrus"
)

func generateConcepts(
	generator *ai.IdeaGenerator,
	predefinedSuperSolutions []string,
	enqueue func(models.RiddleConcept) error,
) (int, error) {
	conceptsEnqueued := 0

	var superSolutions []string
	if len(predefinedSuperSolutions) > 0 {
		logrus.Infof("Using predefined super solutions: %v", predefinedSuperSolutions)
		superSolutions = predefinedSuperSolutions
	} else {
		logrus.Info("Generating super solutions...")
		var err error
		superSolutions, err = generator.GetSuperSolutions()
		if err != nil {
			logrus.Error("Error getting super solutions:", err)
			return 0, err
		}
		logrus.Infof("Generated %d super solutions", len(superSolutions))
	}

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
		if err := enqueue(concept); err != nil {
			logrus.Errorf("Error enqueuing concept for super solution '%s': %v", superSolution, err)
			continue
		}
		logrus.Infof("➕ Concept enqueued for super solution '%s'", superSolution)
		conceptsEnqueued++
	}

	return conceptsEnqueued, nil
}
