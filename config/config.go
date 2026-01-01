package config

import (
	"PiliPili_Frontend/util"
	"github.com/spf13/viper"
	"sort"
)

// Config 保存所有配置值
type Config struct {
	LogLevel            string               // 日志级别
	Encipher            string               // 加密密钥
	EmbyURL             string               // Emby 地址
	EmbyPort            int                  // Emby 端口
	EmbyAPIKey          string               // API 密钥
	
	// --- 多后端配置 ---
	Backends            []BackendConfig      
	
	PlayURLMaxAliveTime int                  // 链接有效期
	ServerPort          int                  // 监听端口
	SpecialMedias       []SpecialMediaConfig // 特殊媒体
}

// BackendConfig 单个后端配置
type BackendConfig struct {
	Name string 
	URL  string 
	Path string 
}

// SpecialMediaConfig 特殊媒体配置
type SpecialMediaConfig struct {
	Key           string
	Name          string
	MediaPath     string
	ItemId        string
	MediaSourceID string
}

var globalConfig Config

func Initialize(configFile string, loglevel string) error {
	viper.SetConfigType("yaml")
	if configFile != "" {
		viper.SetConfigFile(configFile)
	}

	if err := viper.ReadInConfig(); err != nil {
		// 默认兜底配置
		globalConfig = Config{
			LogLevel:            defaultLogLevel(loglevel),
			Encipher:            "vPQC5LWCN2CW2opz",
			EmbyURL:             "http://127.0.0.1",
			EmbyPort:            8096,
			Backends:            []BackendConfig{},
			PlayURLMaxAliveTime: 21600,
			ServerPort:          60001,
			SpecialMedias:       []SpecialMediaConfig{},
		}
	} else {
		globalConfig = Config{
			LogLevel:            getLogLevel(loglevel),
			Encipher:            viper.GetString("Encipher"),
			EmbyURL:             viper.GetString("Emby.url"),
			EmbyPort:            viper.GetInt("Emby.port"),
			EmbyAPIKey:          viper.GetString("Emby.apiKey"),
			Backends:            loadBackends(), // 加载并排序
			PlayURLMaxAliveTime: viper.GetInt("PlayURLMaxAliveTime"),
			ServerPort:          viper.GetInt("Server.port"),
			SpecialMedias:       loadSpecialMedias(),
		}
	}
	return nil
}

// loadBackends 加载后端并按路径长度降序排序（防止短路径误匹配）
func loadBackends() []BackendConfig {
	var backends []BackendConfig
	if err := viper.UnmarshalKey("Backends", &backends); err != nil {
		return []BackendConfig{}
	}
	
	// 核心优化：按 Path 长度降序排序
	// 确保 /mnt/anime/movie (长) 优先于 /mnt/anime (短) 被匹配
	sort.SliceStable(backends, func(i, j int) bool {
		return len(backends[i].Path) > len(backends[j].Path)
	})
	
	return backends
}

func loadSpecialMedias() []SpecialMediaConfig {
	var specialMedias []SpecialMediaConfig
	if err := viper.UnmarshalKey("SpecialMedias", &specialMedias); err != nil {
		return []SpecialMediaConfig{}
	}
	return specialMedias
}

// GetConfig 返回指针，避免结构体拷贝
func GetConfig() *Config {
	return &globalConfig
}

func (config SpecialMediaConfig) IsValid() bool {
	return config.Key != "" && config.Name != "" && config.MediaPath != "" && config.ItemId != "" && config.MediaSourceID != ""
}

func GetFullEmbyURL() string {
	return util.BuildFullURL(globalConfig.EmbyURL, globalConfig.EmbyPort)
}

func defaultLogLevel(loglevel string) string {
	if loglevel != "" { return loglevel }
	return "INFO"
}

func getLogLevel(loglevel string) string {
	if loglevel != "" { return loglevel }
	return viper.GetString("LogLevel")
}
