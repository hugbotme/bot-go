package main

import(
	"fmt"
	aspell "github.com/hugbotme/go-aspell"
	s "strings"
	"bytes"
	"regexp"
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
	var wordBuffer bytes.Buffer
	syntaxNestingLevel := 0
	contentLength := len(content)

	// jvt: start looping content bytes
	for index, b := range content {
		//fmt.Println(string(b))
		if spfp.isMarkdownSyntaxOpeningChar(b) {
			//fmt.Println("entering nesting level")
			syntaxNestingLevel ++
		} else if spfp.isMarkdownSyntaxClosingChar(b) {
			//fmt.Println("leaving nesting level")
			syntaxNestingLevel --

			// jvt: write byte to buffer
			buffer.WriteByte(b)

			// jvt: and continue to next byte
			continue
		}

		// jvt: @todo values under 0 most likely mean invalid markdown, ignoring for now
		if (syntaxNestingLevel > 0) {
			//fmt.Println("in nesting level")
			// jvt: we're ignoring content, just copy
			buffer.WriteByte(b)
		} else {
			var isWordEndingChar bool
			if index > 0 && index < contentLength {
				isWordEndingChar = spfp.isWordEndingChar(b, content[index - 1], content[index + 1])
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

			if (isWordEndingChar) {
				//fmt.Println("word-ending char")
				// jvt: write word-ending byte to buffer
				buffer.WriteByte(b)
			}
		}
	}

	// jvt: and pass back finished string
	return buffer.String()
}

func (spfp spellCheckFileProcessor) isMarkdownSyntaxOpeningChar (b byte) bool {
	char := string(b)
	return char == "[" || char == "(" || char == "`"
}

func (spfp spellCheckFileProcessor) isMarkdownSyntaxClosingChar (b byte) bool {
	char := string(b)
	return char == "]" || char == ")" || char == "´"
}

func (spfp spellCheckFileProcessor) isWordEndingChar (chars ...byte) bool {
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

func (spfp spellCheckFileProcessor) matchLetter (char string) bool {
	matched, _ := regexp.Match("[A-Za-z]", []byte(char))
	return matched
}

func (spfp spellCheckFileProcessor) isLookForwardAndBackChar (char string) bool {
	return char == "'" || char == "-"
}

func (spfp spellCheckFileProcessor) processWord (word string) string {
	spellingCorrect, suggestions := spfp.checkSpelling(word)
	if (spellingCorrect) {
		return word
	} else {
		//fmt.Printf("Incorrect word, suggestions: %s\n", s.Join(suggestions, ", "))

		// jvt: @todo jup....
		if len(suggestions) > 0 {
			fmt.Println("Replacing \"" + word + "\" with \"" + suggestions[0] + "\"")
			fmt.Printf("Alternative suggestions: %s\n", s.Join(suggestions, ", "))
			return suggestions[0]
		}

		return word
	}
}

func (spfp spellCheckFileProcessor) checkSpelling (word string) (bool, []string) {
	if spfp.spellChecker.Check(word) {
		//fmt.Print("OK\n")
		return true, nil
	}

	suggestions := spfp.spellChecker.Suggest(word)
	//fmt.Printf("Spelling mistake:\"" + word + "\" suggestions: %s\n", s.Join(suggestions, ", "))
	return false, suggestions
}
