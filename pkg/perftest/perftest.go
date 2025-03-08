//
// Copyright (c) 2025, NVIDIA CORPORATION. All rights reserved.
//
// See LICENSE.txt for license information
//

package perftest

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gvallee/go_benchmark/pkg/benchmark"
)

func IsPerftestData(content []string) bool {
	for _, line := range content {
		if strings.HasPrefix(line, "Collective:") {
			return true
		}
	}
	return false
}

// ExtractDataFromOutput returns the OSU data from a output file. The first
// array returned represents all the data sizes, while the second returned array
// represents the value for the associated size (the two arrays are assumed to
// be ordered).
func ExtractDataFromOutput(benchmarkOutput []string) ([]float64, []float64, error) {
	var x []float64
	var y []float64

	var val1 float64
	var val2 float64

	var err error

	save := false
	stop := false

	for _, line := range benchmarkOutput {
		val1 = -1.0
		val2 = -1.0

		if line == "" {
			continue
		}
		if !save && strings.Contains(line, "avg") && strings.Contains(line, "min") && strings.Contains(line, "max") {
			save = true
			continue
		}
		if !save {
			// We skip whatever is at the beginning of the file until we reach the OSU header
			continue
		}
		if strings.Contains(line, "more processes have sent help message") || strings.Contains(line, "more process has sent help message") {
			// Open MPI throwing warnings for whatever reason, skipping
			continue
		}
		if strings.HasPrefix(line, "\x00") {
			// Some weird output we get on some platforms (seems to be when OMPI throws out warnings)
			continue
		}

		// We replace all double spaces with a single space to make it easier to identify the real data
		idx := 0
		tokens := strings.Split(line, " ")
		for _, t := range tokens {
			if t == " " || t == "" {
				continue
			}

			idx++

			if val1 == -1.0 && idx == 2 {
				val1, err = strconv.ParseFloat(t, 64)
				if err != nil {
					if len(x) == 0 {
						return nil, nil, fmt.Errorf("unable to convert %s (from %s): %w", t, line, err)
					} else {
						log.Printf("stop parsing, unable to convert %s (from %s): %s", t, line, err)
						stop = true
						break
					}
				}
				x = append(x, val1)
			} else if idx == 3 {
				val2, err = strconv.ParseFloat(t, 64)
				if err != nil {
					if len(y) == 0 {
						return nil, nil, fmt.Errorf("unable to convert %s (from %s): %w", t, line, err)
					} else {
						log.Printf("stop parsing, unable to convert %s (from %s): %s", t, line, err)
						stop = true
						break
					}
				}
				y = append(y, val2)
				break // todo: we need to extend this for sub-benchmarks returning more than one value (see what is done in OpenHPCA)
			}
		}

		if stop {
			break
		}
	}

	return x, y, nil
}

func ParseOutputFile(path string) (*benchmark.Result, error) {
	log.Printf("Parsing result file %s", path)
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	str := string(content)
	fileContent := strings.Split(str, "\n")
	if !IsPerftestData(fileContent) {
		return nil, fmt.Errorf("not ucc_perftest data")
	}
	dataSize, data, err := ExtractDataFromOutput(fileContent)
	if err != nil {
		return nil, fmt.Errorf("unable to parse %s: %w", path, err)
	}
	if len(dataSize) != len(data) {
		return nil, fmt.Errorf("unsupported data format (%s), skipping: %d different sizes with %d values", path, len(dataSize), len(data))
	}

	newResult := new(benchmark.Result)
	for i := 0; i < len(dataSize); i++ {
		newDataPoint := new(benchmark.DataPoint)
		newDataPoint.Size = dataSize[i]
		newDataPoint.Value = data[i]
		newResult.DataPoints = append(newResult.DataPoints, newDataPoint)
	}

	return newResult, nil
}

func GetResultsFromFiles(listFiles []string) (*benchmark.Results, error) {
	res := new(benchmark.Results)
	for _, file := range listFiles {
		newResult, err := ParseOutputFile(file)
		if err != nil {
			return nil, err
		}
		res.Result = append(res.Result, newResult)
	}

	return res, nil
}
