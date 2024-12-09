package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

var docs = `
sftm is a tool to show snowflow id in treemap
usage:
  echos id | sftm [[segment bits]...]
  sftm [[segment bits]...] -f file
`

type ParsingArgs struct {
	segmentsName []string
	segmentsBits []uint8
}

func (a ParsingArgs) transform(id uint32) (s string) {
	var left uint8 = 32
	for i, bits := range a.segmentsBits {
		left = left - bits
		mask := mask(bits, int(left))
		segment := (id & mask) >> left
		s += fmt.Sprintf("%d", segment)
		if i < len(a.segmentsBits)-1 {
			s += "/"
		}
	}
	if left > 0 {
		s += "/"

		mask := mask(left, 0)
		fmt.Println(left, 0, mask)
		segment := id & mask
		fmt.Println(segment)
		s += fmt.Sprintf("%d", segment)
	}
	return s
}

func mask(bits uint8, offset int) uint32 {
	return (uint32(1)<<bits - 1) << offset
}

func main() {
	if len(os.Args) == 0 {
		_, _ = os.Stdout.WriteString(docs)
		return
	}
	// determine segment and segment bits
	// the last arg with -f is file
	// if no -f, read from stdin

	fmt.Println(os.Args[0])

	var file string
	var segmentsName []string
	var segmentsbit []uint8
	for i := 1; i < len(os.Args); i = i + 2 {
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
			segmentsName = append(segmentsName, os.Args[i])
			segmentsbit = append(segmentsbit, uint8(atoi))
		}
	}

	if file != "" {
		// scan all lines from file
		f, err := os.Open(file)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer func(f *os.File) {
			_ = f.Close()
		}(f)
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			id, err := strconv.Atoi(line)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(ParsingArgs{
				segmentsName: segmentsName,
				segmentsBits: segmentsbit,
			}.transform(uint32(id)))
		}
		return
	}

	// scan lines from stdin
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		id, err := strconv.Atoi(line)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(ParsingArgs{
			segmentsName: segmentsName,
			segmentsBits: segmentsbit,
		}.transform(uint32(id)))
	}

}
