using System.Text.RegularExpressions;

namespace BotWorker.Domain.Models.BotMessages;

public partial class BotMessage : MetaData<BotMessage>
{
        public async Task GetCmdResAsync()
        {
            // 已关闭的功能处理
            if (BotCmd.IsClosedCmd(GroupId, CmdName))
            {
                switch (CmdName)
                {
                    case "闲聊":
                        if (!QuestionInfo.GetIsSystem(QuestionInfo.GetQId(CmdPara)))
                            return;
                        break;
                    default:
                        if (CmdName == UserInfo.GetStateRes(User.State))
                            UserInfo.SetState(UserInfo.States.Chat, UserId);
                        if (GroupInfo.GetBool("IsHintClose", GroupId))
                        {
                            if (CmdName.In("剪刀", "石头", "布", "抽奖", "三公") && !CmdPara.IsNum())
                                return;
                            Answer = $"{CmdName.Replace("押", "")}功能已关闭";
                        }
                        await GetAnswerAsync();
                        return;
                }
            }

            if (IsGuild && CmdName.In("结算"))
            {
                IsCancelProxy = true;
                return;
            }

            if (CmdName == "语音播报")
            {
                if (!IsGroup)
                    Answer = "语音播报仅限群内使用";
                else if (IsGuild)
                    Answer = "此版本不支持语音播报";
                else
                {
                    Answer = CmdPara;
                    IsCancelProxy = true;
                }
            }
            else if (CmdName == "生成提示词")
                Answer = $"{{#系统提示词生成器 请以【{CmdPara}】为主题生成一段智能体的系统提示词}}";
            else if (CmdName == "菜单" || CmdName == "帮助")
            {
                Answer = await GetMenuResAsync();
                if (CmdName == "帮助")
                    Answer = "【帮助菜单】\n" + Answer;
            }
            else if (CmdName == "签到")
            { 
                Answer = await TrySignInAsync(false);
                IsCancelProxy = !Answer.IsNull();
            }
            else if (CmdName == "天气")
                Answer = await GetWeatherResAsync(CmdPara);
            else if (CmdName.In("接龙"))
                Answer = await GetJielongRes();
            else if (CmdName == "翻译")
                Answer = await GetTranslateAsync();
            else if (CmdName == "成语")
                Answer = (await Chengyu.GetCyResAsync(this)).ReplaceInvalid();
            else if (CmdName == "爱群主")
                Answer = await GetLampRes();
            else if (CmdName == "爱早喵")
                Answer = await GetLoveZaomiaoRes();
            else if (CmdName == "抽签")
                await GetChouqianAsync();
            else if (CmdName == "解签")
                await GetJieqianAsync();
            else if (CmdName == "笑话")
                Answer = await GetJokeResAsync();
            else if (CmdName == "鬼故事")
                await GetGhostStoryAsync();
            else if (CmdName.In("早安", "午安", "晚安") && CmdPara.IsNull())
                await GetGreetingAsync();
            else if (CmdName == "闲聊")
                await GetAnswerAsync();
            else if ((CmdName == "领积分") && CmdPara.IsNull())
                await GetCreditMoreAsync();
            else if (CmdName == "揍群主")
                Answer = $"揍群主！";
            else if (CmdName == "头衔")
                Answer = await GetSetTitleResAsync();
            else if (CmdName == "我要头衔")
                Answer = await GetSetTitleResAsync();
            else if (CmdName == "变身")
                Answer = await ChangeAgentAsync();
            else if (CmdName == "自动开发")
            {
                var devManager = ServiceProvider.GetRequiredService<BotWorker.Modules.AI.Interfaces.IDevWorkflowManager>();
                var projectPath = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "GeneratedProjects", Guid.NewGuid().ToString("N"));
                var success = await devManager.StartDevProjectAsync(CmdPara, projectPath);
                Answer = success ? $"✅ 自动化开发任务已完成！项目路径：{projectPath}" : "❌ 自动化开发任务执行失败，请检查日志。";
            }
            else if (CmdName.In("画图", "生图", "生成图片"))
                await GetImageResAsync();
            else if (CmdName.In("点歌", "送歌"))
                await GetMusicResAsync();
            else if (CmdName == "mv")
                await GetMusicResAsync("mv");
            else if (CmdName == "dj")
                await GetMusicResAsync("dj");
            else if (CmdName == "计算")
                Answer = await Calc.GetJsRes(CmdPara);
            else if (CmdName == "积分")
                Answer = $"[@:{UserId}]的{{积分类型}}：{{积分}}";
            else if (CmdName == "金币")
                Answer = $"[@:{UserId}]的金币：{await GroupMember.GetCoinsAsync((int)CoinsLog.CoinsType.goldCoins, GroupId, UserId):#0.00}";
            else if (CmdName == "我的宠物")
                Answer = await PetOld.GetMyPetListAsync(GroupId, GroupId, UserId);
            else if (CmdName == "拍砖")
                Answer = await Brick.GetBrickResAsync(this);
            else if (CmdName == "添加待办")
                Answer = await Todo.GetTodoResAsync(GroupId, GroupName, UserId, Name, "+", CmdPara);
            else if (CmdName == "我的待办")
                Answer = await Todo.GetTodoResAsync(GroupId, GroupName, UserId, Name, "todo", CmdPara);
            else if (CmdName.In("钓鱼", "抛竿", "收竿"))
                Answer = await Fishing.GetFishing(GroupId, GroupName, UserId, Name, CmdName, CmdPara);
            else if (CmdName == "大写")
                Answer = RmbDaxie.GetDaxieRes(CmdPara);
            else if (CmdName == "小写")
                Answer = RmbDaxie.GetXiaoxieRes(CmdPara);
            else if (CmdName == "打赏")
                Answer = await GetRewardCreditAsync();
            else if (CmdName == "三公")
                Answer = await GetSanggongResAsync();
            else if (CmdName == "抽奖")
                Answer = await GetLuckyDrawAsync();
            else if (CmdName == "ai")
                await GetAgentResAsync();
            else if (CmdName == "拼音")
                Answer = Pinyin.GetPinyinRes(CmdPara);
            else if (CmdName == "反查")
                Answer = (await Chengyu.GetFanChaResAsync(this)).ReplaceInvalid();
            else if (CmdName == "身份证")
                Answer = CID.GetCidRes(this);
            else if (CmdName == "简体")
                Answer = CmdPara.AsJianti().ReplaceInvalid();
            else if (CmdName == "繁体")
                Answer = CmdPara.AsFanti().ReplaceInvalid();
            else if (CmdName == "md5")
                Answer = CmdPara.MD5().ToLower();
            else if (CmdName == "sha256")
                Answer = CmdPara.Sha256();
            else if (CmdName == "sha384")
                Answer = CmdPara.Sha384();
            else if (CmdName == "sha512")
                Answer = CmdPara.Sha512();
            else if (CmdName == "积分榜")
                Answer = await UserInfo.GetCreditListAsync(GroupId);
            else if (CmdName == "后台")
                Answer = await GetSetupUrlAsync();
            else if (CmdName == "加密")
                Answer = Encrypt.GetEncryptRes(UserInfo.GetGuid(UserId).AsString(), CmdName, CmdPara);
            else if (CmdName == "解密")
                Answer = Encrypt.GetEncryptRes(UserInfo.GetGuid(UserId).AsString(), CmdName, CmdPara).ReplaceInvalid();
            else if (CmdName == "转账")
                Answer = UserInfo.GetTransferBalance(SelfId, GroupId, GroupName, UserId, Name, CmdPara);
            else if (CmdName == "续费")
                Answer = await GetBuyRobotAsync();
            else if (CmdName == "升级")
                Answer = await GetUpgradeAsync();
            else if (CmdName == "降级")
                Answer = await GetCancelSuperAsync();
            else if (CmdName == "结算")
                Answer = await Partner.GetSettleResAsync(SelfId, GroupId, GroupName, UserId, Name);
            else if (CmdName == "兑换礼品")
                Answer = await GetGoodsCreditAsync();
            else if (CmdName == "买入")
                Answer = await GetBuyResAsync();
            else if (CmdName == "赎身")
                Answer = await GetFreeMeAsync();
            else if (CmdName == "买分")
                Answer = await UserInfo.GetBuyCreditAsync(this, SelfId, GroupId, GroupName, UserId, Name, CmdPara);
            else if (CmdName == "卖分")
                Answer = await GetSellCreditAsync();
            else if (CmdName == "加团")
                Answer = await GetBingFansAsync(CmdName);
            else if (CmdName == "退灯牌")
                Answer = await GetBingFansAsync(CmdName);
            else if (CmdName == "抽礼物")
                Answer = await GetGiftResAsync(UserId, CmdPara);
            else if (CmdName == "送礼物")
                Answer = await GetGiftResAsync(UserId, CmdPara);
            else if (CmdName == "逗你玩")
                Answer = await GetDouniwanAsync();
            else if (CmdName.In("全局开启", "全局关闭"))
                Answer = GetCloseAll();
            else if (CmdName == "暗恋")
                Answer = await GetSecretLove();
            else if (CmdName.In("押大", "押小", "押单", "押双", "押围", "押全围", "押点", "押对"))
                Answer = await GetBlockResAsync();
            else if (CmdName == "梭哈")
                Answer = await GetAllInAsync();
            else if (CmdName.In("猜数字", "我猜"))
                Answer = await GetGuessNumAsync();
            else if (CmdName == "todo")
                Answer = await Todo.GetTodoResAsync(GroupId, GroupName, UserId, Name, CmdName, CmdPara);
            else if (CmdName == "报时")
                Answer = $"🕒 现在时间：{DateTime.Now:yyyy年MM月dd日 HH:mm:ss} 星期{"日一二三四五六"[(int)DateTime.Now.DayOfWeek]}";
            else if (CmdName == "倒计时")
                Answer = await CountDown.GetCountDownAsync();
            else if (CmdName == "点歌")
                await GetMusicResAsync();
            else if (CmdName.In("生图", "画图", "生成图片"))
                await GetImageResAsync();
            else if (CmdName.In("红", "和", "蓝"))
            {
                if (CmdPara.IsNum())
                    Answer = await GetRedBlueResAsync(GroupId == 10084);
                else
                {
                    IsCmd = false;
                    CmdName = "闲聊";
                    CmdPara = Message;
                }
            }
            else if (CmdName == "猜拳" || CmdName.In("剪刀", "石头", "布"))
            {
                if (CmdPara.IsNum() || CmdPara.IsNull())
                    Answer = await GetCaiquanAsync();
                else
                    await GetAnswerAsync();
            }
            else if (CmdName.In("设置Key", "开启租赁", "关闭租赁", "我的Key"))
                Answer = await GetAiConfigResAsync();
            else if (CmdName == "尚未实现")
                Answer = $"尚未实现";
            else if (CmdName.In("开启", "关闭") && ((CmdPara.In("闭嘴", "闭嘴模式") && !IsRobotOwner()) || CmdPara.In("私链", "默认提示", "GPT4")))
                Answer = GetTurnOn(CmdName, CmdPara);
            else if (!CmdName.IsNull())
            {
                IsCancelProxy = true;
                if (CmdName == "设置" || CmdName == "提示词")
                {
                    if (CmdName == "提示词")
                    {
                        CmdPara = "提示词 " + CmdPara;
                        CmdName = "设置";
                    }
                    Answer = await SetupResAsync();
                    if (!IsGroup)
                    {
                        if (!Answer.Contains("设置群 "))
                            Answer += "\n设置群 {默认群}";
                    }
                    return;
                }

                Answer = SetupPrivate(true);
                if (Answer != "")
                    return;

                if (CmdName.In("开机", "关机") && CmdPara.IsNull())                
                    Answer = await GroupInfo.SetPowerOnOffAsync(SelfId, GroupId, UserId, CmdName);               
                else if (CmdName.In("开启", "关闭"))
                {
                    CmdPara = CmdPara.Replace("话痨", "话唠").Replace("加黑", "拉黑").Replace("模式", "").Replace("语音回复", "语音").Replace("AI声聊", "语音", StringComparison.CurrentCultureIgnoreCase).Replace("声聊", "语音").Replace("声音", "语音").Replace("语音", "语音回复");
                    CmdPara = CmdPara.Replace("自动撤回", "阅后即焚").Replace("积分系统", "积分").Replace("积分", "积分系统").Replace("回复图片", "图片回复").Replace("回复撤回", "撤回回复");
                    if (CmdPara.In("聊天", "闭嘴", "本群", "官方", "话唠", "终极", "AI", "纯血AI", "猜拳", "猜大小"))
                        await GetShortcutSetAsync();
                    else if (CmdPara.In("欢迎语", "退群提示", "改名提示", "命令前缀", "进群改名", "退群拉黑", "被踢提示", "被踢拉黑", "踢出拉黑", "进群禁言", "道具系统",
                        "宠物系统", "群管系统", "敏感词", "敏感词系统", "简洁", "进群确认", "群链", "邀请统计", "功能提示", "AI", "群主付", "自动签到",
                        "权限提示", "云黑名单", "管理加白", "多人互动", "知识库", "图片回复", "撤回回复", "语音回复", "阅后即焚", "积分系统"))
                    {
                        if (CmdPara.In("群主付") && !IsRobotOwner())
                            Answer = OwnerOnlyMsg;
                        else
                            Answer = await GetTurnOnAsync(CmdName, CmdPara);
                    }
                    else if (CmdPara.In("本群积分"))
                    {
                        Answer = IsRobotOwner() || BotInfo.IsAdmin(SelfId, UserId)
                            ? await GetTurnOnAsync(CmdName, CmdPara)
                            : OwnerOnlyMsg;
                    }
                    else
                        Answer = await GroupInfo.GetSetRobotOpenAsync(GroupId, CmdName, CmdPara);
                }
                else if (CmdName.In("上分", "下分"))
                    Answer = await GroupMember.GetShangFenAsync(SelfId, GroupId, GroupName, UserId, CmdName, CmdPara);
                else if (CmdName.In("拉黑", "取消拉黑", "清空黑名单"))
                    Answer = await GetBlackRes();
                else if (CmdName.In("拉灰", "取消拉灰", "清空灰名单"))
                    Answer = await GetGreyRes();
                else if (CmdName.In("白名单", "取消白名单", "清空白名单"))
                    Answer = GetWhiteRes();
                else if (CmdName == "改名")
                    Answer = await GetChangeName();
                else if (CmdName == "一键改名" && CmdPara == "")
                    await GetChangeNameAllAsync();
                else if (CmdName == "换群")
                    Answer = GetChangeGroup();
                else if (CmdName == "换主人")
                    Answer = GetChangeOwner();
                else if (CmdName == "警告")
                    Answer = await GetWarnRes();
                else if (CmdName == "查警告")
                    Answer = await GroupWarn.GetWarnInfoAsync(GroupId, CmdPara);
                else if (CmdName == "清警告")
                    Answer = await GroupWarn.GetClearResAsync(GroupId, CmdPara);

                if (!Answer.IsNullOrWhiteSpace())
                {
                    Answer = $"{(Group.RobotOwner == UserId ? "【主人】" : "")}{Answer}";
                    if (!IsGroup)
                    {
                        Answer = $"{Answer}{(!Answer.Contains("设置群 ") ? "\n设置群 {默认群}" : "")}";
                    }
                }
            }

