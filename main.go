package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

var files = []string{
	"a_example.txt",
	"b_read_on.txt",
	"c_incunabula.txt",
	"d_tough_choices.txt",
	"e_so_many_books.txt",
	"f_libraries_of_the_world.txt",
}

const nSimulations = 12000

func run(fn string) output {
	var best struct {
		result    []Result
		iteration int
		score     int
		maxScore  int
	}
	for simulationIdx := 0; simulationIdx < nSimulations; simulationIdx++ {

		// read data
		in, err := os.Open(fn)
		dieIf(err)
		// defer in.Close()

		s := bufio.NewScanner(in)
		bf := []byte{}
		s.Buffer(bf, 5e6)

		if !s.Scan() {
			dieIf(errors.New("failed on first line"))
		}
		tmp := lineToIntSlice(s.Text())

		nBooks := tmp[0]
		nLibraries := tmp[1]
		totDays := tmp[2]

		if !s.Scan() {
			dieIf(errors.New("failed on second line"))
		}
		tmp = lineToIntSlice(s.Text())

		// read books
		books := make(Books, 0, nBooks)
		for id, score := range tmp {
			books = append(books, Book{
				ID: id, Score: score,
			})
		}

		var theoreticalMaxScore int
		for _, b := range books {
			theoreticalMaxScore += b.Score
		}

		// load libraries
		libraries := make(Libraries, 0, nLibraries)
		for i := 0; i < nLibraries; i++ {
			if !s.Scan() {
				dieIf(errors.New("failed on first line"))
			}
			tmp = lineToIntSlice(s.Text())

			l := Library{
				ID:               i,
				RegistrationTime: tmp[1],
				BooksPerDay:      tmp[2],
			}

			booksCount := tmp[0]

			if !s.Scan() {
				dieIf(errors.New("failed on second line"))
			}

			// id dei libri contenuti nella biblioteca
			bookIDs := lineToIntSlice(s.Text())

			// raccolgo i libri per la libreria
			bks := make(Books, 0, booksCount)
			for _, bid := range bookIDs {
				bks = append(bks, books[bid])
			}

			// sort library books desc on score
			sort.Sort(bks)
			l.Books = bks

			var totalBooksValue int
			for _, b := range l.Books {
				totalBooksValue += b.Score
			}
			l.TotalBooksValue = totalBooksValue

			libraries = append(libraries, l)
		}
		in.Close()

		// compute library score as random
		rand.Seed(time.Now().UnixNano())
		randomScoring := rand.Perm(len(libraries))
		for i, l := range libraries {
			l.Score = randomScoring[i]
			// l.Score = (totDays - l.RegistrationTime) * l.BooksPerDay * l.TotalBooksValue
			libraries[i] = l
		}

		// sort libraries desc on score
		sort.Sort(libraries)

		// remove duplicates from higher library score down to last library
		for i, l := range libraries {
			var bks Books
			for j, b := range l.Books {
				if j+l.Start > totDays {
					break
				}
				if books[b.ID].Taken {
					continue
				}
				b.Taken = true
				bks = append(bks, b)
				books[b.ID] = b
			}
			l.Books = bks

			libraries[i] = l
		}

		var (
			score, start int
			res          []Result
		)
		for _, l := range libraries {
			if len(l.Books) == 0 {
				// skip empty libraries
				continue
			}

			start += l.RegistrationTime                      // start after the previous one has finished and I registered
			workingDays := totDays - start                   // working days between start and stop of the simulations
			scannedBooksCount := workingDays * l.BooksPerDay // scan books at the current library rate during the working days

			bksid := make([]int, 0, len(l.Books))
			for i, b := range l.Books {
				if i < scannedBooksCount {
					score += b.Score // add score for the books we can actually scan before the end
				}
				bksid = append(bksid, b.ID) // add all the books, just in case
			}
			res = append(res, Result{
				LibraryID:  l.ID,
				BooksCount: len(l.Books),
				BookIDs:    bksid,
			})
		}

		if score > best.score {
			best.result = res
			best.iteration = simulationIdx
			best.score = score
			best.maxScore = theoreticalMaxScore
		}
	}

	// write output
	out, err := os.Create(fn + ".out")
	dieIf(err)
	defer out.Close()

	_, err = out.WriteString(fmt.Sprintf("%v", len(best.result)) + "\n")
	dieIf(err)

	for _, r := range best.result {
		_, err = out.WriteString(fmt.Sprintf("%v %v", r.LibraryID, r.BooksCount) + "\n")
		dieIf(err)

		bks := make([]string, 0, len(r.BookIDs))
		for _, bid := range r.BookIDs {
			bks = append(bks, strconv.Itoa(bid))
		}
		_, err = out.WriteString(strings.Join(bks, " ") + "\n")
		dieIf(err)
	}

	return output{
		iteration: best.iteration,
		p:         best.score,
		max:       best.maxScore,
		fn:        fn,
	}

}

// data types
type Books []Book
type Book struct {
	ID    int
	Score int
	Taken bool
}

func (a Books) Len() int      { return len(a) }
func (a Books) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// books reverse sort
func (a Books) Less(i, j int) bool { return a[i].Score > a[j].Score }

type Libraries []Library
type Library struct {
	ID               int
	Books            Books
	RegistrationTime int
	BooksPerDay      int
	Score            int
	TotalBooksValue  int
	Start            int
}

func (l Libraries) Len() int           { return len(l) }
func (l Libraries) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l Libraries) Less(i, j int) bool { return l[i].Score > l[j].Score }

type Result struct {
	LibraryID  int
	BooksCount int
	BookIDs    []int
}

// helpers and main
func main() {
	t0 := time.Now()

	wgRunners := sync.WaitGroup{}
	wgPrinter := sync.WaitGroup{}
	out := make(chan output, len(files))

	// print result as they arrive, concurrent safe
	wgPrinter.Add(1)
	go func() {
		defer wgPrinter.Done()

		var sumP, sumT int
		for res := range out {
			sumP += res.p
			sumT += res.max
			fmt.Printf("file: %v, points: %v, max: %v, difference: %v, perc. missing: %f%%, iteration: %v/%v \n",
				res.fn, res.p, res.max, res.max-res.p, 100*float64(res.max-res.p)/float64(res.max), res.iteration, nSimulations)
		}

		fmt.Printf("total, points: %v, max: %v, difference: %v, perc. missing: %f%%\n",
			sumP, sumT, sumT-sumP, 100*float64(sumT-sumP)/float64(sumT))
	}()

	// run tasks
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
	p         int
	max       int
	iteration int
	fn        string
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

func dieIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
