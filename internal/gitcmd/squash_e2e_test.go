package gitcmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/KevinYouu/easyGit/internal/testutil"
)

// TestSquashEndToEndPreservesMultiline 端到端回归(编译真实可执行文件后走
// 生产链路 RunInternalRebase → git rebase -i → 隐藏编辑器子命令):
// 合并后的提交信息必须完整保留多行正文与空行。
// 覆盖用户可见缺陷「sq 命令没有保留全部文本,换行的全部丢失」。
func TestSquashEndToEndPreservesMultiline(t *testing.T) {
	runSquashEndToEnd(t, []string{
		"feat: base\n\nbase body",
		"fix: first target\n\nfirst body line 1\nfirst body line 2",
		"chore: second target\n\nsecond body",
	}, "fix: first target\n\nfirst body line 1\nfirst body line 2\n\nchore: second target\n\nsecond body")
}

// TestSquashEndToEndPreservesHashLines 以 # 开头的行必须原样保留:
// git 默认 commit.cleanup=strip 会删掉这类行(Issue 号、模板说明),
// 主题被删空时还会让提交失败;生产链路必须注入 commit.cleanup=verbatim
func TestSquashEndToEndPreservesHashLines(t *testing.T) {
	runSquashEndToEnd(t, []string{
		"feat: base\n\n#0 base note",
		"#123 first target\n\n#456 body note\n    indented line",
		"chore: second target\n\n#789 trailing note",
	}, "#123 first target\n\n#456 body note\n    indented line\n\nchore: second target\n\n#789 trailing note")
}

// runSquashEndToEnd 构造 3 条提交,合并最近两条,断言:
// 默认消息(= buildSquashDefaultMessage)与最终提交信息都等于 want。
func runSquashEndToEnd(t *testing.T, messages []string, want string) {
	t.Helper()
	if testing.Short() {
		t.Skip("端到端用例需编译 easyGit 二进制,短测试跳过")
	}
	if len(messages) != 3 {
		t.Fatalf("需要 3 条提交, got %d", len(messages))
	}

	tempDir, cleanup := testutil.CreateTempGitRepo(t)
	t.Cleanup(cleanup)

	originalDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(originalDir) })
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	hashes := make([]string, 0, len(messages))
	for i, message := range messages {
		path := filepath.Join(tempDir, "file.txt")
		if err := os.WriteFile(path, []byte(strings.Repeat("x", i+1)), 0o644); err != nil {
			t.Fatalf("write file: %v", err)
		}
		testutil.RunGitCommand(t, tempDir, "add", "file.txt")
		testutil.RunGitCommand(t, tempDir, "commit", "-m", message)
		hashes = append(hashes, revParseHead(t, tempDir))
	}

	// 合并最近两条:旧→新的目标列表与基点
	targets := []string{hashes[1], hashes[2]}
	base := hashes[0]

	defaultMessage, err := buildSquashDefaultMessage(targets)
	if err != nil {
		t.Fatalf("buildSquashDefaultMessage() error = %v", err)
	}
	if defaultMessage != want {
		t.Errorf("默认合并消息 =\n%q\nwant\n%q", defaultMessage, want)
	}

	// 生产链路:可执行文件换成真实二进制,由 git 回调其隐藏编辑器子命令
	binary := buildEasyGitBinary(t)
	oldExe := executablePath
	executablePath = func() (string, error) { return binary, nil }
	t.Cleanup(func() { executablePath = oldExe })

	if err := RunInternalRebase(base, "squash", targets, defaultMessage); err != nil {
		t.Fatalf("RunInternalRebase() error = %v", err)
	}

	// 提交数量从 3 降为 2
	countOut, err := gitOutput(tempDir, "rev-list", "--count", "HEAD")
	if err != nil {
		t.Fatalf("rev-list: %v", err)
	}
	if got := strings.TrimSpace(countOut); got != "2" {
		t.Errorf("提交数 = %s, want 2", got)
	}

	// 提交信息一字不差
	messageOut, err := gitOutput(tempDir, "log", "-1", "--format=%B")
	if err != nil {
		t.Fatalf("git log: %v", err)
	}
	if got := strings.TrimRight(messageOut, "\n"); got != want {
		t.Errorf("合并后提交信息 =\n%q\nwant\n%q", got, want)
	}
}

// gitOutput 在指定目录执行 git 并返回标准输出
func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.Output()
	return string(output), err
}

// buildEasyGitBinary 编译当前模块的 easyGit 可执行文件到临时目录
// (端到端用例需要真实二进制作为 GIT_EDITOR/GIT_SEQUENCE_EDITOR)
func buildEasyGitBinary(t *testing.T) string {
	t.Helper()

	name := "easyGit"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary := filepath.Join(t.TempDir(), name)

	// 用例会 chdir 到临时仓库,go build 必须显式指定模块根目录
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate test file failed")
	}
	moduleRoot := filepath.Dir(filepath.Dir(filepath.Dir(file)))

	cmd := exec.Command("go", "build", "-o", binary, "./cmd/easygit")
	cmd.Dir = moduleRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, output)
	}
	return binary
}
