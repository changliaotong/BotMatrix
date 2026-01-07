using BotWorker.Groups;
using System.Text.RegularExpressions;
using BotWorker.Common;
using BotWorker.Common.Exts;
using BotWorker.Core.Database;
using BotWorker.Infrastructure.Tools;
using BotWorker.Bots.Entries;
using BotWorker.Bots.Games;
using BotWorker.Bots.Games.Gift;
using BotWorker.Bots.Models.Office;
using BotWorker.Bots.Groups;
using BotWorker.Core.MetaDatas;
using BotWorker.Bots.Users;
using BotWorker.Agents.Entries;

namespace BotWorker.Bots.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        public static bool IsUrlWhiteListed(string url)
        {
            if (string.IsNullOrWhiteSpace(url)) return false;

            // 去掉两边的空白和可见控制字符
            url = url.Trim();

            // 尝试解析成绝对 URI
            if (!Uri.TryCreate(url, UriKind.Absolute, out var uri))
            {
                // 有些正则抓到的是不带协议的 URL，可以尝试加上 http:// 再解析一次
                if (Uri.TryCreate("http://" + url, UriKind.Absolute, out uri) == false)
                    return false;
            }

            var host = uri.Host.ToLowerInvariant();

            // 域名白名单（按需增加）
            if (host.EndsWith("sz84.com")) return true;
            if (host == "i.y.qq.com") return true;
            if (host == "music.163.com") return true;
            if (host == "mp.weixin.qq.com") return true;
            if (host == "q1.qlogo.cn") return true;
            if (host == "res.qpt.qq.com") return true;
            if (host.EndsWith(".kuwo.cn")) return true;
            if (host.EndsWith(".kuguo.com")) return true;

            // 对于 *.qq.com 下的特定路径或 id 也允许，通过判断原始 url 包含特定片段
            var raw = uri.ToString();
            if (host.EndsWith(".qq.com") && (raw.Contains("51437810") || raw.Contains("1653346663")))
                return true;

            return false;
        }

        //返回人性化结果
        public async Task GetFriendlyResAsync()
        {  
            //兼容旧格式
            Answer = ReplacePlaceholders(Answer);
            
            DoRegister();
            await Ctx.ReplaceAsync(this);            
            Answer = ReplaceDateTime(Answer);
            
            Answer = (await ResolveNestedExpressions(this)).Answer;                        
            
            //========================================================================

            // 转义字符还原
            Answer = Answer.Replace("\\{", "{").Replace("\\}", "}");

            // 提取全部 URL
            var urls = Regex.Matches(Answer, Regexs.Url)
                .Cast<Match>()
                .Select(m => m.Value)
                .ToList();

            bool hasUrl = urls.Any();
            bool onlyWhiteListUrls = hasUrl && urls.All(IsUrlWhiteListed);
            bool isNotVip = !GroupVip.IsYearVIP(GroupId);

            // 原来的问题：只要有一个非白名单就调用无差别的 BlockUrl()
            // 改成调用带回调的版本（或使用无参重载也行）
            if (isNotVip && !onlyWhiteListUrls)
            {
                Answer = Answer.BlockUrl(IsUrlWhiteListed); // 或 Answer.BlockUrl();
            }

            if (IsGuild)
            {
                if (IsMusic || IsAI || (!IsCmd && IsGroup && AnswerId != 0))
                {
                    ProxyBotUin = Group.BotUin;
                    IsProxy = true;
                }              
            }

            IsSend = IsSend && !Answer.IsNull();

            if (Answer.IsMatch(Regexs.BlackWords))                
            {
                IsSend = false;
                Reason += "[违禁词]";
            }

            if (IsGuild || IsRealProxy)
                Answer = Answer.ReplaceSensitive(Regexs.OfficalRejectWords);
            else
            {
                if (User.CreditTotal < -5000)
                {
                    IsSend = false;
                    Reason += "[负分]";
                }

                if (IsGroup && !BotInfo.GetBool("IsGroup", SelfId))
                {
                    IsSend = false;
                    Reason += "[群聊关闭]";
                }

                if (!IsGroup && !BotInfo.GetBool("IsPrivate", SelfId))
                { 
                    IsSend = false;
                    Reason += "[私聊关闭]";
                }
            }

            if (IsProxy && !IsGuild)
            {
                if (GroupOpenid.IsNull())
                    Group.GroupOpenId = GroupOffical.GetGroupOpenid(RealGroupId, ProxyBotUin);

                if (UserOpenId.IsNull())
                    User.UserOpenId = UserGuild.GetUserOpenid(ProxyBotUin, UserId);
            }

            if (Answer.IsMatch(Regexs.Url) || Answer.IsMatch(Regexs.Url2))
            {
                IsCancelProxy = true;

                if (IsGuild)
                {
                    if (GroupId > GroupOffical.MIN_GROUP_ID)                                            
                        Answer = "回复内容包含网址，本机无法发送";
                }

                Reason += "[Url]";
            }
        }

        public void DoRegister()
        {
            // 同步注册（自动包装为异步）
            Ctx.Register("群号", GroupId.ToString);            
            Ctx.Register("名字", Name.ToString);
            Ctx.Register("昵称", Name.ToString);
            Ctx.Register("GroupId", GroupId.ToString);
            Ctx.Register("RealGroupId", GroupId.ToString);
            Ctx.Register("GroupName", (GroupName?? "").ToString);            
            Ctx.Register("UserId", UserId.ToString);
            Ctx.Register("Name", Name.ToString);
            Ctx.Register("BotUin", SelfId.ToString);
            Ctx.Register("BotName", SelfName.ToString); 
            Ctx.Register("SystemPrompt", () => GroupInfo.GetSystemPrompt(GroupId));
            Ctx.Register("系统提示词", () => GroupInfo.GetSystemPrompt(GroupId));

            /*=========================================执行命令支持======================================================*/

            Ctx.Register("知识库开关", () => Group.IsUseKnowledgebase ? "开启" : "关闭");
            Ctx.Register("知识库文件数", () => KnowledgeVectors.CountField("FileId", "GroupId", GroupId).AsString());
            Ctx.Register("知识库字数", () => KnowledgeVectors.GetWhere("sum(len(content))", $"GroupId={GroupId}"));

            Ctx.Register("撤回计数", GetRecallCount);

            foreach (GroupEventType evt in Enum.GetValues(typeof(GroupEventType)))
            {
                if (evt != GroupEventType.撤回)
                    Ctx.Register($"{evt}计数", () => GetEventCount(evt));
            }

            Ctx.Register("菜单", GetMenuRes);
            Ctx.Register("后台", GetSetupUrlAsync);
            Ctx.Register("签到", () => TrySignIn(false));
            Ctx.Register("客服QQ", () => IsGuild || IsProxy ? "1653346663" : "1653346663");
            Ctx.Register("笑话", GetJokeRes);
            Ctx.Register("欢迎语", () => Group.WelcomeMessage);
            Ctx.Register("禁言我", GetMuteMeAsync);
            Ctx.Register("踢我", GetKickmeAsync);
            Ctx.Register("撤回", GetRecallMsgRes);
            Ctx.Register("测试", GetTestIt);
            Ctx.Register("闭嘴", GetShutupRes);

            // ====== 敏感词设置注册简化 ======
            foreach (var key in new[] { "刷屏", "图片", "网址", "脏话", "广告", "推荐群", "推荐好友", "合并转发" })
            {
                Ctx.Register($"{key}设置", () => GroupWarn.GetKeysSet(GroupId, key));
            }
            Ctx.Register("内置词设置", () => GroupWarn.GetKeysSet(GroupId));

            Ctx.Register("随机礼物", () => Gift.GetGiftList(GroupId, UserId));
            Ctx.Register("今日发言榜", () => GroupMsgCount.GetCountList(SelfId, GroupId, UserId, 8).ToString());
            Ctx.Register("昨日发言榜", () => GroupMsgCount.GetCountListY(SelfId, GroupId, UserId, 8).ToString());
            Ctx.Register("今日发言次数", () => GroupMsgCount.GetMsgCount(GroupId, UserId).ToString());
            Ctx.Register("今日发言排名", () => GroupMsgCount.GetCountOrder(GroupId, UserId).ToString());
            Ctx.Register("昨日发言次数", () => GroupMsgCount.GetMsgCountY(GroupId, UserId).ToString());
            Ctx.Register("昨日发言排名", () => GroupMsgCount.GetCountOrderY(GroupId, UserId).ToString());
            Ctx.Register("粉丝团", () => GroupGift.GetFansList(GroupId, UserId));
            Ctx.Register("亲密度值", () => GroupGift.GetFansValue(GroupId, UserId).ToString("N0"));
            Ctx.Register("粉丝排名", () => GroupGift.GetFansOrder(GroupId, UserId).ToString());
            Ctx.Register("粉丝等级", () => GroupGift.GetFansLevel(GroupId, UserId).ToString());
            Ctx.Register("荣誉等级", () => Income.GetClientLevel(UserId).ToString());
            Ctx.Register("荣誉榜", () => Income.GetLevelList(GroupId));
            Ctx.Register("荣誉排名", () => Income.GetLeverOrder(GroupId, UserId));

            // 注册积分类占位符
            Ctx.Register("领积分", () => GetFreeCredit());
            Ctx.Register("积分榜", () => GetCreditList());
            Ctx.Register("积分总榜", () => GetCreditListAll(UserId));
            Ctx.Register("积分类型", () => UserInfo.GetCreditType(GroupId, UserId));
            Ctx.Register("积分排名", () => UserInfo.GetCreditRanking(SelfId, GroupId, UserId).ToString());
            Ctx.Register("积分总排名", () => UserInfo.GetCreditRankingAll(UserId).ToString("N0"));
            Ctx.Register("本群积分", () => UserInfo.GetCredit(GroupId, UserId).ToString("N0"));
            Ctx.Register("本机积分", () => Friend.GetCredit(SelfId, UserId).ToString("N0"));
            Ctx.Register("通用积分", () => UserInfo.GetCredit(UserId).ToString("N0"));
            Ctx.Register("已存积分", () => UserInfo.GetSaveCredit(GroupId, UserId).ToString("N0"));
            Ctx.Register("储存积分", () => UserInfo.GetSaveCredit(GroupId, UserId).ToString("N0"));
            Ctx.Register("积分总额", () => UserInfo.GetTotalCredit(GroupId, UserId).ToString("N0"));
            Ctx.Register("冻结积分", () => UserInfo.GetFreezeCredit(UserId).ToString("N0"));
            Ctx.Register("余额", () => UserInfo.GetBalance(UserId).ToString("N"));
            Ctx.Register("冻结余额", () => UserInfo.GetFreezeBalance(UserId).ToString("N"));

            // 特殊处理积分（判断负分情况）
            Ctx.Register("积分", () =>
            {
                long credit_value = UserInfo.GetCredit(GroupId, UserId);
                var baseValue = credit_value.ToString("N0");
                if (credit_value < 0)
                    return $"{baseValue}\n您已负分{credit_value}，低于-50分后将不能使用机器人";
                return baseValue;
            });

            // 注册金币类
            Ctx.Register("金币", () => GroupMember.GetCoins((int)CoinsLog.CoinsType.goldCoins, GroupId, UserId).AsString("N0"));
            Ctx.Register("金币榜", () => GetCoinsList());
            Ctx.Register("金币总榜", () => GetCoinsListAll(UserId));
            Ctx.Register("金币排名", () => GetCoinsRanking(GroupId, UserId).ToString());
            Ctx.Register("金币总排名", () => GetCoinsRankingAll(UserId).ToString());

            // 你 你2（依赖 RealGroupId 和权限判断）
            Ctx.Register("你", () =>
            {
                if (IsGuild) return $"你";

                if (!IsGroup || IsRealProxy || IsVoiceReply)
                    return IsGroup || Card.IsNull() ? Name : Card;
                
                return $"[@:{UserId}]";
            });

            Ctx.Register("你2", () =>
            {
                if (GroupId == 0)
                    return (IsGuild || IsRealProxy) ? Name : $"{Name}({UserId})";

                return (IsGuild || IsRealProxy) 
                    ? IsVoiceReply ? Card.IsNull() ? Name : Card : $"{Name}({UserOpenId.MaskNo()})"
                    : $"[@:{UserId}]({UserId})";
            });

            // 机器人自己
            Ctx.Register("我", () => $"『{Group.BotName}』");
            Ctx.Register("我2", () => GroupId != 0 && IsVoiceReply ? Group.BotName : $"『{Group.BotName}({SelfId})』");

            // 群信息
            Ctx.Register("群", () => GroupName ?? "");
            Ctx.Register("群2", () => $"{GroupName ?? ""}({GroupId})");
            Ctx.Register("群号", () => GroupId == 0  ? $"{GroupId}（默认群）" : $"{GroupId}");

            var groupOwner = "辉辉";
            var groupOwnerQQ = "51437810";

            Ctx.Register("群主", () => groupOwner);
            Ctx.Register("群主2", () => $"{groupOwner}({groupOwnerQQ})");

            // 主人
            Ctx.Register("主人", () => Group.RobotOwnerName ?? "");
            Ctx.Register("主人2", () =>
            {
                var baseName = Group.RobotOwnerName ?? "";
                var ownerId = Group.RobotOwner;
                var hint = IsPublic ? $"\n群号：{User.DefaultGroup}" : "";
                return $"{baseName}({ownerId}){hint}";
            });

            // 城市 / 群设置
            Ctx.Register("天气预报", () => GetWeatherResAsync(User.CityName ?? ""));
            Ctx.Register("默认城市", () => User.CityName ?? "");
            Ctx.Register("默认群",  User.DefaultGroup.ToString);
            Ctx.Register("默认功能", () => UserInfo.GetStateRes(User.State) ?? "");
            Ctx.Register("默认提示", () => User.IsDefaultHint ? "提示" : "不提示");
            Ctx.Register("闭嘴模式开关", () => User.IsShutup ? "已开启" : "已关闭");
            Ctx.Register("聊天模式", () => GroupInfo.CloudAnswerRes(GroupId) ?? "");
            Ctx.Register("最低积分", Group.BlockMin.ToString);

            // 权限相关
            Ctx.Register("管理权限", () => GroupInfo.GetAdminRightRes(GroupId) ?? "");
            Ctx.Register("使用权限", () => GroupInfo.GetRightRes(GroupId) ?? "");
            Ctx.Register("调教权限", () => GroupInfo.GetTeachRightRes(GroupId) ?? "");
            Ctx.Register("调校权限", () => GroupInfo.GetTeachRightRes(GroupId) ?? "");
            Ctx.Register("教学权限", () => GroupInfo.GetTeachRightRes(GroupId) ?? "");

            // 群加入退出
            Ctx.Register("加群", () => GroupInfo.GetJoinRes(GroupId) ?? "");

            // ====== 快捷方法：提示/开关类注册 ======
            void RegisterHintSwitch(string name, Func<bool> state)
            {
                Ctx.Register($"{name}提示", () => state() ? "提示" : "不提示");
                Ctx.Register($"{name}提示开关", () => state() ? "已开启" : "已关闭");
            }
            void RegisterBlackSwitch(string name, Func<bool> state)
            {
                Ctx.Register($"{name}拉黑", () => state() ? "拉黑" : "不拉黑");
                Ctx.Register($"{name}拉黑开关", () => state() ? "已开启" : "已关闭");
            }

            RegisterHintSwitch("退群", () => Group.IsExitHint);
            RegisterBlackSwitch("退群", () => Group.IsBlackExit);
            RegisterHintSwitch("被踢", () => Group.IsKickHint);
            RegisterBlackSwitch("被踢", () => Group.IsBlackKick);
            RegisterHintSwitch("改名", () => Group.IsChangeHint);

            // 命令格式
            Ctx.Register("命令加#", () => Group.IsRequirePrefix ? "加#" : "不加#");

            // 黑白名单
            Ctx.Register("黑名单列表", GetGroupBlackList);
            Ctx.Register("白名单列表", GetGroupWhiteList);
            Ctx.Register("黑名单人数", () => BlackList.CountWhere($"group_id = {GroupId}").ToString());
            Ctx.Register("白名单人数", () => WhiteList.CountWhere($"group_id = {GroupId}").ToString());

            Ctx.Register("VIP", () => GetVipRes() ?? "");

            // 签到相关
            Ctx.Register("今日签到人数", () => GroupSignIn.SignCount(GroupId).AsString());
            Ctx.Register("昨日签到人数", () => GroupSignIn.SignCountY(GroupId).AsString());
            Ctx.Register("连续签到天数", () => GroupMember.GetSignTimes(GroupId, UserId).ToString());
            Ctx.Register("连续签到等级", () => GroupMember.GetValue("SignLevel", GroupId, UserId) ?? "");
            Ctx.Register("本月签到次数", () => GroupSignIn.SignCountThisMonth(GroupId, UserId).AsString());
            Ctx.Register("签到榜", () => GroupMember.GetSignList(GroupId, 3));
            Ctx.Register("自动签到开关", () => Group.IsAutoSignin ? "已开启" : "已关闭");

            // 群链 & 区块链
            Ctx.Register("群链开关", () => Group.IsBlock ? "已开启" : "已关闭");
            Ctx.Register("私链开关", () => User.IsBlack ? "已开启" : "已关闭");
            Ctx.Register("区块链开关", () => (IsGroup ? Group.IsBlock : User.IsBlock) ? "已开启" : "已关闭");
            Ctx.Register("block_hash16", () =>
            {
                long blockId = Block.GetId(GroupId, UserId);
                return blockId == 0 ? "游戏尚未开始" : Block.GetHash(blockId).Substring(7, 16);
            });
            Ctx.Register("block_hash", () => Block.GetHash(Block.GetId(GroupId, UserId)));
            Ctx.Register("block_type", () => IsGroup ? "群链" : "私链");

            // 宠物系统
            Ctx.Register("身价榜", () => PetOld.GetPriceList(GroupId, GroupId, UserId));
            Ctx.Register("身价", () =>
            {
                if ((GroupId != 0) && !Group.IsPet)
                    return "宠物系统已关闭";
                return PetOld.GetSellPrice(GroupId, UserId).ToString();
            });
            Ctx.Register("身价排名", () => PetOld.GetMyPriceList(GroupId, GroupId, UserId));
            Ctx.Register("我的宠物", () =>
            {
                if ((GroupId != 0) && !Group.IsPet)
                    return "宠物系统已关闭";
                return PetOld.GetMyPetList(GroupId, GroupId, UserId);
            });
            Ctx.Register("赎身", () => GetFreeMe());

            // 积分系统
            Ctx.Register("积分流水", () => Partner.GetCreditList(UserId));
            Ctx.Register("今日积分流水", () => Partner.GetCreditToday(UserId));

            // 金融/余额
            Ctx.Register("余额榜", () => UserInfo.GetBalanceList(GroupId, UserId).ToString());
            Ctx.Register("余额排名", () => UserInfo.GetMyBalanceList(GroupId, UserId).ToString());

            // Tokens / 算力
            Ctx.Register("TOKENS", () => $"{UserInfo.GetTokens(UserId):N0}");
            Ctx.Register("算力", () => $"{UserInfo.GetTokens(UserId):N0}");
            Ctx.Register("算力榜", () => UserInfo.GetTokensList(GroupId, UserId, 3));
            Ctx.Register("算力排名", () => UserInfo.GetTokensRanking(GroupId, UserId).AsString());

            // 合伙人
            Ctx.Register("成为合伙人", () => Partner.BecomePartner(UserId));

            // 运势
            Ctx.Register("今日运势", () => Fortune.Format(Fortune.GenerateFortune(UserId.AsString())));

            // 其他
            Ctx.Register("倒计时", () => CountDown.GetCountDown());
            Ctx.Register("segment", () => "\n");
        }

        public static string ReplaceDateTime(string message)
        {
            DateTime dt = SQLConn.GetDate();

            Dictionary<string, string> map = new()
            {
                ["{年}"] = dt.ToString("yyyy"),
                ["{月}"] = dt.ToString("MM"),
                ["{日}"] = dt.ToString("dd"),
                ["{时}"] = dt.ToString("HH"),
                ["{分}"] = dt.ToString("mm"),
                ["{秒}"] = dt.ToString("ss"),
                ["{星期}"] = "日一二三四五六"[(int)dt.DayOfWeek].ToString()
            };

            if (message.Contains("{农历年}") || message.Contains("{农历月}") || message.Contains("{农历日}"))
            {
                Yinli yinli = new(dt);
                map["{农历年}"] = yinli.GanzhiYearName ?? string.Empty;
                map["{农历月}"] = yinli.MonthName ?? string.Empty;
                map["{农历日}"] = yinli.DayName ?? string.Empty;
            }

            foreach (var kv in map)
                message = message.Replace(kv.Key, kv.Value);

            return message;
        }


        // 递归处理嵌套表达式
        static async Task<BotMessage> ResolveNestedExpressions(BotMessage bm, int maxDepth = 10, int currentDepth = 0)
        {
            // 检查递归深度
            if (currentDepth > maxDepth)
            {
                throw new InvalidOperationException("递归深度超过最大限制，可能存在死循环");
            }

            var matches = bm.Answer.Matches(@"{([^{}]+)}");

            // 如果没有找到嵌套结构，直接返回原始输入
            if (matches.Count == 0)
            {
                return bm;
            }

            foreach (Match match in matches)
            {
                string innerContent = match.Groups[1].Value;

                var inner = bm.DeepCopy();
                inner.Message = innerContent;
                inner.CurrentMessage = innerContent;
                inner.Group = await GroupInfo.LoadAsync(inner.GroupId);
                inner.User = await UserInfo.LoadAsync(inner.UserId);
                inner.Answer = "";

                // 递归处理内层结构
                BotMessage resolvedInner = await ResolveNestedExpressions(inner, maxDepth, currentDepth + 1);

                // 使用处理函数计算内层结构的值
                BotMessage result = resolvedInner;
                await result.HandleMessageAsync();

                // 替换内层结构的值并继续处理外层
                bm.IsAI = bm.IsAI | result.IsAI;
                bm.Answer = bm.Answer.Replace(match.Value, result.Answer);
            }

            await bm.GetFriendlyResAsync();

            return bm;
        }



        private static readonly Dictionary<string, string> PlaceholderReplacements = new()
        {
            { "#积分#", "{积分}" },
            { "#积分榜#", "{积分榜}" },
            { "#积分排名#", "{积分排名}" },
            { "#笑话#", "{笑话}" },
            { "#金币#", "{金币}" },
            { "#你#", "{你}" },
            { "#你2#", "{你2}" },
            { "#我#", "{我}" },
            { "#我2#", "{我2}" },
            { "#群#", "{群}" },
            { "#群2#", "{群2}" },
            { "#群号#", "{群号}" },
            { "#群主#", "{群主}" },
            { "#群主2#", "{群主2}" },
            { "#主人#", "{主人}" },
            { "#主人2#", "{主人2}" },
            { "#天气预报#", "{天气预报}" },
            { "#VIP#", "{VIP}" },
            { "#农历年#", "{农历年}" },
            { "#农历月#", "{农历月}" },
            { "#农历日#", "{农历日}" },
            { "#年#", "{年}" },
            { "#月#", "{月}" },
            { "#日#", "{日}" },
            { "#时#", "{时}" },
            { "#分#", "{分}" },
            { "#秒#", "{秒}" },
            { "{退群加黑}", "{退群拉黑}" },
            { "{退群加黑开关}", "{退群拉黑开关}" },
            { "{被踢加黑}", "{被踢拉黑}" },
            { "{被踢加黑开关}", "{被踢拉黑开关}" },
            { "{QQ}", "{UserId}" },
        };

        private static string ReplacePlaceholders(string value)
        {
            foreach (var replacement in PlaceholderReplacements)
            {
                value = value.Replace(replacement.Key, replacement.Value);
            }
            return value;
        }
    }
}
