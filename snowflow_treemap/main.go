package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var docs = `
sftm is a tool to show snowflow id in treemap
usage:
  echos id | sftm [[segment bits]...]
  sftm [[segment bits]...] -f file
`

func main() {
	if len(os.Args) == 0 {
		_, _ = os.Stdout.WriteString(docs)
		return
	}

	file, args := parse(os.Args)
	args.scan(file, args.transform, printId)

}

type ParsingArgs struct {
	segmentsName []string
	segmentsBits []uint8
}

func (a ParsingArgs) transform(idstr string) string {
	id, err := strconv.Atoi(idstr)
	if err != nil {
		return idstr
	}
	var segment []string
	var left uint8 = 32
	cuts := a.segmentsBits
	for {
		if left <= 0 {
			break
		}
		var bits uint8
		if len(cuts) <= 0 {
			bits = left
		} else {
			bits = cuts[0]
			cuts = cuts[1:]
		}
		left = left - bits
		sprintf := cut(id, bits, int(left))
		segment = append(segment, sprintf)
	}
	return strings.Join(segment, "/")
}

func cut(id int, bits uint8, rightOffset int) string {
	mask := mask(bits, rightOffset)
	segmentNum := uint32(id) & mask >> rightOffset
	sprintf := fmt.Sprintf("%d", segmentNum)
	return sprintf
}

func mask(bits uint8, offset int) uint32 {
	return (uint32(1)<<bits - 1) << offset
}

// scan callback each lines
func (a ParsingArgs) scan(file string, callback ...func(id string) string) {
	var stdin = fileOrStdin(file)
	defer func(f *os.File) {
		_ = f.Close()
	}(stdin)
	scanner := bufio.NewScanner(stdin)
	for scanner.Scan() {
		line := scanner.Text()
		for _, f := range callback {
			line = f(line)
		}
	}
}

// determine segment and segment bits
// the last arg with -f is file
// if no -f, read from stdin
func parse(argstr []string) (file string, args ParsingArgs) {

	for i := 1; i < len(argstr); i = i + 2 {
		if os.Args[i] == "-f" {
			// read from file
			file = os.Args[i+1]
			continue
		} else {
			atoi, err := strconv.Atoi(os.Args[i+1])
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			args.segmentsName = append(args.segmentsName, os.Args[i])
			args.segmentsBits = append(args.segmentsBits, uint8(atoi))
		}
	}
	return
}

func fileOrStdin(file string) *os.File {

	if file != "" {
		return openFile(file)
	} else {
		return os.Stdin
	}
}

var printId = func(id string) string {
	fmt.Println(id)
	return id
}

func openFile(file string) *os.File {
	// scan all lines from file
	f, err := os.Open(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return f
}
