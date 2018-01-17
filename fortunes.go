package gimmemotd

import (
	"bufio"
	"bytes"
	"io"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// LoadFortunes returns a slice of fortune files in a directory.
func LoadFortunes(rootDir string) ([]*os.File, error) {
	files := make([]*os.File, 0)

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "walk %s", path)
		}
		if !info.IsDir() {
			f, err := os.Open(path)
			if err != nil {
				return errors.Wrapf(err, "open %s", info.Name())
			}

			files = append(files, f)
		}

		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "walk fortune dir")
	}

	return files, nil
}

// Fortunes can tell you a fortune.
type Fortunes struct {
	strings []string
}

// MakeFortunes creates an instance of Fortunes.
func MakeFortunes(rs ...io.Reader) (*Fortunes, error) {
	f := &Fortunes{
		strings: make([]string, 0),
	}

	for _, r := range rs {
		ss, err := parseFortune(r)
		if err != nil {
			return nil, errors.Wrap(err, "parse fortune reader")
		}

		f.strings = append(f.strings, ss...)
	}

	return f, nil
}

// Sample samples a fortune.
func (f *Fortunes) Sample() string {
	l := len(f.strings)
	if l == 0 {
		return ""
	}

	return f.strings[rand.Intn(l)]
}

func parseFortune(r io.Reader) ([]string, error) {
	ss := make([]string, 0)

	scanner := bufio.NewScanner(r)

	var buf bytes.Buffer
	for scanner.Scan() {
		s := scanner.Text()
		if s == "%" {
			ss = append(ss, buf.String())
			buf.Reset()
			continue
		}

		_, err := buf.WriteString(s)
		if err != nil {
			return nil, errors.Wrap(err, "write string")
		}
	}

	return ss, nil
}
