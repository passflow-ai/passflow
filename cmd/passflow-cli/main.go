package main

import "github.com/passflow-ai/passflow/cmd/passflow-cli/cmd"

var Version = "dev"

func main() {
	cmd.SetVersion(Version)
	cmd.Execute()
}
