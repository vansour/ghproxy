/*
Git文件加速代理服务器
===================

这是一个用于加速访问GitHub、GitLab、Hugging Face等代码托管平台文件的代理服务器。
主要功能：
1. 代理并加速文件下载
2. 支持多种平台的URL格式转换
3. 提供Web界面生成加速链接
4. 支持wget、curl、git clone等命令
5. 可配置的文件大小限制、日志管理等

支持的平台：
- GitHub (github.com, raw.githubusercontent.com)
- GitLab (gitlab.com)
- Hugging Face (huggingface.co, hf.co)

作者：vansour
*/

package main

import (
	"encoding/json" // JSON数据编码解码
	"fmt"           // 格式化输入输出
	"io"            // 输入输出原语
	"log"           // 日志记录
	"net/http"      // HTTP客户端和服务器
	"net/url"       // URL解析
	"os"            // 操作系统接口
	"strconv"       // 字符串转换
	"strings"       // 字符串操作
	"time"          // 时间相关操作
)

// ==================== 全局变量 ====================

// ==================== 配置结构体 ====================

// Config 主配置结构体，映射config.toml文件中的所有配置项
// 使用toml标签来指定配置文件中对应的字段名
type Config struct {
	// 服务器相关配置
	Server struct {
		Host      string `toml:"host"`      // 监听地址（如0.0.0.0, 127.0.0.1）
		Port      int    `toml:"port"`      // 监听端口号（如8080）
		SizeLimit int    `toml:"sizeLimit"` // 文件大小限制（单位：MB）
	} `toml:"server"`

	// 日志相关配置
	Log struct {
		LogFilePath string `toml:"logFilePath"` // 日志文件存储路径
		MaxLogSize  int    `toml:"maxLogSize"`  // 单个日志文件最大大小（单位：MB）
		Level       string `toml:"level"`       // 日志级别（debug/info/warn/error/none）
	} `toml:"log"`

	// 黑名单配置
	// 用于阻止特定域名或IP的访问
	Blacklist struct {
		Enabled       bool   `toml:"enabled"`       // 是否启用黑名单功能
		BlacklistFile string `toml:"blacklistFile"` // 黑名单文件路径（JSON格式）
	} `toml:"blacklist"`

	// 白名单配置
	// 用于仅允许特定域名或IP的访问（启用时只允许白名单内的访问）
	Whitelist struct {
		Enabled       bool   `toml:"enabled"`       // 是否启用白名单功能
		WhitelistFile string `toml:"whitelistFile"` // 白名单文件路径（JSON格式）
	} `toml:"whitelist"`

	// 速率限制配置
	// 用于防止服务器被过度使用或滥用
	RateLimit struct {
		Enabled       bool `toml:"enabled"`       // 是否启用速率限制
		RatePerMinute int  `toml:"ratePerMinute"` // 每分钟允许的请求数
		Burst         int  `toml:"burst"`         // 突发请求允许数量

		// 带宽限制子配置
		// 用于控制服务器和单个连接的带宽使用
		BandwidthLimit struct {
			Enabled     bool   `toml:"enabled"`     // 是否启用带宽限制
			TotalLimit  string `toml:"totalLimit"`  // 服务器总带宽限制（如"100mbps"）
			TotalBurst  string `toml:"totalBurst"`  // 服务器总带宽突发限制
			SingleLimit string `toml:"singleLimit"` // 单个连接带宽限制
			SingleBurst string `toml:"singleBurst"` // 单个连接带宽突发限制
		} `toml:"bandwidthLimit"`
	} `toml:"rateLimit"`
}

// ==================== 全局配置变量 ====================

// config 全局配置变量，存储从配置文件加载的所有配置信息
// 在程序启动时通过loadConfig函数初始化
var config Config

// ==================== 配置管理函数 ====================

// loadConfig 加载配置文件
// 参数：
//
//	configPath: 配置文件路径（通常是config.toml）
//
// 返回值：
//
//	error: 加载失败时返回错误信息，成功时返回nil
//
// 功能说明：
// 1. 检查配置文件是否存在
// 2. 如果不存在，使用默认配置
// 3. 如果存在，解析TOML格式的配置文件
// 4. 将配置信息加载到全局config变量中
func loadConfig(configPath string) error {
	// 暂时使用默认配置，后续可以添加toml支持
	// TODO: 添加TOML配置文件解析功能
	log.Printf("使用默认配置")
	setDefaultConfig()
	return nil
}

// setDefaultConfig 设置默认配置
// 当配置文件不存在或解析失败时使用
// 所有配置项都使用安全的默认值
func setDefaultConfig() {
	// 服务器配置默认值
	config.Server.Host = "0.0.0.0" // 监听所有网络接口
	config.Server.Port = 8080      // 默认端口8080
	config.Server.SizeLimit = 2048 // 默认文件大小限制2GB

	// 日志配置默认值
	config.Log.LogFilePath = "./logs/ghproxy.log" // 相对于程序目录的日志路径
	config.Log.MaxLogSize = 5                     // 默认单个日志文件最大5MB
	config.Log.Level = "info"                     // 默认日志级别为info

	// 功能开关默认值（默认都关闭，确保安全）
	config.Blacklist.Enabled = false // 默认不启用黑名单
	config.Whitelist.Enabled = false // 默认不启用白名单
	config.RateLimit.Enabled = false // 默认不启用速率限制
}

// ==================== 配置文件生成函数 ====================

// generateConfigFiles 生成配置相关的示例文件
// 根据config.toml中的配置，自动创建相关目录和示例文件
func generateConfigFiles() error {
	log.Printf("开始生成配置相关文件...")

	// 创建日志目录
	if err := createLogDirectory(); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}

	// 创建配置目录
	if err := createConfigDirectory(); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 生成黑名单示例文件
	if err := generateBlacklistFile(); err != nil {
		return fmt.Errorf("生成黑名单文件失败: %v", err)
	}

	// 生成白名单示例文件
	if err := generateWhitelistFile(); err != nil {
		return fmt.Errorf("生成白名单文件失败: %v", err)
	}

	// 生成完整的config.toml示例文件
	if err := generateConfigTomlExample(); err != nil {
		return fmt.Errorf("生成config.toml示例失败: %v", err)
	}

	log.Printf("配置文件生成完成")
	return nil
}

// createLogDirectory 创建日志目录
func createLogDirectory() error {
	logDir := "./logs" // 默认在当前目录下创建logs文件夹
	if config.Log.LogFilePath != "" {
		// 从日志文件路径中提取目录
		logDir = config.Log.LogFilePath[:strings.LastIndex(config.Log.LogFilePath, "/")]
	}

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}
	log.Printf("日志目录已创建: %s", logDir)
	return nil
}

// createConfigDirectory 创建配置目录
func createConfigDirectory() error {
	configDir := "./config" // 在当前目录下创建config文件夹
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	log.Printf("配置目录已创建: %s", configDir)
	return nil
}

// generateBlacklistFile 生成黑名单示例文件
func generateBlacklistFile() error {
	blacklistPath := config.Blacklist.BlacklistFile
	if blacklistPath == "" {
		blacklistPath = "./config/blacklist.json" // 默认在当前目录下的config文件夹
	}

	// 如果文件已存在，不覆盖
	if _, err := os.Stat(blacklistPath); err == nil {
		log.Printf("黑名单文件已存在，跳过生成: %s", blacklistPath)
		return nil
	}

	// 黑名单示例数据
	blacklistExample := map[string]interface{}{
		"domains": []string{
			"malicious-example.com",
			"spam-site.net",
		},
		"ips": []string{
			"192.168.1.100",
			"10.0.0.50",
		},
		"paths": []string{
			"/malicious-path/*",
			"*/dangerous-file.exe",
		},
		"description": "黑名单配置文件 - 在此列出需要阻止访问的域名、IP和路径模式",
		"usage":       "启用黑名单功能需要在config.toml中设置 blacklist.enabled = true",
	}

	return writeJSONFile(blacklistPath, blacklistExample)
}

