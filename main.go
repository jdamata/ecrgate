package main

import (
	cmd "github.com/jdamata/ecrgate/cmd"
)

var version = "dev"

func main() {
	cmd.Execute(version)
}
