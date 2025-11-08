package main

import (
	"github.com/urizennnn/boxify/internal/cli"
	"github.com/urizennnn/boxify/internal/legacy"
)

func main() {
	cli.InitCli()
}

func Oldmain() {
	legacy.Run()
}
