// internal/vfs/vfs.go

package vfs

import (
	"encoding/base64"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Node struct {
	Name     string
	IsDir    bool
	Content  string
	Children map[string]*Node
}

var (
	root    *Node
	cwd     *Node
	cwdPath string
)

type entry struct {
	path    string
	isDir   bool
	content string
}

func Init(path string) error {
	if path == "" {
		return errors.New("vfs init: no CSV file provided")
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("vfs init: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("vfs init: invalid CSV: %w", err)
	}

	if len(records) == 0 {
		return errors.New("vfs init: empty CSV")
	}

	// Проверка заголовка
	header := records[0]
	if len(header) < 2 || header[0] != "path" || header[1] != "type" {
		return errors.New("vfs init: invalid CSV header (expected: path,type,content)")
	}

	root = &Node{
		Name:     "/",
		IsDir:    true,
		Children: make(map[string]*Node),
	}
	cwd = root
	cwdPath = "/"

	var entries []entry

	// Парсим записи
	for i, record := range records[1:] {
		if len(record) < 2 {
			return fmt.Errorf("vfs init: record at line %d has less than 2 fields", i+2)
		}

		path := strings.TrimSpace(record[0])
		typ := strings.TrimSpace(record[1])
		var content string
		if len(record) > 2 {
			content = record[2]
		}

		if path == "" {
			continue
		}

		path = filepath.ToSlash(filepath.Clean(path))
		if !strings.HasPrefix(path, "/") {
			return fmt.Errorf("vfs init: absolute path required, got %q (line %d)", path, i+2)
		}

		var isDir bool
		switch typ {
		case "dir":
			isDir = true
		case "file":
			isDir = false
		default:
			return fmt.Errorf("vfs init: unknown type %q at line %d", typ, i+2)
		}

		entries = append(entries, entry{path: path, isDir: isDir, content: content})
	}

	sortEntriesByDepth(entries)

	for _, e := range entries {
		if err := addNode(e.path, e.isDir, e.content); err != nil {
			return fmt.Errorf("vfs init: %w", err)
		}
	}

	return nil
}

func sortEntriesByDepth(entries []entry) {
	sort.Slice(entries, func(i, j int) bool {
		pathI := entries[i].path
		pathJ := entries[j].path

		// Корень всегда первый
		if pathI == "/" {
			return true
		}
		if pathJ == "/" {
			return false
		}

		compI := strings.Split(strings.TrimPrefix(pathI, "/"), "/")
		compJ := strings.Split(strings.TrimPrefix(pathJ, "/"), "/")

		// Убираем пустые строки (например, из "/home/")
		trimEmpty := func(parts []string) []string {
			var res []string
			for _, p := range parts {
				if p != "" {
					res = append(res, p)
				}
			}
			return res
		}
		compI = trimEmpty(compI)
		compJ = trimEmpty(compJ)

		return len(compI) < len(compJ)
	})
}

func addNode(fullPath string, isDir bool, content string) error {
	if fullPath == "/" {
		if !isDir {
			return errors.New("root must be a directory")
		}
		return nil
	}

	parts := strings.Split(strings.TrimPrefix(fullPath, "/"), "/")
	var cleanParts []string
	for _, p := range parts {
		if p != "" {
			cleanParts = append(cleanParts, p)
		}
	}
	if len(cleanParts) == 0 {
		return nil
	}

	cur := root
	for i, part := range cleanParts {
		isLast := (i == len(cleanParts)-1)

		if child, exists := cur.Children[part]; exists {
			if isLast {
				return fmt.Errorf("duplicate path: %s", fullPath)
			}
			if !child.IsDir {
				return fmt.Errorf("non-directory ancestor in path: %s", fullPath)
			}
			cur = child
		} else {
			if isLast {
				newNode := &Node{
					Name:     part,
					IsDir:    isDir,
					Content:  content,
					Children: make(map[string]*Node),
				}
				cur.Children[part] = newNode
			} else {
				newDir := &Node{
					Name:     part,
					IsDir:    true,
					Children: make(map[string]*Node),
				}
				cur.Children[part] = newDir
				cur = newDir
			}
		}
	}
	return nil
}

func Pwd() string {
	return cwdPath
}

func Cd(path string) error {
	node, fullPath, err := Resolve(path)
	if err != nil {
		return err
	}
	if !node.IsDir {
		return fmt.Errorf("cd: not a directory: %s", path)
	}
	cwd = node
	cwdPath = fullPath
	return nil
}

func Ls(path string) ([]string, error) {
	node, _, err := Resolve(path)
	if err != nil {
		return nil, err
	}
	if !node.IsDir {
		return []string{node.Name}, nil
	}

	var out []string
	if cwdPath != "/" {
		out = append(out, ".", "..")
	}
	for _, child := range node.Children {
		name := child.Name
		if child.IsDir {
			name += "/"
		}
		out = append(out, name)
	}
	return out, nil
}

func Touch(path string) error {
	if root == nil {
		return errors.New("vfs not initialized")
	}

	dirPath, fileName := filepath.Split(path)
	if dirPath == "" {
		dirPath = "."
	}

	parentNode, _, err := Resolve(dirPath)
	if err != nil {
		return fmt.Errorf("cannot access parent directory: %w", err)
	}
	if !parentNode.IsDir {
		return fmt.Errorf("parent is not a directory: %s", dirPath)
	}

	if _, exists := parentNode.Children[fileName]; exists {
		if !parentNode.Children[fileName].IsDir {
			return nil
		}
		return fmt.Errorf("cannot touch '%s': is a directory", path)
	}

	parentNode.Children[fileName] = &Node{
		Name:     fileName,
		IsDir:    false,
		Content:  "",
		Children: nil,
	}

	return nil
}

func Resolve(path string) (*Node, string, error) {
	if root == nil {
		return nil, "", errors.New("vfs not initialized")
	}

	var parts []string
	if strings.HasPrefix(path, "/") {
		parts = splitPath(path)
	} else {
		parts = append(splitPath(cwdPath), splitPath(path)...)
	}

	cur := root
	fullPath := "/"

	for _, p := range parts {
		if p == "" || p == "." {
			continue
		}
		if p == ".." {
			// Упрощённая логика: нельзя выйти выше корня
			cur = root
			fullPath = "/"
			continue
		}

		child, ok := cur.Children[p]
		if !ok {
			return nil, "", fmt.Errorf("no such file or directory: %s", p)
		}
		cur = child
		if fullPath == "/" {
			fullPath += p
		} else {
			fullPath = filepath.Join(fullPath, p)
		}
		fullPath = filepath.ToSlash(fullPath) // для единообразия
	}

	return cur, fullPath, nil
}

func splitPath(p string) []string {
	p = filepath.ToSlash(filepath.Clean(p))
	if p == "/" {
		return []string{}
	}
	return strings.Split(strings.TrimPrefix(p, "/"), "/")
}

func DecodeFile(n *Node) (string, error) {
	if n.IsDir {
		return "", fmt.Errorf("%s is a directory", n.Name)
	}
	// Пытаемся декодировать как base64
	data, err := base64.StdEncoding.DecodeString(n.Content)
	if err != nil {
		// Если не получилось — считаем, что это plain text
		return n.Content, nil
	}
	return string(data), nil
}
