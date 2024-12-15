package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"sort"
	"strconv"
)

type TreeItem struct {
	Parent   *TreeItem
	Children []*TreeItem
	FileInfo fs.FileInfo
	Path     string
	Level    int
}

func prepend(x []*TreeItem, y *TreeItem) []*TreeItem {
	x = append(x, &TreeItem{})
	copy(x[1:], x)
	x[0] = y
	return x
}

func getChildren(treeItem *TreeItem) ([]fs.FileInfo, error) {
	if treeItem.FileInfo.IsDir() {
		dir, err := os.Open(treeItem.Path)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		files, err := dir.Readdir(0)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		return files, nil
	}
	return nil, nil
}

func printLine(out io.Writer, item *TreeItem, levels []bool, isLast bool) {
	var result = ""

	for i := 1; i < item.Level; i++ {
		if levels[i] {
			result += "│"
		}
		result += "\t"
	}
	if isLast {
		result += "└───"
	} else {
		result += "├───"
	}
	result += item.FileInfo.Name()
	if !item.FileInfo.IsDir() {
		if item.FileInfo.Size() == 0 {
			result += " (empty)"
		} else {
			result += " (" + strconv.FormatInt(item.FileInfo.Size(), 10) + "b)"
		}
	}
	result += "\n"
	out.Write([]byte(result))
}

func printTree(out io.Writer, root *TreeItem) error {
	levels := []bool{false}

	if len(root.Children) == 0 {
		return nil
	}

	children := root.Children
	sort.Slice(children, func(i, j int) bool {
		return children[i].FileInfo.Name() < children[j].FileInfo.Name()
	})

	var queue []*TreeItem

	for i := len(children) - 1; i >= 0; i-- {
		queue = prepend(queue, children[i])
	}

	for len(queue) > 0 {
		treeItem := queue[0]
		queue = queue[1:]

		if len(levels) <= treeItem.Level {
			levels = append(levels, true)
		} else {
			levels[treeItem.Level] = true
		}

		children := treeItem.Children
		sort.Slice(children, func(i, j int) bool {
			return children[i].FileInfo.Name() > children[j].FileInfo.Name()
		})

		if !(len(queue) > 0 && treeItem.Level == queue[0].Level) {
			levels[treeItem.Level] = false
			printLine(out, treeItem, levels, true)
		} else {
			printLine(out, treeItem, levels, false)
		}

		for _, child := range children {
			queue = prepend(queue, child)
		}
	}
	return nil
}

func buildTree(path string, printFiles bool) (*TreeItem, error) {
	var stack []*TreeItem

	fileInfo, err := os.Lstat(path)
	if err != nil {
		fmt.Println(err)
	}

	var root = TreeItem{
		FileInfo: fileInfo,
		Path:     path,
		Level:    0,
	}

	stack = append(stack, &root)
	for len(stack) > 0 {
		curr := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		files, err := getChildren(curr)
		if err != nil {
			fmt.Println(err)
		}

		for _, file := range files {
			var childItem = TreeItem{
				Parent:   curr,
				FileInfo: file,
				Path:     curr.Path + `/` + file.Name(),
				Level:    curr.Level + 1,
			}

			if childItem.FileInfo.IsDir() || (printFiles && !childItem.FileInfo.IsDir()) {
				curr.Children = append(curr.Children, &childItem)
				stack = append(stack, &childItem)
			}
		}
	}
	return &root, nil
}

func dirTree(out io.Writer, path string, printFiles bool) error {

	root, err := buildTree(path, printFiles)
	if err != nil {
		return err
	}
	printTree(out, root)
	return nil
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
