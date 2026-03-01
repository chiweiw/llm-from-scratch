package services

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func AutoDetectJDK() []map[string]string {
	return detectJDK(false)
}

func DetectJDK() []map[string]string {
	return detectJDK(true)
}

func detectJDK(includeRegistry bool) []map[string]string {
	jdks := make([]map[string]string, 0)

	javaHome := os.Getenv("JAVA_HOME")
	if javaHome != "" {
		jdks = append(jdks, map[string]string{
			"path":   javaHome,
			"source": "JAVA_HOME",
		})
	}

	pathVar := os.Getenv("PATH")
	if pathVar != "" {
		for _, p := range strings.Split(pathVar, string(os.PathListSeparator)) {
			if strings.Contains(strings.ToLower(p), "java") || strings.Contains(strings.ToLower(p), "jdk") {
				if strings.HasSuffix(p, "bin") {
					parent := filepath.Dir(p)
					jdks = append(jdks, map[string]string{
						"path":   parent,
						"source": "PATH",
					})
				}
			}
		}
	}

	if includeRegistry && runtime.GOOS == "windows" {
		for _, jdkPath := range detectRegistryJDKPaths() {
			if jdkPath == "" {
				continue
			}
			exists := false
			for _, j := range jdks {
				if j["path"] == jdkPath {
					exists = true
					break
				}
			}
			if !exists {
				jdks = append(jdks, map[string]string{
					"path":   jdkPath,
					"source": "Registry",
				})
			}
		}
	}

	seen := make(map[string]bool)
	unique := make([]map[string]string, 0)
	for _, j := range jdks {
		if !seen[j["path"]] {
			seen[j["path"]] = true
			unique = append(unique, j)
		}
	}

	return unique
}

