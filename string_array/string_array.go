package string_array

import (
	"bufio"
	"io"
	"os"
)

func Contains(test string, set []string) bool {
	for _, member := range set {
		if test == member {
			return true
		}
	}

	return false
}

// Compares array1 against a2 and returns the values in array1 that are not present in a2.
func Diff(tests, set []string) []string {

	missing := make([]string, 0)

	for _, test := range tests {
		if Contains(test, set) {
			continue
		}

		missing = append(missing, test)
	}

	return missing
}

func WriteFile(filename string, string_sets ...[]string) error {

	var f *os.File

	if f, err := os.Create(filename); err != nil {
		return err
	}

	defer f.Close()

	for _, set := range string_sets {
		for _, lines := range set {
			if _, err := f.WriteString(line); err != nil {
				return err
			}
		}
	}

	return nil
}

func ReadFile(filename string) (lines []string) {
	var (
		f    *os.File
		line string
		err  error
	)

	if file, err := os.Open(filename); err != nil {
		panic(err)
	}

	defer file.Close()

	reader := bufio.NewReader(file)

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
