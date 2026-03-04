package service

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

func isValidJDK(path string) bool {
	var javaCmd, javacCmd string
	if runtime.GOOS == "windows" {
		javaCmd = filepath.Join(path, "bin", "java.exe")
		javacCmd = filepath.Join(path, "bin", "javac.exe")
	} else {
		javaCmd = filepath.Join(path, "bin", "java")
		javacCmd = filepath.Join(path, "bin", "javac")
	}
	return fileExists(javaCmd) || fileExists(javacCmd)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func detectJDK(includeRegistry bool) []map[string]string {
	jdks := make([]map[string]string, 0)

	javaHome := os.Getenv("JAVA_HOME")
	if javaHome != "" && isValidJDK(javaHome) {
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
					if isValidJDK(parent) {
						jdks = append(jdks, map[string]string{
							"path":   parent,
							"source": "PATH",
						})
					}
				}
			}
		}
	}

	if includeRegistry && runtime.GOOS == "windows" {
		for _, jdkPath := range detectRegistryJDKPaths() {
			if jdkPath == "" {
				continue
			}
			if !isValidJDK(jdkPath) {
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
