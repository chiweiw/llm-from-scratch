package service

import "testing"

func TestParseMavenCommand_IDEACommandLine(t *testing.T) {
	cmd := `D:\Program Files\JetBrains\IntelliJ IDEA 2025.3.3\plugins\maven\lib\maven3\bin\mvn.cmd -Didea.version=2025.3.3 -Dmaven.ext.class.path=D:\Program Files\JetBrains\IntelliJ IDEA 2025.3.3\plugins\maven\lib\maven-event-listener.jar -Djansi.passthrough=true -Dstyle.color=always -s D:\java_tools\apache-maven-3.9.12\conf\settings_sgt0903.xml -Dmaven.repo.local=D:\m2\repository package -f pom.xml`
	result := ParseMavenCommand(cmd)

	if result.MavenPath == "" {
		t.Fatalf("expected mavenPath, got empty")
	}
	if result.RepoLocal != `D:\m2\repository` {
		t.Fatalf("expected repoLocal %q, got %q", `D:\m2\repository`, result.RepoLocal)
	}
	if result.SettingsPath != `D:\java_tools\apache-maven-3.9.12\conf\settings_sgt0903.xml` {
		t.Fatalf("expected settingsPath %q, got %q", `D:\java_tools\apache-maven-3.9.12\conf\settings_sgt0903.xml`, result.SettingsPath)
	}
}

