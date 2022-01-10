package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	procPath = "/proc"
)

func pidof(name string) (int, error) {
	files, err := ioutil.ReadDir(procPath)
	if err != nil {
		return -1, fmt.Errorf("failed to read /proc directory")
	}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		// Open the /proc/xxx/stat file to read the name
		p := filepath.Join(procPath, file.Name(), "stat")

		f, err := os.Open(p)
		if err != nil {
			continue
		}
		defer f.Close()

		r := bufio.NewReader(f)
		scanner := bufio.NewScanner(r)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			var pid2 int
			var name2 string

			n, err := fmt.Sscanf(scanner.Text(), "%d %s", &pid2, &name2)
			if err != nil || n != 2 {
				continue
			}

			name2 = strings.Trim(name2, "()")

			if name == name2 {
				return pid2, nil
			}
		}
	}

	return -1, os.ErrNotExist
}
