using Newtonsoft.Json;
using BotWorker.Infrastructure.Communication;

namespace BotWorker.Domain.Models.Messages.BotMessages;

public partial class BotMessage : MetaData<BotMessage>
{
        public async Task<bool> IsInGroupAsync(long selfId, long group, long target)
        {
            var client = RobotClientFactory.Get(Platform == Platforms.Mirai ? Platforms.Mirai : Platforms.Worker);
            try
            {
                return await client.IsInGroupAsync(selfId, group, target);
            }
            catch (NotSupportedException)
            {
                Console.WriteLine($"[{Platform}] 不支持查询是否在群内");
                return false;
            }
        }

        public async Task MuteAsync(long selfId, long group, long target, int seconds)
        {
            var client = RobotClientFactory.Get(Platform == Platforms.Mirai ? Platforms.Mirai : Platforms.Worker);
            try
            {
                await client.MuteAsync(selfId, group, target, seconds);
            }
            catch (NotSupportedException)
            {
                Console.WriteLine($"[{Platform}] 不支持禁言");
            }
        }

        public async Task KickOutAsync(long selfId, long group, long target)
        {
            var client = RobotClientFactory.Get(Platform == Platforms.Mirai ? Platforms.Mirai : Platforms.Worker);
            try
            {
                await client.KickAsync(selfId, group, target);
            }
            catch (NotSupportedException)
            {
                Console.WriteLine($"[{Platform}] 不支持踢人");
            }
        }

        public async Task RecallAsync(long selfId, long group, string message)
        {
            var client = RobotClientFactory.Get(Platform == Platforms.Mirai ? Platforms.Mirai : Platforms.Worker);
            try
            {
                await client.RecallAsync(selfId, group, message);
            }
            catch (NotSupportedException)
            {
                Console.WriteLine($"[{Platform}] 不支持撤回消息");
            }
        }
        public async Task RecallForwardAsync(long selfId, long group, string message, string forward)
        {
            var client = RobotClientFactory.Get(Platform == Platforms.Mirai ? Platforms.Mirai : Platforms.Worker);
            try
            {
                await client.RecallForwardAsync(selfId, group, message, forward);
            }
            catch (NotSupportedException)
            {
                Console.WriteLine($"[{Platform}] 不支持撤回消息");
            }
        }

        public async Task ChangeNameAsync(long selfId, long group, long target, string newName, string prefixBoy, string prefixGirl, string prefixAdmin)
        {
            var client = RobotClientFactory.Get(Platform == Platforms.Mirai ? Platforms.Mirai : Platforms.Worker);
            try
            {
                await client.ChangeNameAsync(selfId, group, target, newName, prefixBoy, prefixGirl, prefixAdmin);
            }
            catch (NotSupportedException)
            {
                Console.WriteLine($"[{Platform}] 不支持改名");
            }
        }

        public async Task ChangeNameAllAsync(long selfId, long group, string prefixBoy, string prefixGirl, string prefixAdmin, string userId = "")
        {
            var client = RobotClientFactory.Get(Platform == Platforms.Mirai ? Platforms.Mirai : Platforms.Worker);
            try
            {
                await client.ChangeNameAllAsync(selfId, group, prefixBoy, prefixGirl, prefixAdmin);
            }
            catch (NotSupportedException)
            {
                Console.WriteLine($"[{Platform}] 不支持改名");
            }
        }               

        public async Task SetTitleAsync(long selfId, long group, long target, string title)
        {
            var client = RobotClientFactory.Get(Platform == Platforms.Mirai ? Platforms.Mirai : Platforms.Worker);
            try
            {
                await client.SetTitleAsync(selfId, group, target, title);
            }
            catch (NotSupportedException)
            {
                Console.WriteLine($"[{Platform}] 不支持设置头衔");
            }
        }

        public async Task LeaveAsync(long selfId, long group)
        {
            var client = RobotClientFactory.Get(Platform == Platforms.Mirai ? Platforms.Mirai : Platforms.Worker);
            try
            {
                if (SelfInfo.Valid == 22)
                {
                    Console.WriteLine($"[Info] 号码 {SelfInfo.BotName}({SelfInfo.BotUin}) 不支持退群");
                    return;
                }
                await client.LeaveAsync(selfId, group);
            }
            catch (NotSupportedException)
            {
                Console.WriteLine($"[{Platform}] 不支持退群");
            }
        }
}
