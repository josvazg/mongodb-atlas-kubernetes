package source

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func PkgPath(filePath string) (string, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to compute absolute path: %w", err)
	}
	modPath, baseDir, err := ModPath(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to find module path: %w", err)
	}
	if len(baseDir) > len(absPath) {
		return "", fmt.Errorf("basedir %q is %d bytes, expected to be less than %d bytes", baseDir, len(baseDir), len(absPath))
	}
	relPath := absPath[len(baseDir):]
	if !isDir(absPath) {
		relPath = filepath.Dir(relPath)
	}
	return filepath.Join(modPath, relPath), nil
}

func ModPath(filePath string) (string, string, error) {
	if !filepath.IsAbs(filePath) {
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return "", "", fmt.Errorf("failed to compute absolute path: %w", err)
		}
		filePath = absPath
	}
	dir := filepath.Clean(filepath.Dir(filePath))
	for {
		modFile := filepath.Join(dir, "go.mod")
		if line, err := readFirstLine(modFile); err == nil {
			if line != "" {
				return strings.TrimSpace(strings.TrimPrefix(line, "module ")), dir, nil
			}
		}
		if dir == "/" || dir == "." {
			break // Reached root directory without finding go.mod
		}
		dir = filepath.Dir(dir)
	}
	return "", "", fmt.Errorf("no go.mod found")
}

func readFirstLine(filename string) (string, error) {
	file, err := os.Open(filepath.Clean(filename))
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		return scanner.Text(), nil
	}
	return "", fmt.Errorf("failed to read first line: %w", err)
}

func FilePath(pkgPath string) (string, error) {
	if stdlibDir := filePathForStdLibPkg(pkgPath); stdlibDir != "" {
		return stdlibDir, nil
	}
	currentPath, err := filepath.Abs(".")
	if err != nil {
		return "", fmt.Errorf("failed to compute absolute path: %w", err)
	}
	pkg, dir, err := ModPath(currentPath)
	if err != nil {
		return "", fmt.Errorf("could not find mod base: %w", err)
	}
	if ok, _ := matchesMod(pkg, pkgPath); ok {
		return filePathForCurrentMod(pkg, dir, pkgPath)
	}
	return filePathFor3rdPartyMod(dir, pkgPath)
}

func filePathForStdLibPkg(pkgPath string) string {
	fullDir := filepath.Join(GoRoot(), "src", pkgPath)
	if _, err := os.Stat(fullDir); err != nil {
		return ""
	}
	return fullDir
}

func filePathFor3rdPartyMod(dir, pkgPath string) (string, error) {
	f, err := os.Open(filepath.Join(filepath.Clean(dir), "go.mod"))
	if err != nil {
		return "", fmt.Errorf("could not open go.mod: %w", err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	version := ""
	pkgModPath := pkgPath
	for scanner.Scan() {
		var ok bool
		line := scanner.Text()
		ok, pkgModPath = matchesMod(line, pkgPath)
		if !ok {
			continue
		}
		parts := strings.Split(strings.TrimSpace(line), " ")
		if len(parts) < 2 {
			continue
		}
		version = parts[1]
		break
	}
	if version == "" {
		return "", fmt.Errorf("failed to read version for package %q at go.mod", pkgPath)
	}
	fullDir := filepath.Join(GoPath(), "pkg/mod", fmt.Sprintf("%s@%s", pkgModPath, version))
	if _, err := os.Stat(fullDir); err != nil {
		return "", fmt.Errorf("failed to stat mod package sources at %q: %w", fullDir, err)
	}
	return fullDir, nil
}

func filePathForCurrentMod(pkg, dir, pkgPath string) (string, error) {
	relativePath := pkgPath[len(pkg):]
	fullDir := filepath.Join(dir, relativePath)
	if _, err := os.Stat(fullDir); err != nil {
		return "", fmt.Errorf("failed to stat current mod package sources at %q: %w", fullDir, err)
	}
	return fullDir, nil
}

func matchesMod(line, pkgPath string) (bool, string) {
	levels := len(strings.Split(pkgPath, "/"))
	return matchesModLevels(line, pkgPath, levels)
}

func matchesModLevels(line, pkgPath string, level int) (bool, string) {
	if strings.Contains(line, pkgPath) {
		return true, pkgPath
	}
	if level < 3 {
		return false, ""
	}
	return matchesModLevels(line, filepath.Dir(pkgPath), level-1)
}

func GoRoot() string {
	return goEnvVar("GOROOT")
}

func GoPath() string {
	return goEnvVar("GOPATH")
}

func goEnvVar(name string) string {
	goRoot := os.Getenv(name)
	if goRoot != "" {
		return goRoot
	}
	out, err := exec.Command("go", "env", name).CombinedOutput()
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(out))
}

func isDir(absPath string) bool {
	fi, err := os.Stat(absPath)
	return err == nil && fi.IsDir()
}
