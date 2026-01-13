using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Interfaces;
using Microsoft.Extensions.Logging;

namespace BotWorker.Modules.AI.Services
{
    public class SkillService : ISkillService
    {
        private readonly IEnumerable<ISkill> _builtinSkills;
        private readonly ISkillDefinitionRepository _skillRepository;
        private readonly ICodeRunnerService _codeRunner;
        private readonly ILogger<SkillService> _logger;

        public SkillService(
            IEnumerable<ISkill> builtinSkills, 
            ISkillDefinitionRepository skillRepository,
            ICodeRunnerService codeRunner,
            ILogger<SkillService> logger)
        {
            _builtinSkills = builtinSkills;
            _skillRepository = skillRepository;
            _codeRunner = codeRunner;
            _logger = logger;
        }

        public async Task<string> ExecuteSkillAsync(string skillName, string target, string reason, Dictionary<string, string> metadata)
        {
            // 1. 优先从数据库查找技能定义
            BotWorker.Modules.AI.Models.Evolution.SkillDefinition? definition = null;
            try {
                var skillDefinitions = await _skillRepository.GetByActionAsync(skillName);
                definition = skillDefinitions.FirstOrDefault();
            } catch (Exception ex) {
                _logger.LogWarning("[SkillService] Database access failed: {Message}", ex.Message);
            }

            if (definition != null)
            {
                if (definition.IsBuiltin)
                {
                    // 2. 如果是内置技能，从注入的内置技能列表中查找匹配项
                    var builtinSkill = _builtinSkills.FirstOrDefault(s => s.SupportedActions.Any(a => a.Equals(skillName, StringComparison.OrdinalIgnoreCase)));
                    if (builtinSkill != null)
                    {
                        _logger.LogInformation("[SkillService] Executing builtin skill {SkillKey} for action {Action}", definition.SkillKey, skillName);
                        return await builtinSkill.ExecuteAsync(skillName, target, reason, metadata);
                    }
                }
                else
                {
                    // 3. 如果是动态技能（脚本），执行 Python 脚本
                    if (!string.IsNullOrEmpty(definition.ScriptContent))
                    {
                        return await ExecuteDynamicSkillAsync(definition, target, reason, metadata);
                    }
                }
            }

            // [FALLBACK/TEST] 如果数据库中没有，且是测试用的 PYTEST
            if (skillName.Equals("PYTEST", StringComparison.OrdinalIgnoreCase))
            {
                var testDefinition = new BotWorker.Modules.AI.Models.Evolution.SkillDefinition 
                { 
                    SkillKey = "python_test", 
                    ActionName = "PYTEST", 
                    ScriptContent = "import sys\nimport json\n\ndef main():\n    print(f'MOCK Python Result: target={sys.argv[1]}')\n\nif __name__ == '__main__':\n    main()"
                };
                return await ExecuteDynamicSkillAsync(testDefinition, target, reason, metadata);
            }

            // 4. 后备方案：直接在内置技能中查找（兼容未在数据库注册的情况）
            var fallbackSkill = _builtinSkills.FirstOrDefault(s => s.SupportedActions.Any(a => a.Equals(skillName, StringComparison.OrdinalIgnoreCase)));
            if (fallbackSkill != null)
            {
                _logger.LogInformation("[SkillService] Executing fallback builtin skill for action {Action}", skillName);
                return await fallbackSkill.ExecuteAsync(skillName, target, reason, metadata);
            }

            _logger.LogWarning("[SkillService] No skill found for action {Action}", skillName);
            return $"错误：找不到支持行动 {skillName} 的工具。";
        }

        private async Task<string> ExecuteDynamicSkillAsync(BotWorker.Modules.AI.Models.Evolution.SkillDefinition definition, string target, string reason, Dictionary<string, string> metadata)
        {
            _logger.LogInformation("[SkillService] Executing dynamic skill {SkillKey} (Python) for action {Action}", definition.SkillKey, definition.ActionName);

            try
            {
                // 1. 创建临时脚本文件
                var tempDir = Path.Combine(Path.GetTempPath(), "BotMatrix", "Skills");
                Directory.CreateDirectory(tempDir);
                var scriptPath = Path.Combine(tempDir, $"{definition.SkillKey}_{DateTime.Now.Ticks}.py");
                await File.WriteAllTextAsync(scriptPath, definition.ScriptContent);

                // 2. 准备输入数据 (JSON)
                var inputData = new
                {
                    action = definition.ActionName,
                    target = target,
                    reason = reason,
                    metadata = metadata
                };
                var inputJson = System.Text.Json.JsonSerializer.Serialize(inputData);

                // 3. 执行 Python 脚本
                // 注意：这里我们假设 python 已经在环境变量中
                // 我们可以通过 stdin 传递 JSON，或者作为参数。为了安全和长度考虑，推荐 stdin 或 临时文件。
                // 目前 CodeRunnerService.ExecuteCommandAsync 不支持 stdin，我们需要扩展它或在此处手动处理。
                
                // 临时方案：将输入也存入文件，或者作为命令行参数（如果不太长）
                var inputPath = scriptPath + ".input.json";
                await File.WriteAllTextAsync(inputPath, inputJson);

                var command = $"python \"{scriptPath}\" \"{inputPath}\"";
                var result = await _codeRunner.ExecuteCommandAsync(command, AppContext.BaseDirectory);

                // 4. 清理临时文件
                try { File.Delete(scriptPath); File.Delete(inputPath); } catch { }

                if (result.ExitCode == 0)
                {
                    return result.Stdout.Trim();
                }
                else
                {
                    _logger.LogError("[SkillService] Dynamic skill execution failed: {Error}", result.Stderr);
                    return $"动态技能执行失败: {result.Stderr}";
                }
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[SkillService] Exception during dynamic skill execution");
                return $"动态技能执行异常: {ex.Message}";
            }
        }

        public IEnumerable<ISkill> GetAvailableSkills()
        {
            return _builtinSkills;
        }
    }
}
