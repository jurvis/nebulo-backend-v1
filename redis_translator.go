package main

import (
	"fmt"
	"os"
	"bufio"
	"log"
	"strings"
)

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func main() {
	pwd, _ := os.Getwd()
	fmt.Println("Reading lines from source.txt")
	lines, er := readLines(fmt.Sprintf(pwd + "/" + "source.txt"))
	if er != nil {
		log.Fatal("Error reading lines from source.txt")
	}
	os.Remove(fmt.Sprintf(pwd + "/" + "result.txt"))
	f, err := os.OpenFile(fmt.Sprintf(pwd + "/" + "result.txt"), os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	defer f.Close()
	if err != nil {
		return
	}
	w := bufio.NewWriter(f)
	for _, line := range lines {
		words := strings.Split(line, " ")
		statement := fmt.Sprintf("*%d\r\n", len(words))
		for i := 0; i < len(words); i++ {
			word := words[i]
			statement += fmt.Sprintf("$%d\r\n%s\r\n", len(word), word)
		}
		w.WriteString(statement)
	}
	w.Flush()
}