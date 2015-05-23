package main

import(
	"fmt"
	aspell "github.com/hugbotme/go-aspell"
	s "strings"
	"bytes"
)

type spellCheckFileProcessor struct {
	spellChecker aspell.Speller
}

func newSpellCheckFileProcessor() (spellCheckFileProcessor, error) {
	// Initialize the speller
	speller, err := aspell.NewSpeller(map[string]string{
		"lang": "en_US",
	})
	// jvt: be sure to clean up, C lib being used here....
	defer speller.Delete()

	return spellCheckFileProcessor{
		spellChecker: speller,
	}, err
}

/**
 * run a spell check on passed content
 * passes back original content if an error occurs
 */
func (spfp spellCheckFileProcessor) processContent (content []byte) string {
	var buffer bytes.Buffer
	for _, b := range content {
		buffer.WriteByte(b)
	}

	return buffer.String()
}

func (spfp spellCheckFileProcessor) checkSpelling (word string) (bool, []string) {
	if spfp.spellChecker.Check(word) {
		fmt.Print("OK\n")
		return true, nil
	}

	suggestions := spfp.spellChecker.Suggest(word)
	fmt.Printf("Incorrect word, suggestions: %s\n", s.Join(suggestions, ", "))
	return false, suggestions
}

// jvt: @todo more sophisticated "word detection"
/*words := s.Fields(content)
for _, word := range words {
	validWord, word := validate(word)
	if validWord {
		fmt.Printf(word + ": ")
		if speller.Check(word) {
			fmt.Print("OK\n")
		} else {
			suggestions := speller.Suggest(word)
			fmt.Printf("Incorrect word, suggestions: %s\n", s.Join(suggestions, ", "))
		}
	}
}*/
/*
func validate(word string) (bool, string) {
	// jvt: trim punctuation
	word = trimSuffix(word, ".")
	word = trimSuffix(word, ",")
	word = trimSuffix(word, "!")
	word = trimSuffix(word, ":")

	// jvt: check length
	if len(word) < 2 {
		return false, word
	}

	return true, word
}

func trimSuffix(word, suffix string) string {
	if s.HasSuffix(word, suffix) {
		word = word[:len(word)-len(suffix)]
	}

	return word
}*/
