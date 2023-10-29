package util

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/gosimple/slug"
)

func EnsurePathExists(path string, perm fs.FileMode) error {
	stat, err := os.Stat(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}

		err := os.MkdirAll(filepath.Clean(path), perm)
		if err != nil {
			return err
		}

		return nil
	}

	if !stat.IsDir() {
		return errors.New("Folder exists, but is not a directory")
	}

	return nil
}

func GetSlug(in string) string {
	return slug.Make(in)
}

func GetSlugFromTime(t time.Time) string {
	return fmt.Sprintf(t.Format("20060102_1504"))
}

func GetRandomBytes(length int) []byte {
	b := make([]byte, length)
	rand.Read(b)
	return b
}
