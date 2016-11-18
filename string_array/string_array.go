package string_array

import (
	"bufio"
	"io"
	"os"
)

func Contains(set []string, test string) bool {
	for _, member := range set {
		if test == member {
			return true
		}
	}

	return false
}

func sum(values []int) (total int) {
	for _, v := range values {
		total += v
	}
	return
}

func ContainsAll(set []string, tests []string) bool {
	total := len(tests)
	found := make([]int, total)

	for _, s := range set {
		for idx, check := range tests {
			if s == check {
				found[idx] = 1
				if sum(found) == total {
					return true
				}
			}
		}
	}

	return false
}

// Compares array1 against a2 and returns the values in array1 that are not present in a2.
func Diff(tests, set []string) []string {

	missing := make([]string, 0)

	for _, test := range tests {
		if Contains(set, test) {
			continue
		}

		missing = append(missing, test)
	}

	return missing
}

func WriteFile(filename string, string_sets ...[]string) error {

	file, err := os.Create(filename)

	if err != nil {
		return err
	}

	defer file.Close()

	for _, set := range string_sets {
		for _, line := range set {
			if _, err := file.WriteString(line); err != nil {
				return err
			}
		}
	}

	return nil
}

func ReadFile(filename string) (lines []string) {

	file, err := os.Open(filename)

	if err != nil {
		panic(err)
	}

	defer file.Close()

	reader := bufio.NewReader(file)

	var line string

	for {
		if line, err = reader.ReadString('\n'); err != nil {
			break
		}

		lines = append(lines, line)
	}

	if err != io.EOF {
		panic(err)
	}

	return lines
}
