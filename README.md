# PDF Password Brute Force
This is a command-line tool for cracking passwords of PDF files. The tool generates all possible combinations of numbers of a specified length, and tries each combination as the password for the PDF file until a match is found.

## Getting Started

### Prerequisites
- Go (version 1.16 or later)
- [Unidoc](https://github.com/unidoc/unipdf) (version 3 or later)

### Installing
1. Clone the repository

```git
git clone https://github.com/laszter/BruteForce.git
```

2. Install Unidoc

```go
go get github.com/unidoc/unipdf/v3
```

### Usage
```go
go run main.go <pdf_file_path>
```

Replace `<pdf_file_path>` with the path to the PDF file you want to crack. The tool will start generating passwords and trying them against the PDF file until a match is found. Once a match is found, the password and the time taken will be printed to the console output.

Note: This tool is only intended for use on PDF files that you have permission to access. Unauthorized use of this tool may be illegal.

### Customization
You can customize the password length by changing the `length` variable in the `main` function. You can also customize the character set used to generate passwords by changing the `characters` variable.

### Contributing
Contributions are welcome! If you find a bug or want to suggest an improvement, please open an issue or a pull request.

### License
This project is licensed under the MIT License - see the [LICENSE](https://choosealicense.com/licenses/mit/) file for details.