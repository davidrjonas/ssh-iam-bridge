package strarray_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/davidrjonas/ssh-iam-bridge/strarray"
	"github.com/stretchr/testify/assert"
)

type containsTest struct {
	Name     string
	Expected bool
	Test     string
	Set      []string
}

func TestContains(t *testing.T) {
	tests := []containsTest{
		containsTest{"Empty matches empty", true, "", []string{""}},
		containsTest{"Single element match", true, "bbb", []string{"bbb"}},
		containsTest{"Last element match", true, "bbb", []string{"aaa", "bbb"}},
		containsTest{"Middle element match", true, "bbb", []string{"aaa", "bbb", "ccc"}},
		containsTest{"Empty not found", false, "", []string{"a"}},
		containsTest{"No matches", false, "zzz", []string{"aaa"}},
	}

	for _, ct := range tests {
		assert.Equal(t, ct.Expected, strarray.Contains(ct.Set, ct.Test), ct.Name)
	}
}

type containsAllTest struct {
	Name     string
	Expected bool
	Test     []string
	Set      []string
}

func TestContainsAll(t *testing.T) {
	tests := []containsAllTest{
		containsAllTest{"Single match", true, []string{"a"}, []string{"a", "", "b"}},
		containsAllTest{"Empty matches empty", true, []string{""}, []string{"a", "", "b"}},
		containsAllTest{"Multiple strings", true, []string{"a", "b"}, []string{"a", "c", "b"}},
		containsAllTest{"Line endings", true, []string{"a\n", "b\n"}, []string{"a\n", "c", "b\n"}},
		containsAllTest{"One missing", false, []string{"a", "b"}, []string{"a", "c"}},
		containsAllTest{"All missing", false, []string{"a", "b"}, []string{"c", "d"}},
	}

	for _, ct := range tests {
		assert.Equal(t, ct.Expected, strarray.ContainsAll(ct.Set, ct.Test), ct.Name)
	}
}

type diffTest struct {
	Name     string
	Expected []string
	Test     []string
	Set      []string
}

func TestDiff(t *testing.T) {
	tests := []diffTest{
		diffTest{"One element", []string{"a"}, []string{"a"}, []string{"b", "c"}},
		diffTest{"One element match", []string{"a"}, []string{"a", "b"}, []string{"b", "c"}},
		diffTest{"Empty second array", []string{"a", "b"}, []string{"a", "b"}, []string{}},
		diffTest{"Empty first array", []string{}, []string{}, []string{"a", "b"}},
	}

	for _, ct := range tests {
		assert.Equal(t, ct.Expected, strarray.Diff(ct.Test, ct.Set), ct.Name)
	}
}

type uniqueTest struct {
	Name     string
	Expected []string
	Test     []string
}

func TestUnique(t *testing.T) {
	tests := []uniqueTest{
		uniqueTest{"Empty", []string{}, []string{}},
		uniqueTest{"One element", []string{"a"}, []string{"a"}},
		uniqueTest{"One repeat", []string{"a"}, []string{"a", "a"}},
		uniqueTest{"Multi repeat", []string{"a"}, []string{"a", "a", "a"}},
		uniqueTest{"Unsorted", []string{"a", "b"}, []string{"a", "b", "a"}},
	}

	for _, ct := range tests {
		assert.Equal(t, ct.Expected, strarray.Unique(ct.Test), ct.Name)
	}
}

type filterTest struct {
	Name     string
	Expected []string
	Test     []string
	Fn       func(string) bool
}

func failFilterFn(s string) bool {
	return false
}

func passFilterFn(s string) bool {
	return true
}

func passStringFilter(s string) func(string) bool {
	return func(test string) bool {
		return test == s
	}
}

func TestFilter(t *testing.T) {
	tests := []filterTest{
		filterTest{"Empty/pass", []string{}, []string{}, passFilterFn},
		filterTest{"Empty/fail", []string{}, []string{}, failFilterFn},
		filterTest{"Pass all", []string{"a", "b"}, []string{"a", "b"}, passFilterFn},
		filterTest{"Fail all", []string{}, []string{"a", "b"}, failFilterFn},
		filterTest{"Pass one", []string{"a"}, []string{"a", "b"}, passStringFilter("a")},
		filterTest{"Pass one multi", []string{"a", "a"}, []string{"a", "b", "a"}, passStringFilter("a")},
	}

	for _, ct := range tests {
		assert.Equal(t, ct.Expected, strarray.Filter(ct.Test, ct.Fn), ct.Name)
	}
}

func assertFileContents(t *testing.T, expected []byte, filename string, msg string) {
	bytes, err := ioutil.ReadFile(filename)

	if _, ok := err.(*os.PathError); !ok {
		os.Remove(filename)
	}

	assert.Nil(t, err)

	assert.Equal(t, expected, bytes, msg)
}

func TestWriteFile(t *testing.T) {
	filename := "testdata/output"

	strarray.WriteFile(filename, []string{"a\n"})
	assertFileContents(t, []byte{0x61, 0xa}, filename, "One string slice")

	strarray.WriteFile(filename, []string{"a\n", "b\n"})
	assertFileContents(t, []byte{0x61, 0xa, 0x62, 0xa}, filename, "Multiple strings in one slice")

	strarray.WriteFile(filename, []string{"a\n"}, []string{"b\n"})
	assertFileContents(t, []byte{0x61, 0xa, 0x62, 0xa}, filename, "Multiple slices")
}

func TestReadFile_EmptyFilenameReturnsError(t *testing.T) {
	_, err := strarray.ReadFile("")

	assert.IsType(t, &os.PathError{}, err)
}

func TestReadFile_FileContentsAreReturnedWithNilError(t *testing.T) {
	expected := []string{
		"1 Test\n",
		"2 Line Two\n",
	}

	content, err := strarray.ReadFile("testdata/file")

	assert.Nil(t, err)
	assert.Equal(t, expected, content, "ReadFile contents should equal testdata/file")
}
