package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type FileInfoPredicate func(entry os.FileInfo) bool

const (
	entryPrefix           = "├───"
	lastEntryPrefix       = "└───"
	zeroSize              = "empty"
	pattern               = "%s%s\n"
	middleSubfolderPrefix = "│" + "\t"
)

func sortFunction(entries []os.FileInfo) func(i, j int) bool {
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
	_, err := fmt.Fprint(output, strings.Join(collectEntries(path, withFiles), ""))
	if err != nil {
		return err
	}
	return nil
}

func collectEntries(path string, withFiles bool) (result []string) {
	entries := *getEntries(path)
	if len(entries) == 0 {
		return
	}

	sort.Slice(entries, sortFunction(entries))
	if !withFiles {
		entries = filter(entries, func(entry os.FileInfo) bool { return entry.IsDir() })
	}

	for index, entry := range entries {
		length := len(entries)
		if entry.IsDir() {
			subfolders := collectEntries(path+string(os.PathSeparator)+entry.Name(), withFiles)
			getCollectForDir(index, length)(&result, entry.Name(), subfolders)
		} else {
			addFile(index, length, &result, getFileName(entry))
		}
	}
	return
}

func getFileName(entry os.FileInfo) string {
	size := fmt.Sprintf("%db", entry.Size())
	if entry.Size() == 0 {
		size = zeroSize
	}
	return entry.Name() + fmt.Sprintf(" (%s)", size)
}

func getEntries(path string) *[]os.FileInfo {
	open, err := os.Open(path)
	check(err)
	entries, err := open.Readdir(-1)
	check(err)
	return &entries
}

func getCollectForDir(index int, length int) func(temp *[]string, entryName string, subfolders []string) {
	if index != length-1 {
		return func(temp *[]string, entryName string, subfolders []string) {
			addWithSubfolders(temp, entryName, entryPrefix, middleSubfolderPrefix, subfolders)
		}
	} else {
		return func(temp *[]string, entryName string, subfolders []string) {
			addWithSubfolders(temp, entryName, lastEntryPrefix, "\t", subfolders)
		}
	}
}

func addWithSubfolders(temp *[]string, entryName string, prefix string, subfolderPrefix string, subfolders []string) {
	*temp = append(*temp, fmt.Sprintf(pattern, prefix, entryName))
	if len(subfolders) != 0 {
		for _, subfolder := range subfolders {
			*temp = append(*temp, subfolderPrefix+subfolder)
		}
	}
}

func filter(entries []os.FileInfo, filterFunc FileInfoPredicate) (result []os.FileInfo) {
	for _, entry := range entries {
		if filterFunc(entry) {
			result = append(result, entry)
		}
	}
	return
}

func addFile(index int, length int, temp *[]string, entryName string) {
	if index != length-1 {
		*temp = append(*temp, fmt.Sprintf(pattern, entryPrefix, entryName))
	} else {
		*temp = append(*temp, fmt.Sprintf(pattern, lastEntryPrefix, entryName))
	}
}

func check(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
}
