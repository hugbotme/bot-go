package main

import (
	"fmt"
	"log"
	//"gopkg.in/libgit2/git2go.v22"
	"github.com/hugbotme/bot-go/parser"
)


func processHug(url *Hug) {
	fmt.Println("parsing repository: " + url.URL)

	repoName := "karban"

	parser := parser.NewParser("mre", repoName)

	lines, err := parser.GetFileContents("Readme.md")

	if err != nil {
		log.Printf("Error during clone: %v\n", err)
	} else {
		for i, line := range lines {
			fmt.Println(i, line)
		}
	}

	branchname := "bugfix"

	// TODO: ERROR HANDLING
	branch, err := parser.CreateBranch(branchname)
	parser.CommitFile(branch, branchname, "Readme.md", lines, "Fixing some typos")
	parser.PullRequest(branchname, "A friendly pull request")

	files := []string{
		"test string one",
		"another awesome test string",
	}

	processor, err := newSpellCheckFileProcessor()
	if err != nil {
		fmt.Errorf("could not get speller: %s", err.Error())
		return
	}

	for _, file := range files {
		content := processor.processContent([]byte(file))
		fmt.Println("corrected content: " + string(content))
	}
}
