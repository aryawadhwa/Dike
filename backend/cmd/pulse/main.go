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
	webFlag := flag.Bool("web", false, "Start the web audit dashboard")
	portFlag := flag.Int("port", 8080, "Port for the web audit dashboard")
	flag.Parse()

	if err := audit.InitDB(); err != nil {
		fmt.Printf("Warning: Failed to initialize audit database: %v\n", err)
	}
	defer audit.Close()

	if *webFlag {
		go func() {
			if err := web.StartServer(*portFlag); err != nil {
				fmt.Fprintf(os.Stderr, "Web server error: %v\n", err)
			}
		}()
	}

	repl.Start()
}
