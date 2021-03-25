package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

const (
	entryPrefix     = "├───"
	lastEntryPrefix = "└───"
	middleLineStart = "│"
	zeroSize        = "empty"
)

func sortFunction(entries []os.DirEntry) func(i, j int) bool {
	return func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	}
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(output io.Writer, path string, withFiles bool) error {
	_, err := output.Write([]byte(strings.Join(collectEntries(path, withFiles), "\n") + "\n"))
	if err != nil {
		return err
	}
	return nil
}

func collectEntries(path string, withFiles bool) []string {
	//TODO: refactoring

	open, err := os.Open(path)
	check(err)

	entries, err := open.ReadDir(-1)
	check(err)

	var temp []string
	var entriesTemp []os.DirEntry

	if len(entries) == 0 {
		return temp
	}

	sort.Slice(entries, sortFunction(entries))

	if !withFiles {
		for _, ent := range entries {
			if ent.IsDir() {
				entriesTemp = append(entriesTemp, ent)
			}
		}
	} else {
		entriesTemp = entries
	}

	for index, entry := range entriesTemp {
		if !withFiles {
			if !entry.IsDir() {
				continue
			}
		}
		var children []string
		var postfix string
		if entry.IsDir() {
			children = collectEntries(path+string(os.PathSeparator)+entry.Name(), withFiles)
		} else {
			info, err := entry.Info()
			check(err)
			size := strconv.FormatInt(info.Size(), 10) + "b"
			if info.Size() == 0 {
				size = zeroSize
			}
			postfix = fmt.Sprintf(" (%s)", size)
		}
		if index != len(entriesTemp)-1 {
			temp = append(temp, entryPrefix+entry.Name()+postfix)
			if len(children) != 0 {
				for _, child := range children {
					temp = append(temp, middleLineStart+"\t"+child)
				}
			}
		} else {
			temp = append(temp, lastEntryPrefix+entry.Name()+postfix)
			if len(children) != 0 {
				for _, child := range children {
					temp = append(temp, "\t"+child)
				}
			}
		}
	}

	return temp
}

func check(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
}
