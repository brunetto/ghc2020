package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func main() {

	files := []string{
		"a_example",
		"b_small",
		"c_medium",
		"d_quite_big",
		"e_also_big",
	}

	for _, fn := range files {
		run(fn)
	}
}

func run(fn string) {
	// read data
	in, err := os.Open(fn + ".in")
	dieIf(err)

	s := bufio.NewScanner(in)

	if !s.Scan() {
		dieIf(errors.New("failed on first line"))
	}

	fields := lineToIntSlice(s.Text())
	totSlices := fields[0]

	if !s.Scan() {
		dieIf(errors.New("failed on second line"))
	}

	pizzas := lineToIntSlice(s.Text())
	var order []string

	var sum int
	for i := len(pizzas) - 1; i >= 0; i-- {
		if sum+pizzas[i] > totSlices {
			continue
		}
		sum += pizzas[i]
		order = append(order, strconv.Itoa(i))
	}

	fmt.Printf("points: %v, max: %v, difference: %v\n", sum, totSlices, totSlices-sum)

	out, err := os.Create(fn + ".out")
	dieIf(err)
	defer out.Close()

	_, err = out.WriteString(strconv.Itoa(len(order)) + "\n")
	dieIf(err)

	_, err = out.WriteString(strings.Join(order, " ") + "\n")
	dieIf(err)
}

func dieIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func lineToInt64Slice(line string) []int64 {
	fields := strings.Fields(line)
	out := make([]int64, 0, len(fields))
	for _, field := range fields {
		num, err := strconv.ParseInt(field, 10, 64)
		dieIf(err)

		out = append(out, num)
	}
	return out
}
func lineToIntSlice(line string) []int {
	fields := strings.Fields(line)
	out := make([]int, 0, len(fields))
	for _, field := range fields {
		num, err := strconv.ParseInt(field, 10, 64)
		dieIf(err)

		out = append(out, int(num))
	}
	return out
}
