package config

import (
	"PiliPili_Frontend/util"
	"github.com/spf13/viper"
)

// Config 保存所有配置值
type Config struct {
	LogLevel            string               // 日志级别 (e.g., INFO, DEBUG, ERROR)
	Encipher            string               // 加密和混淆使用的密钥
	EmbyURL             string               // Emby 服务器地址 (用于 API 调用)
	EmbyPort            int                  // Emby 服务器端口
	EmbyAPIKey          string               // Emby API 密钥
	
	// --- 修改开始: 支持多后端 ---
	// 删除旧的 BackendURL 和 BackendStorageBasePath
	// BackendURL          string 
	// BackendStorageBasePath string 
	
	// 新增: 后端列表配置
	Backends            []BackendConfig      
	// --- 修改结束 ---

	PlayURLMaxAliveTime int                  // 播放链接最大存活时间 (秒)
	ServerPort          int                  // 本服务运行端口
	SpecialMedias       []SpecialMediaConfig // 特殊媒体配置列表
}

// BackendConfig 定义单个后端的配置 (新增结构体)
type BackendConfig struct {
	Name string // 后端名称 (例如: "NAS-1", "GoogleDrive")
	URL  string // 该后端的推流 URL (例如: "https://stream1.example.com")
	Path string // 该后端对应的 Emby 本地路径前缀 (例如: "/mnt/share1")
}

// SpecialMediaConfig 保存特殊媒体的配置
type SpecialMediaConfig struct {
	Key           string // 特殊媒体的唯一标识键
	Name          string // 描述
	MediaPath     string // 媒体文件路径
	ItemId        string // Emby 中的 Item ID
	MediaSourceID string // Emby 中的 Media Source ID
}

// globalConfig 存储已加载的配置
var globalConfig Config

// Initialize 从提供的配置文件加载配置并初始化日志
func Initialize(configFile string, loglevel string) error {
	viper.SetConfigType("yaml")

	if configFile != "" {
		viper.SetConfigFile(configFile)
	}

	if err := viper.ReadInConfig(); err != nil {
		// 默认配置 (如果读取失败或没有文件)
		globalConfig = Config{
			LogLevel:            defaultLogLevel(loglevel),
			Encipher:            "vPQC5LWCN2CW2opz",
			EmbyURL:             "http://127.0.0.1",
			EmbyPort:            8096,
			EmbyAPIKey:          "",
			Backends:            []BackendConfig{}, // 默认为空
			PlayURLMaxAliveTime: 6 * 60 * 60,
			ServerPort:          60002,
			SpecialMedias:       []SpecialMediaConfig{},
		}
	} else {
		// 从文件加载配置
		globalConfig = Config{
			LogLevel:   getLogLevel(loglevel),
			Encipher:   viper.GetString("Encipher"),
			EmbyURL:    viper.GetString("Emby.url"),
			EmbyPort:   viper.GetInt("Emby.port"),
			EmbyAPIKey: viper.GetString("Emby.apiKey"),
			
			// --- 修改: 加载后端列表 ---
			Backends: loadBackends(), 
			
			PlayURLMaxAliveTime: viper.GetInt("PlayURLMaxAliveTime"),
			ServerPort:          viper.GetInt("Server.port"),
			SpecialMedias:       loadSpecialMedias(),
		}
	}

	return nil
}

// loadBackends 从 viper 解析 Backends 配置 (新增函数)
func loadBackends() []BackendConfig {
	var backends []BackendConfig
	// 尝试从配置文件中的 "Backends" 键读取数组
	if err := viper.UnmarshalKey("Backends", &backends); err != nil {
		return []BackendConfig{}
	}
	return backends
}

// loadSpecialMedias 从 viper 解析 SpecialMedias 配置
func loadSpecialMedias() []SpecialMediaConfig {
	var specialMedias []SpecialMediaConfig

	if err := viper.UnmarshalKey("SpecialMedias", &specialMedias); err != nil {
		return []SpecialMediaConfig{}
	}

	return specialMedias
}

// GetConfig 返回全局配置
func GetConfig() Config {
	return globalConfig
}

// IsValid 检查 SpecialMediaConfig 字段是否有效
func (config SpecialMediaConfig) IsValid() bool {
	return config.Key != "" &&
		config.Name != "" &&
		config.MediaPath != "" &&
		config.ItemId != "" &&
		config.MediaSourceID != ""
}

// GetFullEmbyURL 返回包含端口的完整 Emby URL
func GetFullEmbyURL() string {
	return util.BuildFullURL(globalConfig.EmbyURL, globalConfig.EmbyPort)
}

// defaultLogLevel 如果未指定日志级别，则返回默认值
func defaultLogLevel(loglevel string) string {
	if loglevel != "" {
		return loglevel
	}
	return "INFO"
}

// getLogLevel 从参数或配置文件获取日志级别
func getLogLevel(loglevel string) string {
	if loglevel != "" {
		return loglevel
	}
	return viper.GetString("LogLevel")
}