// generateWhitelistFile 生成白名单示例文件
func generateWhitelistFile() error {
	whitelistPath := config.Whitelist.WhitelistFile
	if whitelistPath == "" {
		whitelistPath = "./config/whitelist.json" // 默认在当前目录下的config文件夹
	}

	// 如果文件已存在，不覆盖
	if _, err := os.Stat(whitelistPath); err == nil {
		log.Printf("白名单文件已存在，跳过生成: %s", whitelistPath)
		return nil
	}

	// 白名单示例数据
	whitelistExample := map[string]interface{}{
		"domains": []string{
			"github.com",
			"gitlab.com",
			"huggingface.co",
			"raw.githubusercontent.com",
			"gist.githubusercontent.com",
			"hf.co",
			"cdn-lfs.huggingface.co",
		},
		"ips": []string{
			"140.82.112.0/20",
			"140.82.114.0/20",
		},
		"paths": []string{
			"*/blob/*",
			"*/raw/*",
			"*/resolve/*",
			"*/archive/*",
		},
		"description": "白名单配置文件 - 只允许访问此列表中的域名、IP和路径模式",
		"usage":       "启用白名单功能需要在config.toml中设置 whitelist.enabled = true",
		"note":        "启用白名单后，只有在此列表中的域名才能被代理访问",
	}

	return writeJSONFile(whitelistPath, whitelistExample)
}

// generateConfigTomlExample 生成完整的config.toml示例文件
func generateConfigTomlExample() error {
	examplePath := "config.toml.example"

	// 如果文件已存在，不覆盖
	if _, err := os.Stat(examplePath); err == nil {
		log.Printf("配置示例文件已存在，跳过生成: %s", examplePath)
		return nil
	}

	configExample := `# Git文件加速代理配置文件
# 详细说明：https://github.com/vansour/ghproxy

# ==================== 服务器配置 ====================
[server]
host = "0.0.0.0"       # 监听地址，0.0.0.0表示监听所有网络接口
port = 8080            # 监听端口
sizeLimit = 2048       # 文件大小限制，单位MB，超过此大小的文件将被拒绝

# ==================== 日志配置 ====================
[log]
logFilePath = "./logs/ghproxy.log"    # 日志文件路径（相对于程序目录）
maxLogSize = 5                        # 单个日志文件最大大小，单位MB
level = "info"                        # 日志级别：debug, info, warn, error, none

# ==================== 黑名单配置 ====================
[blacklist]
enabled = false                              # 是否启用黑名单功能
blacklistFile = "./config/blacklist.json"   # 黑名单文件路径（相对于程序目录）

# ==================== 白名单配置 ====================
[whitelist]
enabled = false                              # 是否启用白名单功能
whitelistFile = "./config/whitelist.json"   # 白名单文件路径（相对于程序目录）

# ==================== 速率限制配置 ====================
[rateLimit]
enabled = false       # 是否启用速率限制
ratePerMinute = 180   # 每分钟允许的请求数
burst = 5             # 突发请求数量

# 带宽限制配置（高级功能）
[rateLimit.bandwidthLimit]
enabled = false           # 是否启用带宽限制
totalLimit = "100mbps"    # 服务器总带宽限制
totalBurst = "100mbps"    # 服务器总带宽突发限制
singleLimit = "10mbps"    # 单个连接带宽限制
singleBurst = "10mbps"    # 单个连接带宽突发限制

# ==================== 使用说明 ====================
# 1. 修改配置后需要重启服务才能生效
# 2. 日志文件会自动轮转，避免文件过大
# 3. 黑名单和白名单不能同时启用
# 4. 速率限制可以有效防止滥用
# 5. 带宽限制需要额外的依赖包支持
# 6. 所有路径都是相对于程序可执行文件的位置
`

	file, err := os.Create(examplePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(configExample)
	if err != nil {
		return err
	}

	log.Printf("配置示例文件已创建: %s", examplePath)
	return nil
}

// writeJSONFile 写入JSON文件的辅助函数
func writeJSONFile(filePath string, data interface{}) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // 设置缩进，使JSON格式更易读
	if err := encoder.Encode(data); err != nil {
		return err
	}

	log.Printf("JSON文件已创建: %s", filePath)
	return nil
}

// ==================== 核心处理函数 ====================

