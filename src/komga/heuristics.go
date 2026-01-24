package komga

import (
	"strings"
	"sync"

	"github.com/pemistahl/lingua-go"
)

var (
	detector     lingua.LanguageDetector
	detectorOnce sync.Once
)

func getDetector() lingua.LanguageDetector {
	detectorOnce.Do(func() {
		detector = lingua.NewLanguageDetectorBuilder().
			FromLanguages(lingua.English, lingua.Russian, lingua.Arabic, lingua.Hebrew, lingua.French, lingua.German, lingua.Swedish).
			WithMinimumRelativeDistance(0.5). // vibes
			Build()
	})
	return detector
}

func IsBookForeign(book BookItem, skipFilter string) (bool, lingua.Language, float64) {
	det := getDetector()

	title := book.Name
	if idx := strings.Index(book.Name, " - "); idx != -1 {
		title = book.Name[idx+3:]
	}

	if skipFilter != "" {
		skips := strings.FieldsFunc(skipFilter, func(r rune) bool {
			return r == '|'
		})
		for _, skip := range skips {
			if strings.Contains(strings.ToLower(title), strings.ToLower(skip)) {
				return false, lingua.English, 1.0
			}
		}
	}

	if language, exists := det.DetectLanguageOf(title); exists {
		confidence := det.ComputeLanguageConfidence(title, language)
		return language != lingua.English, language, confidence
	}

	return false, lingua.English, det.ComputeLanguageConfidence(title, lingua.English)
}

func TitleNoWeirdSymbols(title string) bool {
	return false
}
