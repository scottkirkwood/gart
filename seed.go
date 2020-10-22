package gart

import (
	"fmt"
	"math/rand"
	"os/exec"
	"strings"
	"time"
)

// Seed hold the primary seed used for random numbers
type Seed struct {
	intSeed int64
}

// Jan 1, 2020 (to make filenames a little smaller)
const epoch2020 = 1577836800

// Init initializes the seed (so far)
func Init() Seed {
	intSeed := time.Now().UnixNano() - epoch2020
	rand.Seed(intSeed)
	return Seed{intSeed}
}

// GetSeed returns the rand initialization seed
func (s Seed) GetSeed() int64 {
	return s.intSeed
}

// SetSeed sets the seed given the file seed part of filename
func (s Seed) SetSeed(hexSeed string) error {
	s.intSeed, err = strconv.ParseUint(hexSeed, 16, 64)
	if err != nil {
		return err
	}
	rand.Seed(s.intSeed)
}

// GetFilename returns a string to use for this file
func (s Seed) GetFilename(prefix, ext string) string {
	return fmt.Sprintf("%s%s-%x%s", prefix, getGitHash(), s.intSeed, ext)
}

func getGitHash() string {
	var (
		cmdOut []byte
		err    error
	)
	cmdName := "git"
	cmdArgs := []string{"rev-parse", "--verify", "HEAD"}
	if cmdOut, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		return ""
	}
	return strings.TrimSpace(string(cmdOut))[0:7]
}
