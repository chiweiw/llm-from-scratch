package services

import (
	"fmt"
	"os"
	"strings"
)

type MavenParseResult struct {
	MavenPath    string            `json:"mavenPath"`
	SettingsPath string            `json:"settingsPath"`
	RepoLocal    string            `json:"repoLocal"`
	PomFile      string            `json:"pomFile"`
	Goals        []string          `json:"goals"`
	Properties   map[string]string `json:"properties"`
	ArgsArray    []string          `json:"argsArray"`
}

func ParseMavenCommand(cmdLine string) *MavenParseResult {
	result := &MavenParseResult{
		Goals:      []string{},
		Properties: make(map[string]string),
		ArgsArray:  []string{},
	}

	if cmdLine == "" {
		return result
	}

	args := parseCommandLineArgs(cmdLine)
	if remaining, mavenPath := splitLeadingMavenExecutable(args); mavenPath != "" {
		result.MavenPath = mavenPath
		args = remaining
	}
	if os.Getenv("DEPLOY_TOOL_PARSE_DEBUG") == "1" {
		fmt.Printf("[ParseMavenCommand] input=%q\n", cmdLine)
		fmt.Printf("[ParseMavenCommand] args=%q\n", args)
		fmt.Printf("[ParseMavenCommand] mavenPath=%q\n", result.MavenPath)
	}

	i := 0
	for i < len(args) {
		arg := args[i]

		if arg == "-s" && i+1 < len(args) {
			result.SettingsPath = args[i+1]
			result.ArgsArray = append(result.ArgsArray, "-s", args[i+1])
			i += 2
			continue
		} else if len(arg) > 3 && arg[:3] == "-s=" {
			result.SettingsPath = arg[3:]
			result.ArgsArray = append(result.ArgsArray, arg)
			i += 1
			continue
		}

		const repoPrefix = "-Dmaven.repo.local="
		if strings.HasPrefix(arg, repoPrefix) {
			result.RepoLocal = arg[len(repoPrefix):]
			result.ArgsArray = append(result.ArgsArray, arg)
			i += 1
			continue
		}
		if arg == "-Dmaven.repo.local" && i+1 < len(args) {
			result.RepoLocal = args[i+1]
			result.ArgsArray = append(result.ArgsArray, "-Dmaven.repo.local", args[i+1])
			i += 2
			continue
		}

		if arg == "-f" && i+1 < len(args) {
			result.PomFile = args[i+1]
			result.ArgsArray = append(result.ArgsArray, "-f", args[i+1])
			i += 2
			continue
		} else if len(arg) > 3 && arg[:3] == "-f=" {
			result.PomFile = arg[3:]
			result.ArgsArray = append(result.ArgsArray, arg)
			i += 1
			continue
		}

		if len(arg) > 2 && arg[:2] == "-D" && strings.Contains(arg, "=") {
			key := arg[2:strings.Index(arg, "=")]
			value := arg[strings.Index(arg, "=")+1:]
			result.Properties[key] = value
			result.ArgsArray = append(result.ArgsArray, arg)
			i += 1
			continue
		}

		if arg == "-o" || arg == "--offline" || arg == "-q" || arg == "--quiet" ||
			arg == "-U" || arg == "--update-snapshots" || arg == "-X" || arg == "--debug" ||
			arg == "-B" || arg == "--batch-mode" {
			result.ArgsArray = append(result.ArgsArray, arg)
			i += 1
			continue
		}

		if len(arg) > 0 && arg[0] != '-' {
			result.Goals = append(result.Goals, arg)
			result.ArgsArray = append(result.ArgsArray, arg)
			i += 1
			continue
		}

		result.ArgsArray = append(result.ArgsArray, arg)
		i += 1
	}

	if os.Getenv("DEPLOY_TOOL_PARSE_DEBUG") == "1" {
		fmt.Printf("[ParseMavenCommand] settingsPath=%q repoLocal=%q pomFile=%q goals=%q\n", result.SettingsPath, result.RepoLocal, result.PomFile, result.Goals)
	}

	return result
}

func splitLeadingMavenExecutable(args []string) ([]string, string) {
	if len(args) == 0 {
		return args, ""
	}
	if looksLikeMavenExecutable(args[0]) {
		return args[1:], args[0]
	}
	limit := len(args)
	if limit > 12 {
		limit = 12
	}
	for i := 0; i < limit; i++ {
		candidate := strings.Join(args[:i+1], " ")
		if looksLikeMavenExecutable(candidate) {
			return args[i+1:], candidate
		}
	}
	return args, ""
}

func looksLikeMavenExecutable(arg string) bool {
	a := strings.ToLower(strings.TrimSpace(arg))
	if a == "" {
		return false
	}
	if strings.HasSuffix(a, "mvn") || strings.HasSuffix(a, "mvn.cmd") || strings.HasSuffix(a, "mvn.bat") {
		return true
	}
	return strings.Contains(a, "mvn.cmd") || strings.Contains(a, "mvn.bat") || strings.Contains(a, "bin/mvn") || strings.Contains(a, `bin\mvn`)
}

func parseCommandLineArgs(cmdLine string) []string {
	var args []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false

	for _, ch := range cmdLine {
		switch ch {
		case '\'':
			if !inDoubleQuote {
				inSingleQuote = !inSingleQuote
			} else {
				current.WriteRune(ch)
			}
		case '"':
			if !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
			} else {
				current.WriteRune(ch)
			}
		case ' ':
			if inSingleQuote || inDoubleQuote {
				current.WriteRune(ch)
			} else if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

