package api

import (
	"PiliPili_Frontend/config"
	"PiliPili_Frontend/logger"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// 全局复用 Client，避免高并发下端口耗尽
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
		Client:  globalClient, // 引用全局 Client
	}
}

// GetMediaPath 保持原有逻辑，但底层使用了复用的 Client
func (api *EmbyAPI) GetMediaPath(itemID, mediaSourceID string) (string, error) {
	url := fmt.Sprintf("%s/Items/%s/PlaybackInfo?MediaSourceId=%s&api_key=%s",
		api.EmbyURL, itemID, mediaSourceID, api.APIKey)

	logger.Info("Fetching media path from Emby: %s", url)

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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		MediaSources []struct {
			ID   string `json:"Id"`
			Path string `json:"Path"`
		} `json:"MediaSources"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
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