            long credit = await UserInfo.GetCreditAsync(GroupId, UserId);
            if (credit <= -5000)
            {
                if (CmdName == "闲聊" || User.State == (int)UserInfo.States.Chat && IsGroup)                
                    IsSend = false;               
                else if (CmdName != "签到")
                    Answer = credit < -10000 ? "" : $"你已负分{credit}，不能再发命令";
                //自动切换回闲聊状态；
                if (User.State != (int)UserInfo.States.Chat)
                    UserInfo.SetState(UserInfo.States.Chat, UserId);
            }
            return;
        }

        //得到命令类型及参数
        public static (string, string) GetCmdPara(string text, string regex)
        {
            //去掉通讯工具附加的广告信息
            text = text.RemoveQqTail();

            var cmdName = string.Empty;
            var cmdPara = string.Empty;

            //分析命令类型
            var matches = text.Matches(regex);
            if (matches.Count > 0)
            {
                foreach (Match match in matches)
                {
                    cmdName = match.Groups["cmdName"].Value.Trim();
                    cmdPara = match.Groups["cmdPara"].Value.Trim();
                }

                cmdName = cmdName.AsNarrow().ToLower();
                if (regex == BotCmd.GetRegexCmd())
                    cmdName = BotCmd.GetCmdName(cmdName);
            }
            return (cmdName, cmdPara);
        }
}
