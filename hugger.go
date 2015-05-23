package main

import(
	"fmt"
	//"gopkg.in/libgit2/git2go.v22"
	"github.com/hugbotme/bot-go/parser"
)

func hug(url string) {
	fmt.Println("parsing repository: " + url)

	parser := parser.NewParser()

	parser.ForkRepository("mre", "beacon")

	files := []string{
		"test string one",
		"another awesome test string",
	}

	for _, file := range files {
		content := spellCheck(file)
		fmt.Println("corrected content: " + content)
	}
}
