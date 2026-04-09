package main

import "github.com/jaak-ai/passflow-cli/cmd"

var Version = "dev"

func main() {
	cmd.SetVersion(Version)
	cmd.Execute()
}
