package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func Test() {
	files, err := ioutil.ReadDir("tests")
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		if f.Name() == "out" {
			continue
		}
		fmt.Printf("Testing file: [%s]\n", f.Name())
		got := make(map[string]struct{})
		for _, atom := range Run(fmt.Sprintf("tests/%s", f.Name())) {
			s := fmt.Sprintf("%s.", atom.String())
			got[s] = struct{}{}
		}
		file, err := os.Open(fmt.Sprintf("tests/out/%s", f.Name()))
		if err != nil {
			panic(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		// optionally, resize scanner's capacity for lines over 64K, see next example
		expected := make(map[string]struct{})
		for scanner.Scan() {
			expected[strings.ReplaceAll(strings.TrimSpace(scanner.Text()), " ", "")] = struct{}{}
		}
		if err := scanner.Err(); err != nil {
			panic(err)
		}
		failure := false
		for k := range expected {
			if _, found := got[k]; !found {
				failure = true
				fmt.Println("FAILURE")
				fmt.Printf("Expected %s, but did not get it\n", k)
			}
		}
		for k := range got {
			if _, found := expected[k]; !found {
				failure = true
				fmt.Println("FAILURE")
				fmt.Printf("Got %s, but did not expect it\n", k)
			}
		}
		if !failure {
			fmt.Println("PASS")
		}
		fmt.Println("----")
	}
}
