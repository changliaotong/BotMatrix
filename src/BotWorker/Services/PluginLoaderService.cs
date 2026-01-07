using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Reflection;
using System.Runtime.Loader;
using System.Threading.Tasks;
using BotWorker.Core.Plugin;
using Microsoft.Extensions.Logging;

namespace BotWorker.Services
{
    public interface IPluginLoaderService
    {
        Task LoadAllPluginsAsync();
        Task ReloadPluginsAsync();
    }

    public class PluginLoaderService : IPluginLoaderService
    {
        private readonly PluginManager _pluginManager;
        private readonly ILogger<PluginLoaderService> _logger;
        private readonly string _pluginsDir;
        private FileSystemWatcher? _watcher;

        public PluginLoaderService(
            PluginManager pluginManager,
            ILogger<PluginLoaderService> logger)
        {
            _pluginManager = pluginManager;
            _logger = logger;
            _pluginsDir = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "plugins");
            
            if (!Directory.Exists(_pluginsDir))
            {
                Directory.CreateDirectory(_pluginsDir);
            }

            SetupWatcher();
        }

        private void SetupWatcher()
        {
            _watcher = new FileSystemWatcher(_pluginsDir, "*.dll")
            {
                EnableRaisingEvents = true,
                NotifyFilter = NotifyFilters.LastWrite | NotifyFilters.FileName
            };

            // 使用去抖动处理，避免短时间内多次触发
            DateTime lastWriteTime = DateTime.MinValue;
            _watcher.Changed += async (s, e) => 
            {
                if (DateTime.Now - lastWriteTime < TimeSpan.FromSeconds(1)) return;
                lastWriteTime = DateTime.Now;
                _logger.LogInformation("检测到插件变化: {Path}, 正在重载...", e.FullPath);
                await ReloadPluginsAsync();
            };
        }

        public async Task LoadAllPluginsAsync()
        {
            _logger.LogInformation("开始从目录加载插件: {Dir}", _pluginsDir);
            
            if (!Directory.Exists(_pluginsDir)) return;

            foreach (var dll in Directory.GetFiles(_pluginsDir, "*.dll", SearchOption.AllDirectories))
            {
                await LoadPluginFileAsync(dll);
            }
        }

        private async Task LoadPluginFileAsync(string path)
        {
            try
            {
                // 使用独立的 AssemblyLoadContext 以支持热重载（卸载）
                var alc = new AssemblyLoadContext(Path.GetFileNameWithoutExtension(path), true);
                
                using var fs = File.OpenRead(path);
                var assembly = alc.LoadFromStream(fs);

                var pluginTypes = assembly.GetTypes()
                    .Where(t => typeof(IPlugin).IsAssignableFrom(t) && !t.IsInterface && !t.IsAbstract);

                foreach (var type in pluginTypes)
                {
                    try
                    {
                        if (Activator.CreateInstance(type) is IPlugin plugin)
                        {
                            _logger.LogInformation("加载插件: {Name} ({Description})", plugin.Name, plugin.Description);
                            await _pluginManager.LoadPluginAsync(plugin);
                        }
                    }
                    catch (Exception ex)
                    {
                        _logger.LogError(ex, "实例化插件类失败: {Type}", type.FullName);
                    }
                }
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "加载插件文件失败: {Path}", path);
            }
        }

        public async Task ReloadPluginsAsync()
        {
            _logger.LogInformation("正在重载所有插件...");
            // 在实际复杂的系统中，这里需要先停止并卸载旧的插件
            // 目前简单实现为再次加载新发现的插件
            await LoadAllPluginsAsync();
        }
    }
}
