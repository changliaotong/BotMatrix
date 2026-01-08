using System;
using System.Threading.Tasks;
using BotWorker.Domain.Interfaces;
using BotWorker.Common;
using BotWorker.Common.Extensions;
using BotWorker.Modules.Plugins;

using BotWorker.Infrastructure.Communication.OneBot;

namespace BotWorker.Application.Messaging.Pipeline
{
    /// <summary>
    /// 内置命令中间件：处理一些不属于插件系统的内置命令（如：菜单、状态等）
    /// </summary>
    public class BuiltinCommandMiddleware : IMiddleware
    {
        public async Task InvokeAsync(IPluginContext context, RequestDelegate next)
        {
            if (context is PluginContext pluginCtx && pluginCtx.Event is BotMessageEvent botMsgEvent)
            {
                var botMsg = botMsgEvent.BotMessage;
                
                // 处理内置指令 (复刻自 CommandMessage.cs)
                var isCmdMsg = botMsg.CurrentMessage.IsMatch(BotCmd.GetRegexCmd());
                botMsg.IsCmd = isCmdMsg;

                if (botMsg.IsCmd)
                {
                    (botMsg.CmdName, botMsg.CmdPara) = BotMessage.GetCmdPara(botMsg.CurrentMessage, BotCmd.GetRegexCmd());

                    // 如果是已迁移到插件的管理指令，则跳过内置处理，让插件系统处理
                    if (botMsg.CmdName.In("踢", "禁言", "取消禁言", "设置头衔", "开机", "关机", "设置欢迎语", "欢迎语", "改名提示", "拉黑", "取消拉黑", "黑名单", "被踢拉黑", "退群拉黑", "敏感词系统",
                        "领养宝宝", "我的宝宝", "宝宝改名", "宝宝学习", "宝宝打工", "宝宝互动", "宝宝商城", "购买", "拐卖宝宝说明", "开启宝宝系统", "关闭宝宝系统", "抛弃宝宝",
                        "求婚", "接受求婚", "拒绝求婚", "我要离婚", "办理结婚证", "办理离婚证", "我的婚姻", "婚姻面板", "发喜糖", "发红包", "吃喜糖", "购买婚纱", "购买婚戒", "我的对象", "另一半签到", "另一半抢楼", "另一半抢红包", "领取结婚福利", "我的甜蜜爱心", "赠送甜蜜爱心", "使用甜蜜抽奖", "甜蜜爱心说明"))
                    {
                        await next(context);
                        return;
                    }

                    // 指令有效性检查
                    if (IsInvalidCommand(botMsg))
                    {
                        botMsg.IsCmd = false;
                        botMsg.CmdName = "闲聊";
                        botMsg.CmdPara = botMsg.CurrentMessage;
                    }
                    else
                    {
                        if (botMsg.IsRefresh) botMsg.HandleRefresh();
                        else await botMsg.GetCmdResAsync();

                        if (!string.IsNullOrEmpty(botMsg.Answer)) return;
                    }
                }

                // 2. 确认指令状态
                await botMsg.ConfirmCmdAsync();
                if (!string.IsNullOrEmpty(botMsg.Answer)) return;

                // 3. 默认降级处理
                botMsg.CmdPara = botMsg.CurrentMessage;
                await botMsg.GetCmdResAsync();

                if (botMsg.IsRefresh && !string.IsNullOrEmpty(botMsg.Answer))
                {
                    return;
                }
            }

            await next(context);
        }

        private bool IsInvalidCommand(BotWorker.Domain.Models.Messages.BotMessages.BotMessage botMsg)
        {
            return (botMsg.CmdName.In("续费", "暗恋", "换群", "换主人", "警告") && !string.IsNullOrEmpty(botMsg.CmdPara) && !botMsg.CmdPara.IsNum())
                || (botMsg.CmdName.In("剪刀", "石头", "布", "抽奖", "三公", "牛牛", "牌九", "骰子") && !string.IsNullOrEmpty(botMsg.CmdPara) && (botMsg.CmdPara.Trim() != "梭哈") && !botMsg.CmdPara.IsNum())
                || (botMsg.CmdName.In("菜单", "领积分", "签到", "爱群主", "笑话", "鬼故事", "早安", "午安", "晚安", "揍群主", "升级", "降级", "结算", "一键改名") && !string.IsNullOrEmpty(botMsg.CmdPara))
                || (botMsg.CmdName.In("计算") && !botMsg.CmdPara.IsMatch(Regexs.Formula));
        }
    }
}
