namespace BotWorker.Domain.Models.BotMessages;

public partial class BotMessage : MetaData<BotMessage>
{
        private async Task LogGuildEventAsync(string eventName)
        {
            var guildEvent = new GuildEvent
            {
                GroupId = GroupId,
                GroupName = GroupName,
                BotUin = SelfId,
                BotName = SelfName,
                UserId = UserId,
                UserName = Name,
                EventType = EventType,
                EventName = eventName,
            };

            await GuildEvent.AppendAsync(guildEvent, ["Id"]);
        }

        public async Task OnFriendAddAsync()
        {            
            await LogGuildEventAsync("加机器人好友");

            Answer = $"我来了！输入 / 可唤出菜单";

            (int i, _) = await AddCreditAsync(5000, EventType);
            if (i != -1) 
                Answer += $"\n积分：+5000，累计：{{积分}}";

            var addTokensRes = await AddTokensAsync(30000, EventType);
            if (addTokensRes.Result != -1)
                Answer += $"\n算力：+30000，累计：{{算力}}";

            Answer += "\n注意：删除机器人会收回赠送的积分算力哦";
        }

        public async Task OnFriendDelAsync()
        {
            Answer = "删除机器人好友";
            await LogGuildEventAsync("删除机器人好友");

            _ = await MinusTokensAsync(30000, EventType);
            _ = await MinusCreditAsync(5000, EventType);
            IsSend = false;          
        }

        public async Task OnGroupAddRobotAsync(bool isNewGroup)
        {
            await LogGuildEventAsync("加机器人进群");

            if (!GroupVip.IsVip(GroupId))
            {
                GroupInfo.SetRobotOwner(GroupId, UserId);
                GroupInfo.SetPowerOn(GroupId);
            }

            Answer = "我来了，输入 / 或艾特我可唤出菜单";

            if (isNewGroup)
            {
                (int i, _) = await AddCreditAsync(5000, EventType);
                if (i != -1)
                    Answer += $"\n积分：+5000，累计：{{积分}}";

                var addTokensRes = await AddTokensAsync(30000, EventType);
                if (addTokensRes.Result != -1)
                    Answer += $"\n算力：+30000，累计：{{算力}}";

                Answer += "\n注意：踢出机器人会收回赠送的积分算力哦";
            }
        }

        public async Task OnGroupDelRobot()
        {
            await LogGuildEventAsync("踢机器人出群");

            Answer = "我被踢了";
            IsSend = false;
            
            (int i, _) = await MinusCreditAsync(5000, EventType);
            if (i != -1)            
                Answer += $"\n积分：-5000，累计：{{积分}}";
            
            var minusTokensRes = await MinusTokensAsync(30000, EventType);
            if (minusTokensRes.Result != -1)
                Answer += $"\n算力：-30000，累计：{{算力}}";
        }

        /// <summary>
        /// 频道新成员加入
        /// </summary>
        private async Task OnGuildMemberAddAsync()
        {
            Answer = $"欢迎<@{UserOpenId}>加入【{GuildName}】，输入 / 或艾特我可唤出菜单";
            (int i, _) = await AddCreditAsync(5000, EventType);
            if (i != -1)
                Answer += $"\n积分：+5000，累计：{{积分}}";

            var addTokensRes = await AddTokensAsync(30000, EventType);
            if (addTokensRes.Result != -1)
                Answer += $"\n算力：+30000，累计：{{算力}}";

            Answer += "\n注意：退出频道会收回赠送的积分算力哦";
        }
}
