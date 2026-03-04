package utils

import (
	"os"
	"os/exec"
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

func detectRegistryJDKPaths() []string {
	if runtime.GOOS != "windows" {
		return nil
	}
	regKeys := []string{
		`HKLM\SOFTWARE\JavaSoft\Java Development Kit`,
		`HKLM\SOFTWARE\JavaSoft\JDK`,
		`HKLM\SOFTWARE\Eclipse Adoptium\JDK`,
	}
	var homes []string
	for _, key := range regKeys {
		out, err := exec.Command("reg", "query", key, "/s", "/v", "JavaHome").CombinedOutput()
		if err != nil {
			continue
		}
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if strings.Contains(strings.ToLower(line), "javahome") {
				fields := splitFieldsPreserveTail(line)
				if len(fields) >= 3 {
					val := strings.TrimSpace(strings.Join(fields[2:], " "))
					if val != "" {
						homes = append(homes, val)
					}
				}
			}
		}
	}
	return homes
}

func splitFieldsPreserveTail(s string) []string {
	var fields []string
	current := []rune{}
	spaceRun := false
	for _, r := range s {
		if r == ' ' || r == '\t' {
			if !spaceRun {
				if len(current) > 0 {
					fields = append(fields, strings.TrimSpace(string(current)))
					current = current[:0]
				}
				spaceRun = true
			}
		} else {
			if spaceRun {
				spaceRun = false
			}
			current = append(current, r)
		}
	}
	if len(current) > 0 {
		fields = append(fields, strings.TrimSpace(string(current)))
	}
	return fields
}
