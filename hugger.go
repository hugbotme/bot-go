package main

import(
	"fmt"
)

func hug(url string) {
	fmt.Println("parsing repository: " + url)

	githubClient := getGithubClient()




	// jvt: @todo parse repository
	files := []string{
		"test string one",
		"another awesome test string",
	}

	for _, file := range files {
		content := spellCheck(file)
		fmt.Println("corrected content: " + content)
	}
}
