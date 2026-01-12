using System.Runtime.Loader;

namespace BotWorker.Application.Services
{
    public class ModuleLoader
    {
        private readonly List<IPlugin> _modules = [];

        public IReadOnlyList<IPlugin> Modules => _modules;

        public static IEnumerable<IPlugin> LoadFromFolder(string pluginFolder, IEnumerable<string>? disabledModules = null)
        {
            disabledModules ??= [];
            var modules = new List<IPlugin>();

            if (!Directory.Exists(pluginFolder))
                return modules;

            foreach (var dll in Directory.GetFiles(pluginFolder, "*.dll"))
            {
                try
                {
                    var alc = new AssemblyLoadContext(Path.GetFileNameWithoutExtension(dll), true);
                    using var fs = File.OpenRead(dll);
                    var assembly = alc.LoadFromStream(fs);

                    var types = assembly.GetTypes().Where(t => typeof(IPlugin).IsAssignableFrom(t) && !t.IsInterface && !t.IsAbstract);
                    foreach (var type in types)
                    {
                        var instance = (IPlugin)Activator.CreateInstance(type)!;
                        if (disabledModules.Contains(instance.Metadata.Name, StringComparer.OrdinalIgnoreCase))
                            continue;

                        modules.Add(instance);
                        Console.WriteLine($"加载外部插件模块: {instance.Metadata.Name}");
                    }
                }
                catch (Exception ex)
                {
                    Console.WriteLine($"加载插件失败 {dll}: {ex.Message}");
                }
            }

            return modules;
        }

        public void ConfigureServices(IServiceCollection services, IConfiguration configuration)
        {
            foreach (var module in _modules)
            {
                // 注意：这里需要 IRobot 实例，或者我们将 RegisterServices 移出接口
                // 暂时注释掉，等待接口对齐
                // module.InitAsync(robot); 
            }
        }
    }
}


