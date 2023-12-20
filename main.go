package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

var usageDetails = `
If no argument is given, the amounts of occurrences of each status code
are counted and displayed.

If one or more arguments are given, lines are filtered and printed only
if they match the filter. Arguments may either be single status codes,
like '200' or ranges, like '200:299'.

If the -v flag is given, the filter is inverted. Only lines with status
codes, that don't match the given arguments, will be printed.
`

var vFlag bool

// This is incomplete, but more is not needed here.
type httpline struct {
	Resp string
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [-v] [FILTER]...\n", os.Args[0])
		fmt.Fprint(flag.CommandLine.Output(), usageDetails)
	}
	flag.BoolVar(&vFlag, "v", false, "Invert filter.")
	flag.Parse()
}

func main() {
	if len(flag.Args()) == 0 {
		printStats()
		return
	}
	filter()
}

func printStats() {
	invalidLines := 0
	statusCodeCounts := make(map[int]int)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		statusCode, err := extractStatusCode(scanner.Bytes())
		if err != nil {
			invalidLines++
		} else {
			statusCodeCounts[statusCode]++
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Reading standard input failed: %v\n", err)
		os.Exit(1)
	}
	if invalidLines > 0 {
		fmt.Println("Invalid lines:", invalidLines)
	}
	seenStatusCodes := make([]int, 0, len(statusCodeCounts))
	for k, _ := range statusCodeCounts {
		seenStatusCodes = append(seenStatusCodes, k)
	}
	sort.Ints(seenStatusCodes)
	for _, statusCode := range seenStatusCodes {
		fmt.Printf("Status code %d: %dx\n", statusCode, statusCodeCounts[statusCode])
	}
}

func filter() {
	desiredStatusCodes := getDesiredStatusCodes()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		statusCode, err := extractStatusCode(scanner.Bytes())
		if _, ok := desiredStatusCodes[statusCode]; err == nil && ok {
			fmt.Println(scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Reading standard input failed: %v\n", err)
		os.Exit(1)
	}
}

func getDesiredStatusCodes() map[int]bool {
	desiredStatusCodes := make(map[int]bool)
	if vFlag {
		for i := 100; i < 600; i++ {
			desiredStatusCodes[i] = true
		}
	}
	for _, arg := range flag.Args() {
		split := strings.Split(arg, ":")
		if len(split) == 1 {
			statusCode, err := strconv.Atoi(arg)
			if err != nil || statusCode < 100 || statusCode > 599 {
				fmt.Fprintf(os.Stderr, "Error: Invalid argument '%s'.\n", arg)
				flag.Usage()
				os.Exit(1)
			}
			if vFlag {
				delete(desiredStatusCodes, statusCode)
			} else {
				desiredStatusCodes[statusCode] = true
			}
		} else if len(split) == 2 {
			minStatusCode, err := strconv.Atoi(split[0])
			if err != nil || minStatusCode < 100 || minStatusCode > 599 {
				fmt.Fprintf(os.Stderr, "Error: Invalid argument '%s'.\n", arg)
				flag.Usage()
				os.Exit(1)
			}
			maxStatusCode, err := strconv.Atoi(split[1])
			if err != nil || maxStatusCode < 100 || maxStatusCode > 599 || minStatusCode > maxStatusCode {
				fmt.Fprintf(os.Stderr, "Error: Invalid argument '%s'.\n", arg)
				flag.Usage()
				os.Exit(1)
			}
			for statusCode := minStatusCode; statusCode <= maxStatusCode; statusCode++ {
				if vFlag {
					delete(desiredStatusCodes, statusCode)
				} else {
					desiredStatusCodes[statusCode] = true
				}
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error: Invalid argument '%s'.\n", arg)
			flag.Usage()
			os.Exit(1)
		}
	}
	return desiredStatusCodes
}

func extractStatusCode(line []byte) (int, error) {
	var hl httpline
	err := json.Unmarshal(line, &hl)
	if err != nil {
		return 0, err
	} else if hl.Resp == "" {
		return 0, fmt.Errorf("response is empty")
	}
	statusLine := strings.SplitN(hl.Resp, "\n", 2)[0]
	fields := strings.Fields(statusLine)
	if len(fields) < 2 {
		return 0, fmt.Errorf("invalid status line")
	}
	statusCode, err := strconv.Atoi(fields[1])
	if err == nil && (statusCode < 100 || statusCode > 599) {
		err = fmt.Errorf("invalid status code %d", statusCode)
	}
	return statusCode, err
}
