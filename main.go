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

	fields := lineToInt64Slice(s.Text())
	totSlices := fields[0]
	nTypes := fields[1]

	if !s.Scan() {
		dieIf(errors.New("failed on second line"))
	}

	pizzas := lineToInt64Slice(s.Text())

	// check data
	if int64(len(pizzas)) != nTypes {
		dieIf(errors.Errorf("expected %v types of pizzas, got %v", nTypes, len(pizzas)))
	}

	var sum, total, id int64
	for i, n := range pizzas {
		sum += n
		if sum > totSlices {
			break
		}
		id = int64(i)
		total = sum
	}
	if sum < totSlices {
		dieIf(errors.Errorf("not enough slices in the pizza types: %v < %v", sum, totSlices))
	}

	fmt.Println(id, total)

	out, err := os.Create(fn + ".out")
	dieIf(err)
	defer out.Close()

	_, err = out.WriteString(strconv.Itoa(int(id+1)) + "\n")
	dieIf(err)

	ids := make([]string, 0, id+1)
	for i := 0; i < int(id+1); i++ {
		ids = append(ids, strconv.Itoa(i))
	}

	_, err = out.WriteString(strings.Join(ids, " ") + "\n")
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