// proxyHandler 核心代理处理函数
// 这是整个代理服务器的核心，处理所有的HTTP请求
//
// 参数：
//
//	w: HTTP响应写入器，用于向客户端发送响应
//	r: HTTP请求对象，包含客户端发送的所有请求信息
//
// 功能说明：
// 1. 处理特殊路径（如favicon.ico）
// 2. 解析和验证目标URL
// 3. 转换不同平台的URL格式
// 4. 代理请求到目标服务器
// 5. 检查文件大小限制
// 6. 返回响应给客户端
func proxyHandler(w http.ResponseWriter, r *http.Request) {
	// 直接把 /favicon.ico 交给文件系统
	// 这样可以让浏览器正常显示网站图标
	if r.URL.Path == "/favicon.ico" {
		http.ServeFile(w, r, "favicon.ico")
		return
	}

	// ========== 第一步：获取和处理请求路径 ==========

	// 直接从RequestURI获取完整路径，这样可以避免Go的路径清理
	// RequestURI包含原始的请求路径，不会被Go的HTTP库自动"清理"
	// 这对于代理服务器来说很重要，因为我们需要保持URL的原始格式
	requestURI := r.RequestURI

	// 去掉开头的 "/"，因为我们要把剩余部分作为目标URL
	// 例如："/https://github.com/user/repo" -> "https://github.com/user/repo"
	requestPath := strings.TrimPrefix(requestURI, "/")

	// 添加调试日志，记录请求信息便于调试和监控
	log.Printf("收到请求: %s", requestURI)
	log.Printf("处理路径: %s", requestPath)

	// 处理URL解码问题
	// 浏览器可能会对URL进行编码，我们需要将其解码回原始格式
	// 例如：%3A -> :, %2F -> /
	if decodedPath, err := url.QueryUnescape(requestPath); err == nil {
		requestPath = decodedPath
		log.Printf("解码后路径: %s", requestPath)
	}

	// ========== 第二步：处理根路径请求（显示Web界面） ==========

	// 如果是根路径或空路径，返回使用说明页面
	// 这个页面提供了一个友好的Web界面，用户可以输入URL并生成加速链接
	if requestPath == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		// 下面是完整的HTML页面，包含样式和JavaScript
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Git文件加速代理</title>
    <link rel="icon" type="image/x-icon" href="/favicon.ico">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css">
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            min-height: 100vh;
            color: #333;
        }
        
        .container {
            max-width: 900px;
            margin: 0 auto;
            padding: 20px;
        }
        
        .header {
            text-align: center;
            color: white;
            margin-bottom: 40px;
        }
        
        .header h1 {
            font-size: 2.5rem;
            margin-bottom: 10px;
            font-weight: 700;
        }
        
        .header p {
            font-size: 1.1rem;
            opacity: 0.9;
        }
        
        .main-panel {
            background: white;
            border-radius: 16px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            padding: 40px;
            margin-bottom: 30px;
        }
        
        .input-section {
            margin-bottom: 30px;
        }
        
        .input-section label {
            display: block;
            margin-bottom: 10px;
            font-weight: 600;
            color: #333;
        }
        
        .url-input {
            width: 100%%;
            padding: 15px 20px;
            border: 2px solid #e1e5e9;
            border-radius: 10px;
            font-size: 16px;
            transition: all 0.3s ease;
        }
        
        .url-input:focus {
            outline: none;
            border-color: #667eea;
            box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
        }
        
        .generate-btn {
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
            border: none;
            padding: 15px 30px;
            border-radius: 10px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.3s ease;
            margin-top: 15px;
            width: 100%%;
        }
        
        .generate-btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 10px 20px rgba(102, 126, 234, 0.3);
        }
        
        .results {
            margin-top: 30px;
        }
        
        .result-tabs {
            display: flex;
            border-bottom: 2px solid #e9ecef;
            margin-bottom: 20px;
        }
        
        .tab-btn {
            flex: 1;
            padding: 12px 16px;
            background: none;
            border: none;
            border-bottom: 3px solid transparent;
            cursor: pointer;
            font-size: 14px;
            font-weight: 500;
            color: #6c757d;
            transition: all 0.3s ease;
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 8px;
        }
        
        .tab-btn:hover {
            color: #495057;
            background: #f8f9fa;
        }
        
        .tab-btn.active {
            color: #667eea;
            border-bottom-color: #667eea;
            background: #f8f9fa;
        }
        
        .result-item {
            background: #f8f9fa;
            border: 1px solid #e9ecef;
            border-radius: 10px;
            padding: 20px;
        }
        
        .result-item h3 {
            color: #495057;
            margin-bottom: 10px;
            font-size: 1.1rem;
        }
        
        .result-code {
            background: #f1f3f4;
            border: 1px solid #dadce0;
            border-radius: 6px;
            padding: 12px;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            font-size: 14px;
            word-break: break-all;
            position: relative;
            min-height: 20px;
        }
        
        .result-code span {
            display: block;
            min-height: 20px;
        }
        
        .result-code span:not(:empty) {
            padding-right: 80px;
        }
        
        .copy-btn {
            position: absolute;
            top: 10px;
            right: 10px;
            background: #667eea;
            color: white;
            border: none;
            padding: 5px 10px;
            border-radius: 4px;
            font-size: 12px;
            cursor: pointer;
            transition: background 0.3s ease;
            opacity: 0;
            visibility: hidden;
        }
        
        .result-code span:not(:empty) + .copy-btn {
            opacity: 1;
            visibility: visible;
        }
        
        .copy-btn:hover {
            background: #5a6fd8;
        }
        
        .platforms {
            background: white;
            border-radius: 16px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            padding: 30px;
        }
        
        .platforms h2 {
            text-align: center;
            color: #333;
            margin-bottom: 20px;
        }
        
        .platform-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
        }
        
        .platform-card {
            background: #f8f9fa;
            border-radius: 10px;
            padding: 20px;
            text-align: center;
        }
        
        .platform-card h3 {
            color: #495057;
            margin-bottom: 10px;
        }
        
        .platform-card p {
            color: #6c757d;
            font-size: 0.9rem;
        }
        
        .features {
            background: white;
            border-radius: 16px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            padding: 30px;
            margin-bottom: 30px;
        }
        
        .features h2 {
            text-align: center;
            color: #333;
            margin-bottom: 20px;
        }
        
        .feature-list {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
        }
        
        .feature-item {
            background: #f8f9fa;
            border-radius: 10px;
            padding: 20px;
        }
        
        .feature-item h3 {
            color: #495057;
            margin-bottom: 10px;
            font-size: 1.1rem;
        }
        
        .feature-item p {
            color: #6c757d;
            font-size: 0.9rem;
            line-height: 1.5;
        }
        
        .toast {
            position: fixed;
            top: 20px;
            right: 20px;
            background: #28a745;
            color: white;
            padding: 15px 20px;
            border-radius: 8px;
            display: none;
            z-index: 1000;
        }
        
        /* ========== Footer 样式 ========== */
        .footer {
            background: white;
            border-radius: 16px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            padding: 30px;
            margin-top: 30px;
            text-align: center;
            border-top: 2px solid #e9ecef;
        }
        
        .footer-content {
            display: flex;
            justify-content: center;
            align-items: center;
            gap: 30px;
            flex-wrap: wrap;
        }
        
        .footer-links {
            display: flex;
            gap: 20px;
            align-items: center;
        }
        
        .footer-link {
            display: inline-flex;
            align-items: center;
            gap: 8px;
            text-decoration: none;
            color: #667eea;
            font-weight: 500;
            padding: 8px 16px;
            border-radius: 8px;
            transition: all 0.3s ease;
            border: 2px solid transparent;
        }
        
        .footer-link:hover {
            color: #5a6fd8;
            background: #f8f9fa;
            border-color: #e9ecef;
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(102, 126, 234, 0.2);
        }
        
        .footer-link i {
            font-size: 18px;
        }
        
        .copyright {
            color: #6c757d;
            font-size: 14px;
            margin: 0;
        }
        
        @media (max-width: 768px) {
            .container {
                padding: 15px;
            }
            
            .main-panel {
                padding: 25px;
            }
            
            .header h1 {
                font-size: 2rem;
            }
            
            .footer-content {
                flex-direction: column;
                gap: 20px;
            }
            
            .footer-links {
                flex-direction: column;
                gap: 15px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 Git文件加速代理</h1>
            <p>支持 GitHub、GitLab、Hugging Face 三大平台文件加速访问</p>
        </div>
        
        <div class="main-panel">
            <div class="input-section">
                <label for="original-url">输入原始链接：</label>
                <input type="text" id="original-url" class="url-input" 
                       placeholder="例如：https://github.com/user/repo/blob/main/file.txt"
                       oninput="generateLinksRealtime()">
            </div>
            
            <div id="results" class="results">
                <div class="result-tabs">
                    <button class="tab-btn active" onclick="switchTab('browser')">
                        <span>🌐</span> 浏览器访问
                    </button>
                    <button class="tab-btn" onclick="switchTab('wget')">
                        <span>📥</span> wget 下载
                    </button>
                    <button class="tab-btn" onclick="switchTab('curl')">
                        <span>📦</span> curl 下载
                    </button>
                    <button class="tab-btn" onclick="switchTab('git')">
                        <span>🔀</span> git clone
                    </button>
                </div>
                
                <div class="result-item">
                    <div class="result-code">
                        <span id="result-content"></span>
                        <button class="copy-btn" onclick="copyResult()">复制</button>
                    </div>
                </div>
            </div>
        </div>
        
        <div class="platforms">
            <h2>支持的平台</h2>
            <div class="platform-grid">
                <div class="platform-card">
                    <h3>GitHub</h3>
                    <p>支持仓库文件、Raw文件、Gist等</p>
                </div>
                <div class="platform-card">
                    <h3>GitLab</h3>
                    <p>支持项目文件和Raw文件</p>
                </div>
                <div class="platform-card">
                    <h3>Hugging Face</h3>
                    <p>支持模型和数据集文件</p>
                </div>
            </div>
        </div>
        
        <!-- Footer 版权信息和链接 -->
        <div class="footer">
            <div class="footer-content">
                <p class="copyright">© 2024-2025 Git文件加速代理. All rights reserved.</p>
                <div class="footer-links">
                    <a href="https://github.com/vansour/ghproxy" target="_blank" class="footer-link">
                        <i class="fab fa-github"></i>
                        GitHub 仓库
                    </a>
                    <a href="https://hub.docker.com/r/vansour/ghproxy" target="_blank" class="footer-link">
                        <i class="fab fa-docker"></i>
                        Docker 镜像
                    </a>
                </div>
            </div>
        </div>
    </div>
    
    <div id="toast" class="toast">复制成功！</div>
    
    <script>
        // 存储所有生成的链接
        let generatedLinks = {
            browser: '',
            wget: '',
            curl: '',
            git: ''
        };
        
        // 当前活跃的标签
        let currentTab = 'browser';
        
        function switchTab(tabName) {
            // 更新标签按钮状态
            document.querySelectorAll('.tab-btn').forEach(btn => {
                btn.classList.remove('active');
            });
            event.target.closest('.tab-btn').classList.add('active');
            
            // 更新当前标签
            currentTab = tabName;
            
            // 更新显示内容
            updateResultContent();
        }
        
        function updateResultContent() {
            const resultContent = document.getElementById('result-content');
            resultContent.textContent = generatedLinks[currentTab];
        }
        
        function generateLinksRealtime() {
            const originalUrl = document.getElementById('original-url').value.trim();
            
            // 清空所有链接
            generatedLinks = {
                browser: '',
                wget: '',
                curl: '',
                git: ''
            };
            
            // 如果输入为空，清空显示
            if (!originalUrl) {
                updateResultContent();
                return;
            }
            
            // 检查URL格式
            if (!originalUrl.startsWith('http://') && !originalUrl.startsWith('https://')) {
                generatedLinks[currentTab] = '请输入完整的URL（包含http://或https://）';
                updateResultContent();
                return;
            }
            
            // 检查是否是支持的域名
            try {
                const url = new URL(originalUrl);
                const supportedDomains = [
                    'github.com', 'gitlab.com', 'huggingface.co',
                    'raw.githubusercontent.com', 'gist.githubusercontent.com',
                    'hf.co', 'cdn-lfs.huggingface.co'
                ];
                
                if (!supportedDomains.some(domain => url.hostname === domain || url.hostname.endsWith('.' + domain))) {
                    generatedLinks[currentTab] = '只支持GitHub、GitLab、Hugging Face相关域名';
                    updateResultContent();
                    return;
                }
                
                // 特殊处理Hugging Face - 仅支持文件下载
                if (url.hostname === 'huggingface.co' || url.hostname === 'hf.co') {
                    if (!url.pathname.includes('/resolve/') && !url.pathname.includes('/blob/')) {
                        generatedLinks[currentTab] = 'Hugging Face 链接需要包含具体文件路径（/blob/ 或 /resolve/）';
                        updateResultContent();
                        return;
                    }
                }
                
                // 特殊处理GitHub - 仅支持文件下载
                if (url.hostname === 'github.com') {
                    const path = url.pathname;
                    // 只允许文件路径和gist，不允许直接访问仓库根路径
                    const isFilePath = path.includes('/blob/') || path.includes('/raw/') || path.includes('/tree/');
                    // 允许gist
                    const isGist = path.includes('/gist/');
                    
                    if (!isFilePath && !isGist) {
                        generatedLinks[currentTab] = 'GitHub 链接仅支持文件下载路径（/blob/, /raw/, /tree/）或gist，git clone请使用git命令';
                        updateResultContent();
                        return;
                    }
                }
                
                // 特殊处理GitLab - 仅支持文件下载
                if (url.hostname === 'gitlab.com') {
                    const path = url.pathname;
                    // 只允许文件路径，不允许直接访问仓库根路径
                    const isFilePath = path.includes('/-/blob/') || path.includes('/-/raw/') || path.includes('/-/tree/');
                    
                    if (!isFilePath) {
                        generatedLinks[currentTab] = 'GitLab 链接仅支持文件下载路径（/-/blob/, /-/raw/, /-/tree/），git clone请使用git命令';
                        updateResultContent();
                        return;
                    }
                }
            } catch (e) {
                generatedLinks[currentTab] = 'URL格式无效';
                updateResultContent();
                return;
            }
            
            // 获取当前域名和端口
            const proxyHost = window.location.host;
            const proxyProtocol = window.location.protocol;
            const baseUrl = proxyProtocol + '//' + proxyHost;
            
            // 生成加速链接
            const acceleratedUrl = baseUrl + '/' + originalUrl;
            
            // 存储各种格式的链接
            generatedLinks.browser = acceleratedUrl;
            generatedLinks.wget = 'wget "' + acceleratedUrl + '"';
            generatedLinks.curl = 'curl -L "' + acceleratedUrl + '"';
            
            // Git clone处理
            if (originalUrl.includes('github.com') || originalUrl.includes('gitlab.com')) {
                let gitUrl = originalUrl;
                
                // 检查是否是不支持git clone的链接类型
                if (gitUrl.includes('/archive/') || 
                    gitUrl.includes('/releases/') || 
                    gitUrl.includes('/tarball/') ||
                    gitUrl.includes('/zipball/') ||
                    gitUrl.includes('/raw/') ||
                    gitUrl.includes('/-/raw/') ||
                    gitUrl.includes('/gist/')) {
                    generatedLinks.git = '此链接不支持 git clone（archive/release/raw文件请使用浏览器或下载命令）';
                } else {
                    // 处理GitHub/GitLab仓库链接
                    if (gitUrl.includes('/blob/') || gitUrl.includes('/tree/')) {
                        // 提取仓库根URL
                        gitUrl = gitUrl.split('/blob/')[0].split('/tree/')[0];
                    }
                    
                    // 确保URL是指向仓库根目录的
                    const parts = gitUrl.split('/');
                    if (parts.length >= 5) {
                        gitUrl = parts[0] + '//' + parts[2] + '/' + parts[3] + '/' + parts[4];
                        
                        // 如果URL已经以.git结尾，不再添加.git
                        if (!gitUrl.endsWith('.git')) {
                            gitUrl += '.git';
                        }
                        
                        const acceleratedGitUrl = baseUrl + '/' + gitUrl;
                        generatedLinks.git = 'git clone ' + acceleratedGitUrl;
                    } else {
                        generatedLinks.git = '此链接不支持 git clone（URL格式无效）';
                    }
                }
            } else {
                generatedLinks.git = '此链接不支持 git clone（仅支持 GitHub/GitLab 仓库）';
            }
            
            // 更新当前显示的内容
            updateResultContent();
        }
        
        function generateLinks() {
            // 保持兼容性，直接调用实时生成函数
            generateLinksRealtime();
            
            // 滚动到结果区域
            document.getElementById('results').scrollIntoView({ behavior: 'smooth' });
        }
        
        function copyResult() {
            const text = generatedLinks[currentTab];
            
            navigator.clipboard.writeText(text).then(function() {
                showToast();
            }).catch(function(err) {
                // 降级方案
                const textArea = document.createElement('textarea');
                textArea.value = text;
                document.body.appendChild(textArea);
                textArea.select();
                document.execCommand('copy');
                document.body.removeChild(textArea);
                showToast();
            });
        }
        
        function showToast() {
            const toast = document.getElementById('toast');
            toast.style.display = 'block';
            setTimeout(function() {
                toast.style.display = 'none';
            }, 2000);
        }
        
        // 页面加载时的示例
        window.addEventListener('load', function() {
            // 可以在这里添加示例链接
            const examples = [
                'https://github.com/vansour/bbr/blob/main/bbr.sh',
                'https://gitlab.com/gitlab-org/gitlab/-/blob/master/README.md',
                'https://huggingface.co/microsoft/DialoGPT-medium/resolve/main/README.md'
            ];
            
            // 随机显示一个示例
            const randomExample = examples[Math.floor(Math.random() * examples.length)];
            document.getElementById('original-url').placeholder = '例如：' + randomExample;
        });
    </script>
</body>
</html>
		`)
		return
	}

	// ========== 第三步：URL格式验证和修复 ==========

	// 检查是否是有效的URL
	// 处理Go路由器自动清理双斜杠的问题
	// Go的HTTP路由器可能会将"https://"变成"https:/"，我们需要修复这个问题
	if strings.HasPrefix(requestPath, "https:/") && !strings.HasPrefix(requestPath, "https://") {
		requestPath = "https://" + strings.TrimPrefix(requestPath, "https:/")
		log.Printf("修复https URL: %s", requestPath)
	} else if strings.HasPrefix(requestPath, "http:/") && !strings.HasPrefix(requestPath, "http://") {
		requestPath = "http://" + strings.TrimPrefix(requestPath, "http:/")
		log.Printf("修复http URL: %s", requestPath)
	}

	// 额外处理：检查URL中是否有被错误清理的协议部分
	// 有时可能出现"https:/domain.com"这样的格式，需要修复为"https://domain.com"
	if strings.Contains(requestPath, ":/") && !strings.Contains(requestPath, "://") {
		// 查找协议部分并修复
		parts := strings.Split(requestPath, ":/")
		if len(parts) == 2 {
			protocol := parts[0]
			remainder := parts[1]
			// 只处理标准的HTTP/HTTPS协议
			if protocol == "https" || protocol == "http" {
				requestPath = protocol + "://" + remainder
				log.Printf("修复协议分隔符: %s", requestPath)
			}
		}
	}

	// 最终验证：确保URL格式正确
	// 如果还是没有正确的协议前缀，返回错误
	if !strings.HasPrefix(requestPath, "http://") && !strings.HasPrefix(requestPath, "https://") {
		http.Error(w, "无效的URL格式，请使用完整的URL", http.StatusBadRequest)
		return
	}

	// ========== 第四步：解析和转换目标URL ==========

	// 解析目标URL，将字符串转换为url.URL结构体
	// 这样可以方便地访问URL的各个部分（协议、域名、路径等）
	targetURL, err := url.Parse(requestPath)
	if err != nil {
		http.Error(w, "URL解析失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 处理URL转换（GitHub、GitLab、Hugging Face）
	// 不同平台有不同的URL格式，需要转换为可以直接下载的raw格式
	// 例如：GitHub的blob链接转换为raw.githubusercontent.com链接
	targetURL = convertURL(targetURL)

	// ========== 第五步：安全验证 ==========

	// 验证是否是支持的域名
	// 只允许代理已知的安全域名，防止被滥用为通用代理
	if !isSupportedDomain(targetURL.Host) {
		http.Error(w, "只支持GitHub、GitLab、Hugging Face相关域名", http.StatusForbidden)
		return
	}

	// 特殊验证Hugging Face文件下载
	if targetURL.Host == "huggingface.co" {
		if !strings.Contains(targetURL.Path, "/resolve/") && !strings.Contains(targetURL.Path, "/raw/") {
			http.Error(w, "Hugging Face 链接需要包含具体文件路径（/resolve/ 或 /raw/）", http.StatusBadRequest)
			return
		}
	}

	// 特殊验证GitHub - 仅支持文件下载，git clone应通过git命令使用
	if targetURL.Host == "github.com" {
		path := targetURL.Path
		// 只允许文件路径和gist，不允许直接访问仓库根路径
		isFilePath := strings.Contains(path, "/blob/") || strings.Contains(path, "/raw/") || strings.Contains(path, "/tree/")
		// 检查是否是gist
		isGist := strings.Contains(path, "/gist/")

		if !isFilePath && !isGist {
			http.Error(w, "GitHub 链接仅支持文件下载路径（/blob/, /raw/, /tree/）或gist，git clone请使用git命令", http.StatusBadRequest)
			return
		}
	}

	// ========== 第六步：平台特定验证 ==========

	// 特殊验证Hugging Face文件下载
	// Hugging Face有特定的URL格式要求，确保是文件下载而不是页面浏览
	if targetURL.Host == "huggingface.co" {
		if !strings.Contains(targetURL.Path, "/resolve/") && !strings.Contains(targetURL.Path, "/raw/") {
			http.Error(w, "Hugging Face 链接需要包含具体文件路径（/resolve/ 或 /raw/）", http.StatusBadRequest)
			return
		}
	}

	// 特殊验证GitHub - 仅支持文件下载，git clone应通过git命令使用
	// 防止用户通过浏览器代理访问整个仓库，只允许具体文件
	if targetURL.Host == "github.com" {
		path := targetURL.Path
		// 只允许文件路径和gist，不允许直接访问仓库根路径
		isFilePath := strings.Contains(path, "/blob/") || strings.Contains(path, "/raw/") || strings.Contains(path, "/tree/")
		// 检查是否是gist（GitHub代码片段）
		isGist := strings.Contains(path, "/gist/")

		if !isFilePath && !isGist {
			http.Error(w, "GitHub 链接仅支持文件下载路径（/blob/, /raw/, /tree/）或gist，git clone请使用git命令", http.StatusBadRequest)
			return
		}
	}

	// 特殊验证GitLab - 仅支持文件下载，git clone应通过git命令使用
	// 与GitHub类似，只允许文件下载，不允许仓库浏览
	if targetURL.Host == "gitlab.com" {
		path := targetURL.Path
		// 只允许文件路径，不允许直接访问仓库根路径
		// GitLab的URL格式：/-/blob/, /-/raw/, /-/tree/
		isFilePath := strings.Contains(path, "/-/blob/") || strings.Contains(path, "/-/raw/") || strings.Contains(path, "/-/tree/")

		if !isFilePath {
			http.Error(w, "GitLab 链接仅支持文件下载路径（/-/blob/, /-/raw/, /-/tree/），git clone请使用git命令", http.StatusBadRequest)
			return
		}
	}

	// 记录最终的目标URL
	log.Printf("目标URL: %s", targetURL.String())

	// ========== 第七步：创建HTTP客户端和请求 ==========

	// 创建HTTP客户端，自定义重定向策略
	// 这里配置了安全的重定向处理，防止被重定向到不安全的域名
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// 防止无限重定向攻击
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}

			// 检查重定向目标是否为支持的域名
			// 这是一个重要的安全措施，防止通过重定向访问内网或其他不安全的地址
			if !isSupportedDomain(req.URL.Host) {
				log.Printf("重定向到不支持的域名: %s", req.URL.Host)
				return fmt.Errorf("redirect to unsupported domain: %s", req.URL.Host)
			}

			// 记录重定向过程便于调试
			log.Printf("跟随重定向: %s -> %s", via[len(via)-1].URL.String(), req.URL.String())
			return nil
		},
	}

	// 创建HTTP请求
	// 复制原始请求的方法（GET/POST等）和请求体
	req, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		http.Error(w, "创建请求失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// ========== 第八步：设置请求头 ==========

	// 复制原始请求的头部，但排除一些代理相关的头部
	// 这些头部应该由代理服务器重新生成，而不是直接转发
	for key, values := range r.Header {
		// 排除这些头部：
		// - Host: 应该是目标服务器的域名
		// - X-Forwarded-For: 代理链信息，由代理服务器添加
		// - X-Real-Ip: 真实IP信息，由代理服务器添加
		if key != "Host" && key != "X-Forwarded-For" && key != "X-Real-Ip" {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}

	// 设置User-Agent，模拟Windows用户以获取正确的下载文件
	// 某些网站可能会根据User-Agent返回不同的内容或限制访问
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	// 添加更多浏览器头部来避免被检测为机器人
	// 这些头部让请求看起来更像是来自真实的浏览器
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")    // 接受的语言
	req.Header.Set("Accept-Encoding", "gzip, deflate, br") // 接受的编码格式
	req.Header.Set("DNT", "1")                             // Do Not Track请求
	req.Header.Set("Connection", "keep-alive")             // 保持连接
	req.Header.Set("Upgrade-Insecure-Requests", "1")       // 升级不安全请求
	// 现代浏览器的安全相关头部
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")

	// ========== 第九步：发送请求并获取响应 ==========

	// 发送HTTP请求到目标服务器
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "请求失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close() // 确保响应体被正确关闭

	// ========== 第十步：处理响应 ==========

	// 复制响应头到客户端
	// 将目标服务器的响应头转发给客户端，保持原始响应的完整性
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// 检查文件大小限制
	// 根据配置文件中的sizeLimit设置，拒绝过大的文件下载
	// 这可以防止服务器资源被耗尽，也可以避免用户下载超大文件
	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		if size, err := strconv.ParseInt(contentLength, 10, 64); err == nil {
			// 将配置中的MB转换为字节进行比较
			maxSize := int64(config.Server.SizeLimit * 1024 * 1024)
			if size > maxSize {
				// 如果文件大小超过限制，返回413错误（请求实体过大）
				http.Error(w, fmt.Sprintf("文件大小 %d MB 超过限制 %d MB", size/(1024*1024), config.Server.SizeLimit), http.StatusRequestEntityTooLarge)
				return
			}
			// 记录文件大小信息
			log.Printf("文件大小: %d MB", size/(1024*1024))
		}
	}

	// 设置HTTP状态码
	// 将目标服务器的状态码转发给客户端
	w.WriteHeader(resp.StatusCode)

	// ========== 第十一步：传输响应体 ==========

	// 复制响应体数据
	// 这是整个代理过程的核心：将目标服务器的响应数据流式传输给客户端
	// 使用io.Copy可以高效地处理大文件，不会将整个文件加载到内存中
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		// 记录传输错误，可能是网络中断或客户端断开连接
		log.Printf("复制响应体失败: %v", err)
	}

	// ========== 第十二步：记录访问日志 ==========

	// 记录完整的访问日志，包含客户端IP、原始请求、目标URL和响应状态
	// 这对于监控、调试和分析服务使用情况非常重要
	log.Printf("[%s] %s -> %s (Status: %d)",
		r.RemoteAddr,       // 客户端IP地址
		requestURI,         // 原始请求URI
		targetURL.String(), // 实际访问的目标URL
		resp.StatusCode)    // HTTP响应状态码
}

// ==================== API相关结构体 ====================

// GenerateLinksRequest API请求结构体
// 用于接收客户端发送的生成加速链接请求
type GenerateLinksRequest struct {
	OriginalURL string `json:"original_url"` // 原始URL（GitHub、GitLab、Hugging Face等）
}

// GenerateLinksResponse API响应结构体
// 用于返回生成的各种格式的加速链接给客户端
type GenerateLinksResponse struct {
	Success     bool   `json:"success"`         // 请求是否成功
	BrowserLink string `json:"browser_link"`    // 浏览器访问链接
	WgetCommand string `json:"wget_command"`    // wget下载命令
	CurlCommand string `json:"curl_command"`    // curl下载命令
	GitCommand  string `json:"git_command"`     // git clone命令
	Error       string `json:"error,omitempty"` // 错误信息（仅在失败时返回）
}

// ==================== API处理函数 ====================

// generateLinksAPI 生成加速链接的API处理函数
// 路径：/api/generate
// 方法：POST
//
// 功能说明：
// 1. 接收包含原始URL的JSON请求
// 2. 验证URL格式和平台支持
// 3. 生成各种格式的加速链接（浏览器、wget、curl、git）
// 4. 返回JSON格式的响应
//
// 这个API主要供Web界面的JavaScript调用，实现实时链接生成功能
func generateLinksAPI(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json") // 返回JSON格式
	// CORS设置，允许跨域访问（主要是为了支持前端JavaScript调用）
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// 处理预检请求（CORS）
	// 浏览器在发送跨域POST请求前会先发送OPTIONS请求
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 只接受POST请求
	if r.Method != "POST" {
		response := GenerateLinksResponse{
			Success: false,
			Error:   "只支持POST请求",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	var req GenerateLinksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := GenerateLinksResponse{
			Success: false,
			Error:   "请求格式错误",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	originalURL := strings.TrimSpace(req.OriginalURL)
	if originalURL == "" {
		response := GenerateLinksResponse{
			Success: false,
			Error:   "原始URL不能为空",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	if !strings.HasPrefix(originalURL, "http://") && !strings.HasPrefix(originalURL, "https://") {
		response := GenerateLinksResponse{
			Success: false,
			Error:   "请输入完整的URL（包含http://或https://）",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// 获取请求主机信息
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, r.Host)

	// 生成加速链接
	acceleratedURL := baseURL + "/" + originalURL

	// 特殊验证Hugging Face文件下载
	if strings.Contains(originalURL, "huggingface.co") {
		if !strings.Contains(originalURL, "/resolve/") && !strings.Contains(originalURL, "/blob/") {
			response := GenerateLinksResponse{
				Success: false,
				Error:   "Hugging Face 链接需要包含具体文件路径（/blob/ 或 /resolve/）",
			}
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	// 特殊验证GitHub - 仅支持文件下载和git clone
	if strings.Contains(originalURL, "github.com") {
		if u, err := url.Parse(originalURL); err == nil {
			path := u.Path
			// 检查是否是仓库根路径（用于git clone）- 格式应为 /user/repo 或 /user/repo/
			pathParts := strings.Split(strings.Trim(path, "/"), "/")
			isRepoRoot := len(pathParts) == 2 && pathParts[0] != "" && pathParts[1] != "" && !strings.Contains(path, ".")
			// 检查是否是文件路径
			isFilePath := strings.Contains(path, "/blob/") || strings.Contains(path, "/raw/") || strings.Contains(path, "/tree/")
			// 检查是否是gist
			isGist := strings.Contains(path, "/gist/")

			if !isRepoRoot && !isFilePath && !isGist {
				response := GenerateLinksResponse{
					Success: false,
					Error:   "GitHub 链接仅支持仓库根路径（git clone）或文件路径（/blob/, /raw/, /tree/）",
				}
				json.NewEncoder(w).Encode(response)
				return
			}
		}
	}

	// 特殊验证GitLab - 仅支持文件下载和git clone
	if strings.Contains(originalURL, "gitlab.com") {
		if u, err := url.Parse(originalURL); err == nil {
			path := u.Path
			// 检查是否是仓库根路径（用于git clone）- 格式应为 /user/repo 或 /user/repo/
			pathParts := strings.Split(strings.Trim(path, "/"), "/")
			isRepoRoot := len(pathParts) == 2 && pathParts[0] != "" && pathParts[1] != "" && !strings.Contains(path, ".")
			// 检查是否是文件路径
			isFilePath := strings.Contains(path, "/-/blob/") || strings.Contains(path, "/-/raw/") || strings.Contains(path, "/-/tree/")

			if !isRepoRoot && !isFilePath {
				response := GenerateLinksResponse{
					Success: false,
					Error:   "GitLab 链接仅支持仓库根路径（git clone）或文件路径（/-/blob/, /-/raw/, /-/tree/）",
				}
				json.NewEncoder(w).Encode(response)
				return
			}
		}
	}

	// 提取文件名
	fileName := "downloaded_file"
	if lastSlash := strings.LastIndex(originalURL, "/"); lastSlash != -1 {
		if lastSlash+1 < len(originalURL) {
			fileName = originalURL[lastSlash+1:]
		}
	}
	if fileName == "" || strings.Contains(fileName, "?") {
		fileName = "downloaded_file"
	}

	// 生成各种命令
	wgetCmd := fmt.Sprintf(`wget "%s"`, acceleratedURL)
	curlCmd := fmt.Sprintf(`curl -L "%s"`, acceleratedURL)

	// Git clone处理
	gitCmd := "此链接不支持 git clone（仅支持 GitHub/GitLab 仓库）"
	if strings.Contains(originalURL, "github.com") || strings.Contains(originalURL, "gitlab.com") {
		gitURL := originalURL

		// 检查是否是不支持git clone的链接类型
		if strings.Contains(gitURL, "/archive/") ||
			strings.Contains(gitURL, "/releases/") ||
			strings.Contains(gitURL, "/tarball/") ||
			strings.Contains(gitURL, "/zipball/") ||
			strings.Contains(gitURL, "/raw/") ||
			strings.Contains(gitURL, "/-/raw/") ||
			strings.Contains(gitURL, "/gist/") {
			gitCmd = "此链接不支持 git clone（archive/release/raw文件请使用浏览器或下载命令）"
		} else {
			// 处理仓库链接
			if strings.Contains(gitURL, "/blob/") || strings.Contains(gitURL, "/tree/") {
				gitURL = strings.Split(gitURL, "/blob/")[0]
				gitURL = strings.Split(gitURL, "/tree/")[0]
			}

			// 确保URL是指向仓库根目录的
			parts := strings.Split(gitURL, "/")
			if len(parts) >= 5 {
				// 保留 https://domain/user/repo 部分
				gitURL = strings.Join(parts[:5], "/")

				// 如果URL已经以.git结尾，不再添加.git
				if !strings.HasSuffix(gitURL, ".git") {
					gitURL += ".git"
				}

				acceleratedGitURL := baseURL + "/" + gitURL
				gitCmd = fmt.Sprintf("git clone %s", acceleratedGitURL)
			}
		}
	}

	response := GenerateLinksResponse{
		Success:     true,
		BrowserLink: acceleratedURL,
		WgetCommand: wgetCmd,
		CurlCommand: curlCmd,
		GitCommand:  gitCmd,
	}

	json.NewEncoder(w).Encode(response)
}

// ==================== 日志管理函数 ====================

// setupLogRotation 设置日志轮转功能
//
// 功能说明：
// 1. 根据配置文件设置日志文件路径和大小限制
// 2. 自动创建日志目录（如果不存在）
// 3. 检查当前日志文件大小，超过限制时自动备份
// 4. 设置日志同时输出到文件和控制台
// 5. 通过时间戳命名备份文件，便于管理
//
// 日志轮转策略：
// - 当日志文件超过配置的最大大小时，自动重命名为 "原文件名.时间戳"
// - 创建新的日志文件继续记录
// - 这样可以防止单个日志文件过大，便于日志分析和管理
func setupLogRotation() {
	// 将配置中的MB转换为字节数
	maxLogSize := int64(config.Log.MaxLogSize * 1024 * 1024)

	// 使用配置中的日志路径
	logPath := config.Log.LogFilePath

	// 确保日志目录存在
	// 从完整路径中提取目录部分
	logDir := strings.TrimSuffix(logPath, "/ghproxy.log")
	if logDir == "" {
		// 如果提取失败，使用默认目录
		logDir = "/data/ghproxy/log"
	}
	// 创建目录，权限755（所有者可读写执行，组和其他用户可读执行）
	os.MkdirAll(logDir, 0755)

	// 检查日志文件大小，实现日志轮转
	if info, err := os.Stat(logPath); err == nil {
		if info.Size() > maxLogSize {
			// 备份当前日志文件
			// 使用Unix时间戳作为后缀，确保文件名唯一
			backupPath := fmt.Sprintf("%s.%d", logPath, time.Now().Unix())
			os.Rename(logPath, backupPath)
			log.Printf("日志文件已备份为: %s", backupPath)
		}
	}

	// 设置日志输出到文件
	// 使用追加模式打开文件，如果文件不存在则创建
	// 权限644（所有者可读写，组和其他用户可读）
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("无法创建日志文件: %v", err)
		return
	}

	// 设置日志同时输出到文件和控制台
	// 这样既可以实时查看日志，又可以持久化保存
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	log.Printf("日志文件设置为: %s", logPath)
}

// ==================== 主函数 ====================

// main 程序入口函数
//
// 执行流程：
// 1. 加载配置文件（支持通过命令行参数指定）
// 2. 设置日志系统
// 3. 打印启动信息
// 4. 创建和配置HTTP服务器
// 5. 启动服务器并开始监听请求
//
// 命令行用法：
//
//	./ghproxy                  # 使用默认配置文件 config.toml
//	./ghproxy custom.toml      # 使用指定的配置文件
func main() {
	// ========== 第一步：配置初始化 ==========

	// 确定配置文件路径
	// 默认使用当前目录下的 config.toml
	// 如果提供了命令行参数，则使用指定的配置文件
	configPath := "config.toml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	// 加载配置文件
	// 如果加载失败，程序会终止并显示错误信息
	if err := loadConfig(configPath); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// ========== 第1.5步：生成配置相关文件 ==========

	// 自动生成配置相关的目录和示例文件
	// 包括日志目录、配置目录、黑名单白名单示例等
	if err := generateConfigFiles(); err != nil {
		log.Printf("警告：生成配置文件时出现错误: %v", err)
		// 不终止程序，继续运行
	}

	// ========== 第二步：日志系统初始化 ==========

	// 设置日志轮转，确保日志文件不会无限增长
	setupLogRotation()

	// ========== 第三步：显示启动信息 ==========

	// 打印服务器配置信息
	fmt.Printf("Git文件加速代理\n")
	fmt.Printf("监听地址: %s:%d\n", config.Server.Host, config.Server.Port)
	fmt.Printf("文件大小限制: %d MB\n", config.Server.SizeLimit)
	fmt.Printf("支持平台: GitHub, GitLab, Hugging Face\n")
	fmt.Printf("Web界面: http://%s:%d\n", config.Server.Host, config.Server.Port)
	fmt.Printf("=" + strings.Repeat("=", 50) + "\n")

	// ========== 第四步：创建HTTP服务器 ==========

	// 创建服务器监听地址
	serverAddr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)

	// 创建自定义的HTTP服务器
	// 使用自定义的处理器来避免Go标准库的路径清理问题
	// 这对代理服务器很重要，因为我们需要保持URL的原始格式
	server := &http.Server{
		Addr: serverAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 路由分发：根据请求路径选择不同的处理器

			// API路由：处理生成加速链接的API请求
			// 路径：/api/generate
			if strings.HasPrefix(r.URL.Path, "/api/generate") {
				generateLinksAPI(w, r)
				return
			}

			// 代理路由：处理所有其他请求（文件代理下载）
			// 这是服务器的核心功能，代理访问GitHub、GitLab等平台的文件
			proxyHandler(w, r)
		}),
	}

	// ========== 第五步：启动服务器 ==========

	// 打印启动成功信息
	fmt.Printf("Git文件加速代理启动成功！\n")
	fmt.Printf("使用方法: http://%s:%d/完整的文件URL\n", config.Server.Host, config.Server.Port)

	// 启动HTTP服务器并开始监听请求
	// 这是一个阻塞调用，程序会一直运行直到服务器停止或出现致命错误
	// 如果启动失败（如端口被占用），log.Fatal会终止程序并输出错误信息
	log.Fatal(server.ListenAndServe())
}

// ==================== URL转换函数 ====================

// convertURL 转换各种平台的URL为raw格式
//
// 参数：
//
//	u: 需要转换的URL对象
//
// 返回值：
//
//	*url.URL: 转换后的URL对象
//
// 功能说明：
// 不同的代码托管平台有不同的URL格式：
// - GitHub: 需要将blob链接转换为raw.githubusercontent.com
// - GitLab: 需要将blob链接转换为raw链接
// - Hugging Face: 需要将blob链接转换为resolve链接
//
// 这样转换后的URL可以直接下载文件内容，而不是显示网页
func convertURL(u *url.URL) *url.URL {
	switch u.Host {
	case "github.com":
		return convertGitHubURL(u)
	case "gitlab.com":
		return convertGitLabURL(u)
	case "huggingface.co":
		return convertHuggingFaceURL(u)
	}
	return u
}

// convertGitHubURL 转换GitHub URL为raw格式
//
// 参数：
//
//	u: GitHub的URL对象
//
// 返回值：
//
//	*url.URL: 转换后的URL对象
//
// 转换规则：
// 1. 将github.com的blob链接转换为raw.githubusercontent.com
// 2. 移除路径中的"/blob/"部分
// 3. 保持其他类型的路径不变（如仓库根路径、tree路径等）
//
// 示例转换：
//
//	输入: https://github.com/user/repo/blob/main/file.txt
//	输出: https://raw.githubusercontent.com/user/repo/main/file.txt
//
// 这样转换后的URL可以直接下载文件内容
func convertGitHubURL(u *url.URL) *url.URL {
	if u.Host == "github.com" {
		path := u.Path
		// 只转换blob链接为raw格式，保持其他路径不变
		if strings.Contains(path, "/blob/") {
			// 例: /user/repo/blob/branch/file -> /user/repo/branch/file
			newPath := strings.Replace(path, "/blob/", "/", 1)
			u.Host = "raw.githubusercontent.com"
			u.Path = newPath
		}
		// 对于仓库根路径、tree路径等，保持原样以支持git clone
	}
	return u
}

// convertGitLabURL 转换GitLab URL为raw格式
//
// 参数：
//
//	u: GitLab的URL对象
//
// 返回值：
//
//	*url.URL: 转换后的URL对象
//
// 转换规则：
// 1. 将gitlab.com的blob链接转换为raw链接
// 2. 将路径中的"/-/blob/"替换为"/-/raw/"
// 3. 保持其他类型的路径不变
//
// 示例转换：
//
//	输入: https://gitlab.com/user/repo/-/blob/main/file.txt
//	输出: https://gitlab.com/user/repo/-/raw/main/file.txt
func convertGitLabURL(u *url.URL) *url.URL {
	if u.Host == "gitlab.com" {
		path := u.Path
		// 只转换blob链接为raw链接，保持其他路径不变
		if strings.Contains(path, "/-/blob/") {
			// 例: /user/repo/-/blob/branch/file -> /user/repo/-/raw/branch/file
			newPath := strings.Replace(path, "/-/blob/", "/-/raw/", 1)
			u.Path = newPath
		}
		// 对于仓库根路径、tree路径等，保持原样以支持git clone
	}
	return u
}

// 转换Hugging Face URL为resolve格式
func convertHuggingFaceURL(u *url.URL) *url.URL {
	if u.Host == "huggingface.co" {
		path := u.Path
		// 将huggingface.co的blob链接转换为resolve链接
		if strings.Contains(path, "/blob/") {
			// 例: /model/blob/main/file -> /model/resolve/main/file
			newPath := strings.Replace(path, "/blob/", "/resolve/", 1)
			u.Path = newPath
		}
		// 确保路径包含文件下载相关的路径
		if !strings.Contains(path, "/resolve/") && !strings.Contains(path, "/raw/") {
			// 对于没有resolve的路径，检查是否为文件下载路径
			parts := strings.Split(strings.Trim(path, "/"), "/")
			if len(parts) >= 3 {
				// 格式应为: /model/main/file 或 /datasets/dataset/main/file
				// 在模型名和分支之间插入resolve
				if parts[0] == "datasets" && len(parts) >= 4 {
					// 数据集格式: /datasets/dataset/resolve/main/file
					newParts := []string{parts[0], parts[1], "resolve"}
					newParts = append(newParts, parts[2:]...)
					u.Path = "/" + strings.Join(newParts, "/")
				} else {
					// 模型格式: /model/resolve/main/file
					newParts := []string{parts[0], "resolve"}
					newParts = append(newParts, parts[1:]...)
					u.Path = "/" + strings.Join(newParts, "/")
				}
			}
		}
	}
	return u
}

// ==================== 安全验证函数 ====================

// isSupportedDomain 检查是否是支持的代码托管平台域名
//
// 参数：
//
//	host: 需要检查的域名
//
// 返回值：
//
//	bool: 如果域名被支持返回true，否则返回false
//
// 功能说明：
// 这是一个重要的安全函数，用于防止代理服务器被滥用为通用代理。
// 只有在白名单中的域名才会被允许代理访问。
//
// 支持的平台和域名：
// 1. GitHub相关：
//   - github.com: 主站
//   - raw.githubusercontent.com: 原始文件服务
//   - gist.githubusercontent.com: Gist文件服务
//   - codeload.github.com: 下载服务
//   - api.github.com: API服务
//
// 2. GitLab相关：
//   - gitlab.com: 主站
//   - gitlab.io: GitLab Pages
//
// 3. Hugging Face相关：
//   - huggingface.co: 主站
//   - hf.co: 短域名
//   - cdn-lfs.huggingface.co: LFS CDN
//   - cas-bridge.xethub.hf.co: CDN桥接服务
//   - cdn-lfs.hf.co: LFS CDN短域名
func isSupportedDomain(host string) bool {
	// 定义所有允许的域名白名单
	allowedDomains := []string{
		// GitHub相关域名
		"raw.githubusercontent.com",
		"github.com",
		"gist.githubusercontent.com",
		"codeload.github.com",
		"api.github.com",

		// GitLab相关域名
		"gitlab.com",
		"gitlab.io",

		// Hugging Face相关域名
		"huggingface.co",
		"hf.co",                   // Hugging Face短域名
		"cdn-lfs.huggingface.co",  // Hugging Face LFS CDN
		"cas-bridge.xethub.hf.co", // Hugging Face CDN桥接
		"cdn-lfs.hf.co",           // Hugging Face LFS CDN短域名
	}

	// 检查域名是否在白名单中
	for _, domain := range allowedDomains {
		if host == domain {
			return true
		}
	}

	// 域名不在白名单中，返回false
	return false
}
