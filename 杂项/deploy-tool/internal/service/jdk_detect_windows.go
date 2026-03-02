//go:build windows

package service

import "golang.org/x/sys/windows/registry"

func detectRegistryJDKPaths() []string {
	var paths []string
	regPaths := []string{
		`SOFTWARE\JavaSoft\Java Development Kit`,
		`SOFTWARE\JavaSoft\JDK`,
		`SOFTWARE\Eclipse Adoptium\JDK`,
	}
	for _, regPath := range regPaths {
		paths = append(paths, readRegistryJavaHome(regPath)...)
	}
	return paths
}

func readRegistryJavaHome(regPath string) []string {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, regPath, registry.READ)
	if err != nil {
		return nil
	}
	defer key.Close()

	names, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return nil
	}

	var homes []string
	for _, name := range names {
		subKey, err := registry.OpenKey(key, name, registry.READ)
		if err != nil {
			continue
		}
		javaHome, _, err := subKey.GetStringValue("JavaHome")
		subKey.Close()
		if err == nil && javaHome != "" {
			homes = append(homes, javaHome)
		}
	}
	return homes
}

