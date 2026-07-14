package roadmap

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// TestRoadmapService 测试路线图服务
func TestRoadmapService(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建测试服务
	svc := &RoadmapService{
		configFile: filepath.Join(tempDir, ConfigFileName),
	}
	svc.config = svc.getDefaultConfig()

	// 先保存默认配置到文件
	if err := svc.saveConfig(svc.config); err != nil {
		t.Fatalf("saveConfig failed: %v", err)
	}

	// 测试配置文件存在
	t.Run("ConfigFileExists", func(t *testing.T) {
		// 检查文件是否存在
		if _, err := os.Stat(svc.configFile); os.IsNotExist(err) {
			t.Fatal("config file was not created")
		}
	})

	// 测试加载配置
	t.Run("LoadConfig", func(t *testing.T) {
		config, err := svc.loadConfig()
		if err != nil {
			t.Fatalf("loadConfig failed: %v", err)
		}

		// 验证默认数据
		if config.Version != "1.0" {
			t.Errorf("expected version 1.0, got %s", config.Version)
		}

		if len(config.Items) == 0 {
			t.Error("expected default items, got none")
		}
	})

	// 测试提交新建议
	t.Run("SubmitSuggestion", func(t *testing.T) {
		req := SubmitSuggestionRequest{
			Title:       "测试功能",
			Description: "这是一个测试功能",
			Priority:    PriorityHigh,
		}

		item, err := svc.SubmitSuggestion(req)
		if err != nil {
			t.Fatalf("SubmitSuggestion failed: %v", err)
		}

		if item.Title != req.Title {
			t.Errorf("expected title %s, got %s", req.Title, item.Title)
		}

		if item.Status != StatusPlanning {
			t.Errorf("expected status planning, got %s", item.Status)
		}

		if item.Priority != req.Priority {
			t.Errorf("expected priority %s, got %s", req.Priority, item.Priority)
		}

		// 验证ID不为空且唯一
		if item.ID == "" {
			t.Error("expected non-empty ID")
		}
	})

	// 测试提交空标题（应该失败）
	t.Run("SubmitSuggestionEmptyTitle", func(t *testing.T) {
		req := SubmitSuggestionRequest{
			Title:       "",
			Description: "描述",
			Priority:    PriorityMedium,
		}

		_, err := svc.SubmitSuggestion(req)
		if err == nil {
			t.Fatal("expected error for empty title, got nil")
		}
	})

	// 测试提交超长标题（应该失败）
	t.Run("SubmitSuggestionLongTitle", func(t *testing.T) {
		req := SubmitSuggestionRequest{
			Title:       strings.Repeat("a", MaxTitleLength+1),
			Description: "描述",
			Priority:    PriorityMedium,
		}

		_, err := svc.SubmitSuggestion(req)
		if err == nil {
			t.Fatal("expected error for long title, got nil")
		}
	})

	// 测试投票
	t.Run("Vote", func(t *testing.T) {
		// 先获取一个项目
		config, _ := svc.loadConfig()
		if len(config.Items) == 0 {
			t.Fatal("no items to vote on")
		}

		testItem := config.Items[0]
		initialUp := testItem.Votes.Up

		// 投票支持
		voteReq := VoteRequest{
			ItemID: testItem.ID,
			Vote:   stringPtr("up"),
		}

		updated, err := svc.Vote(voteReq)
		if err != nil {
			t.Fatalf("Vote failed: %v", err)
		}

		if updated.Votes.Up != initialUp+1 {
			t.Errorf("expected up votes %d, got %d", initialUp+1, updated.Votes.Up)
		}

		// 取消投票
		voteReq.Vote = nil
		updated, err = svc.Vote(voteReq)
		if err != nil {
			t.Fatalf("Cancel vote failed: %v", err)
		}

		if updated.Votes.Up != initialUp {
			t.Errorf("expected up votes %d after cancel, got %d", initialUp, updated.Votes.Up)
		}
	})

	// 测试投票空ID（应该失败）
	t.Run("VoteEmptyID", func(t *testing.T) {
		voteReq := VoteRequest{
			ItemID: "",
			Vote:   stringPtr("up"),
		}

		_, err := svc.Vote(voteReq)
		if err == nil {
			t.Fatal("expected error for empty item ID, got nil")
		}
	})

	// 测试添加评论
	t.Run("AddComment", func(t *testing.T) {
		config, _ := svc.loadConfig()
		if len(config.Items) == 0 {
			t.Fatal("no items to comment on")
		}

		testItem := config.Items[0]
		initialComments := len(testItem.Comments)

		commentReq := CommentRequest{
			ItemID:  testItem.ID,
			Content: "这是测试评论",
		}

		updated, err := svc.AddComment(commentReq)
		if err != nil {
			t.Fatalf("AddComment failed: %v", err)
		}

		if len(updated.Comments) != initialComments+1 {
			t.Errorf("expected %d comments, got %d", initialComments+1, len(updated.Comments))
		}

		lastComment := updated.Comments[len(updated.Comments)-1]
		if lastComment.Content != commentReq.Content {
			t.Errorf("expected comment content %s, got %s", commentReq.Content, lastComment.Content)
		}
	})

	// 测试添加空评论（应该失败）
	t.Run("AddCommentEmpty", func(t *testing.T) {
		config, _ := svc.loadConfig()
		if len(config.Items) == 0 {
			t.Fatal("no items to comment on")
		}

		commentReq := CommentRequest{
			ItemID:  config.Items[0].ID,
			Content: "",
		}

		_, err := svc.AddComment(commentReq)
		if err == nil {
			t.Fatal("expected error for empty comment, got nil")
		}
	})

	// 测试添加超长评论（应该失败）
	t.Run("AddCommentTooLong", func(t *testing.T) {
		config, _ := svc.loadConfig()
		if len(config.Items) == 0 {
			t.Fatal("no items to comment on")
		}

		commentReq := CommentRequest{
			ItemID:  config.Items[0].ID,
			Content: strings.Repeat("a", MaxCommentLength+1),
		}

		_, err := svc.AddComment(commentReq)
		if err == nil {
			t.Fatal("expected error for long comment, got nil")
		}
	})

	// 测试更新状态
	t.Run("UpdateStatus", func(t *testing.T) {
		config, _ := svc.loadConfig()
		if len(config.Items) == 0 {
			t.Fatal("no items to update")
		}

		testItem := config.Items[0]

		updated, err := svc.UpdateStatus(testItem.ID, StatusInProgress)
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}

		if updated.Status != StatusInProgress {
			t.Errorf("expected status in_progress, got %s", updated.Status)
		}
	})

	// 测试更新状态空ID（应该失败）
	t.Run("UpdateStatusEmptyID", func(t *testing.T) {
		_, err := svc.UpdateStatus("", StatusInProgress)
		if err == nil {
			t.Fatal("expected error for empty ID, got nil")
		}
	})

	// 测试导出 JSON
	t.Run("ExportToJSON", func(t *testing.T) {
		jsonStr, err := svc.ExportToJSON()
		if err != nil {
			t.Fatalf("ExportToJSON failed: %v", err)
		}

		if jsonStr == "" {
			t.Error("expected non-empty JSON string")
		}

		// 验证 JSON 包含版本字段（Go JSON 序列化后字段名小写）
		if !strings.Contains(jsonStr, "version") {
			t.Errorf("JSON output missing version field, got: %s", jsonStr)
		}

		// 验证 JSON 包含项目列表
		if !strings.Contains(jsonStr, "items") {
			t.Error("JSON output missing items field")
		}
	})

	// 测试 GetItem 返回正确的指针
	t.Run("GetItemPointer", func(t *testing.T) {
		config, _ := svc.loadConfig()
		if len(config.Items) == 0 {
			t.Fatal("no items to get")
		}

		testItem := config.Items[0]
		item, err := svc.GetItem(testItem.ID)
		if err != nil {
			t.Fatalf("GetItem failed: %v", err)
		}

		// 验证返回的指针指向正确的数据
		if item.ID != testItem.ID {
			t.Errorf("expected ID %s, got %s", testItem.ID, item.ID)
		}

		// 修改返回的项目不应影响原始数据（因为每次loadConfig都读取新数据）
		// 但验证指针确实指向有效内存
		item.Title = "modified"

		// 重新获取，确认数据未被持久化修改（因为没调用save）
		item2, err := svc.GetItem(testItem.ID)
		if err != nil {
			t.Fatalf("GetItem second call failed: %v", err)
		}
		if item2.Title == "modified" {
			t.Error("GetItem returned pointer to shared data, should be independent")
		}
	})

	// 测试并发投票（验证锁机制正确性）
	t.Run("ConcurrentVote", func(t *testing.T) {
		// 为并发测试创建独立服务实例，避免与其他测试的数据竞争
		concurrentTempDir := t.TempDir()
		concurrentSvc := &RoadmapService{
			configFile: filepath.Join(concurrentTempDir, ConfigFileName),
		}
		concurrentSvc.config = concurrentSvc.getDefaultConfig()
		_ = concurrentSvc.saveConfig(concurrentSvc.config)

		if len(concurrentSvc.config.Items) == 0 {
			t.Fatal("no items to vote on")
		}

		testItem := concurrentSvc.config.Items[0]
		initialUp := testItem.Votes.Up

		var wg sync.WaitGroup
		numGoroutines := 10

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				voteReq := VoteRequest{
					ItemID: testItem.ID,
					Vote:   stringPtr("up"),
				}
				_, err := concurrentSvc.Vote(voteReq)
				if err != nil {
					t.Errorf("Vote failed: %v", err)
				}
			}()
		}

		wg.Wait()

		// 验证最终投票数（每个goroutine都投up，但UserVote会被覆盖，最终只算1票）
		// 因为所有goroutine共享同一个UserVote字段，最后一次写入会覆盖前面的
		// 所以预期是 initialUp + 1（只有最后一次投票有效）
		updated, err := concurrentSvc.GetItem(testItem.ID)
		if err != nil {
			t.Fatalf("GetItem failed: %v", err)
		}

		// 由于UserVote是单用户投票，并发投票时只有最后一次有效
		// 但Up计数应该只增加1（因为每次投票都会先取消旧投票）
		expectedUp := initialUp + 1
		if updated.Votes.Up != expectedUp {
			t.Errorf("expected up votes %d after concurrent voting (single user vote), got %d", expectedUp, updated.Votes.Up)
		}
	})
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}

// BenchmarkLoadConfig 性能测试
func BenchmarkLoadConfig(b *testing.B) {
	tempDir := b.TempDir()
	svc := &RoadmapService{
		configFile: filepath.Join(tempDir, ConfigFileName),
	}
	svc.config = svc.getDefaultConfig()
	_ = svc.saveConfig(svc.config)

	b.ResetTimer()
	for b.Loop() {
		_, err := svc.loadConfig()
		if err != nil {
			b.Fatalf("loadConfig failed: %v", err)
		}
	}
}
