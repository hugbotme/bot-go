package main

import(
	"fmt"
	"log"
	//"gopkg.in/libgit2/git2go.v22"
	"github.com/hugbotme/bot-go/parser"
	"github.com/libgit2/git2go"
)

func hug(url string) {
	fmt.Println("parsing repository: " + url)

	parser := parser.NewParser()

	repoName := "karban"

	repo, _, err := parser.ForkRepository("mre", repoName)

	if err != nil {
		log.Printf("Error during fork: %v\n", err)
	}

	log.Printf("Forked repo:" + *repo.CloneURL)

	repoClone, err := git.Clone(*repo.CloneURL, "cloned_projects/" + repoName, &git.CloneOptions{})


if err != nil {
log.Printf("Error during clone: %v\n", err)
}


	log.Printf("%v", repoClone)

lines, err := parser.ReadLines("project-clones/karban/README.md")
if err != nil {
log.Fatalf("ReadLines: %s", err)
}
for i, line := range lines {
fmt.Println(i, line)
}




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
		content := processor.processContent(file)
		fmt.Println("corrected content: " + content)
	}
}
