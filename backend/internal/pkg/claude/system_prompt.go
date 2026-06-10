package claude

import (
	"crypto/sha256"
	_ "embed"
	"fmt"
	"strings"
	"text/template"
	"time"
)

//go:embed system_prompt_static.txt
var SystemPromptStatic string

//go:embed system_prompt_dynamic.txt
var systemPromptDynamicTemplate string

// SystemPromptOpts 用于生成动态 system prompt 的参数
type SystemPromptOpts struct {
	Platform  string // "darwin", "linux", "win32"
	Shell     string // "bash", "zsh", "powershell"
	OSVersion string // "Darwin 24.6.0", "Linux 6.1.0", "Windows 10 Pro 10.0.19045"
	WorkDir   string // "/home/user/project"
	ModelName string // "Opus 4.6 (1M context)"
	ModelID   string // "claude-opus-4-6[1m]"
	IsGitRepo string // "true" or "false"
}

// BuildDynamicPrompt 根据选项生成动态部分的 system prompt
func BuildDynamicPrompt(opts SystemPromptOpts) string {
	tmpl, err := template.New("dynamic").Parse(systemPromptDynamicTemplate)
	if err != nil {
		return ""
	}

	data := struct {
		WorkingDir string
		IsGitRepo  string
		Platform   string
		Shell      string
		OSVersion  string
		ModelName  string
		ModelID    string
	}{
		WorkingDir: opts.WorkDir,
		IsGitRepo:  opts.IsGitRepo,
		Platform:   opts.Platform,
		Shell:      opts.Shell,
		OSVersion:  opts.OSVersion,
		ModelName:  opts.ModelName,
		ModelID:    opts.ModelID,
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, data); err != nil {
		return ""
	}
	return sb.String()
}

// DefaultOptsForPlatform 根据平台指纹生成默认的动态选项
func DefaultOptsForPlatform(stainlessOS, stainlessArch string, accountID int64, model string) SystemPromptOpts {
	platform, shell, osVersion := mapPlatformInfo(stainlessOS)
	workDir := generateWorkingDir(stainlessOS, accountID)
	modelName, modelID := mapModelInfo(model)

	return SystemPromptOpts{
		Platform:  platform,
		Shell:     shell,
		OSVersion: osVersion,
		WorkDir:   workDir,
		ModelName: modelName,
		ModelID:   modelID,
		IsGitRepo: "true",
	}
}

// mapPlatformInfo 根据 X-Stainless-OS 映射平台信息
func mapPlatformInfo(stainlessOS string) (platform, shell, osVersion string) {
	switch strings.ToLower(stainlessOS) {
	case "macos":
		return "darwin", "zsh", "Darwin 24.6.0"
	case "windows_nt":
		return "win32", "powershell", "Windows 10 Pro 10.0.19045"
	default: // Linux and others
		return "linux", "bash", "Linux 6.1.0"
	}
}

// generateWorkingDir 基于账号 ID 生成确定性的工作目录路径
func generateWorkingDir(stainlessOS string, accountID int64) string {
	// 使用 accountID 生成确定性但多样的项目名
	hash := sha256.Sum256([]byte(fmt.Sprintf("workdir:%d", accountID)))
	projectNames := []string{
		"app", "project", "service", "api", "backend",
		"web", "platform", "core", "client", "server",
	}
	userNames := []string{
		"developer", "user", "dev", "admin", "engineer",
	}
	projectIdx := int(hash[0]) % len(projectNames)
	userIdx := int(hash[1]) % len(userNames)

	switch strings.ToLower(stainlessOS) {
	case "macos":
		return fmt.Sprintf("/Users/%s/%s", userNames[userIdx], projectNames[projectIdx])
	case "windows_nt":
		return fmt.Sprintf("C:\\Users\\%s\\%s", userNames[userIdx], projectNames[projectIdx])
	default:
		return fmt.Sprintf("/home/%s/%s", userNames[userIdx], projectNames[projectIdx])
	}
}

// mapModelInfo 根据模型名映射显示名称和 ID
func mapModelInfo(model string) (displayName, modelID string) {
	m := strings.ToLower(model)
	switch {
	case strings.Contains(m, "fable"):
		return "Fable 5 (1M context)", "claude-fable-5[1m]"
	case strings.Contains(m, "mythos"):
		return "Mythos 5 (1M context)", "claude-mythos-5[1m]"
	case strings.Contains(m, "opus-4-8") || strings.Contains(m, "opus-4.8"):
		return "Opus 4.8 (1M context)", "claude-opus-4-8[1m]"
	case strings.Contains(m, "opus-4-7") || strings.Contains(m, "opus-4.7"):
		return "Opus 4.7 (1M context)", "claude-opus-4-7[1m]"
	case strings.Contains(m, "opus"):
		return "Opus 4.8 (1M context)", "claude-opus-4-8[1m]"
	case strings.Contains(m, "haiku"):
		return "Haiku 4.5", "claude-haiku-4-5-20251001"
	default: // sonnet and others
		return "Sonnet 4.6 (1M context)", "claude-sonnet-4-6[1m]"
	}
}

// CurrentDate 返回当前日期字符串
func CurrentDate() string {
	return time.Now().Format("2006-01-02")
}
