package main

import (
	"fmt"
	"os"
	"strings"
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

func main() {
	// Check if the file path is provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: bruteforce <pdf_file_path>")
		return
	}
	filePath := os.Args[1]

	// Open the PDF file with a known password
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		fmt.Println("Error reading PDF:", err)
		return
	}

	characters := []rune("0123456789")
	length := 1
	passwords := make(chan string)

	go func() {
		for {
			generatePasswords(characters, length, "", passwords)
			length++
		}
	}()

	startTime := time.Now()
	progress := ""
	for password := range passwords {
		result, _ := pdfReader.Decrypt([]byte(password)) // Check the password here
		if result {
			fmt.Printf("\r%s", strings.Repeat(" ", len(progress)))
			fmt.Printf("\rPassword: %s\nTime taken: %s\n", password, time.Since(startTime))
			break
		}
		progress = fmt.Sprintf("Trying password: %s", password)
		fmt.Printf("\r%s", progress)
	}
}
