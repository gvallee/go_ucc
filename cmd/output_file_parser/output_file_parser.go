//
// Copyright (c) 2025, NVIDIA CORPORATION. All rights reserved.
//
// See LICENSE.txt for license information
//

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gvallee/go_ucc/pkg/perftest"
)

func printHelpMessage(cmdName string) {
	fmt.Printf("%s parses file with the OSU benchmark output. This is meant as a test, for example in order to test the parser with new version of OSU or new tests in the OSU suite", cmdName)
	fmt.Println("\nUsage:")
	flag.PrintDefaults()
}

func main() {
	help := flag.Bool("h", false, "Help message")
	osu_file := flag.String("perftest-file", "", "Path to the file that contains the ucc_perftest data to parse")

	flag.Parse()

	if *help {
		cmdName := filepath.Base(os.Args[0])
		printHelpMessage(cmdName)
		os.Exit(0)
	}

	res, err := perftest.ParseOutputFile(*osu_file)
	if err != nil {
		fmt.Printf("ParseOutputFile() failed: %s", err)
		os.Exit(1)
	}

	for _, d := range res.DataPoints {
		fmt.Printf("%f\t\t%f\n", d.Size, d.Value)
	}
}
