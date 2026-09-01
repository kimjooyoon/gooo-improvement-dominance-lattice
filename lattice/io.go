package lattice

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

func DigestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func DigestFile(path string) (string, []byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", nil, err
	}
	return DigestBytes(data), data, nil
}

func ReadJSON(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}

func WriteJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return WriteBytes(path, data)
}

func WriteBytes(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".gooo-lattice-*")
	if err != nil { return err }
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o644); err != nil { _ = tmp.Close(); return err }
	if _, err := tmp.Write(data); err != nil { _ = tmp.Close(); return err }
	if err := tmp.Close(); err != nil { return err }
	return os.Rename(tmpName, path)
}

func EnsureCallerOwnedOutput(path string) error {
	if path == "" || !filepath.IsAbs(path) { return fmt.Errorf("output must be an absolute caller-owned path") }
	info, err := os.Stat(path)
	if err != nil { return fmt.Errorf("output directory: %w", err) }
	if !info.IsDir() { return fmt.Errorf("output is not a directory") }
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil { return err }
	working, err := os.Getwd()
	if err != nil { return err }
	working, err = filepath.EvalSymlinks(working)
	if err != nil { return err }
	root := findGitRoot(working)
	if root != "" && pathWithin(root, resolved) { return fmt.Errorf("caller-owned output may not be inside input repository") }
	return nil
}

func findGitRoot(start string) string {
	for {
		info, err := os.Stat(filepath.Join(start, ".git"))
		if err == nil && (info.IsDir() || info.Mode().IsRegular()) { return start }
		parent := filepath.Dir(start)
		if parent == start { return "" }
		start = parent
	}
}

func pathWithin(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && len(rel) > 3 && rel[:3] != ".."+string(filepath.Separator)
}

func ValidateGeneratedGo(path string) error {
	data, err := os.ReadFile(path)
	if err != nil { return err }
	file, err := parser.ParseFile(token.NewFileSet(), path, data, parser.AllErrors)
	if err != nil { return fmt.Errorf("generated evaluator is not valid Go: %w", err) }
	if file.Name == nil || file.Name.Name != "generated" { return fmt.Errorf("generated evaluator package must be generated") }
	return nil
}
