package stream

import (
	"PiliPili_Frontend/api"
	"PiliPili_Frontend/config"
	"PiliPili_Frontend/logger"
	"PiliPili_Frontend/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/singleflight" // 需 go get
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var cache *Cache
var globalTimeChecker util.TimeChecker
var sfGroup singleflight.Group // 请求合并组

func init() {
	var err error
	cache, err = NewCache(30 * time.Minute)
	if err != nil {
		logger.Error("Failed to initialize cache: %v", err)
		os.Exit(1)
	}
	globalTimeChecker = util.TimeChecker{}
	logger.Info("TimeChecker initialized successfully")
}

func HandleStreamRequest(c *gin.Context) {
	logger.Info("Handling stream request...")
	logRequestDetails(c)

	itemID, mediaSourceID, mediaPath, isSpecialDate := fetchParameters(c)
	if itemID == "" || mediaSourceID == "" {
		return
	}

	if _, found := handleCache(c, itemID, mediaSourceID); found {
		return
	}

	var err error
	mediaPath, err = fetchMediaPathIfNeeded(itemID, mediaSourceID, mediaPath, isSpecialDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	streamingURL, err := generateAndCacheURL(itemID, mediaSourceID, mediaPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("Redirecting to streaming URL: %s", streamingURL)
	c.Header("Location", streamingURL)
	c.Status(http.StatusFound)
}

func fetchParameters(c *gin.Context) (string, string, string, bool) {
	currentTime := time.Now()
	specialConfig := getMediaForSpecialDate(currentTime)
	if specialConfig.IsValid() {
		logger.Info("Special date detected.")
		return specialConfig.ItemId, specialConfig.MediaSourceID, specialConfig.MediaPath, true
	}

	itemID := c.Param("itemID")
	mediaSourceID := c.Query("MediaSourceId")
	if itemID == "" || mediaSourceID == "" {
		logger.Warn("Missing itemID or MediaSourceId")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing itemID or MediaSourceId"})
		return "", "", "", false
	}
	return itemID, mediaSourceID, "", false
}

// 优化：fetchMediaPath 增加 Singleflight 防击穿
func fetchMediaPath(itemID, mediaSourceID string) (string, error) {
	key := fmt.Sprintf("mp:%s:%s", itemID, mediaSourceID)

	// 合并并发请求，同一时刻只有一个请求会真正打到 Emby
	v, err, _ := sfGroup.Do(key, func() (interface{}, error) {
		embyAPI := api.NewEmbyAPI()
		return embyAPI.GetMediaPath(itemID, mediaSourceID)
	})

	if err != nil {
		logger.Error("Failed to fetch media path (merged): %v", err)
		return "", err
	}
	// 不再在此处做路径截取，保留完整原始路径供 generateStreamingURL 匹配
	return v.(string), nil
}

// 优化：多后端匹配 + 快速拼接
func generateStreamingURL(mediaPath, itemID, mediaSourceID string) (string, error) {
	cfg := config.GetConfig()
	
	var selectedBackend config.BackendConfig
	var finalPath string
	matched := false

	// config.go 中已保证 Backends 按 Path 长度降序排列
	for _, backend := range cfg.Backends {
		if strings.HasPrefix(mediaPath, backend.Path) {
			selectedBackend = backend
			// 截取路径
			finalPath = mediaPath[len(backend.Path):]
			finalPath = strings.TrimPrefix(finalPath, "/")
			matched = true
			logger.Info("Matched backend: %s", backend.Name)
			break
		}
	}

	if !matched {
		logger.Error("No matching backend found for: %s", mediaPath)
		return "", fmt.Errorf("no matching backend configuration")
	}

	signatureInstance, _ := GetSignatureInstance()
	expireAt := time.Now().Unix() + int64(cfg.PlayURLMaxAliveTime)
	signature, err := signatureInstance.Encrypt(itemID, mediaSourceID, expireAt)
	if err != nil {
		return "", err
	}

	backendBaseURL := strings.TrimSuffix(selectedBackend.URL, "/")
	
	// 性能优化：Builder 拼接
	var b strings.Builder
	b.Grow(len(backendBaseURL) + len(finalPath) + len(signature) + 20)
	b.WriteString(backendBaseURL)
	b.WriteString("?path=")
	b.WriteString(url.QueryEscape(finalPath))
	b.WriteString("&signature=")
	b.WriteString(signature)

	return b.String(), nil
}

// ... 下面函数保持不变：getMediaForSpecialDate, getMediaForMissingMedia, handleCache, validateSignature, generateAndCacheURL, fetchMediaPathIfNeeded
// (为节省篇幅省略，请使用原有逻辑，只需注意 fetchMediaPathIfNeeded 内部调用的是我们修改过的 fetchMediaPath)

func getMediaForSpecialDate(t time.Time) config.SpecialMediaConfig {
	// ... 保持原样 ...
	// 需注意引用 config.GetConfig().SpecialMedias
	specialMedias := config.GetConfig().SpecialMedias
	// ... 循环逻辑保持原样 ...
	for _, media := range specialMedias {
		switch media.Key {
		case "ChineseNewYearEve":
			if globalTimeChecker.IsChineseNewYearEve(t) { return media }
		case "October1":
			if globalTimeChecker.IsOctober1Morning(t) { return media }
		case "December13":
			if globalTimeChecker.IsDecember13Morning(t) { return media }
		case "September18":
			if globalTimeChecker.IsSeptember18Morning(t) { return media }
		}
	}
	return config.SpecialMediaConfig{}
}

func getMediaForMissingMedia() config.SpecialMediaConfig {
	specialMedias := config.GetConfig().SpecialMedias
	for _, media := range specialMedias {
		if media.Key == "MediaMissing" { return media }
	}
	return config.SpecialMediaConfig{}
}

func handleCache(c *gin.Context, itemID, mediaSourceID string) (string, bool) {
	cacheKey := fmt.Sprintf("%s:%s", itemID, mediaSourceID)
	if cachedURL, found := cache.Get(cacheKey); found {
		logger.Info("Cache hit for key: %s", cacheKey)
		if validateSignature(cachedURL) {
			logger.Debug("Valid signature cache hit")
			c.Header("Location", cachedURL)
			c.Status(http.StatusFound)
			return cachedURL, true
		}
	}
	return "", false
}

func fetchMediaPathIfNeeded(itemID, mediaSourceID, mediaPath string, isSpecialDate bool) (string, error) {
	if !isSpecialDate {
		var err error
		mediaPath, err = fetchMediaPath(itemID, mediaSourceID)
		if err != nil {
			missing := getMediaForMissingMedia()
			if missing.ItemId != "" {
				return missing.MediaPath, nil // 使用默认媒体
			}
			return "", err
		}
	}
	return mediaPath, nil
}

func generateAndCacheURL(itemID, mediaSourceID, mediaPath string) (string, error) {
	streamingURL, err := generateStreamingURL(mediaPath, itemID, mediaSourceID)
	if err != nil { return "", err }
	
	cacheKey := fmt.Sprintf("%s:%s", itemID, mediaSourceID)
	_ = cache.Set(cacheKey, streamingURL)
	return streamingURL, nil
}

func validateSignature(cachedURL string) bool {
	// ... 保持原样 ...
	signatureStart := "signature="
	index := strings.Index(cachedURL, signatureStart)
	if index == -1 { return false }
	signature := cachedURL[index+len(signatureStart):]
	inst, _ := GetSignatureInstance()
	decoded, err := inst.Decrypt(signature)
	if err != nil { return false }
	expireAt, ok := decoded["expireAt"].(float64)
	if !ok || int64(expireAt) <= time.Now().Unix() { return false }
	return true
}

func logRequestDetails(c *gin.Context) {
	logger.Debug("Request Headers: %v", c.Request.Header)
	// Body 读取已移至 Middleware 处理
}
