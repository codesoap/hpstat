package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

var usageDetails = `Usage:
    %s
    %s -i
    %s [-v] FILTER...

If no argument is given, the amounts of occurrences of each status code
are counted and displayed.

If the -i flag is given, all invalid lines will be printed.

If one or more arguments are given, lines are filtered and printed only
if they match the filter. Arguments may either be single status codes,
like '200' or ranges, like '200:299'.

If the -v flag is given, the filter is inverted. Only lines with status
codes, that don't match the given arguments, will be printed.
`

var iFlag bool
var vFlag bool

// This is incomplete, but more is not needed here.
type httpline struct {
	Resp string
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usageDetails, os.Args[0], os.Args[0], os.Args[0])
		os.Exit(1)
	}
	flag.BoolVar(&iFlag, "i", false, "Show invalid requests.")
	flag.BoolVar(&vFlag, "v", false, "Invert filter.")
	flag.Parse()
	if iFlag && (len(flag.Args()) != 0 || vFlag) {
		flag.Usage()
	}
}

func main() {
	if iFlag {
		printInvalidLines()
	}
	if len(flag.Args()) == 0 {
		printStats()
	} else {
		filter()
	}
}

func printStats() {
	invalidLines := 0
	statusCodeCounts := make(map[int]int)
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Reading standard input failed: %v\n", err)
			os.Exit(1)
		}
		statusCode, err := extractStatusCode(line)
		if err != nil {
			invalidLines++
		} else {
			statusCodeCounts[statusCode]++
		}
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

func printInvalidLines() {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Reading standard input failed: %v\n", err)
			os.Exit(1)
		}
		if _, err := extractStatusCode(line); err != nil {
			fmt.Print(string(line))
		}
	}
}

func filter() {
	desiredStatusCodes := getDesiredStatusCodes()
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Reading standard input failed: %v\n", err)
			os.Exit(1)
		}
		statusCode, err := extractStatusCode(line)
		if _, ok := desiredStatusCodes[statusCode]; err == nil && ok {
			fmt.Print(string(line))
		}
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
			}
			maxStatusCode, err := strconv.Atoi(split[1])
			if err != nil || maxStatusCode < 100 || maxStatusCode > 599 || minStatusCode > maxStatusCode {
				fmt.Fprintf(os.Stderr, "Error: Invalid argument '%s'.\n", arg)
				flag.Usage()
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
