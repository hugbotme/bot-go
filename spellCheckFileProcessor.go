package main

import (
	"bytes"
	"fmt"
	aspell "github.com/hugbotme/go-aspell"
	"io/ioutil"
	"regexp"
	s "strings"
)

type spellCheckFileProcessor struct {
	spellChecker  aspell.Speller
	stopWords     []string
	probableWords []string
}

func newSpellCheckFileProcessor(stopWordsFile string, probableWordsFile string) (spellCheckFileProcessor, error) {
	// Initialize the speller
	speller, err := aspell.NewSpeller(map[string]string{
		"lang": "en_US",
		//"personal": stopWordsFile,
	})
	// jvt: be sure to clean up, C lib being used here....
	defer speller.Delete()

	// jvt: read stop words file to array
	// jvt: @todo error handling?
	stopWordsContent, _ := ioutil.ReadFile(stopWordsFile)

	// jvt: read probable words words file to array
	// jvt: @todo error handling?
	probableWordsContent, _ := ioutil.ReadFile(probableWordsFile)

	return spellCheckFileProcessor{
		spellChecker:  speller,
		stopWords:     s.Split(string(stopWordsContent), "\n"),
		probableWords: s.Split(string(probableWordsContent), "\n"),
	}, err
}

/**
 * run a spell check on passed content
 * passes back original content if an error occurs
 */
func (spfp spellCheckFileProcessor) processContent(content []byte) string {
	var buffer bytes.Buffer
	var wordBuffer bytes.Buffer
	syntaxNestingLevel := 0
	contentLength := len(content)

	// jvt: start looping content bytes
	for index, b := range content {
		//fmt.Println(string(b))
		if spfp.isMarkdownSyntaxOpeningChar(b) {
			//fmt.Println("entering nesting level")
			syntaxNestingLevel++
		} else if spfp.isMarkdownSyntaxClosingChar(b) {
			//fmt.Println("leaving nesting level")
			syntaxNestingLevel--

			// jvt: write byte to buffer
			buffer.WriteByte(b)

			// jvt: and continue to next byte
			continue
		}

		// jvt: @todo values under 0 most likely mean invalid markdown, ignoring for now
		if syntaxNestingLevel > 0 {
			//fmt.Println("in nesting level")
			// jvt: we're ignoring content, just copy
			buffer.WriteByte(b)
		} else {
			var isWordEndingChar bool
			if index > 0 && index < contentLength-2 {
				isWordEndingChar = spfp.isWordEndingChar(b, content[index-1], content[index+1])
			} else {
				isWordEndingChar = spfp.isWordEndingChar(b)
			}

			// jvt: check for end of word
			if wordBuffer.Len() > 0 && isWordEndingChar {
				//fmt.Println("found word " + wordBuffer.String())
				// jvt: process word & write back to buffer
				buffer.WriteString(spfp.processWord(wordBuffer.String()))

				// jvt: reset word buffer
				wordBuffer.Reset()
			} else if !isWordEndingChar {
				//fmt.Println("in word")
				// jvt: we're in a word, copy current byte to word buffer
				wordBuffer.WriteByte(b)
			}

			if isWordEndingChar {
				//fmt.Println("word-ending char")
				// jvt: write word-ending byte to buffer
				buffer.WriteByte(b)
			}
		}
	}

	// jvt: and pass back finished string
	return buffer.String()
}

func (spfp spellCheckFileProcessor) isMarkdownSyntaxOpeningChar(b byte) bool {
	char := string(b)
	return char == "[" || char == "(" || char == "`"
}

func (spfp spellCheckFileProcessor) isMarkdownSyntaxClosingChar(b byte) bool {
	char := string(b)
	return char == "]" || char == ")" || char == "´"
}

func (spfp spellCheckFileProcessor) isWordEndingChar(chars ...byte) bool {
	// jvt: @todo huh? byte -> string -> byte array type case is fine, but byte to byte array type cast not? missing something stupid here....
	// jvt: we always get the first param
	char := string(chars[0])

	var matched bool
	if len(chars) > 1 && spfp.isLookForwardAndBackChar(char) {
		//fmt.Println("checking for contraction " + char + string(chars[1]) + string(chars[2]))
		// jvt: try to detect contraction
		matched = spfp.matchLetter(string(chars[1])) && spfp.matchLetter(string(chars[2]))
	} else {
		matched = spfp.matchLetter(char)
	}

	return !matched
}

func (spfp spellCheckFileProcessor) matchLetter(char string) bool {
	matched, _ := regexp.Match("[A-Za-z]", []byte(char))
	return matched
}

func (spfp spellCheckFileProcessor) isLookForwardAndBackChar(char string) bool {
	return char == "'" || char == "-"
}

func (spfp spellCheckFileProcessor) processWord(word string) string {
	// jvt: check for stop word
	if spfp.checkForStopword(word) {
		return word
	}

	spellingCorrect, suggestions := spfp.checkSpelling(word)
	if spellingCorrect {
		return word
	} else {
		//fmt.Printf("Incorrect word, suggestions: %s\n", s.Join(suggestions, ", "))

		// jvt: @todo jup....
		if len(suggestions) > 0 {
			//fmt.Printf("suggestions: %s\n", s.Join(suggestions, ", "))
			preferredWord := spfp.checkForPreferred(suggestions)
			fmt.Println("Replacing \"" + word + "\" with \"" + preferredWord + "\"")
			return preferredWord
		}

		return word
	}
}

func (spfp spellCheckFileProcessor) checkForStopword(word string) bool {
	for _, stopword := range spfp.stopWords {
		if word == stopword {
			return true
		}
	}

	return false
}

func (spfp spellCheckFileProcessor) checkForPreferred(suggestions []string) string {
	for _, suggestion := range suggestions {
		for _, preferred := range spfp.probableWords {
			if suggestion == preferred {
				fmt.Println("found preferred word: " + preferred)
				return preferred
			}
		}
	}

	// jvt: if we didn't find anything preferred, just return first suggestion
	return suggestions[0]
}

func (spfp spellCheckFileProcessor) checkSpelling(word string) (bool, []string) {
	if spfp.spellChecker.Check(word) {
		//fmt.Print("OK\n")
		return true, nil
	}

	suggestions := spfp.spellChecker.Suggest(word)
	//fmt.Printf("Spelling mistake:\"" + word + "\" suggestions: %s\n", s.Join(suggestions, ", "))
	return false, suggestions
}
