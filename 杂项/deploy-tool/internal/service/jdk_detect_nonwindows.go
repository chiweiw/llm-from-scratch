//go:build !windows

package service

func detectRegistryJDKPaths() []string {
	return nil
}

