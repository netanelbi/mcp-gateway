package secret

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/docker/mcp-gateway/pkg/config"
)

const DefaultSecretsFile = "secrets.env"

// FileSecrets represents a file-based secrets store
type FileSecrets struct {
	Path string
}

// NewFileSecrets creates a new FileSecrets instance
// Uses the default secrets file in ~/.docker/mcp/secrets.env
func NewFileSecrets() (*FileSecrets, error) {
	path, err := config.FilePath(DefaultSecretsFile)
	if err != nil {
		return nil, err
	}
	return &FileSecrets{Path: path}, nil
}

// List returns all secret names from the file
func (f *FileSecrets) List(ctx context.Context) ([]StoredSecret, error) {
	secrets, err := f.readAll(ctx)
	if err != nil {
		if os.IsNotExist(err) {
			return []StoredSecret{}, nil
		}
		return nil, err
	}

	var result []StoredSecret
	for name := range secrets {
		result = append(result, StoredSecret{
			Name:     name,
			Provider: "file",
		})
	}

	// Sort by name for consistent output
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

// Set sets a secret value in the file
func (f *FileSecrets) Set(ctx context.Context, name, value string) error {
	secrets, err := f.readAll(ctx)
	if err != nil {
		if os.IsNotExist(err) {
			secrets = make(map[string]string)
		} else {
			return err
		}
	}

	secrets[name] = value
	return f.writeAll(secrets)
}

// Delete removes a secret from the file
func (f *FileSecrets) Delete(ctx context.Context, name string) error {
	secrets, err := f.readAll(ctx)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("secret %s not found", name)
		}
		return err
	}

	if _, ok := secrets[name]; !ok {
		return fmt.Errorf("secret %s not found", name)
	}

	delete(secrets, name)
	return f.writeAll(secrets)
}

// DeleteAll removes all secrets from the file
func (f *FileSecrets) DeleteAll(ctx context.Context) error {
	return f.writeAll(make(map[string]string))
}

// readAll reads all secrets from the file
func (f *FileSecrets) readAll(ctx context.Context) (map[string]string, error) {
	secrets := make(map[string]string)

	buf, err := os.ReadFile(f.Path)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(buf))
	for scanner.Scan() {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue // Skip invalid lines
		}

		secrets[key] = value
	}

	return secrets, scanner.Err()
}

// writeAll writes all secrets to the file
func (f *FileSecrets) writeAll(secrets map[string]string) error {
	// Ensure directory exists
	dir := filepath.Dir(f.Path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// Sort keys for consistent output
	var keys []string
	for k := range secrets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	for _, k := range keys {
		buf.WriteString(fmt.Sprintf("%s=%s\n", k, secrets[k]))
	}

	return os.WriteFile(f.Path, buf.Bytes(), 0o600)
}

// StoredSecret represents a secret stored in the file (matches desktop.StoredSecret interface)
type StoredSecret struct {
	Name     string `json:"name"`
	Provider string `json:"provider,omitempty"`
}
