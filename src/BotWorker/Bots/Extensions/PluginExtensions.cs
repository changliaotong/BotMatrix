using System.Reflection;
using System.Runtime.Loader;
using System.Text;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using sz84.Core.Interfaces;

namespace sz84.Bots.Extensions
{
    public static class PluginExtensions
    {
        private static readonly Dictionary<string, IBotModule> LoadedModules = new(StringComparer.OrdinalIgnoreCase);
        private static readonly Dictionary<string, List<string>> DependencyGraph = new();

        public static IServiceCollection AddGameModules(this IServiceCollection services, IConfiguration config)
        {
            var enabled = config.GetSection("GameModules:EnabledModules").Get<List<string>>() ?? new();

            // 加载本地程序集模块 + 外部插件 DLL 模块
            var modules = LoadModulesFromAppDomain()
                .Concat(LoadModulesFromPluginFolder("plugins"))
                .ToList();

            // 建图、检测依赖
            foreach (var module in modules)
            {
                var name = module.Metadata.Name;
                LoadedModules[name] = module;
                DependencyGraph[name] = module.Metadata.RequiredModules.ToList();
            }

            DetectCycles();

            var resolved = new HashSet<string>();
            foreach (var name in enabled)
                ResolveModule(name, services, resolved, new Stack<string>());

            ExportDependencyGraphToDot("module_graph.dot");

            return services;
        }

        private static void ResolveModule(string name, IServiceCollection services, HashSet<string> resolved, Stack<string> stack)
        {
            if (resolved.Contains(name)) return;
            if (!LoadedModules.ContainsKey(name))
                throw new Exception($"模块 {name} 未找到！");

            stack.Push(name);
            foreach (var dep in LoadedModules[name].Metadata.RequiredModules)
                ResolveModule(dep, services, resolved, stack);
            stack.Pop();

            LoadedModules[name].RegisterServices(services, null!);
            resolved.Add(name);
            Console.WriteLine($"✅ 注册模块: {name}");
        }

        private static void DetectCycles()
        {
            var visited = new HashSet<string>();
            var recStack = new HashSet<string>();

            bool Visit(string node)
            {
                if (!visited.Add(node)) return false;
                if (!DependencyGraph.ContainsKey(node)) return false;

                recStack.Add(node);
                foreach (var neighbor in DependencyGraph[node])
                {
                    if (recStack.Contains(neighbor) || Visit(neighbor))
                        throw new Exception($"❌ 循环依赖：{node} → {neighbor}");
                }
                recStack.Remove(node);
                return false;
            }

            foreach (var node in DependencyGraph.Keys)
                Visit(node);
        }

        private static void ExportDependencyGraphToDot(string filePath)
        {
            var sb = new StringBuilder("digraph G {\n");
            foreach (var kv in DependencyGraph)
                foreach (var dep in kv.Value)
                    sb.AppendLine($"  \"{kv.Key}\" -> \"{dep}\";");
            sb.AppendLine("}");
            File.WriteAllText(filePath, sb.ToString());
        }

        private static IEnumerable<IBotModule> LoadModulesFromAppDomain()
        {
            return AppDomain.CurrentDomain.GetAssemblies()
                .SelectMany(a => SafeTypes(a))
                .Where(t => typeof(IBotModule).IsAssignableFrom(t) && !t.IsAbstract)
                .Select(t => (IBotModule)Activator.CreateInstance(t)!);
        }

        private static IEnumerable<IBotModule> LoadModulesFromPluginFolder(string folder)
        {
            if (!Directory.Exists(folder)) return Enumerable.Empty<IBotModule>();
            var modules = new List<IBotModule>();

            foreach (var dll in Directory.GetFiles(folder, "*.dll"))
            {
                var asm = AssemblyLoadContext.Default.LoadFromAssemblyPath(Path.GetFullPath(dll));
                var types = SafeTypes(asm)
                    .Where(t => typeof(IBotModule).IsAssignableFrom(t) && !t.IsAbstract);
                modules.AddRange(types.Select(t => (IBotModule)Activator.CreateInstance(t)!));
            }
            return modules;
        }

        private static IEnumerable<Type> SafeTypes(Assembly asm)
        {
            try { return asm.GetTypes(); }
            catch { return Enumerable.Empty<Type>(); }
        }
    }

}
