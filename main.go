package main

import "log"

func main() {

}

func dieIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
