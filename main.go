package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

func main() {
	t0 := time.Now()
	files := []string{
		"a_example",
		"b_small",
		"c_medium",
		"d_quite_big",
		"e_also_big",
	}

	wgRunners := sync.WaitGroup{}
	wgPrinter := sync.WaitGroup{}
	out := make(chan output, 5)

	wgPrinter.Add(1)
	go func() {
		defer wgPrinter.Done()

		var sumP, sumT int
		for res := range out {
			sumP += res.p
			sumT += res.max
			fmt.Printf("file: %v, points: %v, max: %v, difference: %v\n", res.fn, res.p, res.max, res.max-res.p)
		}

		fmt.Printf("total, points: %v, max: %v, difference: %v, perc. missing: %f%%: \n",
			sumP, sumT, sumT-sumP, 100*float64(sumT-sumP)/float64(sumT))
	}()

	for _, fn := range files {
		wgRunners.Add(1)

		go func(wg *sync.WaitGroup, fn string, out chan output) {
			defer wg.Done()

			out <- run(fn)
		}(&wgRunners, fn, out)
	}

	wgRunners.Wait()
	close(out)

	wgPrinter.Wait()

	fmt.Println()
	log.Println("done in ", time.Since(t0))
}

type output struct {
	p   int
	max int
	fn  string
}

func run(fn string) output {
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

	out, err := os.Create(fn + ".out")
	dieIf(err)
	defer out.Close()

	_, err = out.WriteString(strconv.Itoa(len(order)) + "\n")
	dieIf(err)

	_, err = out.WriteString(strings.Join(order, " ") + "\n")
	dieIf(err)

	return output{
		p:   sum,
		max: totSlices,
		fn:  fn,
	}
}

func dieIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
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
