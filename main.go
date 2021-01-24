// Copyright 2021 Chris Thunes
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

type StackDump struct {
	Text string
}

type Thread struct {
	Header string
	Name   string
	Stack  string
}

func (dump *StackDump) parseThread(lines []string) (*Thread, error) {
	header := lines[0]
	stack := strings.Join(lines[1:], "\n")

	// Extract name and remove quotes
	startName := strings.IndexByte(header, '"')
	endName := strings.LastIndexByte(header, '"')
	if startName < 0 || endName <= startName {
		return nil, errors.New("no thread name found")
	}
	name := header[startName+1 : endName]

	thread := &Thread{
		Header: header,
		Name:   name,
		Stack:  stack,
	}

	return thread, nil
}

func (dump *StackDump) ParseThreads() ([]*Thread, error) {
	var threads []*Thread
	var thread []string
	for _, l := range strings.Split(dump.Text, "\n") {
		if len(l) > 0 && (l[0] == ' ' || l[0] == '\t') {
			if len(thread) > 0 {
				thread = append(thread, l)
			}
		} else {
			if len(thread) > 0 {
				parsed, err := dump.parseThread(thread)
				if err != nil {
					return nil, err
				}
				threads = append(threads, parsed)
				thread = nil
			}
			if len(l) > 0 && strings.Contains(l, "nid=") && strings.Contains(l, "tid=") {
				thread = append(thread, l)
			}
		}
	}
	if len(thread) > 0 {
		parsed, err := dump.parseThread(thread)
		if err != nil {
			return nil, err
		}
		threads = append(threads, parsed)
	}
	return threads, nil
}

func main() {
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		usage := "usage: %s [-c] [-v] <pattern>\n"
		fmt.Fprintf(out, usage, os.Args[0])
		flag.PrintDefaults()
	}

	usageError := func(format string, a ...interface{}) {
		fmt.Fprintf(os.Stderr, "error: "+format, a...)
		fmt.Fprint(os.Stderr, "\n\n")
		flag.Usage()
		os.Exit(1)
	}

	invert := false
	countOnly := false
	matchName := false

	flag.BoolVar(&invert, "v", invert, "invert matching")
	flag.BoolVar(&countOnly, "c", countOnly, "print number of matching threads")
	flag.BoolVar(&matchName, "name", matchName, "match on thread name only")
	flag.Parse()

	if flag.NArg() < 1 {
		usageError("missing argument")
	} else if flag.NArg() > 1 {
		usageError("too many arguments")
	}

	pattern, err := regexp.Compile(flag.Arg(0))
	if err != nil {
		log.Fatalf("invalid pattern: %v", err)
	}

	stackBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("error reading stdin: %v", err)
	}

	stack := &StackDump{string(stackBytes)}
	threads, err := stack.ParseThreads()

	if err != nil {
		log.Fatalf("error parsing stack dump: %v", err)
	}

	count := 0
	first := true
	for _, thread := range threads {
		text := thread.Header
		if len(thread.Stack) > 0 {
			text += "\n" + thread.Stack
		}

		var matches bool
		if matchName {
			matches = pattern.MatchString(thread.Name)
		} else {
			matches = pattern.MatchString(text)
		}

		if matches != invert {
			if countOnly {
				count += 1
			} else {
				if first {
					first = false
				} else {
					fmt.Println()
				}
				fmt.Println(text)
			}
		}
	}
	if countOnly {
		fmt.Println(count)
	}
}
