package java

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Switcher struct {
	previousJavaHome string
}

func NewSwitcher() *Switcher {
	return &Switcher{
		previousJavaHome: os.Getenv("JAVA_HOME"),
	}
}

func (s *Switcher) SetJavaHome(javaHome string) error {
	// Store previous value for backup
	s.previousJavaHome = os.Getenv("JAVA_HOME")

	// Set environment variable for current session
	err := os.Setenv("JAVA_HOME", javaHome)
	if err != nil {
		return err
	}

	// Update PATH to include new Java bin directory
	err = s.updatePath(javaHome)
	if err != nil {
		return err
	}

	// Set persistently based on platform
	switch runtime.GOOS {
	case "windows":
		return s.setWindowsPersistent(javaHome)
	case "linux", "darwin":
		return s.setUnixPersistent(javaHome)
	default:
		return fmt.Errorf("プラットフォーム %s はサポートされていません", runtime.GOOS)
	}
}

func (s *Switcher) GetCurrentJavaHome() string {
	return os.Getenv("JAVA_HOME")
}

func (s *Switcher) GetPreviousJavaHome() string {
	return s.previousJavaHome
}

func (s *Switcher) RestorePrevious() error {
	if s.previousJavaHome == "" {
		return fmt.Errorf("復元可能な前のJAVA_HOMEがありません")
	}
	return s.SetJavaHome(s.previousJavaHome)
}

// Windows永続化: レジストリ経由
func (s *Switcher) setWindowsPersistent(javaHome string) error {
	// システム環境変数を設定（管理者権限が必要）
	cmd := exec.Command("reg", "add", "HKLM\\SYSTEM\\CurrentControlSet\\Control\\Session Manager\\Environment",
		"/v", "JAVA_HOME", "/t", "REG_EXPAND_SZ", "/d", javaHome, "/f")

	if err := cmd.Run(); err != nil {
		// 管理者権限がない場合はユーザー環境変数に設定
		cmd = exec.Command("reg", "add", "HKCU\\Environment",
			"/v", "JAVA_HOME", "/t", "REG_EXPAND_SZ", "/d", javaHome, "/f")

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("レジストリ設定に失敗しました: %v", err)
		}
	}

	// PATH環境変数も更新
	javaBinPath := filepath.Join(javaHome, "bin")

	// PowerShellでPATH更新
	pathUpdateScript := fmt.Sprintf(`
		$currentPath = [Environment]::GetEnvironmentVariable('PATH', 'User')
		$pathParts = $currentPath -split ';' | Where-Object { $_ -notlike '*java*' }
		$newPath = '%s;' + ($pathParts -join ';')
		[Environment]::SetEnvironmentVariable('PATH', $newPath, 'User')
		[Environment]::SetEnvironmentVariable('JAVA_HOME', '%s', 'User')
	`, javaBinPath, javaHome)

	cmd = exec.Command("powershell", "-Command", pathUpdateScript)
	return cmd.Run()
}

// Unix/Linux/macOS永続化: シェル設定ファイル更新
func (s *Switcher) setUnixPersistent(javaHome string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("ホームディレクトリの取得に失敗: %v", err)
	}

	// 対象ファイルを特定
	configFiles := []string{".bashrc", ".zshrc", ".profile"}
	var targetFile string

	for _, file := range configFiles {
		path := fmt.Sprintf("%s/%s", homeDir, file)
		if _, err := os.Stat(path); err == nil {
			targetFile = path
			break
		}
	}

	if targetFile == "" {
		// デフォルトで.bashrcを作成
		targetFile = fmt.Sprintf("%s/.bashrc", homeDir)
	}

	// 既存のJAVA_HOME関連設定を削除（macOS/Linux互換）
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		// macOSはsed -i ''が必要
		cmd = exec.Command("sed", "-i", "", "/export JAVA_HOME=/d", targetFile)
		cmd.Run() // エラーは無視（ファイルが存在しない場合など）
		cmd = exec.Command("sed", "-i", "", "/# JavaSwitcher PATH/d", targetFile)
		cmd.Run()
	} else {
		// Linux/その他
		cmd = exec.Command("sed", "-i", "/export JAVA_HOME=/d", targetFile)
		cmd.Run()
		cmd = exec.Command("sed", "-i", "/# JavaSwitcher PATH/d", targetFile)
		cmd.Run()
	}

	// 新しいJAVA_HOME設定を追加
	exportLines := fmt.Sprintf(`export JAVA_HOME=%s
export PATH=$JAVA_HOME/bin:$PATH  # JavaSwitcher PATH`, javaHome)

	cmd = exec.Command("sh", "-c", fmt.Sprintf("echo '%s' >> %s", exportLines, targetFile))

	return cmd.Run()
}

// PATH環境変数の更新
func (s *Switcher) updatePath(javaHome string) error {
	javaBinPath := filepath.Join(javaHome, "bin")

	// 現在のPATHを取得
	currentPath := os.Getenv("PATH")

	// Javaのbinパスが既に含まれているかチェック
	if strings.Contains(currentPath, javaBinPath) {
		return nil // 既に含まれている
	}

	// プラットフォーム別のPATH区切り文字
	pathSeparator := ":"
	if runtime.GOOS == "windows" {
		pathSeparator = ";"
	}

	// 他のJavaパスを削除
	pathParts := strings.Split(currentPath, pathSeparator)
	cleanedParts := []string{}

	for _, part := range pathParts {
		// Javaのbinディレクトリでない場合のみ追加
		if !s.isJavaBinPath(part) {
			cleanedParts = append(cleanedParts, part)
		}
	}

	// 新しいJavaのbinパスを先頭に追加
	newPath := javaBinPath + pathSeparator + strings.Join(cleanedParts, pathSeparator)

	// 現在のセッションで設定
	return os.Setenv("PATH", newPath)
}

// Javaのbinパスかどうかを判定
func (s *Switcher) isJavaBinPath(path string) bool {
	path = strings.ToLower(path)
	return strings.Contains(path, "java") &&
		   (strings.Contains(path, "bin") || strings.Contains(path, "javapath"))
}

// PowerShellスクリプト生成（Windows向け追加オプション）
func (s *Switcher) GenerateWindowsScript(javaHome string) (string, error) {
	javaBinPath := filepath.Join(javaHome, "bin")
	script := fmt.Sprintf(`# JavaSwitcher Auto-generated Script
$env:JAVA_HOME = "%s"
$env:PATH = "%s;" + ($env:PATH -split ";" | Where-Object { $_ -notlike "*java*" } | Join-String -Separator ";")

[Environment]::SetEnvironmentVariable("JAVA_HOME", "%s", "User")
Write-Host "JAVA_HOME set to: %s"
Write-Host "PATH updated to prioritize: %s"
Write-Host "Please restart your terminal to apply changes globally."
`, javaHome, javaBinPath, javaHome, javaHome, javaBinPath)

	return script, nil
}