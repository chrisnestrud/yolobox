package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()
	if cfg.Image != "ghcr.io/finbarr/yolobox:latest" {
		t.Errorf("expected default image ghcr.io/finbarr/yolobox:latest, got %s", cfg.Image)
	}
	if cfg.Runtime != "" {
		t.Errorf("expected empty default runtime, got %s", cfg.Runtime)
	}
}

func TestMergeConfig(t *testing.T) {
	dst := Config{
		Runtime: "docker",
		Image:   "old-image",
	}
	src := Config{
		Image:     "new-image",
		SSHAgent:  true,
		NoNetwork: true,
		Scratch:   true,
	}

	mergeConfig(&dst, src)

	if dst.Runtime != "docker" {
		t.Errorf("expected runtime to stay docker, got %s", dst.Runtime)
	}
	if dst.Image != "new-image" {
		t.Errorf("expected image to be new-image, got %s", dst.Image)
	}
	if !dst.SSHAgent {
		t.Error("expected SSHAgent to be true")
	}
	if !dst.NoNetwork {
		t.Error("expected NoNetwork to be true")
	}
	if !dst.Scratch {
		t.Error("expected Scratch to be true")
	}
}

func TestMergeConfigCustomize(t *testing.T) {
	dst := Config{}
	src := Config{
		Customize: CustomizeConfig{
			Packages:   []string{"maven", "default-jdk"},
			Dockerfile: ".yolobox.Dockerfile",
		},
	}

	mergeConfig(&dst, src)

	if len(dst.Customize.Packages) != 2 {
		t.Fatalf("expected 2 customize packages, got %v", dst.Customize.Packages)
	}
	if dst.Customize.Dockerfile != ".yolobox.Dockerfile" {
		t.Fatalf("expected customize dockerfile to merge, got %q", dst.Customize.Dockerfile)
	}
}

