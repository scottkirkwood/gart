package gart

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/fogleman/gg"
)

// Noisy safe write
func (s Seed) SafeWrite(ctx *gg.Context, prefix, ext string) error {
	fname := s.GetFilename(prefix, ext)
	if err := safeWrite(ctx, fname); err != nil {
		fmt.Printf("Problem saving %s: %v\n", fname, err)
		return err
	}
	fmt.Printf("Saved to %s\n", fname)
	return nil
}

// safeWrite writes to a temp file then renames atomically
func safeWrite(ctx *gg.Context, fname string) error {
	ext := path.Ext(fname)
	tmpfile, err := ioutil.TempFile("./", "gart.*"+ext)
	if err != nil {
		return err
	}
	// todo: Change depending on extension
	if err := ctx.SavePNG(tmpfile.Name()); err != nil {
		os.Remove(tmpfile.Name())
		return err
	}
	return os.Rename(tmpfile.Name(), fname)
}
