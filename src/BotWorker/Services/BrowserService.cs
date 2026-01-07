using System;
using System.Threading.Tasks;
using Microsoft.Playwright;

namespace BotWorker.Services
{
    public interface IBrowserService
    {
        Task<string> NavigateAsync(string url);
        Task<string> TakeScreenshotAsync(string url);
    }

    public class BrowserService : IBrowserService, IAsyncDisposable
    {
        private IPlaywright? _playwright;
        private IBrowser? _browser;
        private IBrowserContext? _context;
        private bool _isStarted;

        public async Task StartAsync()
        {
            if (_isStarted) return;
            _playwright = await Playwright.CreateAsync();
            _browser = await _playwright.Chromium.LaunchAsync(new BrowserTypeLaunchOptions { Headless = true });
            _context = await _browser.NewContextAsync();
            _isStarted = true;
        }

        public async Task<string> NavigateAsync(string url)
        {
            await StartAsync();
            var page = await _context!.NewPageAsync();
            try
            {
                await page.GotoAsync(url);
                return await page.InnerTextAsync("body");
            }
            finally
            {
                await page.CloseAsync();
            }
        }

        public async Task<string> TakeScreenshotAsync(string url)
        {
            await StartAsync();
            var page = await _context!.NewPageAsync();
            try
            {
                await page.GotoAsync(url);
                var bytes = await page.ScreenshotAsync(new PageScreenshotOptions { FullPage = true });
                return Convert.ToBase64String(bytes);
            }
            finally
            {
                await page.CloseAsync();
            }
        }

        public async ValueTask DisposeAsync()
        {
            if (_context != null) await _context.DisposeAsync();
            if (_browser != null) await _browser.DisposeAsync();
            _playwright?.Dispose();
        }
    }
}