func TestResolvePath(t *testing.T) {
	home, _ := os.UserHomeDir()
	projectDir := "/project"

	tests := []struct {
		input    string
		expected string
	}{
		{"~/foo", filepath.Join(home, "foo")},
		{"~", home},
		{"./bar", "/project/bar"},
		{"/absolute/path", "/absolute/path"},
		{"relative", "relative"}, // non-dotted relative paths pass through
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := resolvePath(tt.input, projectDir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("resolvePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestResolvePathEmpty(t *testing.T) {
	_, err := resolvePath("", "/project")
	if err == nil {
		t.Error("expected error for empty path")
	}
}

func TestResolveMount(t *testing.T) {
	home, _ := os.UserHomeDir()
	projectDir := "/project"

	tests := []struct {
		input    string
		expected string
	}{
		{"./src:/app/src", "/project/src:/app/src"},
		{"~/secrets:/secrets:ro", filepath.Join(home, "secrets") + ":/secrets:ro"},
		{"/absolute:/dst", "/absolute:/dst"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := resolveMount(tt.input, projectDir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("resolveMount(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestResolveMountInvalid(t *testing.T) {
	_, err := resolveMount("no-colon", "/project")
	if err == nil {
		t.Error("expected error for invalid mount")
	}
}

func TestResolvedRuntimeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "auto"},
		{"docker", "docker"},
		{"podman", "podman"},
		{"colima", "docker"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := resolvedRuntimeName(tt.input)
			if result != tt.expected {
				t.Errorf("resolvedRuntimeName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuildRunArgs(t *testing.T) {
	cfg := Config{
		Image:  "test-image",
		Env:    []string{"FOO=bar"},
		Mounts: []string{},
	}

	args, _, err := buildRunArgs(cfg, "/test/project", []string{"bash"}, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	argsStr := strings.Join(args, " ")

	// Check essential flags are present
	if !strings.Contains(argsStr, "-it") {
		t.Error("expected -it flag for interactive mode")
	}
	if !strings.Contains(argsStr, "-w /test/project") {
		t.Error("expected -w /test/project (workdir should be actual project path)")
	}
	if !strings.Contains(argsStr, "YOLOBOX=1") {
		t.Error("expected YOLOBOX=1 env var")
	}
	if !strings.Contains(argsStr, "YOLOBOX_PROJECT_PATH=/test/project") {
		t.Error("expected YOLOBOX_PROJECT_PATH env var")
	}
	if !strings.Contains(argsStr, "FOO=bar") {
		t.Error("expected FOO=bar env var")
	}
	if !strings.Contains(argsStr, "test-image") {
		t.Error("expected test-image")
	}

	// Check volume mounts
	if !strings.Contains(argsStr, "yolobox-home:/home/yolo") {
		t.Error("expected yolobox-home volume")
	}
	if !strings.Contains(argsStr, "yolobox-cache:/var/cache") {
		t.Error("expected yolobox-cache volume")
	}

	// Verify no --network flag when using default network
	if strings.Contains(argsStr, "--network") {
		t.Error("expected no --network flag for default network behavior")
	}
}

func TestBuildRunArgsNoYolo(t *testing.T) {
	cfg := Config{
		Image:  "test-image",
		NoYolo: true,
	}

	args, _, err := buildRunArgs(cfg, "/test/project", []string{"bash"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	argsStr := strings.Join(args, " ")
	if !strings.Contains(argsStr, "YOLOBOX=1") {
		t.Error("expected YOLOBOX=1 env var to be present")
	}
	if !strings.Contains(argsStr, "NO_YOLO=1") {
		t.Error("expected NO_YOLO=1 env var to be present")
	}
}

func TestBuildRunArgsRuntimeArgs(t *testing.T) {
	cfg := Config{
		Image:       "test-image",
		RuntimeArgs: []string{"--security-opt", "seccomp=unconfined"},
	}

	args, _, err := buildRunArgs(cfg, "/test/project", []string{"bash"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	imageIdx := -1
	for i, arg := range args {
		if arg == "test-image" {
			imageIdx = i
			break
		}
	}
	if imageIdx == -1 {
		t.Fatalf("test-image not found in args: %v", args)
	}
	if imageIdx < 2 {
		t.Fatalf("runtime args missing before image: %v", args)
	}
	if args[imageIdx-2] != "--security-opt" || args[imageIdx-1] != "seccomp=unconfined" {
		t.Fatalf("runtime args not placed before image: %v", args)
	}
}

func TestValidateRuntimeOptions(t *testing.T) {
	cfg := Config{
		CPUs:    "2.5",
		Memory:  "4g",
		ShmSize: "1GiB",
	}

	if err := validateRuntimeOptions(cfg); err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}
}

func TestValidateRuntimeOptionsInvalid(t *testing.T) {
	tests := []Config{
		{CPUs: "zero"},
		{CPUs: "-1"},
		{Memory: "a lot"},
		{ShmSize: "123x"},
		{RuntimeArgs: []string{"", " "}},
	}

	for _, cfg := range tests {
		if err := validateRuntimeOptions(cfg); err == nil {
			t.Fatalf("expected error for cfg %#v", cfg)
		}
	}
}

func TestParseMultilineInput(t *testing.T) {
	input := "alpha\n beta\r\n\ngamma"
	values := parseMultilineInput(input)
	expected := []string{"alpha", "beta", "gamma"}
	if len(values) != len(expected) {
		t.Fatalf("expected %d values, got %d (%v)", len(expected), len(values), values)
	}
	for i, val := range expected {
		if values[i] != val {
			t.Fatalf("expected %q at index %d, got %q", val, i, values[i])
		}
	}
}

func TestBuildRunArgsNoNetwork(t *testing.T) {
	cfg := Config{
		Image:     "test-image",
		NoNetwork: true,
	}

	args, _, err := buildRunArgs(cfg, "/test/project", []string{"bash"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	argsStr := strings.Join(args, " ")
	if !strings.Contains(argsStr, "--network none") {
		t.Error("expected --network none for NoNetwork")
	}
}

func TestBuildRunArgsReadonlyProject(t *testing.T) {
	cfg := Config{
		Image:           "test-image",
		ReadonlyProject: true,
	}

	args, _, err := buildRunArgs(cfg, "/test/project", []string{"bash"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	argsStr := strings.Join(args, " ")
	if !strings.Contains(argsStr, "/test/project:/test/project:ro") {
		t.Error("expected /test/project:/test/project:ro for ReadonlyProject")
	}
	if !strings.Contains(argsStr, "yolobox-output:/output") {
		t.Error("expected yolobox-output volume for ReadonlyProject")
	}
}

func TestBuildRunArgsNonInteractive(t *testing.T) {
	cfg := Config{
		Image: "test-image",
	}

	args, _, err := buildRunArgs(cfg, "/test/project", []string{"echo", "hello"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	argsStr := strings.Join(args, " ")
	if strings.Contains(argsStr, "-it") {
		t.Error("expected no -it flag for non-interactive mode")
	}
}

func TestBuildRunArgsScratch(t *testing.T) {
	cfg := Config{
		Image:   "test-image",
		Scratch: true,
	}

	args, _, err := buildRunArgs(cfg, "/test/project", []string{"bash"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	argsStr := strings.Join(args, " ")
	if strings.Contains(argsStr, "yolobox-home:/home/yolo") {
		t.Error("expected no yolobox-home volume with Scratch")
	}
	if strings.Contains(argsStr, "yolobox-cache:/var/cache") {
		t.Error("expected no yolobox-cache volume with Scratch")
	}
	// Verify project mount is still present (at real path)
	if !strings.Contains(argsStr, "/test/project:/test/project") {
		t.Error("expected project mount to still be present with Scratch")
	}
	// Verify no /output volume without ReadonlyProject
	if strings.Contains(argsStr, "/output") {
		t.Error("expected no /output volume without ReadonlyProject")
	}
}

func TestBuildRunArgsScratchWithReadonlyProject(t *testing.T) {
	cfg := Config{
		Image:           "test-image",
		Scratch:         true,
		ReadonlyProject: true,
	}

	args, _, err := buildRunArgs(cfg, "/test/project", []string{"bash"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	argsStr := strings.Join(args, " ")
	// Should have anonymous /output volume, not named yolobox-output
	if strings.Contains(argsStr, "yolobox-output:/output") {
		t.Error("expected anonymous /output volume with Scratch, got named volume")
	}
	if !strings.Contains(argsStr, "-v /output") {
		t.Error("expected anonymous /output volume for readonly-project with Scratch")
	}
}

func TestParseFlagsScratch(t *testing.T) {
	cfg, rest, err := parseBaseFlags("run", []string{"--scratch", "echo", "hello"}, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.Scratch {
		t.Error("expected Scratch to be true after parsing --scratch flag")
	}
	if len(rest) != 2 || rest[0] != "echo" || rest[1] != "hello" {
		t.Errorf("expected remaining args [echo hello], got %v", rest)
	}
}

func TestParseFlagsCustomize(t *testing.T) {
	projectDir := t.TempDir()
	cfg, rest, err := parseBaseFlags("run", []string{
		"--packages", "default-jdk, maven",
		"--customize-file", ".yolobox.Dockerfile",
		"--rebuild-image",
		"java", "--version",
	}, projectDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Customize.Packages) != 2 || cfg.Customize.Packages[0] != "default-jdk" || cfg.Customize.Packages[1] != "maven" {
		t.Fatalf("unexpected customize packages: %v", cfg.Customize.Packages)
	}
	if cfg.Customize.Dockerfile != ".yolobox.Dockerfile" {
		t.Fatalf("unexpected customize dockerfile: %q", cfg.Customize.Dockerfile)
	}
	if !cfg.RebuildImage {
		t.Fatal("expected RebuildImage to be true")
	}
	if len(rest) != 2 || rest[0] != "java" || rest[1] != "--version" {
		t.Fatalf("unexpected remaining args: %v", rest)
	}
}

func TestParseFlagsCustomizeInvalidPackage(t *testing.T) {
	_, _, err := parseBaseFlags("run", []string{"--packages", "default-jdk,$(evil)", "java"}, t.TempDir())
	if err == nil {
		t.Fatal("expected invalid package name error")
	}
}

func TestHasCustomization(t *testing.T) {
	if hasCustomization(Config{}) {
		t.Fatal("expected empty config to have no customization")
	}
	if !hasCustomization(Config{Customize: CustomizeConfig{Packages: []string{"default-jdk"}}}) {
		t.Fatal("expected packages customization to be detected")
	}
	if !hasCustomization(Config{Customize: CustomizeConfig{Dockerfile: ".yolobox.Dockerfile"}}) {
		t.Fatal("expected dockerfile customization to be detected")
	}
}

func TestGenerateCustomDockerfile(t *testing.T) {
	dockerfile, err := generateCustomDockerfile("ghcr.io/finbarr/yolobox:latest", []string{"maven", "default-jdk", "maven"}, "USER root\nRUN echo hi\nUSER yolo\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(dockerfile, "FROM ghcr.io/finbarr/yolobox:latest") {
		t.Fatalf("expected base image in generated Dockerfile:\n%s", dockerfile)
	}
	if !strings.Contains(dockerfile, "apt-get install -y --no-install-recommends default-jdk maven") {
		t.Fatalf("expected sorted package install in generated Dockerfile:\n%s", dockerfile)
	}
	if !strings.Contains(dockerfile, "RUN echo hi") {
		t.Fatalf("expected fragment content in generated Dockerfile:\n%s", dockerfile)
	}
}

func TestGenerateCustomDockerfileRejectsInvalidPackage(t *testing.T) {
	_, err := generateCustomDockerfile("base", []string{"$(evil)"}, "")
	if err == nil {
		t.Fatal("expected invalid package to be rejected")
	}
}

func TestCustomImageTagStable(t *testing.T) {
	tagA := customImageTag("sha256:base", "FROM base\n", []string{"maven", "default-jdk"})
	tagB := customImageTag("sha256:base", "FROM base\n", []string{"default-jdk", "maven", "maven"})
	if tagA != tagB {
		t.Fatalf("expected normalized package order to yield same tag, got %q vs %q", tagA, tagB)
	}
}

func TestResolveCustomizeFile(t *testing.T) {
	projectDir := t.TempDir()
	got, err := resolveCustomizeFile(".yolobox.Dockerfile", projectDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != filepath.Join(projectDir, ".yolobox.Dockerfile") {
		t.Fatalf("unexpected resolved customize file: %q", got)
	}
}

func TestLoadCustomizeFragment(t *testing.T) {
	projectDir := t.TempDir()
	path := filepath.Join(projectDir, ".yolobox.Dockerfile")
	if err := os.WriteFile(path, []byte("USER root\nRUN echo hi\n"), 0644); err != nil {
		t.Fatalf("failed to write test customize file: %v", err)
	}

	got, err := loadCustomizeFragment(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "USER root\nRUN echo hi" {
		t.Fatalf("unexpected customize fragment contents: %q", got)
	}
}

func TestStringSliceFlag(t *testing.T) {
	var s stringSliceFlag

	if err := s.Set("first"); err != nil {
		t.Fatalf("Set(first) failed: %v", err)
	}
	if err := s.Set("second"); err != nil {
		t.Fatalf("Set(second) failed: %v", err)
	}

	if len(s) != 2 {
		t.Errorf("expected 2 values, got %d", len(s))
	}
	if s[0] != "first" || s[1] != "second" {
		t.Errorf("unexpected values: %v", s)
	}
	if s.String() != "first,second" {
		t.Errorf("unexpected String(): %s", s.String())
	}
}

func TestAutoPassthroughEnvVars(t *testing.T) {
	// Check that common API keys are in the list
	expected := []string{
		"ANTHROPIC_API_KEY",
		"CLAUDE_CODE_OAUTH_TOKEN",
		"OPENAI_API_KEY",
		"COPILOT_GITHUB_TOKEN",
		"GITHUB_TOKEN",
		"GH_TOKEN",
	}

	for _, key := range expected {
		found := false
		for _, v := range autoPassthroughEnvVars {
			if v == key {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %s in autoPassthroughEnvVars", key)
		}
	}
}

func TestToolShortcuts(t *testing.T) {
	// Check that expected tools are shortcuts
	expected := []string{
		"claude",
		"codex",
		"gemini",
		"opencode",
		"copilot",
	}

	for _, tool := range expected {
		if !isToolShortcut(tool) {
			t.Errorf("expected %s to be a tool shortcut", tool)
		}
	}

	// Check that non-tools are not shortcuts
	nonTools := []string{"run", "help", "version", "setup", "foo"}
	for _, cmd := range nonTools {
		if isToolShortcut(cmd) {
			t.Errorf("expected %s NOT to be a tool shortcut", cmd)
		}
	}
}

func TestBuildRunArgsNetwork(t *testing.T) {
	cfg := Config{
		Image:   "test-image",
		Network: "dev_network",
	}

	args, _, err := buildRunArgs(cfg, "/test/project", []string{"echo"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	argsStr := strings.Join(args, " ")
	if !strings.Contains(argsStr, "--network dev_network") {
		t.Error("expected --network dev_network")
	}
}

func TestBuildRunArgsPod(t *testing.T) {
	cfg := Config{
		Image:     "test-image",
		Pod:       "dev-pod",
		Network:   "ignored_network",
		NoNetwork: true,
	}

	args, _, err := buildRunArgs(cfg, "/test/project", []string{"echo"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	argsStr := strings.Join(args, " ")
	if !strings.Contains(argsStr, "--pod dev-pod") {
		t.Error("expected --pod dev-pod")
	}
	if strings.Contains(argsStr, "--network ignored_network") || strings.Contains(argsStr, "--network none") {
		t.Errorf("expected pod mode to suppress network flags, got: %s", argsStr)
	}
}

func TestBuildRunArgsResourceLimits(t *testing.T) {
	cfg := Config{
		Image:   "test-image",
		CPUs:    "4",
		Memory:  "8g",
		ShmSize: "2g",
		GPUs:    "all",
	}

	args, _, err := buildRunArgs(cfg, "/test/project", []string{"bash"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	argsStr := strings.Join(args, " ")
	for _, expect := range []string{
		"--cpus 4",
		"--memory 8g",
		"--shm-size 2g",
		"--gpus all",
	} {
		if !strings.Contains(argsStr, expect) {
			t.Fatalf("expected run args to contain %s, got %s", expect, argsStr)
		}
	}
}

func TestBuildRunArgsDeviceSecurity(t *testing.T) {
	cfg := Config{
		Image:   "test-image",
		Devices: []string{"/dev/kvm:/dev/kvm"},
		CapAdd:  []string{"SYS_PTRACE"},
		CapDrop: []string{"MKNOD"},
		GPUs:    "all",
	}

	args, _, err := buildRunArgs(cfg, "/test/project", []string{"bash"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	argsStr := strings.Join(args, " ")
	for _, expect := range []string{
		"--device /dev/kvm:/dev/kvm",
		"--cap-add SYS_PTRACE",
		"--cap-drop MKNOD",
		"--gpus all",
	} {
		if !strings.Contains(argsStr, expect) {
			t.Fatalf("expected run args to contain %s, got %s", expect, argsStr)
		}
	}
}

func TestMergeConfigNetwork(t *testing.T) {
	dst := Config{
		Runtime: "docker",
		Image:   "old-image",
	}
	src := Config{
		Network: "my_network",
	}

	mergeConfig(&dst, src)

	if dst.Network != "my_network" {
		t.Errorf("expected Network to be my_network, got %s", dst.Network)
	}
}

func TestMergeConfigPod(t *testing.T) {
	dst := Config{
		Runtime: "docker",
		Image:   "old-image",
	}
	src := Config{
		Pod: "my_pod",
	}

	mergeConfig(&dst, src)

	if dst.Pod != "my_pod" {
		t.Errorf("expected Pod to be my_pod, got %s", dst.Pod)
	}
}

func TestParseFlagsNetwork(t *testing.T) {
	cfg, rest, err := parseBaseFlags("run", []string{"--network", "mynet", "echo"}, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Network != "mynet" {
		t.Errorf("expected Network=mynet, got %s", cfg.Network)
	}
	if len(rest) != 1 || rest[0] != "echo" {
		t.Errorf("expected remaining args [echo], got %v", rest)
	}
}

func TestParseFlagsPod(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	cfg, rest, err := parseBaseFlags("run", []string{"--pod", "mypod", "echo"}, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Pod != "mypod" {
		t.Errorf("expected Pod=mypod, got %s", cfg.Pod)
	}
	if len(rest) != 1 || rest[0] != "echo" {
		t.Errorf("expected remaining args [echo], got %v", rest)
	}
}

func TestParseFlagsResourceLimits(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	args := []string{
		"--cpus", "4",
		"--memory", "8g",
		"--shm-size", "1g",
		"--gpus", "all",
		"--runtime-arg", "--cpu-shares=512",
		"echo",
	}

	cfg, rest, err := parseBaseFlags("run", args, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.CPUs != "4" {
		t.Errorf("unexpected CPU limit: %+v", cfg.CPUs)
	}
	if cfg.Memory != "8g" || cfg.ShmSize != "1g" {
		t.Errorf("unexpected memory limits: %+v", cfg)
	}
	if cfg.GPUs != "all" {
		t.Errorf("expected GPUs=all, got %s", cfg.GPUs)
	}
	if len(cfg.RuntimeArgs) != 1 || cfg.RuntimeArgs[0] != "--cpu-shares=512" {
		t.Fatalf("expected runtime args to be captured, got %+v", cfg.RuntimeArgs)
	}
	if len(rest) != 1 || rest[0] != "echo" {
		t.Errorf("expected remaining args [echo], got %v", rest)
	}
}

func TestParseFlagsDeviceSecurity(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	args := []string{
		"--device", "/dev/kvm:/dev/kvm",
		"--cap-add", "SYS_PTRACE",
		"--cap-drop", "MKNOD",
		"--runtime-arg", "--security-opt",
		"--runtime-arg", "seccomp=unconfined",
		"--gpus", "all",
		"echo",
	}

	cfg, rest, err := parseBaseFlags("run", args, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectSliceEqual(t, cfg.Devices, []string{"/dev/kvm:/dev/kvm"})
	expectSliceEqual(t, cfg.CapAdd, []string{"SYS_PTRACE"})
	expectSliceEqual(t, cfg.CapDrop, []string{"MKNOD"})
	expectSliceEqual(t, cfg.RuntimeArgs, []string{"--security-opt", "seccomp=unconfined"})
	if cfg.GPUs != "all" {
		t.Errorf("expected GPUs=all, got %s", cfg.GPUs)
	}
	if len(rest) != 1 || rest[0] != "echo" {
		t.Errorf("expected remaining args [echo], got %v", rest)
	}
}

func expectSliceEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("expected slice %v, got %v", want, got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("expected slice %v, got %v", want, got)
		}
	}
}

func TestParseFlagsNetworkConflict(t *testing.T) {
	_, _, err := parseBaseFlags("run", []string{"--network", "mynet", "--no-network", "echo"}, t.TempDir())
	if err == nil {
		t.Error("expected error for --network with --no-network")
	}
	if err != nil && !strings.Contains(err.Error(), "cannot use --network with --no-network") {
		t.Errorf("expected conflict error message, got: %v", err)
	}
}

func TestParseFlagsPodNetworkConflict(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	_, _, err := parseBaseFlags("run", []string{"--pod", "mypod", "--network", "mynet", "echo"}, t.TempDir())
	if err == nil {
		t.Error("expected error for --pod with --network")
	}
	if err != nil && !strings.Contains(err.Error(), "cannot use --pod with --network") {
		t.Errorf("expected conflict error message, got: %v", err)
	}
}

func TestParseFlagsPodNoNetworkConflict(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	_, _, err := parseBaseFlags("run", []string{"--pod", "mypod", "--no-network", "echo"}, t.TempDir())
	if err == nil {
		t.Error("expected error for --pod with --no-network")
	}
	if err != nil && !strings.Contains(err.Error(), "cannot use --pod with --no-network") {
		t.Errorf("expected conflict error message, got: %v", err)
	}
}

func TestSplitToolArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantYolobox []string
		wantTool    []string
	}{
		{
			name:        "tool flag only",
			args:        []string{"--resume"},
			wantYolobox: nil,
			wantTool:    []string{"--resume"},
		},
		{
			name:        "tool flag with value",
			args:        []string{"--resume", "abc123"},
			wantYolobox: nil,
			wantTool:    []string{"--resume", "abc123"},
		},
		{
			name:        "yolobox flag then tool flag",
			args:        []string{"--no-network", "--resume"},
			wantYolobox: []string{"--no-network"},
			wantTool:    []string{"--resume"},
		},
		{
			name:        "yolobox pod flag with value then tool flag",
			args:        []string{"--pod", "mypod", "--resume"},
			wantYolobox: []string{"--pod", "mypod"},
			wantTool:    []string{"--resume"},
		},
		{
			name:        "yolobox flag with value then tool flag",
			args:        []string{"--env", "FOO=bar", "--resume"},
			wantYolobox: []string{"--env", "FOO=bar"},
			wantTool:    []string{"--resume"},
		},
		{
			name:        "yolobox flag with equals then tool flag",
			args:        []string{"--env=FOO=bar", "--resume"},
			wantYolobox: []string{"--env=FOO=bar"},
			wantTool:    []string{"--resume"},
		},
		{
			name:        "multiple yolobox flags then tool args",
			args:        []string{"--no-network", "--scratch", "--resume", "abc123"},
			wantYolobox: []string{"--no-network", "--scratch"},
			wantTool:    []string{"--resume", "abc123"},
		},
		{
			name:        "customization flags stay with yolobox",
			args:        []string{"--packages", "default-jdk,maven", "--rebuild-image", "--resume"},
			wantYolobox: []string{"--packages", "default-jdk,maven", "--rebuild-image"},
			wantTool:    []string{"--resume"},
		},
		{
			name:        "explicit separator",
			args:        []string{"--no-network", "--", "--help"},
			wantYolobox: []string{"--no-network"},
			wantTool:    []string{"--help"},
		},
		{
			name:        "non-flag arg",
			args:        []string{"somefile.txt"},
			wantYolobox: nil,
			wantTool:    []string{"somefile.txt"},
		},
		{
			name:        "yolobox flag then non-flag arg",
			args:        []string{"--scratch", "somefile.txt"},
			wantYolobox: []string{"--scratch"},
			wantTool:    []string{"somefile.txt"},
		},
		{
			name:        "no args",
			args:        []string{},
			wantYolobox: nil,
			wantTool:    nil,
		},
		{
			name:        "only yolobox flags",
			args:        []string{"--scratch", "--no-network"},
			wantYolobox: []string{"--scratch", "--no-network"},
			wantTool:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotYolobox, gotTool := splitToolArgs(tt.args)

			if len(gotYolobox) != len(tt.wantYolobox) {
				t.Errorf("yolobox args: got %v, want %v", gotYolobox, tt.wantYolobox)
			} else {
				for i := range gotYolobox {
					if gotYolobox[i] != tt.wantYolobox[i] {
						t.Errorf("yolobox args[%d]: got %q, want %q", i, gotYolobox[i], tt.wantYolobox[i])
					}
				}
			}

			if len(gotTool) != len(tt.wantTool) {
				t.Errorf("tool args: got %v, want %v", gotTool, tt.wantTool)
			} else {
				for i := range gotTool {
					if gotTool[i] != tt.wantTool[i] {
						t.Errorf("tool args[%d]: got %q, want %q", i, gotTool[i], tt.wantTool[i])
					}
				}
			}
		})
	}
}

func TestDetectTimezone(t *testing.T) {
	// Test with TZ env var set
	t.Setenv("TZ", "Europe/London")
	tz := detectTimezone()
	if tz != "Europe/London" {
		t.Errorf("expected Europe/London from TZ env, got %q", tz)
	}

	// Test without TZ env var (falls back to /etc/localtime)
	t.Setenv("TZ", "")
	tz = detectTimezone()
	// On most systems /etc/localtime exists; just verify it doesn't crash
	// and returns either a valid timezone or empty string
	if tz != "" && !strings.Contains(tz, "/") {
		t.Errorf("expected IANA timezone with '/' or empty string, got %q", tz)
	}
}

func TestBuildRunArgsTimezone(t *testing.T) {
	t.Setenv("TZ", "America/New_York")
	cfg := Config{
		Image: "test-image",
	}

	args, _, err := buildRunArgs(cfg, "/test/project", []string{"bash"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	argsStr := strings.Join(args, " ")
	if !strings.Contains(argsStr, "TZ=America/New_York") {
		t.Error("expected TZ=America/New_York in args")
	}
}

func TestPreprocessClaudeConfig(t *testing.T) {
	// Use temp dir as HOME so the function writes to tmpDir/.yolobox/tmp/
	// instead of the real ~/.yolobox/tmp/ (which may not exist or be writable)
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	srcPath := filepath.Join(tmpDir, ".claude.json")

	// Config with installMethod that should be removed
	srcContent := `{
  "numStartups": 10,
  "installMethod": "native",
  "autoUpdates": false
}`
	if err := os.WriteFile(srcPath, []byte(srcContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Run preprocessing
	resultPath := preprocessClaudeConfig(srcPath)
	if resultPath == "" {
		t.Fatal("preprocessClaudeConfig returned empty path")
	}

	// Verify the file uses a unique name (not the fixed "claude-config.json")
	baseName := filepath.Base(resultPath)
	if baseName == "claude-config.json" {
		t.Error("expected unique temp file name, got fixed claude-config.json")
	}
	if !strings.HasPrefix(baseName, "claude-config-") || !strings.HasSuffix(baseName, ".json") {
		t.Errorf("expected temp file matching claude-config-*.json, got %s", baseName)
	}

	// Read the result
	result, err := os.ReadFile(resultPath)
	if err != nil {
		t.Fatalf("failed to read result file: %v", err)
	}

	resultStr := string(result)

	// Should NOT contain installMethod
	if strings.Contains(resultStr, "installMethod") {
		t.Errorf("result should not contain installMethod, got: %s", resultStr)
	}

	// Should still contain other fields
	if !strings.Contains(resultStr, "numStartups") {
		t.Errorf("result should contain numStartups, got: %s", resultStr)
	}
	if !strings.Contains(resultStr, "autoUpdates") {
		t.Errorf("result should contain autoUpdates, got: %s", resultStr)
	}
}

func TestConcurrentPreprocessClaudeConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	srcPath := filepath.Join(tmpDir, ".claude.json")

	srcContent := `{"numStartups": 10, "installMethod": "native"}`
	if err := os.WriteFile(srcPath, []byte(srcContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Call preprocessClaudeConfig concurrently and verify unique paths
	const n = 10
	paths := make([]string, n)
	done := make(chan int, n)
	for i := 0; i < n; i++ {
		go func(idx int) {
			paths[idx] = preprocessClaudeConfig(srcPath)
			done <- idx
		}(i)
	}
	for i := 0; i < n; i++ {
		<-done
	}

	// All paths should be non-empty and unique
	seen := make(map[string]bool)
	for i, p := range paths {
		if p == "" {
			t.Fatalf("preprocessClaudeConfig[%d] returned empty path", i)
		}
		if seen[p] {
			t.Errorf("duplicate temp path from concurrent calls: %s", p)
		}
		seen[p] = true
	}
}

func TestDirContainsSymlinks(t *testing.T) {
	// Directory with no symlinks
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "subdir", "nested.txt"), []byte("world"), 0644); err != nil {
		t.Fatal(err)
	}

	if dirContainsSymlinks(dir) {
		t.Error("expected no symlinks in plain directory")
	}

	// Directory with a symlink
	target := filepath.Join(dir, "file.txt")
	link := filepath.Join(dir, "link.txt")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if !dirContainsSymlinks(dir) {
		t.Error("expected symlinks detected")
	}
}

func TestDirContainsSymlinksNested(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(sub, "real.txt")
	if err := os.WriteFile(target, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(sub, "link.txt")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	if !dirContainsSymlinks(dir) {
		t.Error("expected nested symlink to be detected")
	}
}

func TestCopyDirDereferenced(t *testing.T) {
	// Create source directory with a symlink
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "regular.txt"), []byte("regular"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create an external file to symlink to
	external := t.TempDir()
	if err := os.WriteFile(filepath.Join(external, "external.txt"), []byte("external-content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Symlink from inside src to external file
	if err := os.Symlink(filepath.Join(external, "external.txt"), filepath.Join(src, "linked.txt")); err != nil {
		t.Fatal(err)
	}

	// Symlink to an external directory
	extDir := filepath.Join(external, "subdir")
	if err := os.MkdirAll(extDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(extDir, "deep.txt"), []byte("deep-content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(extDir, filepath.Join(src, "linked-dir")); err != nil {
		t.Fatal(err)
	}

	// Copy with dereference
	dst := filepath.Join(t.TempDir(), "copy")
	if err := copyDirDereferenced(src, dst); err != nil {
		t.Fatalf("copyDirDereferenced failed: %v", err)
	}

	// Verify regular file was copied
	data, err := os.ReadFile(filepath.Join(dst, "regular.txt"))
	if err != nil || string(data) != "regular" {
		t.Errorf("regular.txt: got %q, err %v", data, err)
	}

	// Verify symlinked file was dereferenced (copied as regular file)
	data, err = os.ReadFile(filepath.Join(dst, "linked.txt"))
	if err != nil || string(data) != "external-content" {
		t.Errorf("linked.txt: got %q, err %v", data, err)
	}
	info, _ := os.Lstat(filepath.Join(dst, "linked.txt"))
	if info.Mode()&os.ModeSymlink != 0 {
		t.Error("linked.txt should be a regular file, not a symlink")
	}

	// Verify symlinked directory was dereferenced
	data, err = os.ReadFile(filepath.Join(dst, "linked-dir", "deep.txt"))
	if err != nil || string(data) != "deep-content" {
		t.Errorf("linked-dir/deep.txt: got %q, err %v", data, err)
	}
	info, _ = os.Lstat(filepath.Join(dst, "linked-dir"))
	if info.Mode()&os.ModeSymlink != 0 {
		t.Error("linked-dir should be a regular directory, not a symlink")
	}
}

func TestCopyDirDereferencedSkipsBrokenSymlinks(t *testing.T) {
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "good.txt"), []byte("good"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("/nonexistent/path/file.txt", filepath.Join(src, "broken.txt")); err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(t.TempDir(), "copy")
	if err := copyDirDereferenced(src, dst); err != nil {
		t.Fatalf("copyDirDereferenced should not fail on broken symlinks: %v", err)
	}

	// good.txt should exist
	if _, err := os.Stat(filepath.Join(dst, "good.txt")); err != nil {
		t.Error("good.txt should have been copied")
	}

	// broken.txt should be skipped
	if _, err := os.Stat(filepath.Join(dst, "broken.txt")); err == nil {
		t.Error("broken.txt should have been skipped")
	}
}

func TestStageDirResolvingSymlinks(t *testing.T) {
	src := t.TempDir()
	external := t.TempDir()

	if err := os.WriteFile(filepath.Join(src, "config.json"), []byte(`{"key":"value"}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(external, "shared.json"), []byte(`{"shared":true}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(external, "shared.json"), filepath.Join(src, "shared.json")); err != nil {
		t.Fatal(err)
	}

	staged, err := stageDirResolvingSymlinks(src)
	if err != nil {
		t.Fatalf("stageDirResolvingSymlinks failed: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(staged)
	}()

	// Verify the staged directory contains dereferenced files
	data, err := os.ReadFile(filepath.Join(staged, "shared.json"))
	if err != nil || string(data) != `{"shared":true}` {
		t.Errorf("staged shared.json: got %q, err %v", data, err)
	}

	// Verify it's a regular file, not a symlink
	info, _ := os.Lstat(filepath.Join(staged, "shared.json"))
	if info.Mode()&os.ModeSymlink != 0 {
		t.Error("staged shared.json should be a regular file")
	}
}

func TestFindSSHAgentSocketLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-only test")
	}

	// With SSH_AUTH_SOCK set, should return it directly
	t.Setenv("SSH_AUTH_SOCK", "/tmp/ssh-test/agent.123")
	sock, err := findSSHAgentSocket()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sock != "/tmp/ssh-test/agent.123" {
		t.Errorf("expected /tmp/ssh-test/agent.123, got %s", sock)
	}

	// Without SSH_AUTH_SOCK, should error
	t.Setenv("SSH_AUTH_SOCK", "")
	_, err = findSSHAgentSocket()
	if err == nil {
		t.Error("expected error when SSH_AUTH_SOCK is empty")
	}
}

func TestFindSSHAgentSocketMacOSNoSSHAuthSock(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS-only test")
	}

	// On macOS, SSH_AUTH_SOCK is not used directly — the function should
	// detect the Docker runtime and return the VM-internal path or error.
	// Without Docker running, it may error — that's the expected behavior.
	t.Setenv("SSH_AUTH_SOCK", "")
	sock, err := findSSHAgentSocket()
	if err != nil {
		// Expected when Docker/Colima isn't configured
		return
	}
	// If it succeeds, the socket path should be non-empty
	if sock == "" {
		t.Error("expected non-empty socket path")
	}
}
