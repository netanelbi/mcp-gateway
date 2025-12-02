package secret

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/docker/mcp-gateway/cmd/docker-mcp/secret-management/formatting"
)

type ListOptions struct {
	JSON bool
}

func List(ctx context.Context, opts ListOptions) error {
	fs, err := NewFileSecrets()
	if err != nil {
		return err
	}

	l, err := fs.List(ctx)
	if err != nil {
		return err
	}

	if opts.JSON {
		if len(l) == 0 {
			l = []StoredSecret{} // Guarantee empty list (instead of displaying null)
		}
		jsonData, err := json.MarshalIndent(l, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(jsonData))
		return nil
	}
	var rows [][]string
	for _, v := range l {
		rows = append(rows, []string{v.Name, v.Provider})
	}
	formatting.PrettyPrintTable(rows, []int{40, 120})
	return nil
}
