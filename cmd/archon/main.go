package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var (
		projectPath string
	)
	flag.StringVar(&projectPath, "project", "", "Path to an Archon project (optional)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Archon CLI (MVP)\n")
		fmt.Fprintf(os.Stderr, "Usage: archon [--project path] <command> [args]\n")
		fmt.Fprintf(os.Stderr, "Commands: open | index | snapshot | export (stubs)\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}
	cmd := flag.Arg(0)
	switch cmd {
	case "open", "index", "snapshot", "export":
		fmt.Println("(stub)", cmd, projectPath)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		os.Exit(2)
	}
}
