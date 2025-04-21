package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/davecgh/go-spew/spew"
	"go.temporal.io/sdk/client"
)

func main() {
	fmt.Println("Welcome to kilcron Demo ...")
	Run()
}

var c client.Client

func Run() {
	// Create a new ServeMux
	mux := http.NewServeMux()

	// Define a handler function
	defaultHandler := func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/demo/debug/", http.StatusFound)
		return
	}

	// Attach handler function to the ServeMux
	mux.HandleFunc("/", defaultHandler)
	mux.HandleFunc("/demo/debug/", debugAccessHandler)

	// HTTP Server Setup ..
	server := &http.Server{
		Addr:    ":8888",
		Handler: mux,
	}

	// Running the HTTP server in a go routine
	go func() {
		if err := server.ListenAndServe(); err != nil {
			fmt.Println("Server error:", err)
		}
	}()

	// Create the Temporal client
	var err error
	c, err = client.NewLazyClient(client.Options{})
	if err != nil {
		spew.Dump(err)
		log.Fatalln("Unable to create Temporal client", err)
	}
	defer c.Close()

	// Prepare for handling signals
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Wait for interrupt signal
	interruptSignal := <-interrupt
	fmt.Printf("Received %s, shutting down.\n", interruptSignal)

	// Shutdown the server gracefully
	if err := server.Shutdown(context.Background()); err != nil {
		fmt.Println("Error shutting down:", err)
	} else {
		fmt.Println("Server shutdown gracefully.")
	}

	return
}
