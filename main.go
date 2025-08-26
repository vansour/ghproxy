package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// 版本信息，通过构建时的ldflags设置
var (
	Version   = "2025.08.26.0551-test"
	BuildTime = "2025-08-26 05:51:08 UTC"
)

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	// 直接从RequestURI获取完整路径，这样可以避免Go的路径清理
	requestURI := r.RequestURI

	// 去掉开头的 "/"
	requestPath := strings.TrimPrefix(requestURI, "/")

	// 添加调试日志
	log.Printf("收到请求: %s", requestURI)
	log.Printf("处理路径: %s", requestPath)

	// 处理URL解码问题
	if decodedPath, err := url.QueryUnescape(requestPath); err == nil {
		requestPath = decodedPath
		log.Printf("解码后路径: %s", requestPath)
	}

	// 如果是根路径或空路径，返回使用说明
	if requestPath == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Git代码文件加速代理服务</title>
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
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 Git代码文件加速代理</h1>
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

	// 检查是否是有效的URL
	// 处理Go路由器自动清理双斜杠的问题
	if strings.HasPrefix(requestPath, "https:/") && !strings.HasPrefix(requestPath, "https://") {
		requestPath = "https://" + strings.TrimPrefix(requestPath, "https:/")
		log.Printf("修复https URL: %s", requestPath)
	} else if strings.HasPrefix(requestPath, "http:/") && !strings.HasPrefix(requestPath, "http://") {
		requestPath = "http://" + strings.TrimPrefix(requestPath, "http:/")
		log.Printf("修复http URL: %s", requestPath)
	}

	// 额外处理：检查URL中是否有被错误清理的协议部分
	if strings.Contains(requestPath, ":/") && !strings.Contains(requestPath, "://") {
		// 查找协议部分并修复
		parts := strings.Split(requestPath, ":/")
		if len(parts) == 2 {
			protocol := parts[0]
			remainder := parts[1]
			if protocol == "https" || protocol == "http" {
				requestPath = protocol + "://" + remainder
				log.Printf("修复协议分隔符: %s", requestPath)
			}
		}
	}

	if !strings.HasPrefix(requestPath, "http://") && !strings.HasPrefix(requestPath, "https://") {
		http.Error(w, "无效的URL格式，请使用完整的URL", http.StatusBadRequest)
		return
	}

	// 解析目标URL
	targetURL, err := url.Parse(requestPath)
	if err != nil {
		http.Error(w, "URL解析失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 处理URL转换（GitHub、GitLab、Hugging Face）
	targetURL = convertURL(targetURL)

	// 验证是否是支持的域名
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

	// 特殊验证GitLab - 仅支持文件下载，git clone应通过git命令使用
	if targetURL.Host == "gitlab.com" {
		path := targetURL.Path
		// 只允许文件路径，不允许直接访问仓库根路径
		isFilePath := strings.Contains(path, "/-/blob/") || strings.Contains(path, "/-/raw/") || strings.Contains(path, "/-/tree/")

		if !isFilePath {
			http.Error(w, "GitLab 链接仅支持文件下载路径（/-/blob/, /-/raw/, /-/tree/），git clone请使用git命令", http.StatusBadRequest)
			return
		}
	}

	log.Printf("目标URL: %s", targetURL.String())

	// 创建HTTP客户端，自定义重定向策略
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// 允许跟随重定向，但需要检查重定向目标域名
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}

			// 检查重定向目标是否为支持的域名
			if !isSupportedDomain(req.URL.Host) {
				log.Printf("重定向到不支持的域名: %s", req.URL.Host)
				return fmt.Errorf("redirect to unsupported domain: %s", req.URL.Host)
			}

			log.Printf("跟随重定向: %s -> %s", via[len(via)-1].URL.String(), req.URL.String())
			return nil
		},
	} // 创建请求
	req, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		http.Error(w, "创建请求失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 复制原始请求的头部，但排除一些不需要的
	for key, values := range r.Header {
		if key != "Host" && key != "X-Forwarded-For" && key != "X-Real-Ip" {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}

	// 设置User-Agent，模拟Windows用户以获取正确的下载文件
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	// 添加更多浏览器头部来避免被检测为机器人
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "请求失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// 设置状态码
	w.WriteHeader(resp.StatusCode)

	// 复制响应体
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("复制响应体失败: %v", err)
	}

	// 记录访问日志
	log.Printf("[%s] %s -> %s (Status: %d)",
		r.RemoteAddr,
		requestURI,
		targetURL.String(),
		resp.StatusCode)
}

// API结构体
type GenerateLinksRequest struct {
	OriginalURL string `json:"original_url"`
}

type GenerateLinksResponse struct {
	Success     bool   `json:"success"`
	BrowserLink string `json:"browser_link"`
	WgetCommand string `json:"wget_command"`
	CurlCommand string `json:"curl_command"`
	GitCommand  string `json:"git_command"`
	Error       string `json:"error,omitempty"`
}

// API处理函数
func generateLinksAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

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

// 设置日志轮转功能
func setupLogRotation() {
	const maxLogSize = 5 * 1024 * 1024 // 5MB

	// 根据环境选择日志路径
	var logPath string
	if _, err := os.Stat("/app/logs"); err == nil {
		// Docker环境
		logPath = "/app/logs/server.log"
		os.MkdirAll("/app/logs", 0755)
	} else {
		// 系统环境
		logPath = "/var/log/ghproxy/server.log"
		os.MkdirAll("/var/log/ghproxy", 0755)
	}

	// 检查日志文件大小
	if info, err := os.Stat(logPath); err == nil {
		if info.Size() > maxLogSize {
			// 备份当前日志
			os.Rename(logPath, fmt.Sprintf("%s.%d", logPath, time.Now().Unix()))
		}
	}

	// 设置日志输出到文件
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("无法创建日志文件: %v", err)
		return
	}

	// 设置日志输出到文件和控制台
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
}

func main() {
	// 设置日志轮转 - 限制为5MB
	setupLogRotation()

	// 打印版本信息
	fmt.Printf("Git代码文件加速代理服务 v%s\n", Version)
	fmt.Printf("构建时间: %s\n", BuildTime)
	fmt.Printf("监听端口: :8080\n")
	fmt.Printf("支持平台: GitHub, GitLab, Hugging Face\n")
	fmt.Printf("Web界面: http://127.0.0.1:8080\n")
	fmt.Printf("=" + strings.Repeat("=", 50) + "\n")

	// 创建自定义的处理器来避免Go的路径清理问题
	server := &http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 特殊处理API路由
			if strings.HasPrefix(r.URL.Path, "/api/generate") {
				generateLinksAPI(w, r)
				return
			}
			// 所有其他请求都走代理处理器
			proxyHandler(w, r)
		}),
	}

	fmt.Printf("Git代码文件代理服务启动成功！\n")
	log.Printf("服务版本: %s, 构建时间: %s", Version, BuildTime)
	fmt.Printf("使用方法: http://127.0.0.1:8080/完整的文件URL\n")

	log.Fatal(server.ListenAndServe())
}

// 转换各种平台的URL为raw格式
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

// 转换GitHub URL为raw格式
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

// 转换GitLab URL为raw格式
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

// 检查是否是支持的代码托管平台域名
func isSupportedDomain(host string) bool {
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

	for _, domain := range allowedDomains {
		if host == domain {
			return true
		}
	}
	return false
}
