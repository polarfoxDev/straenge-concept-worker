package ai

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

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

func makeWordUnsafe(word string) string {
	for i, specialCharacter := range specialCharacterMap {
		word = strings.ReplaceAll(word, strconv.Itoa(i), specialCharacter)
	}
	return word
}

type IdeaGenerator struct {
	openAiApiKey string
	language     string
}

func (gen *IdeaGenerator) Login(apiKey string) {
	gen.openAiApiKey = apiKey
}

func (gen *IdeaGenerator) SetLanguage(lang string) {
	gen.language = lang
}

func (gen *IdeaGenerator) GetSuperSolutions() ([]string, error) {
	promptDE := `
		Erstelle mindestens 40 Kategorien als JSON-Array (eine Zeile, keine Zusätze). Regeln:
		1. genau ein Wort pro Kategorie (Komposita erlaubt),
		2. keine inhaltlichen Überschneidungen,
		3. Länge 6–30 Zeichen,
		4. keine Wörter aus der Blacklist,
		5. überwiegend allgemein, aber auch spezielle Begriffe (z. B. SpanischeKüche, Süßwasserfische, Deutschrap, bestimmter Film),
		6. mindestens 2 ungewöhnliche/kuriose Kategorien,
		7. nur allgemein bekannte Begriffe, keine Neuschöpfungen.
		Beispiele (Blacklist): Früchte, Gemüse, Musikinstrumente, Ostern, Kaninchen, Computer, Architektur, Philosophie, Küchengeräte, GameOfThrones, Fischarten, Programmiersprachen, Automarken, Deutschrap, SpanischeKüche, Weltmusik
	`
	promptSV := `
		Skapa minst 40 kategorier som en JSON-array (en rad, inga tillägg). Regler:
		1. exakt ett ord per kategori (sammansättningar tillåtna),
		2. inga innehållsliga överlapp,
		3. längd 6–30 tecken,
		4. inga ord från svartlistan,
		5. mest allmänna men även specifika begrepp (t.ex. SpanskMat, Sötvattensfiskar, TyskRap, en viss film),
		6. minst 2 ovanliga/knasiga kategorier,
		7. endast allmänt kända begrepp, inga nyskapade ord.
		Exempel (svartlista): frukter, grönsaker, musikinstrument, kaniner, datorer, arkitektur, filosofi, köksredskap, GameOfThrones, fiskarter, programmeringsspråk, bilmärken, tyskrap, spanskmat, världsmusik
	`
	var prompt string
	if gen.language == "de" {
		prompt = promptDE
	} else {
		prompt = promptSV
	}
	result, err := gen.reasoningRequest(prompt, "gpt-5", "high")
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

func (gen *IdeaGenerator) GetThemeBySuperSolution(superSolution string) (string, error) {
	unsafeSuperSolution := makeWordUnsafe(superSolution)
	promptDE := `
		Formuliere eine rätselhafte Kurzbeschreibung zum Oberbegriff ` + unsafeSuperSolution + `.
		Regeln:
		1. Maximal 4 Wörter oder 30 Zeichen.
		2. Keine Wortteile, Wortstämme oder Synonyme des Oberbegriffs.
		3. Auf Deutsch, gern metaphorisch oder als Wortspiel.
		4. Sie soll zum Knobeln anregen.
		5. Gib nur die Beschreibung zurück – eine Zeile, ohne Anführungszeichen oder Zusatztext.
		Beispiele:
		- Musikinstrumente -> Klangquellen
		- Süßwasserfische -> Am Haken!
	`
	promptSV := `
		Formulera en gåtfull kort beskrivning av överbegreppet ` + unsafeSuperSolution + `.
		Regler:
		1. Maximal 4 ord eller 30 tecken.
		2. Inga orddelar, ordstammar eller synonymer till överbegreppet.
		3. På svenska, gärna metaforiskt eller som ett ordspel.
		4. Den ska få en att klura.
		5. Ge endast beskrivningen - en rad, utan citattecken eller tilläggstext.
		Exempel:
		- Musikinstrument -> Ljudkällor
		- Sötvattensfiskar -> På kroken!
	`
	var prompt string
	if gen.language == "de" {
		prompt = promptDE
	} else {
		prompt = promptSV
	}
	result, err := gen.fastRequest(prompt, "gpt-5-chat-latest")
	logrus.Debug("raw gpt result: " + result)
	return result, err
}

func (gen *IdeaGenerator) GetWordPoolBySuperSolution(superSolution string) ([]string, error) {
	unsafeSuperSolution := makeWordUnsafe(superSolution)
	promptDE := `
		Nenne 10–30 Unterbegriffe zum Thema + unsafeSuperSolution +.
		Regeln:
		1. Überwiegend geläufige Begriffe, die eine Durchschnittsperson kennt.
		2. Ausnahmen: Bei Themen wie Automarken oder Programmiersprachen sind bekannte Marken- oder Fremdwörter erlaubt.
		3. Meist ein Wort (Komposita erlaubt); wenige Ausnahmen mit max. 3 Wörtern.
		4. Begriffe dürfen ähnlich, aber nicht identisch sein; Wiederholungen vermeiden.
		5. Gib nur ein gültiges JSON-Array in einer Zeile zurück – ohne Zusatztext oder Codeblock.
		6. Qualität vor Quantität: 10–15 gute Begriffe sind ausreichend, wenn mehr nicht sinnvoll sind.
		7. Bevorzuge kurze Begriffe (4–8 Zeichen), wenn möglich.
		Beispiel (Thema „Automarken“):
		["Volkswagen","Toyota","Ford", ...]
	`
	promptSV := `
		Nämn 10–30 underbegrepp till temat + unsafeSuperSolution +.
		Regler:
		1. Mest vanliga begrepp som en genomsnittsperson känner till.
		2. Undantag: För teman som Bilmärken eller Programmeringsspråk är kända varumärken eller utländska ord tillåtna.
		3. Vanligen ett ord (sammansättningar tillåtna); några få får ha högst 3 ord.
		4. Begrepp får vara liknande men inte identiska; undvik onödiga upprepningar.
		5. Ge endast en giltig JSON-array på en rad – utan extra text eller kodblock.
		6. Kvalitet före kvantitet: 10–15 bra begrepp räcker om fler inte är rimliga.
		7. Föredra korta begrepp (4–8 tecken) när det är möjligt.
		Exempel (tema ”Bilmärken”):
		["Volkswagen","Toyota","Ford", ...]
	`
	var prompt string
	if gen.language == "de" {
		prompt = promptDE
	} else {
		prompt = promptSV
	}
	result, err := gen.reasoningRequest(prompt, "gpt-5", "high")
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

func (gen *IdeaGenerator) reasoningRequest(query string, model string, effort string) (string, error) {
	body := map[string]any{
		"model":            model,
		"reasoning_effort": effort,
		"messages":         []any{map[string]any{"role": "system", "content": query}},
	}
	return gen.rawRequest(body)
}

func (gen *IdeaGenerator) fastRequest(query string, model string) (string, error) {
	body := map[string]any{
		"model":    model,
		"messages": []any{map[string]any{"role": "user", "content": query}},
	}
	return gen.rawRequest(body)
}

func (gen *IdeaGenerator) rawRequest(body map[string]any) (string, error) {
	logrus.Debug("raw gpt request: " + string(body["messages"].([]any)[0].(map[string]any)["content"].(string)))
	client := resty.New().SetTimeout(time.Duration(5 * time.Minute))

	response, err := client.R().
		SetAuthToken(gen.openAiApiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(apiEndpoint)

	if err != nil {
		log.Fatalf("Error while sending send the request: %v", err)
		return "", err
	}

	responseBody := response.Body()

	var data map[string]any
	err = json.Unmarshal(responseBody, &data)
	if err != nil {
		logrus.Error("Error while decoding JSON response:", err)
		return "", err
	}

	logrus.Debug("raw gpt result: " + string(responseBody))
	content := data["choices"].([]any)[0].(map[string]any)["message"].(map[string]any)["content"].(string)
	return content, nil
}
