package gart

import (
	"fmt"
	"math/rand"
	"os/exec"
	"time"
)

// Seed hold the primary seed used for random numbers
type Seed struct {
	intSeed int64
}

func newSeed() SeedSeed {
	intSeed := time.Now().UnixNano()
	rand.Seed(intSeed)
	return Seed{intSeed}
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
	return string(cmdOut)
}

func (s Seed) getSeed() int64 {
	return s.intSeed
}

func (s Seed) getFilename(prefix, ext string) string {
	return fmt.Sprintf("%s%s-%x%s", prefix, getGitHash(), s.intSeed, ext)
}
