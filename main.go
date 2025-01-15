package main

import (
	"github.com/tmc/langchaingo/textsplitter"
	"io"
	"os"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	f, err := os.Open("output.md")
	if err != nil {
		panic(err)
	}
	s := textsplitter.NewMarkdownTextSplitter()
	b, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	tokens, err := s.SplitText(string(b))
	if err != nil {
		panic(err)
	}
	for _, token := range tokens {
		println(token)
	}

}
