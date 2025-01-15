package main

import "wh40k/cmd"

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
