package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	sourceFile := flag.String(
		"i",
		"-",
		"Specifies an input source file.  Use - for stdin.",
	)
	flag.Parse()

	var source io.Reader
	if *sourceFile == "-" {
		source = os.Stdin
	} else {
		var err error
		source, err = os.Open(*sourceFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	tree, err := Parser.Parse(source)
	if err != nil {
		log.Fatal(err)
	}

	lists := make(chan List)
	go transformPRGM(tree, lists)
	for a := range lists {
		v, err := eval(a, global_env)
		if err != nil {
			log.Fatal(err)
		}
		if v != nil {
			fmt.Println(v)
		}
	}
}
