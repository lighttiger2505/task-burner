package ui

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
)

func Ask(query string) (string, error) {
	var l sync.Mutex
	l.Lock()
	defer l.Unlock()

	var scanner *bufio.Scanner
	if scanner == nil {
		scanner = bufio.NewScanner(os.Stdin)
	}
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	if query != "" {
		if _, err := fmt.Fprint(os.Stdout, query+" "); err != nil {
			return "", err
		}
	}

	result := make(chan string, 1)
	go func() {
		var line string
		if scanner.Scan() {
			line = scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			log.Printf("ui: scan err: %s", err)
			return
		}
		result <- line
	}()

	select {
	case line := <-result:
		return line, nil
	case <-sigCh:
		// Print a newline so that any further output starts properly
		// on a new line.
		fmt.Fprintln(os.Stdout)
		return "", errors.New("interrupted")
	}
}
