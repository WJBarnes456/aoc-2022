package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Treating files and directories as separate types makes the typing simpler
// No respect for everything being a file here
// We use a map for directories to avoid overwriting directories we already traversed
type Directory struct {
	name        string
	files       []File
	directories map[string]*Directory
	parent      *Directory
}

type File struct {
	name string
	size int
}

func (f *File) Size() int {
	return f.size
}

func (d *Directory) Size() int {
	total := 0
	for _, f := range d.files {
		total += f.Size()
	}

	for _, d2 := range d.directories {
		total += d2.Size()
	}

	return total
}

// Returns whether there's a next line, what it is if so, then the list of files and directory names, then an error
// This is because we don't know if we've reached the end of an ls until we reach it
func parseLs(scanner *bufio.Scanner) (bool, string, []File, []string, error) {
	files, directories := []File{}, []string{}

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")

		if parts[0] == "$" && len(parts) == 3 {
			return true, line, files, directories, nil
		}

		if parts[0] == "dir" && len(parts) == 2 {
			directories = append(directories, parts[1])
		} else {
			// In all other cases, assume it's a file
			// this isn't very defensive, but it'll work
			size, err := strconv.ParseInt(parts[0], 10, 32)

			if err != nil {
				return false, "", files, directories, fmt.Errorf("failed to parse file %s size: %v", parts[1], err)
			}

			files = append(files, File{parts[1], int(size)})
		}
	}

	return false, "", files, directories, nil
}

func parseCd(parts []string, cwd **Directory, root *Directory) error {
	if parts[2] == "/" {
		(*cwd) = root
		return nil
	}

	if parts[2] == ".." {
		(*cwd) = (*cwd).parent
		return nil
	}

	dest, exists := (*cwd).directories[parts[2]]

	if !exists {
		return fmt.Errorf("attempted to descend to non-existent directory %v", dest)
	}

	(*cwd) = dest
	return nil
}

func parseFilesystem(r io.Reader) (Directory, error) {
	scanner := bufio.NewScanner(r)

	root := Directory{"/", []File{}, map[string]*Directory{}, nil}

	var cwd *Directory

	validLine := scanner.Scan()
	line := scanner.Text()

	for validLine {
		parts := strings.Split(line, " ")

		if parts[0] != "$" {
			return root, fmt.Errorf("parser encountered a non-command %s", line)
		}

		if parts[1] == "cd" && len(parts) == 3 {
			parseCd(parts, &cwd, &root)
			validLine, line = scanner.Scan(), scanner.Text()
			continue
		}

		if parts[1] == "ls" && len(parts) == 2 {
			// parseLs is a bit of a beast, because you can't tell if you've seen the end of an ls until you've actually seen it
			validNextLine, nextLine, files, directoryNames, err := parseLs(scanner)

			if err != nil {
				return root, fmt.Errorf("failed to parse ls: %v", err)
			}

			cwd.files = files

			for _, name := range directoryNames {
				_, dirExists := cwd.directories[name]
				if !dirExists {
					cwd.directories[name] = &Directory{name, []File{}, map[string]*Directory{}, cwd}
				}
			}

			validLine, line = validNextLine, nextLine
			continue
		}

		return root, fmt.Errorf("invalid command %s", line)
	}

	return root, nil
}

func allDirectories(rootDir *Directory) []*Directory {
	// Breadth-first search over the nodes
	nodes := []*Directory{rootDir}
	for i := 0; i < len(nodes); i++ {
		node := nodes[i]
		for _, d := range node.directories {
			nodes = append(nodes, d)
		}
	}
	return nodes
}

func part1(rootDir *Directory) int {
	dirs := allDirectories(rootDir)

	total := 0
	for _, d := range dirs {
		size := d.Size()
		if size < 100000 {
			total += size
		}
	}

	return total
}

func part2(rootDir *Directory) int {
	total := 70000000
	required := 30000000

	// we need usedSpace -x + required <= total
	// smallest x such that -x <= total - required - usedSpace
	// i.e. x such that x >= usedSpace + required - total

	requiredSize := rootDir.Size() + required - total

	dirs := allDirectories(rootDir)

	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].Size() < dirs[j].Size()
	})

	for _, d := range dirs {
		if d.Size() >= requiredSize {
			return d.Size()
		}
	}

	return 0
}

func run() error {
	rootDir, err := parseFilesystem(os.Stdin)

	if err != nil {
		return fmt.Errorf("failed to parse filesystem: %v", err)
	}

	fmt.Println("File system:", rootDir)
	fmt.Println("Root size:", rootDir.Size())

	part1 := part1(&rootDir)

	fmt.Println("Part 1:", part1)

	part2 := part2(&rootDir)

	fmt.Println("Part 2:", part2)

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error in day 7:", err)
		os.Exit(1)
	}
}
