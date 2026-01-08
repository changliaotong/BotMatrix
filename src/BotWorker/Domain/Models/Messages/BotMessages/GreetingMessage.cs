namespace BotWorker.Domain.Models.Messages.BotMessages;

public partial class BotMessage : MetaData<BotMessage>
{

        public async Task GetGreetingAsync()
        {
            var greetingType = CmdName switch
            {
                "早安" => 0,
                "午安" => 1,
                "晚安" => 2,
                _ => 0,
            };

            DateTime now = GetDateTime();
            int hour = now.Hour;

            if (CmdName == "早安")
            {
                if (hour >= 3 && hour < 4)
                    Answer = "你是凌晨的追梦人，还是早起的奋斗者？无论如何，愿你迎接美好的一天！🌅";
                else if (hour >= 4 && hour < 5)
                    Answer = "天还未亮，你已醒来。愿今天的努力换来更大的回报，早安！✨";
                else if (hour >= 5 && hour < 6)
                    Answer = "东方泛起鱼肚白，新的一天正悄悄到来，早安呀！🌄";
                else if (hour >= 6 && hour < 7)
                    Answer = "清晨的第一缕阳光洒在窗前，愿你拥有满满的正能量！🌞";
                else if (hour >= 7 && hour < 8)
                    Answer = "太阳刚升起，新的挑战和机遇也随之而来，早安，勇敢前行！☀️";
                else if (hour >= 8 && hour < 9)
                    Answer = "早餐吃了吗？充满活力的一天从美味开始！🍞🥛";
                else if (hour >= 9 && hour < 10)
                    Answer = "早安！工作或学习都要加把劲，愿你今日收获满满！💼📚";
                else if (hour >= 10 && hour < 11)
                    Answer = "虽然起得不早，但依然可以活力满满地开启新的一天哦~ 😄";
                else if (hour >= 11 && hour < 12)
                    Answer = "快到中午啦，别忘了保持好心情，事情都会慢慢变好的！☘️";
                else
                    Answer = "虽然已经不算早，但一天的精彩才刚刚开始，早安！🌈";

                Answer = $"✅ {Answer}你是本群第{GreetingRecords.GetCount(GroupId, 0)}全服第{GreetingRecords.GetCount(0)}位早起者！😄";
            }
            else if (CmdName == "午安")
            {
                if (now.Hour >= 10 && now.Hour < 11)
                    Answer = "午安！别忘了吃午饭哦，补充能量，下午继续加油！🍱";
                if (hour >= 11 && hour < 13)
                    Answer = "午安！中午到了，记得按时吃饭哦，休息片刻继续加油！🍱";
                else if (hour >= 13 && hour < 14)
                    Answer = "饭后小憩，午后阳光温暖惬意，愿你心情舒畅~ 😌";
                else if (hour >= 14 && hour < 15)
                    Answer = "困意袭来？眯一会儿或伸个懒腰，继续迎接下午的挑战吧！☕";
                else if (hour >= 15 && hour < 16)
                    Answer = "午后时光正好，来杯茶，感受片刻的宁静。午安~ 🍵";
                else if (hour >= 16 && hour < 17)
                    Answer = "临近傍晚，工作/学习是否接近尾声？保持专注再冲一波！💪";
                else if (hour >= 17 && hour < 18)
                    Answer = "夕阳西下，光影交织，是时候放慢脚步，享受傍晚的宁静。🌇";
                else
                    Answer = "午安也是一种祝福，不管几点都希望你一切顺利、安好如初~ ✨";
                Answer = $"✅ {Answer}你是本群第{GreetingRecords.GetCount(GroupId, 1)}全服第{GreetingRecords.GetCount(1)}位饭困者 😴";
            }
            else if (CmdName == "晚安")
            {
                if (now.Hour >= 17 && now.Hour < 19)
                    Answer = "夜幕降临，是时候放松一下，享受美好的夜晚！🌃";
                else if (hour >= 19 && hour < 20)
                    Answer = "华灯初上，忙碌了一天的你，值得一段静谧时光。🌃";
                else if (hour >= 20 && hour < 21)
                    Answer = "晚安！夜色温柔，希望你今晚有甜甜的梦~ 🌙";
                else if (hour >= 21 && hour < 22)
                    Answer = "一天就要结束了，洗个热水澡，早点休息吧~ 🛁";
                else if (hour >= 22 && hour < 23)
                    Answer = "闭上眼睛，卸下烦恼，明天会更好，晚安好梦！✨";
                else if (hour >= 23 && hour < 0)
                    Answer = "夜已深，是时候对今天说声“辛苦啦”，晚安！💤";
                else if (hour >= 0 && hour < 1)
                    Answer = "已经是凌晨了，还没睡的话记得早点休息哦，身体最重要！🌌";
                else if (hour >= 1 && hour < 2)
                    Answer = "夜猫子你好~ 安静的夜里也请照顾好自己，晚安~ 🦉";
                else if (hour >= 2 && hour < 3)
                    Answer = "凌晨的时光容易让人沉思，也容易让人疲惫。早点睡吧，朋友。💤";
                else if (hour >= 3 && hour < 4)
                    Answer = "已经凌晨三点了，太阳都快醒了~ 快去睡吧！🌄";
                else if (hour >= 4 && hour < 5)
                    Answer = "夜将尽，梦将启，如果你还未入睡，现在也不晚，晚安~ 🌠";
                else
                    Answer = "💤 晚安，现在是个不错的睡觉时间~ 祝你好梦！✨";
                Answer = $"✅ {Answer}你是本群第{GreetingRecords.GetCount(GroupId, 2)}全服第{GreetingRecords.GetCount(2)}位追梦人！💤";
            }

            if (GreetingRecords.Exists(GroupId, UserId, greetingType))
                Answer = $"今天已经问候过{CmdName}了";            
            else if (((CmdName == "早安" && now.Hour >= 3 && now.Hour < 12) || (CmdName == "午安" && now.Hour >= 10 && now.Hour < 18) || (CmdName == "晚安" && now.Hour >= 17 || now.Hour < 5)))
            {
                int i = GreetingRecords.Append(SelfId, GroupId, GroupName, UserId, Name, greetingType);
                if (i == -1)
                    Answer = RetryMsg;
                else if (Group.IsCreditSystem)
                {
                    var creditAdd = 50;
                    (i, long credit) = UserInfo.AddCredit(SelfId, GroupId, GroupName, UserId, Name, creditAdd, $"{CmdName}加分");
                    if (i != -1)
                        Answer += $"\n💎 积分：+{creditAdd}，累计：{credit:N0}";
                }
            }
            else
                Answer = $"请在正确的时间段发送问候语：\n" +
                        "🌞 早安：3:00 ~ 12:00\n" +
                        "☀️ 午安：10:00 ~ 18:00\n" +
                        "🌙 晚安：17:00 ~ 5:00";

            //if ((IsOffical || IsNapCat || IsMirai) && !Answer.IsNull())            
            //await SendMessageAsync();               

            //Answer = "";
            //IsCmd = false;
            //CmdName = "闲聊";
            //CmdPara = CmdName;
            //await GetAnswerAsync();
        }
}
