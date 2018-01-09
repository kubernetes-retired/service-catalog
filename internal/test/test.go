package test

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
)

// UpdateGolden writes out the golden files with the latest values, rather than failing the test.
// Example: go test ./cmd/svcat --update
var UpdateGolden = flag.Bool("update", false, "update golden files")

// buildTestdataPath returns the full path to a testdata file.
// * relpath - relative path to the file in the test's testdata directory.
func buildTestdataPath(relpath string) (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "unable to get the current working directory")
	}

	path := filepath.Join(pwd, "testdata", relpath)
	return path, nil
}

// GetTestdata returns the contents of a testdata file.
// * relpath - relative path to the file in the test's testdata directory.
func GetTestdata(relpath string) (fullpath string, contents []byte, err error) {
	fullpath, err = buildTestdataPath(relpath)
	if err != nil {
		return "", nil, err
	}

	contents, err = ioutil.ReadFile(fullpath)
	return fullpath, contents, errors.Wrapf(err, "unable to read testdata %s", fullpath)
}

// AssertEqualsGoldenFile asserts that the value equals the contents of the golden file.
// When the go test -update flag is present, the golden file is updated to match, rather than failing the test.
func AssertEqualsGoldenFile(t *testing.T, goldenFile string, got string) {
	t.Helper()

	path, want, err := GetTestdata(goldenFile)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	gotB := []byte(got)
	if !bytes.Equal(want, gotB) {
		if *UpdateGolden {
			err := ioutil.WriteFile(path, gotB, 0666)
			if err != nil {
				t.Fatalf("%+v", errors.Wrapf(err, "unable to update golden file %s", path))
			}
		} else {
			t.Fatalf("does not match golden file %s\n\nWANT:\n%q\n\nGOT:\n%q\n", path, want, gotB)
		}
	}
}
