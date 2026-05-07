package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aryawadhwa/dike/pkg/repl"
	"github.com/aryawadhwa/dike/pkg/web"
	"github.com/aryawadhwa/dike/pkg/audit"
)

func main() {
	jsonFlag := flag.Bool("json", false, "Run in headless mode and output JSON")
	cmdFlag := flag.String("cmd", "", "Command to execute in headless mode")
	dirFlag := flag.String("dir", "", "Working directory for headless mode")
	webFlag := flag.Bool("web", false, "Start the web audit dashboard")
	portFlag := flag.Int("port", 8080, "Port for the web audit dashboard")
	flag.Parse()

	if err := audit.InitDB(); err != nil {
		fmt.Printf("Warning: Failed to initialize audit database: %v\n", err)
	}
	defer audit.Close()

	if *jsonFlag {
		if *cmdFlag == "" {
			fmt.Fprintln(os.Stderr, "Error: --cmd is required when using --json")
			os.Exit(1)
		}
		// Headless mode now logs to the DB since InitDB was called above.
		repl.HeadlessExecute(*cmdFlag, *dirFlag)
		os.Exit(0)
	}

	if *webFlag {
		go func() {
			if err := web.StartServer(*portFlag); err != nil {
				fmt.Fprintf(os.Stderr, "Web server error: %v\n", err)
			}
		}()
	}

	repl.Start()
}
