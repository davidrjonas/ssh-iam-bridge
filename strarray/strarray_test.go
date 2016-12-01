package strarray_test

import (
	"os"
	"testing"

	"github.com/davidrjonas/ssh-iam-bridge/strarray"
)

func TestReadFile_EmptyFilenameReturnsError(t *testing.T) {
	_, err := strarray.ReadFile("")

	if _, ok := err.(*os.PathError); !ok {
		t.Error("Empty path should return an os.PathError")
	}
}

func TestReadFile_FileContentsAreReturnedWithNilError(t *testing.T) {
	content, err := strarray.ReadFile("testdata/file")

	if err != nil {
		t.Error("Error should be nil but was", err)
	}

	expected := []string{
		"1 Test\n",
		"2 Line Two\n",
	}

	for idx, line := range expected {
		if idx > len(content[idx]) {
			t.Errorf("Line %d is beyond the end of content", idx+1)
		}

		if content[idx] != line {
			t.Errorf("got unexpected content on line %d; \"%s\"", idx+1, content[idx])
		}

	}
}
