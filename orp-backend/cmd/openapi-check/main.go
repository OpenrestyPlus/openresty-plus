package main

import (
	"fmt"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"

	"net.daoke/orp-backend/internal/httpapi"
)

var pathEntry = regexp.MustCompile(`(?m)^    '(/[^']+)': \{`)
var methodEntry = regexp.MustCompile(`(?m)^      (get|post|put|delete|patch): \{`)
var pathParam = regexp.MustCompile(`\{[^}]+\}`)

func main() {
	source, err := os.ReadFile("../orp-frontend/apps/web-antd/src/api/orp/openapi-spec.ts")
	if err != nil {
		fmt.Fprintln(os.Stderr, "run this command from the orp-backend directory:", err)
		os.Exit(2)
	}
	entries := pathEntry.FindAllSubmatchIndex(source, -1)
	mux := httpapi.NewMux(nil)
	missing := []string{}
	for index, entry := range entries {
		end := len(source)
		if index+1 < len(entries) {
			end = entries[index+1][0]
		}
		path := string(source[entry[2]:entry[3]])
		block := source[entry[1]:end]
		methods := methodEntry.FindAllSubmatch(block, -1)
		for _, method := range methods {
			verb := strings.ToUpper(string(method[1]))
			route := "/api" + pathParam.ReplaceAllString(path, "1")
			request := httptest.NewRequest(verb, route, nil)
			_, pattern := mux.Handler(request)
			if pattern == "" {
				missing = append(missing, verb+" "+path)
			}
		}
	}
	if len(missing) > 0 {
		fmt.Fprintln(os.Stderr, "OpenAPI operations without a Go route:")
		for _, item := range missing {
			fmt.Fprintln(os.Stderr, " -", item)
		}
		os.Exit(1)
	}
	fmt.Printf("OpenAPI route check passed (%d paths).\n", len(entries))
}
