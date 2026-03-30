package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/unidoc/unipdf/v3/model"
)

// generatePasswords generates all passwords of the given length iteratively using
// an index counter, avoiding recursion and repeated string allocations.
// Returns false if the done signal was received before exhausting all combinations.
func generatePasswords(charset []byte, length int, passwords chan<- []byte, done <-chan struct{}) bool {
	n := len(charset)
	indices := make([]int, length)
	buf := make([]byte, length)

	for {
		// Build password from current indices
		for i, idx := range indices {
			buf[i] = charset[idx]
		}
		pwd := make([]byte, length)
		copy(pwd, buf)

		select {
		case <-done:
			return false
		case passwords <- pwd:
		}

		// Increment the index counter (right-to-left, like an odometer)
		pos := length - 1
		for pos >= 0 {
			indices[pos]++
			if indices[pos] < n {
				break
			}
			indices[pos] = 0
			pos--
		}
		if pos < 0 {
			return true // exhausted all combinations of this length
		}
	}
}

func worker(filePath string, passwords <-chan []byte, found chan<- string, count *atomic.Int64, last *atomic.Value, wg *sync.WaitGroup, done <-chan struct{}) {
	defer wg.Done()

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

			// Update stats atomically — no channel overhead
			count.Add(1)
			last.Store(string(password))

			result, _ := pdfReader.Decrypt(password)
			if result {
				select {
				case found <- string(password):
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

	characters := []byte("0123456789")
	// Large buffer so the generator is never blocked waiting for workers
	passwords := make(chan []byte, numWorkers*500)
	found := make(chan string, 1)
	done := make(chan struct{})

	var count atomic.Int64
	var last atomic.Value

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(filePath, passwords, found, &count, &last, &wg, done)
	}

	// Generate passwords iteratively across increasing lengths
	go func() {
		defer close(passwords)
		for length := 1; ; length++ {
			select {
			case <-done:
				return
			default:
				if !generatePasswords(characters, length, passwords, done) {
					return
				}
			}
		}
	}()

	startTime := time.Now()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// Progress display — reads atomic values, no extra goroutine or channel needed
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if pwd, ok := last.Load().(string); ok && pwd != "" {
					n := count.Load()
					fmt.Printf("\rTrying: %s (speed: %.0f pwd/s, total: %d)    ",
						pwd, float64(n)/time.Since(startTime).Seconds(), n)
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
	fmt.Printf("Total passwords tested: %d\n", count.Load())

	wg.Wait()
}
