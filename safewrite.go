package gart

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// SafeWrite noisily saves to tmp file and then moves for gg
func (s Seed) SafeWrite(ctx *Context, prefix, ext string) error {
	fname := s.GetFilename(prefix, ext)
	if err := safeWrite(ctx, fname); err != nil {
		fmt.Printf("Problem saving %s: %v\n", fname, err)
		return err
	}
	fmt.Printf("Saved to %s\n", fname)
	return nil
}

// safeWrite writes to a temp file then renames atomically
func safeWrite(ctx *Context, fname string) error {
	ext := path.Ext(fname)
	tmpfile, err := ioutil.TempFile("./", "gart.*"+ext)
	if err != nil {
		return err
	}
	// todo: Change depending on extension
	if ext == ".png" {
		if err := ctx.WritePNG(tmpfile.Name()); err != nil {
			os.Remove(tmpfile.Name())
			return err
		}
	} else if ext == ".svg" {
		if err := ctx.WriteSVG(tmpfile.Name()); err != nil {
			os.Remove(tmpfile.Name())
			return err
		}
	} else if ext == ".pdf" {
		if err := ctx.WritePDF(tmpfile.Name()); err != nil {
			os.Remove(tmpfile.Name())
			return err
		}
	} else {
		return fmt.Errorf("unsupported file format %s", ext)
	}
	return os.Rename(tmpfile.Name(), fname)
}
