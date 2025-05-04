package ai

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"slices"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

const (
	apiEndpoint = "https://api.openai.com/v1/chat/completions"
)

var specialCharacterMap = []string{"Ä", "Ö", "Ü", "ẞ"}

func makeWordSafe(word string) string {
	word = strings.ToUpper(word)
	word = strings.ReplaceAll(word, " ", "")
	word = strings.ReplaceAll(word, "-", "")
	word = strings.ReplaceAll(word, "ß", "ẞ")
	for i, specialCharacter := range specialCharacterMap {
		word = strings.ReplaceAll(word, specialCharacter, strconv.Itoa(i))
	}
	return word
}

type IdeaGenerator struct {
	openAiApiKey string
}

func (gen *IdeaGenerator) Login(apiKey string) {
	gen.openAiApiKey = apiKey
}

func (gen *IdeaGenerator) GetSuperSolutions() ([]string, error) {
	prompt := `
		Erstelle mindestens 50 Oberbegriffe (Kategorien) nach folgenden Regeln:

		1. Jeder Begriff besteht aus genau einem Wort (Komposita wie "Weltmusik" sind erlaubt).
		2. Die Begriffe dürfen sich inhaltlich nicht überschneiden.
		3. Die Begriffe müssen mindestens 6 und höchstens 30 Zeichen lang sein.
		4. Verwende keinerlei Wörter aus der unten stehenden Blacklist.
		5. Wähle überwiegend allgemeinere Kategorien, füge aber auch speziellere Kategorien hinzu (z. B. "SpanischeKüche", "Süßwasserfische", "Deutschrap" oder ein bestimmter Film).
		6. Überrasche mich mit mindestens 10 ungewöhnlichen oder schrulligen Kategorien.
		7. Gib ausschließlich ein gültiges JSON-Array in einer Zeile zurück, ohne sonstigen Text. Verzichte auch auf einen Codeblock.
		
		Beispiele (Blacklist): FRÜCHTE, GEMÜSE, MUSIKINSTRUMENTE, OSTERN, KANINCHEN, COMPUTER, ARCHITEKTUR, PHILOSOPHIE, KÜCHENGERÄTE, GAMEOFTHRONES, FISCHARTEN, PROGRAMMIERSPRACHEN, AUTOMARKEN, DEUTSCHRAP, SPANISCHEKÜCHE, WELTMUSIK`
	result, err := gen.rawRequest(prompt)
	if err != nil {
		logrus.Error("Error getting super solutions")
		return nil, err
	}
	logrus.Debug("raw gpt result: " + result)
	var items []string
	err = json.Unmarshal([]byte(strings.ToUpper(result)), &items)
	if err != nil {
		logrus.Error("Error parsing JSON:", err)
	}
	allowedItems := []string{}
	for _, item := range items {
		safeItem := makeWordSafe(item)
		if len(safeItem) >= 6 {
			alreadyInList := slices.Contains(allowedItems, safeItem)
			if !alreadyInList {
				allowedItems = append(allowedItems, safeItem)
			}
		}
	}
	return allowedItems, err
}

func (gen *IdeaGenerator) GetThemeBySuperSolution(unsafeSuperSolution string) (string, error) {
	prompt := `
		Formuliere eine rätselhafte Kurzbeschreibung zum Oberbegriff ` + unsafeSuperSolution + `.
		
		Regeln:
		1. Länge: höchstens 4 Wörter oder 30 Zeichen.
		2. Keine direkten Wortteile, Wortstämme oder Synonyme des Oberbegriffs.
		3. Deutsch, gern metaphorisch oder als Wortspiel.
		4. Man soll ein wenig knobeln müssen, um den Oberbegriff zu erraten.
		5. Gib ausschließlich diese Beschreibung in einer Zeile zurück - ohne Anführungszeichen, Zusatztext oder Formatierung.
		
		Beispiele
		- Musikinstrumente -> Klangquellen
		- Süßwasserfische -> Am Haken!
		`
	result, err := gen.rawRequest(prompt)
	logrus.Debug("raw gpt result: " + result)
	return result, err
}

func (gen *IdeaGenerator) GetWordPoolBySuperSolution(unsafeSuperSolution string) ([]string, error) {
	prompt := `
		Nenne mir etwa 10-30 Unterbegriffe zum Thema ` + unsafeSuperSolution + ` nach diesen Regeln:

		1. Nutze überwiegend geläufige Begriffe, die eine Durchschnittsperson kennt.
		2. **Ausnahmen**: Wenn das Thema es erfordert (z. B. "Automarken", "Programmiersprachen"), sind bekannte Markennamen oder nicht-deutsche Wörter ausdrücklich erlaubt.
		3. Meistens soll jeder Begriff aus einem Wort bestehen (Komposita erlaubt); einige wenige dürfen aus höchstens drei Wörtern bestehen.
		4. Begriffe dürfen sich ähneln, aber nicht identisch sein; vermeide unnötige Wiederholungen.
		5. Gib ausschließlich ein gültiges JSON-Array in einer Zeile zurück - ohne einleitenden oder nachfolgenden Text. Verzichte auch auf einen Codeblock.
		6. Qualität vor Quantität: Wenn du nur 10 Begriffe kennst, die allgemein verständlich und relevant sind, ist das auch in Ordnung.

		Beispielformat (für das Thema "Automarken"):
		["Volkswagen","Toyota","Ford", ...]
		`
	result, err := gen.rawRequest(prompt)
	if err != nil {
		logrus.Error("Error getting word pool")
		return nil, err
	}
	logrus.Debug("raw gpt result: " + result)
	var items []string
	err = json.Unmarshal([]byte(strings.ToUpper(result)), &items)
	if err != nil {
		logrus.Error("Error parsing JSON:", err)
	}
	allowedItems := []string{}
	for _, item := range items {
		safeItem := makeWordSafe(item)
		if len(safeItem) >= 4 {
			alreadyInList := slices.Contains(allowedItems, safeItem)
			if !alreadyInList {
				allowedItems = append(allowedItems, safeItem)
			}
		}
	}
	return allowedItems, err
}

func (gen *IdeaGenerator) rawRequest(query string) (string, error) {
	client := resty.New()

	response, err := client.R().
		SetAuthToken(gen.openAiApiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]any{
			"model":      "gpt-4o",
			"messages":   []any{map[string]any{"role": "system", "content": query}},
			"max_tokens": 500,
		}).
		Post(apiEndpoint)

	if err != nil {
		log.Fatalf("Error while sending send the request: %v", err)
		return "", err
	}

	body := response.Body()

	var data map[string]any
	err = json.Unmarshal(body, &data)
	if err != nil {
		logrus.Error("Error while decoding JSON response:", err)
		return "", err
	}

	// Extract the content from the JSON response
	content := data["choices"].([]any)[0].(map[string]any)["message"].(map[string]any)["content"].(string)
	return content, nil
}
