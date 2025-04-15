package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gothicenemy/software-architecture-3/painter"
	"github.com/gothicenemy/software-architecture-3/painter/lang"
	"github.com/gothicenemy/software-architecture-3/ui"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
)

const HttpPort = ":17000"

func main() {
	log.Println("Starting Painter application (final structure)...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var (
		painterLoop *painter.Loop
		window      *ui.Window
		server      *http.Server
	)

	driverDone := make(chan struct{})
	shutdownRequest := make(chan struct{})

	go func() {
		select {
		case sig := <-sigChan:
			log.Printf("Received OS signal: %v. Initiating shutdown...", sig)
			close(shutdownRequest)
		case <-driverDone:
			log.Println("Signal handler noticed driver finished.")
		}
	}()

	go func() {
		time.Sleep(200 * time.Millisecond)
		if painterLoop == nil {
			log.Println("HTTP Server: Painter loop is nil, server not starting.")
			return
		}

		parser := &lang.Parser{}
		server = &http.Server{Addr: HttpPort}

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if painterLoop == nil {
				http.Error(w, "Painter loop not ready", http.StatusServiceUnavailable)
				return
			}
			if r.Method != http.MethodPost {
				http.Error(w, "Only POST method is accepted", http.StatusMethodNotAllowed)
				return
			}
			defer r.Body.Close()
			cmds, err := parser.Parse(r.Body)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error parsing commands: %v", err), http.StatusBadRequest)
				return
			}
			for _, cmd := range cmds {
				painterLoop.Post(cmd)
			}
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "Commands processed")
		})

		log.Printf("Starting HTTP server on port %s", HttpPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server ListenAndServe error: %v", err)
			select {
			case <-shutdownRequest:
			default:
				close(shutdownRequest)
			}
		}
		log.Println("HTTP server stopped.")
	}()

	driver.Main(func(s screen.Screen) {
		log.Println("Driver started.")
		painterLoop = painter.NewLoop(s)
		go painterLoop.Start()
		log.Println("Painter loop created and started.")

		window = ui.NewWindow(s, painterLoop)
		if window == nil {
			log.Println("Window creation failed.")
			return
		}

		go func() {
			select {
			case <-shutdownRequest:
				log.Println("Shutdown request received, stopping window...")
				window.Stop()
			case <-window.Closed():
				log.Println("Window closed by user, signaling shutdown.")
				select {
				case <-shutdownRequest:
				default:
					close(shutdownRequest)
				}
			}
		}()

		window.Loop()
		log.Println("Window loop finished.")
	})

	close(driverDone)

	log.Println("Starting graceful shutdown...")

	if server != nil {
		log.Println("Shutting down HTTP server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		} else {
			log.Println("HTTP server gracefully stopped.")
		}
		cancel()
	} else {
		log.Println("HTTP server was not running.")
	}

	if painterLoop != nil {
		log.Println("Requesting painter loop stop...")
		painterLoop.Stop()
		log.Println("Painter loop confirmed stopped.")
	}

	log.Println("Painter application finished.")
}
