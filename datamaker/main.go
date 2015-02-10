package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	var token string
	flag.StringVar(&token, "t", "", "Token to enforce. Leave empty to keep original token.")
	flag.Parse()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		path := getClearPath(line)
		if token != "" {
			path = replaceToken(path, token)
		}
		fmt.Println(path)
	}
}

func replaceToken(path string, newToken string) string {
	// Replace the token
	pattern := regexp.MustCompile(`token=[a-zA-z0-9]+`)
	loc := pattern.FindStringIndex(path)
	if loc != nil {
		path = path[:loc[0]] + "token=" + newToken + path[loc[1]:]
	} else {
	}

	return path
}

func getClearPath(line string) string {
	array := strings.Split(line, " ")
	path := ""
	for i, work := range array {
		if work == "\"GET" {
			path = array[i+1]
			break
		}
	}

	return path
}
