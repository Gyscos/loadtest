package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Find the requested URL in nginx logs, and change the tokens with a fixed one.
func main() {
	var token string
	flag.StringVar(&token, "t", "", "Token to enforce. Leave empty to keep original token.")
	flag.Parse()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		path := getClearPath(line)
		path = apifyPath(path)
		if path == "" {
			continue
		}
		if token != "" {
			path = replaceToken(path, token)
		}
		fmt.Println(path)
	}
}

func apifyPath(path string) string {
	if strings.HasPrefix(path, "/api") {
		return path
	}

	if strings.HasPrefix(path, "/v2") {
		return "/api" + path[len("/v2"):] + "&version=2"
	}

	if strings.HasPrefix(path, "/v3") {
		return "/api" + path[len("/v3"):] + "&version=3"
	}

	return ""
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
	if strings.HasPrefix(line, "/") {
		return line
	}
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
