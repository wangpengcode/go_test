package pathutil

import (
	"os"
	"path/filepath"
)

// FirstExistingFile 返回第一个“存在且是文件”的路径。
func FirstExistingFile(paths ...string) (string, bool) {
	for _, p := range paths {
		if p == "" {
			continue
		}
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		if info.Mode().IsRegular() {
			return p, true
		}
	}
	return "", false
}

// FirstExistingDir 返回第一个“存在且是目录”的路径。
func FirstExistingDir(paths ...string) (string, bool) {
	for _, p := range paths {
		if p == "" {
			continue
		}
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		if info.IsDir() {
			return p, true
		}
	}
	return "", false
}

// FindUp 从 startDir 开始，向上逐级父目录查找 filename。
// 返回值是“包含该文件的目录路径”。
func FindUp(startDir, filename string) (string, bool) {
	dir := startDir
	for {
		if dir == "" || dir == string(filepath.Separator) {
			return "", false
		}
		_, err := os.Stat(filepath.Join(dir, filename))
		if err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// ModuleRootFrom 尝试定位 Go 模块根目录（包含 go.mod 的目录）。
// 查找起点是 startPath 所在的目录。
func ModuleRootFrom(startPath string) (string, bool) {
	start := startPath
	if start == "" {
		return "", false
	}
	if info, err := os.Stat(start); err == nil && info.IsDir() {
		// start 本身就是目录，保持不变
	} else {
		start = filepath.Dir(start)
	}
	return FindUp(start, "go.mod")
}

// ResolveMaybeRelativeToRoot 会把相对路径 p 按 root 拼成新路径来尝试解析；
// 只有“解析后路径存在”才会返回解析后的结果。
// 如果 p 本身就存在（以当前工作目录为基准），则直接返回 p。
func ResolveMaybeRelativeToRoot(p, root string) string {
	if p == "" || filepath.IsAbs(p) {
		return p
	}
	if _, err := os.Stat(p); err == nil {
		return p
	}
	if root == "" {
		return p
	}
	candidate := filepath.Join(root, p)
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	return p
}
