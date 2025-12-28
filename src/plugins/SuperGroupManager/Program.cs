using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using System.Text.RegularExpressions;
using BotMatrix.SDK;

namespace SuperGroupManager
{
    class Program
    {
        static async Task Main(string[] args)
        {
            var app = new BotMatrixPlugin();

            // 1. 入群欢迎逻辑
            app.On("on_group_increase", async ctx => {
                string groupId = ctx.Event.Payload["group_id"]?.ToString() ?? "";
                string userId = ctx.Event.Payload["user_id"]?.ToString() ?? "";
                
                ctx.Reply($"🌟 欢迎新成员 [at:user_id={userId}] 加入本群！\n请阅读群公告，遵守群规。");
                return;
            });

            // 2. 关键词监控中间件
            app.Use(next => async ctx => {
                if (ctx.Event.Name == "on_group_message") {
                    string text = ctx.Event.Payload.ContainsKey("text") ? ctx.Event.Payload["text"]?.ToString() ?? "" : "";
                    
                    // 示例敏感词列表 (实际应从 SessionStore 加载)
                    var forbiddenWords = new[] { "广告", "加群", "发票", "代开" };
                    
                    foreach (var word in forbiddenWords) {
                        if (text != null && text.Contains(word)) {
                            string messageId = ctx.Event.Payload["message_id"]?.ToString() ?? "";
                            string userId = ctx.Event.Payload["from"]?.ToString() ?? "";
                            string groupId = ctx.Event.Payload["group_id"]?.ToString() ?? "";

                            // 撤回消息
                            ctx.DeleteMessage(messageId);
                            
                            // 警告系统 (使用 SessionStore 记录警告次数)
                            // 这里简化演示，直接回复并禁言
                            ctx.Reply($"⚠️ 检测到违规词：{word}\n用户 [at:user_id={userId}] 已被撤回并禁言 10 分钟。");
                            
                            // 禁言 10 分钟 (600秒)
                            ctx.CallAction("mute_user", new Dictionary<string, object> {
                                { "group_id", groupId },
                                { "user_id", userId },
                                { "duration", 600 }
                            });
                            return; // 拦截不再向下执行
                        }
                    }
                }
                await next(ctx);
            });

            // 3. 交互式配置面板 (超级亮点)
            app.OnIntent("group_config", async ctx => {
                try {
                    string userId = ctx.Event.Payload["from"]?.ToString() ?? "";
                    
                    // 权限校验 (模拟)
                    if (userId != "admin_user_id") { // 实际应检查是否为群主/管理员
                        // ctx.Reply("❌ 只有群管理员才能执行此操作。");
                        // return;
                    }

                    var menu = "🛠️ 超级群管配置面板\n" +
                               "1. 开启/关闭 入群欢迎\n" +
                               "2. 编辑 敏感词库\n" +
                               "3. 设置 自动禁言时长\n" +
                               "q. 退出设置\n\n" +
                               "请输入选项数字：";

                    var choiceCtx = await ctx.AskAsync(menu, timeoutMs: 30000);
                    string choice = choiceCtx.Event.Payload["text"]?.ToString() ?? "";

                    switch (choice) {
                        case "1":
                            var statusCtx = await ctx.AskAsync("请输入 1 开启，0 关闭：");
                            ctx.Reply($"✅ 设置成功！入群欢迎已{(statusCtx.Event.Payload["text"]?.ToString() == "1" ? "开启" : "关闭")}。");
                            break;
                        case "2":
                            var wordCtx = await ctx.AskAsync("请输入要添加的敏感词：");
                            string newWord = wordCtx.Event.Payload["text"]?.ToString() ?? "";
                            // ctx.Session.Set("forbidden_words", newWord); // 实际应追加到列表
                            ctx.Reply($"✅ 已添加敏感词：{newWord}");
                            break;
                        case "q":
                            ctx.Reply("👋 已退出配置。");
                            break;
                        default:
                            ctx.Reply("⚠️ 无效选项。");
                            break;
                    }
                } catch (TimeoutException) {
                    ctx.Reply("⏰ 响应超时，已自动退出配置模式。");
                }
            });

            // 4. 黑名单查询
            app.Command("/blacklist", async ctx => {
                ctx.Reply("🔍 正在从分布式存储检索黑名单列表...");
                // 模拟延迟
                await Task.Delay(500);
                ctx.Reply("🚫 当前黑名单：\n- user_888 (滥发广告)\n- user_999 (辱骂他人)");
            });

            // 5. 帮助指令
            app.Command("/help", async ctx => {
                ctx.Reply("🛡️ SuperGroupManager 帮助菜单\n" +
                          "--------------------------\n" +
                          "/blacklist - 查看封禁列表\n" +
                          "群设置 - 进入交互式管理面板\n" +
                          "关键词监控 - 自动撤回敏感词并禁言");
            });

            Console.WriteLine("SuperGroupManager started...");
            await app.RunAsync();
        }
    }
}
