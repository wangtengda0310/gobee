package gitlab

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GitLabResourceChecker GitLab资源检查器
type GitLabResourceChecker struct {
	gitlabURL   string
	projectPath string
	accessToken string
	apiURL      string
	httpClient  *http.Client
}

// FileInfo 文件信息
type FileInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Path string `json:"path"`
	Mode string `json:"mode"`
}

// NewGitLabResourceChecker 创建新的GitLab资源检查器
func NewGitLabResourceChecker(gitlabURL, projectPath, accessToken string) *GitLabResourceChecker {
	encodedProject := url.PathEscape(projectPath)

	return &GitLabResourceChecker{
		gitlabURL:   strings.TrimSuffix(gitlabURL, "/"),
		projectPath: encodedProject,
		accessToken: accessToken,
		apiURL:      fmt.Sprintf("%s/api/v4", strings.TrimSuffix(gitlabURL, "/")),
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GetLatestCommit 获取最新commit信息
func (c *GitLabResourceChecker) GetLatestCommit(branch string) (string, error) {
	url_ := fmt.Sprintf("%s/api/v4/projects/%s/repository/commits?ref_name=%s&per_page=1",
		c.gitlabURL, c.projectPath, branch)

	//resp, err := c.makeRequest("GET", url_, nil)
	req, err := http.NewRequest("GET", url_, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("PRIVATE-TOKEN", c.accessToken)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}

	var commits []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&commits)
	resp.Body.Close()

	if len(commits) > 0 {
		return commits[0]["id"].(string), nil
	}
	return "", fmt.Errorf("no commits found")
}

// ListAllVoiceFilesWithLinkHeader 使用Link头分页获取所有文件
// branch 分支
// dirPath := "Master/Card/Audio/Voice"
// ext 文件后缀名
func (c *GitLabResourceChecker) ListAllVoiceFilesWithLinkHeader(branch, dirPath, ext string) ([]FileInfo, error) {
	baseURL := fmt.Sprintf("%s/projects/%s/repository/tree", c.apiURL, c.projectPath)

	// 获取最新提交
	ref, err := c.GetLatestCommit(branch)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("path", dirPath)
	params.Set("ref", ref)
	params.Set("per_page", "100")

	currentURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	var allFiles []FileInfo

	for currentURL != "" {
		req, err := http.NewRequest("GET", currentURL, nil)
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %w", err)
		}

		req.Header.Set("PRIVATE-TOKEN", c.accessToken)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("请求失败: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("API错误: %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("读取响应失败: %w", err)
		}

		var files []FileInfo
		if err := json.Unmarshal(body, &files); err != nil {
			return nil, fmt.Errorf("解析JSON失败: %w", err)
		}

		// 如果有后缀提供，这里筛选后缀文件
		suffix := ext
		if ext != "" && !strings.HasPrefix(suffix, ".") {
			suffix = "." + suffix
		}
		for _, file := range files {
			if file.Type == "blob" && (strings.HasSuffix(strings.ToLower(file.Name), suffix) || suffix == "") {
				allFiles = append(allFiles, file)
			}
		}

		// 解析Link头获取下一页
		currentURL = ""
		if linkHeader := resp.Header.Get("Link"); linkHeader != "" {
			links := strings.Split(linkHeader, ",")
			for _, link := range links {
				parts := strings.Split(link, ";")
				if len(parts) >= 2 && strings.Contains(parts[1], `rel="next"`) {
					currentURL = strings.Trim(parts[0], " <>")
					break
				}
			}
		}

		time.Sleep(100 * time.Millisecond) // 避免请求过快
	}

	return allFiles, nil
}
