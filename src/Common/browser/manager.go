package browser

import (
	clog "BotMatrix/common/log"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
	"go.uber.org/zap"
)

// BrowserManager 管理 Playwright 浏览器实例
type BrowserManager struct {
	pw           *playwright.Playwright
	browser      playwright.Browser
	context      playwright.BrowserContext
	mu           sync.RWMutex
	headless     bool
	downloadPath string
}

// NewBrowserManager 创建一个新的浏览器管理器
func NewBrowserManager(headless bool, downloadPath string) (*BrowserManager, error) {
	return &BrowserManager{
		headless:     headless,
		downloadPath: downloadPath,
	}, nil
}

// Start 启动 Playwright 和浏览器
func (m *BrowserManager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.pw != nil {
		return nil // 已经启动
	}

	// 1. 启动 Playwright
	// 如果需要自动安装驱动，可能需要执行 playwright.Install()，但这通常是 setup 阶段做的
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("could not start playwright: %v", err)
	}
	m.pw = pw

	// 2. 启动浏览器 (Chromium)
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(m.headless),
		Args: []string{
			"--no-sandbox",
			"--disable-setuid-sandbox",
			"--disable-dev-shm-usage",
		},
	})
	if err != nil {
		// 尝试停止 pw
		_ = pw.Stop()
		return fmt.Errorf("could not launch browser: %v", err)
	}
	m.browser = browser

	// 3. 创建默认上下文
	ctx, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		Viewport: &playwright.Size{
			Width:  1280,
			Height: 800,
		},
	})
	if err != nil {
		_ = browser.Close()
		_ = pw.Stop()
		return fmt.Errorf("could not create browser context: %v", err)
	}
	m.context = ctx

	clog.Info("BrowserManager started successfully", zap.Bool("headless", m.headless))
	return nil
}

// Stop 停止浏览器和 Playwright
func (m *BrowserManager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.context != nil {
		if err := m.context.Close(); err != nil {
			clog.Error("Failed to close browser context", zap.Error(err))
		}
		m.context = nil
	}

	if m.browser != nil {
		if err := m.browser.Close(); err != nil {
			clog.Error("Failed to close browser", zap.Error(err))
		}
		m.browser = nil
	}

	if m.pw != nil {
		if err := m.pw.Stop(); err != nil {
			clog.Error("Failed to stop playwright", zap.Error(err))
		}
		m.pw = nil
	}
	return nil
}

// EnsureStarted 确保浏览器已启动
func (m *BrowserManager) EnsureStarted() error {
	m.mu.RLock()
	started := m.pw != nil
	m.mu.RUnlock()

	if !started {
		return m.Start()
	}
	return nil
}

// Navigate 访问 URL 并返回页面内容
func (m *BrowserManager) Navigate(ctx context.Context, url string) (string, error) {
	if err := m.EnsureStarted(); err != nil {
		return "", err
	}

	m.mu.Lock()
	page, err := m.context.NewPage()
	m.mu.Unlock()
	if err != nil {
		return "", fmt.Errorf("could not create page: %v", err)
	}
	defer page.Close()

	// 设置超时
	timeout := 30000.0 // 30s
	_, err = page.Goto(url, playwright.PageGotoOptions{
		Timeout:   playwright.Float(timeout),
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		return "", fmt.Errorf("could not goto url %s: %v", url, err)
	}

	// 等待一小会儿确保动态内容加载（可选，更稳健的做法是 WaitSelector）
	time.Sleep(2 * time.Second)

	// 获取文本内容
	// 简单策略：获取 body 的 innerText
	content, err := page.Locator("body").InnerText()
	if err != nil {
		return "", fmt.Errorf("could not get page content: %v", err)
	}

	return content, nil
}

// Screenshot 截图
func (m *BrowserManager) Screenshot(ctx context.Context, url string) ([]byte, error) {
	if err := m.EnsureStarted(); err != nil {
		return nil, err
	}

	m.mu.Lock()
	page, err := m.context.NewPage()
	m.mu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("could not create page: %v", err)
	}
	defer page.Close()

	_, err = page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		return nil, fmt.Errorf("could not goto url %s: %v", url, err)
	}

	bytes, err := page.Screenshot(playwright.PageScreenshotOptions{
		FullPage: playwright.Bool(true),
		Type:     playwright.ScreenshotTypePng,
	})
	if err != nil {
		return nil, fmt.Errorf("could not take screenshot: %v", err)
	}

	return bytes, nil
}

// ExtractContent 提取特定选择器的内容
func (m *BrowserManager) ExtractContent(ctx context.Context, url string, selector string) (string, error) {
	if err := m.EnsureStarted(); err != nil {
		return "", err
	}

	m.mu.Lock()
	page, err := m.context.NewPage()
	m.mu.Unlock()
	if err != nil {
		return "", fmt.Errorf("could not create page: %v", err)
	}
	defer page.Close()

	_, err = page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		return "", fmt.Errorf("could not goto url %s: %v", url, err)
	}

	content, err := page.Locator(selector).InnerText()
	if err != nil {
		return "", fmt.Errorf("could not get content for selector %s: %v", selector, err)
	}

	return content, nil
}
