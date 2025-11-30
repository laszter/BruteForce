package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/unidoc/unipdf/v3/model"
)

func generatePasswords(charset []rune, length int, currentPassword string, passwords chan string) {
	if length == 0 {
		passwords <- currentPassword
	} else {
		for _, c := range charset {
			generatePasswords(charset, length-1, currentPassword+string(c), passwords)
		}
	}
}

func worker(_ int, filePath string, passwords <-chan string, found chan<- string, progress chan<- string, wg *sync.WaitGroup, done <-chan struct{}) {
	defer wg.Done()

	// Each worker opens its own PDF reader
	f, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		return
	}

	for {
		select {
		case <-done:
			return
		case password, ok := <-passwords:
			if !ok {
				return
			}
			// Send progress update
			select {
			case progress <- password:
			default:
			}

			result, _ := pdfReader.Decrypt([]byte(password))
			if result {
				select {
				case found <- password:
				case <-done:
				}
				return
			}
		}
	}
}

func main() {
	// Check if the file path is provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: bruteforce <pdf_file_path>")
		return
	}
	filePath := os.Args[1]

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Println("Error: File does not exist:", filePath)
		return
	}

	numWorkers := runtime.NumCPU()
	fmt.Printf("Starting brute force with %d workers...\n", numWorkers)

	characters := []rune("0123456789")
	passwords := make(chan string, numWorkers*2)
	found := make(chan string, 1)
	progress := make(chan string, numWorkers*10)
	done := make(chan struct{})

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(i, filePath, passwords, found, progress, &wg, done)
	}

	// Generate passwords
	go func() {
		defer close(passwords)
		length := 1
		for {
			select {
			case <-done:
				return
			default:
				generatePasswords(characters, length, "", passwords)
				length++
			}
		}
	}()

	startTime := time.Now()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	lastPassword := ""
	passwordCount := 0

	// Monitor progress
	go func() {
		for {
			select {
			case <-done:
				return
			case pwd := <-progress:
				lastPassword = pwd
				passwordCount++
			case <-ticker.C:
				if lastPassword != "" {
					fmt.Printf("\rTrying password: %s (speed: %.0f pwd/s)    ",
						lastPassword, float64(passwordCount)/time.Since(startTime).Seconds())
				}
			}
		}
	}()

	// Wait for result
	password := <-found
	close(done)
	fmt.Printf("\r%s\n", strings.Repeat(" ", 80))
	fmt.Printf("Password found: %s\n", password)
	fmt.Printf("Time taken: %s\n", time.Since(startTime))
	fmt.Printf("Total passwords tested: %d\n", passwordCount)

	wg.Wait()
}
