package appconfig

import (
	"os"
	"path/filepath"
	"testing"
)

// writeFile 在 dir 下创建名为 name 的文件（自动建目录）。
func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", dir, err)
	}
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile(%s): %v", p, err)
	}
	return p
}

// writeGitDir 创建 .git，gitPath 为 .git 的完整路径。
// isDir=true 创建目录，false 创建 worktree gitdir 指针文件（内容为 gitdirContent）。
func writeGitDir(t *testing.T, gitPath string, isDir bool, gitdirContent string) {
	t.Helper()
	if isDir {
		if err := os.MkdirAll(gitPath, 0755); err != nil {
			t.Fatalf("MkdirAll(%s): %v", gitPath, err)
		}
		return
	}
	// 创建文件：只确保父目录存在，避免把 .git 自身建成目录
	if err := os.MkdirAll(filepath.Dir(gitPath), 0755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", filepath.Dir(gitPath), err)
	}
	if err := os.WriteFile(gitPath, []byte(gitdirContent), 0644); err != nil {
		t.Fatalf("WriteFile(%s): %v", gitPath, err)
	}
}

func TestFindConfigFile_UpwardSearchFound(t *testing.T) {
	// 结构: root/.git(目录) root/.rain-qa-func.json root/sub/deep
	// 从 deep 向上应找到 root 下的配置
	root := t.TempDir()
	writeGitDir(t, filepath.Join(root, ".git"), true, "")
	want := writeFile(t, root, ConfigFileName, "{}")
	deep := filepath.Join(root, "sub", "deep")
	if err := os.MkdirAll(deep, 0755); err != nil {
		t.Fatal(err)
	}

	got, r := findConfigFile(deep)
	if !r.found {
		t.Fatalf("期望 found=true，实际 false（搜索: %s）", r.searchedDirs)
	}
	if got != want {
		t.Errorf("路径: 期望 %s，实际 %s", want, got)
	}
	if r.viaWorktree {
		t.Errorf("不应触发 worktree 跳转")
	}
}

func TestFindConfigFile_WorktreeJump(t *testing.T) {
	// 结构: 两棵独立的目录树
	//   主仓库 mainRepo/.git(目录) mainRepo/.rain-qa-func.json
	//   worktree wt/.git(文件: gitdir: mainRepo/.git/worktrees/wt) wt/sub
	// 从 wt/sub 向上无法到达 mainRepo，须靠 .git 文件跳转
	mainRepo := t.TempDir()
	writeGitDir(t, filepath.Join(mainRepo, ".git"), true, "")
	want := writeFile(t, mainRepo, ConfigFileName, "{}")

	wt := t.TempDir()
	gitdirLine := "gitdir: " + filepath.ToSlash(filepath.Join(mainRepo, ".git", "worktrees", "wt"))
	writeGitDir(t, filepath.Join(wt, ".git"), false, gitdirLine)
	deep := filepath.Join(wt, "sub")
	if err := os.MkdirAll(deep, 0755); err != nil {
		t.Fatal(err)
	}

	got, r := findConfigFile(deep)
	if !r.found {
		t.Fatalf("期望 found=true，实际 false（搜索: %s）", r.searchedDirs)
	}
	if got != want {
		t.Errorf("路径: 期望 %s，实际 %s", want, got)
	}
	if !r.viaWorktree {
		t.Errorf("期望 viaWorktree=true，实际 false")
	}
	// worktreeMainRepo 字段记录的主仓库根应能被 configInDir 命中
	if r.worktreeMainRepo == "" || configInDir(r.worktreeMainRepo) == "" {
		t.Errorf("worktreeMainRepo 无效: %q", r.worktreeMainRepo)
	}
}

func TestFindConfigFile_NotFoundFallback(t *testing.T) {
	// 无配置文件，逐级向上到根都找不到，应回退 startDir 下的默认路径
	root := t.TempDir()
	deep := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(deep, 0755); err != nil {
		t.Fatal(err)
	}

	got, r := findConfigFile(deep)
	if r.found {
		t.Errorf("期望 found=false，实际 true")
	}
	want := filepath.Join(deep, ConfigFileName)
	if got != want {
		t.Errorf("回退路径: 期望 %s，实际 %s", want, got)
	}
	if r.searchedDirs == "" {
		t.Errorf("未找到时应记录搜索轨迹")
	}
}

func TestConfigInDir(t *testing.T) {
	dir := t.TempDir()

	// 不存在
	if got := configInDir(dir); got != "" {
		t.Errorf("空目录应返回空串，实际 %q", got)
	}
	// 存在
	p := writeFile(t, dir, ConfigFileName, "{}")
	if got := configInDir(dir); got != p {
		t.Errorf("期望 %s，实际 %s", p, got)
	}
	// 是目录而非文件（同名目录）— 应返回空
	os.Remove(p)
	if err := os.Mkdir(p, 0755); err != nil {
		t.Fatal(err)
	}
	if got := configInDir(dir); got != "" {
		t.Errorf("同名目录应返回空串，实际 %q", got)
	}
}

func TestWorktreeMainRepo(t *testing.T) {
	mainRepo := t.TempDir()
	wt := t.TempDir()

	// 正常 worktree gitdir 指针
	gitdirLine := "gitdir: " + filepath.ToSlash(filepath.Join(mainRepo, ".git", "worktrees", "wt"))
	writeGitDir(t, filepath.Join(wt, ".git"), false, gitdirLine)
	if got := worktreeMainRepo(wt); got != filepath.Clean(mainRepo) {
		t.Errorf("期望 %s，实际 %q", filepath.Clean(mainRepo), got)
	}

	// .git 是目录（普通仓库）— 应返回空
	plain := t.TempDir()
	writeGitDir(t, filepath.Join(plain, ".git"), true, "")
	if got := worktreeMainRepo(plain); got != "" {
		t.Errorf("普通仓库应返回空，实际 %q", got)
	}

	// .git 不存在 — 应返回空
	empty := t.TempDir()
	if got := worktreeMainRepo(empty); got != "" {
		t.Errorf("无 .git 应返回空，实际 %q", got)
	}

	// .git 文件内容格式异常（无 / .git / 段）— 应返回空
	bad := t.TempDir()
	writeGitDir(t, filepath.Join(bad, ".git"), false, "gitdir: /some/strange/path")
	if got := worktreeMainRepo(bad); got != "" {
		t.Errorf("异常 gitdir 应返回空，实际 %q", got)
	}
}

func TestSectionLoadSaveRoundTrip(t *testing.T) {
	// 用临时目录覆盖 globalFilePath，测试 Section 的 Load/Save 往返
	root := t.TempDir()
	defer func(p string) {
		globalMu.Lock()
		globalFilePath = p
		globalMu.Unlock()
	}(FilePath())
	globalMu.Lock()
	globalFilePath = filepath.Join(root, ConfigFileName)
	globalMu.Unlock()

	sec := New("test_section")

	// 初始不存在
	if sec.Exists() {
		t.Fatal("新 section 不应存在")
	}

	// Save
	type cfg struct {
		Name string `json:"name"`
		Port int    `json:"port"`
	}
	if err := sec.Save(&cfg{Name: "alpha", Port: 8080}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if !sec.Exists() {
		t.Fatal("Save 后 section 应存在")
	}

	// Load
	var got cfg
	if err := sec.Load(&got); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Name != "alpha" || got.Port != 8080 {
		t.Errorf("Load 数据不符: %+v", got)
	}

	// 另一 section 不受影响、Exists 返回 false
	other := New("other_section")
	if other.Exists() {
		t.Fatal("其他 section 不应存在")
	}
}
