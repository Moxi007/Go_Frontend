package api

import (
	"Go_Frontend/config"
	"Go_Frontend/logger"
	"errors"
	"fmt"
	"net/http"
	"time"

	// ✅ 优化: 引入高性能 JSON 库替代标准库
	// 需先运行: go get github.com/goccy/go-json
	"github.com/goccy/go-json"
)

// 全局复用 Client
var globalClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	},
}

type EmbyAPI struct {
	EmbyURL string
	APIKey  string
	Client  *http.Client
}

func NewEmbyAPI() *EmbyAPI {
	cfg := config.GetConfig()
	return &EmbyAPI{
		EmbyURL: config.GetFullEmbyURL(),
		APIKey:  cfg.EmbyAPIKey,
		Client:  globalClient,
	}
}

// GetMediaPath 获取媒体路径
func (api *EmbyAPI) GetMediaPath(itemID, mediaSourceID string) (string, error) {
	url := fmt.Sprintf("%s/Items/%s/PlaybackInfo?MediaSourceId=%s&api_key=%s",
		api.EmbyURL, itemID, mediaSourceID, api.APIKey)

	logger.Debug("Fetching media path: %s", url)

	resp, err := api.Client.Get(url)
	if err != nil {
		logger.Error("Failed to fetch media path: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Received non-200 response from Emby: %d", resp.StatusCode)
		return "", errors.New("failed to fetch media path")
	}

	// ✅ 优化: 流式解析 JSON
	// 避免 io.ReadAll 读取整个 Body 到内存，大幅降低 GC 压力
	var result struct {
		MediaSources []struct {
			ID   string `json:"Id"`
			Path string `json:"Path"`
		} `json:"MediaSources"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Error("Failed to decode Emby response: %v", err)
		return "", err
	}

	for _, source := range result.MediaSources {
		if source.ID == mediaSourceID {
			logger.Info("Found media path: %s", source.Path)
			return source.Path, nil
		}
	}

	return "", errors.New("media source not found")
}
