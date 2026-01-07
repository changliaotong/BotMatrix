using System.Runtime.Loader;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using sz84.Core.Interfaces;

namespace sz84.Bots.Plugins
{
    public class ModuleLoader
    {
        private readonly List<IBotModule> _modules = [];

        public IReadOnlyList<IBotModule> Modules => _modules;

        public static IEnumerable<IBotModule> LoadFromFolder(string pluginFolder, IEnumerable<string>? disabledModules = null)
        {
            disabledModules ??= [];
            var modules = new List<IBotModule>();

            if (!Directory.Exists(pluginFolder))
                return modules;

            foreach (var dll in Directory.GetFiles(pluginFolder, "*.dll"))
            {
                try
                {
                    var alc = new AssemblyLoadContext(Path.GetFileNameWithoutExtension(dll), true);
                    using var fs = File.OpenRead(dll);
                    var assembly = alc.LoadFromStream(fs);

                    var types = assembly.GetTypes().Where(t => typeof(IBotModule).IsAssignableFrom(t) && !t.IsInterface && !t.IsAbstract);
                    foreach (var type in types)
                    {
                        var instance = (IBotModule)Activator.CreateInstance(type)!;
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
                module.RegisterServices(services, configuration);
            }
        }
    }

}
