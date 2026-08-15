package update

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/KevinYouu/easyGit/internal/i18n"
)

const (
	repoOwner         = "KevinYouu"
	repoName          = "easyGit"
	githubAPIURL      = "https://api.github.com"
	githubDownloadURL = "https://github.com"
	installScriptURL  = "https://raw.githubusercontent.com/KevinYouu/easyGit/main/install.ps1"
	githubUserAgent   = "easyGit-updater"
	checksumFileName  = "checksums.txt"
	// 响应头等待超时：连接建立后迟迟不返回响应头则放弃
	httpResponseHeaderTimeout = 10 * time.Second
	// 整体传输超时：覆盖大文件下载耗时，慢速网络下比仅响应头超时更宽松
	httpClientTimeout = 60 * time.Second
)

// GitHubRelease 表示 GitHub Release API 的响应（仅解析需要的字段）
type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

// ReleaseClient 负责与 GitHub Release API 交互，baseURL 可注入便于测试
type ReleaseClient struct {
	apiBaseURL      string
	downloadBaseURL string
	client          *http.Client
}

// NewReleaseClient 创建指向 GitHub 官方服务的发布客户端
func NewReleaseClient() *ReleaseClient {
	return newReleaseClient(githubAPIURL, githubDownloadURL)
}

func newReleaseClient(apiBaseURL, downloadBaseURL string) *ReleaseClient {
	// 基于默认传输克隆，仅覆盖响应头超时，保留连接池等默认调优
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = httpResponseHeaderTimeout
	return &ReleaseClient{
		apiBaseURL:      apiBaseURL,
		downloadBaseURL: downloadBaseURL,
		client: &http.Client{
			Timeout:   httpClientTimeout,
			Transport: transport,
		},
	}
}

// LatestVersion 查询仓库最新 release 的版本号（tag_name）
func (c *ReleaseClient) LatestVersion() (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", c.apiBaseURL, repoOwner, repoName)
	resp, err := c.get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s: %d", i18n.T("update.api_error_status"), resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	if release.TagName == "" {
		return "", fmt.Errorf("%s", i18n.T("update.missing_tag_name"))
	}
	return release.TagName, nil
}

// AssetName 生成指定版本与平台的安装包文件名
func (c *ReleaseClient) AssetName(version, platform string) string {
	return fmt.Sprintf("easyGit_%s_%s.zip", version, platform)
}

// AssetURL 生成指定资产的下载地址
func (c *ReleaseClient) AssetURL(version, assetName string) string {
	return fmt.Sprintf("%s/%s/%s/releases/download/%s/%s",
		c.downloadBaseURL, repoOwner, repoName, version, assetName)
}

// Checksums 获取指定版本的校验和表（文件名 → SHA256）
func (c *ReleaseClient) Checksums(version string) (map[string]string, error) {
	url := c.AssetURL(version, checksumFileName)
	resp, err := c.get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %d", i18n.T("update.download_failed_status"), resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return parseChecksums(string(content)), nil
}

// DownloadFile 将 url 指向的文件下载到 destPath
func (c *ReleaseClient) DownloadFile(url, destPath string) error {
	resp, err := c.get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %d", i18n.T("update.download_failed_status"), resp.StatusCode)
	}

	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

func (c *ReleaseClient) get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", githubUserAgent)
	return c.client.Do(req)
}

// parseChecksums 解析 goreleaser 生成的 checksums.txt（每行: <sha256>  <文件名>）
func parseChecksums(content string) map[string]string {
	checksums := make(map[string]string)
	for line := range strings.SplitSeq(content, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		checksums[fields[1]] = fields[0]
	}
	return checksums
}

// verifyChecksum 校验文件 SHA256 是否与校验表一致。
// 校验失败的错误信息包含实际与期望哈希，便于诊断篡改或下载截断。
func verifyChecksum(checksums map[string]string, assetName, filePath string) error {
	expected, exists := checksums[assetName]
	if !exists {
		return fmt.Errorf("%s: %s", i18n.T("update.checksum_missing"), assetName)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return err
	}
	actual := hex.EncodeToString(hasher.Sum(nil))
	if actual != expected {
		return fmt.Errorf("%s %s",
			fmt.Sprintf(i18n.T("update.checksum_mismatch"), assetName),
			fmt.Sprintf(i18n.T("update.checksum_mismatch_detail"), actual, expected))
	}
	return nil
}
