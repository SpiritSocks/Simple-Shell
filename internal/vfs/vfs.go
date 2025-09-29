package vfs

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Node — узел виртуальной файловой системы
type Node struct {
	Name     string           `json:"name"`
	IsDir    bool             `json:"is_dir"`
	Content  string           `json:"content,omitempty"`  // base64 для файлов
	Children map[string]*Node `json:"children,omitempty"` // для директорий
}

var (
	root    *Node
	cwd     *Node
	cwdPath string
)

// Init — загрузка VFS из JSON
func Init(path string) error {
	if path == "" {
		return errors.New("vfs init: no JSON file provided")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("vfs init: %w", err)
	}

	var r Node
	if err := json.Unmarshal(data, &r); err != nil {
		return fmt.Errorf("vfs init: invalid JSON: %w", err)
	}

	if !r.IsDir {
		return errors.New("vfs init: root must be a directory")
	}

	root = &r
	cwd = root
	cwdPath = "/"
	return nil
}

// Pwd — вернуть текущий путь
func Pwd() string {
	return cwdPath
}

// Cd — перейти в директорию
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

// Ls — список содержимого
func Ls(path string) ([]string, error) {
	node, _, err := Resolve(path)
	if err != nil {
		return nil, err
	}
	if !node.IsDir {
		return []string{node.Name}, nil
	}

	var out []string
	if cwd != root {
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

// --- helpers ---

// найти узел по пути
func Resolve(path string) (*Node, string, error) {
	if root == nil {
		return nil, "", errors.New("vfs not initialized")
	}
	var parts []string
	if strings.HasPrefix(path, "/") {
		parts = SplitPath(path)
	} else {
		parts = append(SplitPath(cwdPath), SplitPath(path)...)
	}
	cur := root
	fullPath := "/"
	for _, p := range parts {
		if p == "" || p == "." {
			continue
		}
		if p == ".." {
			// упрощение: всегда возвращаем root при попытке выйти выше
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
	}
	return cur, fullPath, nil
}

func SplitPath(p string) []string {
	return strings.Split(filepath.Clean(p), string(os.PathSeparator))
}

// получить содержимое файла
func DecodeFile(n *Node) (string, error) {
	if n.IsDir {
		return "", fmt.Errorf("%s is a directory", n.Name)
	}
	data, err := base64.StdEncoding.DecodeString(n.Content)
	if err != nil {
		// fallback: treat as raw text
		return n.Content, nil
	}
	return string(data), nil
}
