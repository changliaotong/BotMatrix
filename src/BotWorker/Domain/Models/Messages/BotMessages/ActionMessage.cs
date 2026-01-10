using Newtonsoft.Json;
using BotWorker.Infrastructure.Communication;

namespace BotWorker.Domain.Models.Messages.BotMessages;

public partial class BotMessage : MetaData<BotMessage>
{
    private IRobotClient RobotClient => RobotClientFactory.Get(Platform == Platforms.Mirai ? Platforms.Mirai : Platforms.Worker);

    public async Task<bool> IsInGroupAsync(long selfId, long group, long target)
    {
        try
        {
            return await RobotClient.IsInGroupAsync(selfId, group, target);
        }
        catch (NotSupportedException)
        {
            Console.WriteLine($"[{Platform}] 不支持查询是否在群内");
            return false;
        }
    }

    public async Task MuteAsync(long selfId, long group, long target, int seconds)
    {
        try
        {
            await RobotClient.MuteAsync(selfId, group, target, seconds);
        }
        catch (NotSupportedException)
        {
            Console.WriteLine($"[{Platform}] 不支持禁言");
        }
    }

    public async Task KickOutAsync(long selfId, long group, long target)
    {
        try
        {
            await RobotClient.KickAsync(selfId, group, target);
        }
        catch (NotSupportedException)
        {
            Console.WriteLine($"[{Platform}] 不支持踢人");
        }
    }

    public async Task RecallAsync(long selfId, long group, string message)
    {
        try
        {
            await RobotClient.RecallAsync(selfId, group, message);
        }
        catch (NotSupportedException)
        {
            Console.WriteLine($"[{Platform}] 不支持撤回消息");
        }
    }
    public async Task RecallForwardAsync(long selfId, long group, string message, string forward)
    {
        try
        {
            await RobotClient.RecallForwardAsync(selfId, group, message, forward);
        }
        catch (NotSupportedException)
        {
            Console.WriteLine($"[{Platform}] 不支持撤回消息");
        }
    }

    public async Task ChangeNameAsync(long selfId, long group, long target, string newName, string prefixBoy, string prefixGirl, string prefixAdmin)
    {
        try
        {
            await RobotClient.ChangeNameAsync(selfId, group, target, newName, prefixBoy, prefixGirl, prefixAdmin);
        }
        catch (NotSupportedException)
        {
            Console.WriteLine($"[{Platform}] 不支持改名");
        }
    }

    public async Task ChangeNameAllAsync(long selfId, long group, string prefixBoy, string prefixGirl, string prefixAdmin, string userId = "")
    {
        try
        {
            await RobotClient.ChangeNameAllAsync(selfId, group, prefixBoy, prefixGirl, prefixAdmin);
        }
        catch (NotSupportedException)
        {
            Console.WriteLine($"[{Platform}] 不支持改名");
        }
    }               

    public async Task SetTitleAsync(long selfId, long group, long target, string title)
    {
        try
        {
            await RobotClient.SetTitleAsync(selfId, group, target, title);
        }
        catch (NotSupportedException)
        {
            Console.WriteLine($"[{Platform}] 不支持设置头衔");
        }
    }

    public async Task SetGroupAdminAsync(long selfId, long group, long target, bool admin)
    {
        try
        {
            await RobotClient.SetGroupAdminAsync(selfId, group, target, admin);
        }
        catch (NotSupportedException)
        {
            Console.WriteLine($"[{Platform}] 不支持设置管理员");
        }
    }

    public async Task SetGroupWholeMuteAsync(long selfId, long group, bool mute)
    {
        try
        {
            await RobotClient.SetGroupWholeMuteAsync(selfId, group, mute);
        }
        catch (NotSupportedException)
        {
            Console.WriteLine($"[{Platform}] 不支持全群禁言");
        }
    }

    public async Task LeaveAsync(long selfId, long group)
    {
        try
        {
            if (SelfInfo.Valid == 22)
            {
                Console.WriteLine($"[Info] 号码 {SelfInfo.BotName}({SelfInfo.BotUin}) 不支持退群");
                return;
            }
            await RobotClient.LeaveAsync(selfId, group);
        }
        catch (NotSupportedException)
        {
            Console.WriteLine($"[{Platform}] 不支持退群");
        }
    }

    public async Task SetGroupCardAsync(long selfId, long group, long target, string card)
        {
            try
            {
                await RobotClient.SetGroupCardAsync(selfId, group, target, card);
            }
            catch (NotSupportedException)
            {
                Console.WriteLine($"[{Platform}] 不支持设置群名片");
            }
        }
}
