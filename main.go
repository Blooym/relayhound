package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/gosuri/uilive"
)

type Status int64

const (
	NoData          int = 0
	Found           int = 1
	ConnectionError int = 2
)

type HostStatus struct {
	URL    string
	Status Status
}

func main() {
	// Init
	var hosts []string
	flag.Func("hosts", "WebSocket URL(s) to monitor for the target data", func(s string) error {
		hosts = append(hosts, s)
		return nil
	})
	targetFlag := flag.String("target", "", "Target data to search for")
	flag.Parse()
	if len(hosts) == 0 {
		log.Fatal("--hosts is required")
	}
	if *targetFlag == "" {
		log.Fatal("--target is required")
	}
	target := *targetFlag
	fmt.Printf("Configured hosts: %v\n", hosts)
	fmt.Printf("Target data: %v\n", target)
	fmt.Println()

	// Initialize status tracking
	statuses := make([]*HostStatus, len(hosts))
	for i, host := range hosts {
		statuses[i] = &HostStatus{
			URL:    host,
			Status: Status(NoData),
		}
	}
	writer := uilive.New()
	writer.Start()
	defer writer.Stop()
	var statusMtx sync.Mutex
	updateChan := make(chan struct{}, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signal
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		cancel()
	}()

	// Status updater
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-updateChan:
				statusMtx.Lock()
				for _, status := range statuses {
					var icon string
					switch status.Status {
					case Status(Found):
						icon = "FOUND:"
					case Status(NoData):
						icon = "NO_DATA:"
					case Status(ConnectionError):
						icon = "CONN_ERR:"
					}
					fmt.Fprintf(writer, "%s %s\n", icon, status.URL)
				}
				statusMtx.Unlock()
			}
		}
	}()
	updateChan <- struct{}{}

	// Begin connections to each relay.
	var wg sync.WaitGroup
	for i, host := range hosts {
		wg.Add(1)
		go func(wsURL string, index int) {
			defer wg.Done()
			result := listenWebSocket(ctx, wsURL, target)

			statusMtx.Lock()
			statuses[index].Status = result
			statusMtx.Unlock()

			select {
			case updateChan <- struct{}{}:
			default:
			}
		}(host, i)
	}
	wg.Wait()
}

func listenWebSocket(ctx context.Context, wsURL string, target string) Status {
	fullURL := wsURL + "/xrpc/com.atproto.sync.subscribeRepos"
	conn, _, err := websocket.DefaultDialer.Dial(fullURL, nil)
	if err != nil {
		return Status(ConnectionError)
	}
	defer conn.Close()

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return Status(NoData)
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				return Status(ConnectionError)
			}
			messageStr := string(message)
			if strings.Contains(messageStr, target) {
				return Status(Found)
			}
		}
	}
}
