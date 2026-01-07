using sz84.Bots.Entries;
using sz84.Bots.Extensions;
using sz84.Core.MetaDatas;
using sz84.Bots.Groups;
using System.Threading.Tasks;
using BotWorker.Common.Exts;

namespace sz84.Bots.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        //群刷屏
        public async Task GetRefreshRes()
        {
            IsCancelProxy = true;

            //刷屏
            bool isBlack = GroupWarn.ExistsKey(GroupId, "刷屏", "拉黑");
            bool isKick = GroupWarn.ExistsKey(GroupId, "刷屏", "踢出");
            bool isMute = GroupWarn.ExistsKey(GroupId, "刷屏", "禁言");
            bool isWarn = GroupWarn.ExistsKey(GroupId, "刷屏", "警告");
            bool isMinus = GroupWarn.ExistsKey(GroupId, "刷屏", "扣分");            
            if (!isBlack && !isKick && !isMute && !isWarn && !isMinus)
                return;

            await RecallAsync(SelfId, RealGroupId, MsgId);
            
            int times = 1;
            if (BotInfo.DictTimes.ContainsKey("a" + UserId.ToString()))
            {
                //a 提示请勿刷屏次数
                times = int.Parse(BotInfo.DictTimes["a" + UserId.ToString()]) + 1;
                BotInfo.DictTimes["a" + UserId.ToString()] = times.ToString();
                if (times > 9)
                {
                    if (isBlack || (isKick && Group.IsBlackKick))
                    {                        
                        await KickOutAsync(SelfId, RealGroupId, UserId);
                        int i = AddBlack(UserId, "刷屏拉黑");
                        if (i == -1)
                        {
                            Answer = RetryMsg;
                            return;
                        }
                        Answer = $"[@:{UserId}]刷屏将被拉黑并T飞\n已拉黑！";
                        GroupEvent.Append(this, $"拉黑", $"刷屏拉黑");
                        return;
                    }
                    if (isKick)
                    {
                        await KickOutAsync(SelfId, RealGroupId, UserId);
                        Answer = $"✅ [@:{UserId}]刷屏将被T飞";
                        GroupEvent.Append(this, $"踢出", $"刷屏踢出");
                        return;
                    }
                    if (isMute && Group.MuteRefreshCount != 0)
                    {
                        if (UserPerm < 2)
                        {
                            Answer = "刷屏禁言：不能禁言群主与管理";
                            IsSend = false;
                            return;
                        }
                        int muteTime = GroupInfo.GetInt("dissay_refresh_count", GroupId);                       
                        await MuteAsync(SelfId, RealGroupId, UserId, muteTime * 60);
                        Answer = $"✅ [@:{UserId}]刷屏将被禁言{muteTime}分钟！";
                        GroupEvent.Append(this, $"禁言", $"刷屏禁言{muteTime}分钟");
                        return;
                    }
                }
            }
            else
                BotInfo.DictTimes["a" + UserId.ToString()] = "1";

            Answer = $"请勿刷屏{times}";
            Answer += isWarn ? AddWarn(UserId, Name, SelfId) : "";
            GroupEvent.Append(this, $"警告", $"刷屏警告");

            if (isMinus)
            {
                (int i, long creditValue) = MinusCredit((long)(Math.Pow(2, times - 1) * 10), "刷屏扣分");
                Answer = i == -1
                    ? "刷屏扣分失败"
                    : Answer + $"\n积分：-{(Math.Pow(2, times - 1) * 10)} 累计：{creditValue}";
                GroupEvent.Append(this, $"扣分", $"刷屏扣分");
            }
        }

        //官方刷屏处理，只判断有回复消息的情况。
        public void HandleRefresh()
        {
            IsCancelProxy = true;

            int times = 1;
            if (BotInfo.DictTimes.ContainsKey("a" + UserId.ToString()))
            {
                //a 连续警告次数
                times = int.Parse(BotInfo.DictTimes["a" + UserId.ToString()]) + 1;
                BotInfo.DictTimes["a" + UserId.ToString()] = times.ToString();
            }
            else
                BotInfo.DictTimes["a" + UserId.ToString()] = "1";


            if (times >= 20)
            {
                int i = BlackList.AddBlackList(SelfId, BotInfo.GroupIdDef, $"{GroupName}({GroupId})", SelfId, "机器人", UserId, $"群{GroupId}内刷屏拉黑");
                if (i != -1)
                {
                    Answer = $"✅ [@:{UserId}]因连续刷屏已被列入官方黑名单";
                    GroupEvent.Append(this, $"拉黑", $"连续刷屏20次被列入官方黑名单");
                }
            }
            else
            {
                long minus = (long)(Math.Pow(2, times - 1) * 10);
                (int i, long creditValue) = MinusCredit(minus, "刷屏扣分");
                if (i != -1)
                {
                    Answer = $"请勿刷屏，{times}次警告！\n积分：-{minus}分 累计：{creditValue}";
                    GroupEvent.Append(this, $"扣分", $"刷屏扣分:-{minus}");
                }
            }
        }
    }
}
