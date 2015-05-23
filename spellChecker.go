package main

import(
	"fmt"
	aspell "github.com/hugbotme/go-aspell"
	s "strings"
)

/**
 * run a spell check on passed content
 * passes back original content if an error occurs
 */
func spellCheck(content string) string {
	// Initialize the speller
	speller, err := aspell.NewSpeller(map[string]string{
		"lang": "en_US",
	})
	if err != nil {
		fmt.Errorf("could not get speller: %s", err.Error())
		return content
	}
	// jvt: be sure to clean up, C lib being used here....
	defer speller.Delete()

	// jvt: @todo more sophisticated "word detection"
	words := s.Split(content, " ")
	for _, word := range words {
		fmt.Println(word + ": ")
		if speller.Check(word) {
			fmt.Print("OK\n")
		} else {
			fmt.Printf("Incorrect word, suggestions: %s\n", s.Join(speller.Suggest(word), ", "))
		}
	}

	return content
}
