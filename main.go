package main

import (
	"bufio"
	"fmt"
	"log"
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

func run(fn string) output {
	// read data
	in, err := os.Open(fn)
	dieIf(err)
	defer in.Close()

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

	// read libraries
	libraries := make(Libraries, 0, nLibraries)
	for i := 0; i < nLibraries; i++ {
		if !s.Scan() {
			dieIf(errors.New("failed on first line"))
		}
		tmp = lineToIntSlice(s.Text())

		l := Library{
			ID:               i,
			BooksCount:       tmp[0],
			RedistrationTime: tmp[1],
			BooksPerDay:      tmp[2],
		}
		if !s.Scan() {
			dieIf(errors.New("failed on first line"))
		}

		bookIDs := lineToIntSlice(s.Text())
		bks := make(Books, 0, l.BooksCount)
		for _, bid := range bookIDs {
			l.BookIDs = append(l.BookIDs, bid)
			bks = append(bks, books[bid])
		}
		sort.Sort(bks)
		l.Books = bks

		l.BooksCount = len(l.Books)

		var totalBooksValue int
		for _, b := range l.Books {
			totalBooksValue += b.Score
		}
		l.TotalBooksValue = totalBooksValue

		// score della library
		var score int = (totDays - l.RedistrationTime) * l.BooksPerDay * l.TotalBooksValue

		l.Score = score

		libraries = append(libraries, l)
	}

	sort.Sort(libraries)

	// rimuovo i duplicati dopo aver calcolato il potenziale
	for i, l := range libraries {
		var bks Books
		var bksIds []int
		for _, b := range l.Books {
			if books[b.ID].Taken {
				continue
			}
			b.Taken = true
			bksIds = append(bksIds, b.ID)
			bks = append(bks, b)
			books[b.ID] = b
		}
		l.BookIDs = bksIds
		l.Books = bks
		l.BooksCount = len(l.BookIDs)

		libraries[i] = l
	}

	res := []Result{}
	for _, l := range libraries {
		if l.BooksCount == 0 {
			continue
		}
		bksid := make([]int, 0, l.BooksCount)
		for _, b := range l.Books {
			bksid = append(bksid, b.ID)
		}
		res = append(res, Result{
			LibraryID:  l.ID,
			BooksCount: l.BooksCount,
			BookIDs:    bksid,
		})
	}

	// write output
	out, err := os.Create(fn + ".out")
	dieIf(err)
	defer out.Close()

	_, err = out.WriteString(fmt.Sprintf("%v", len(res)) + "\n")
	dieIf(err)

	score := map[int]int{}
	for _, r := range res {
		_, err = out.WriteString(fmt.Sprintf("%v %v", r.LibraryID, r.BooksCount) + "\n")
		dieIf(err)

		bks := make([]string, 0, len(r.BookIDs))
		for _, bid := range r.BookIDs {
			score[bid] = books[bid].Score
			bks = append(bks, strconv.Itoa(bid))
		}
		_, err = out.WriteString(strings.Join(bks, " ") + "\n")
		dieIf(err)
	}

	var totalScore int
	for _, s := range score {
		totalScore += s
	}

	return output{
		p:   totalScore,
		max: theoreticalMaxScore,
		fn:  fn,
	}

}

// =============================================================

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
	BooksCount       int
	BookIDs          []int
	Books            Books
	RedistrationTime int
	BooksPerDay      int
	Score            int
	TotalBooksValue  int
}

func (l Libraries) Len() int           { return len(l) }
func (l Libraries) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l Libraries) Less(i, j int) bool { return l[i].Score > l[j].Score }

type Result struct {
	LibraryID  int
	BooksCount int
	BookIDs    []int
}

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
			fmt.Printf("file: %v, points: %v, max: %v, difference: %v\n", res.fn, res.p, res.max, res.max-res.p)
		}

		fmt.Printf("total, points: %v, max: %v, difference: %v, perc. missing: %f%%: \n",
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
	p   int
	max int
	fn  string
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
