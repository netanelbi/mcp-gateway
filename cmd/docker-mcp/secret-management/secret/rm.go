package secret

import (
	"context"
	"errors"
	"fmt"
)

type RmOpts struct {
	All bool
}

func Remove(ctx context.Context, names []string, opts RmOpts) error {
	fs, err := NewFileSecrets()
	if err != nil {
		return err
	}

	if opts.All && len(names) == 0 {
		l, err := fs.List(ctx)
		if err != nil {
			return err
		}
		for _, secret := range l {
			names = append(names, secret.Name)
		}
	}

	if opts.All && len(names) == 0 {
		fmt.Println("no secrets to remove")
		return nil
	}

	var errs []error
	for _, name := range names {
		if err := fs.Delete(ctx, name); err != nil {
			errs = append(errs, err)
			fmt.Printf("failed removing secret %s: %v\n", name, err)
			continue
		}
		fmt.Printf("removed secret %s\n", name)
	}
	return errors.Join(errs...)
}
