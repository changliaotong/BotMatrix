namespace BotWorker.Infrastructure.Utils
{
    /// <summary>
    /// 正则表达式
    /// </summary>
    public class Regexs
    {

        //官机敏感词
        public static string OfficalRejectWords => @"啪啪啪|黄爆|跑骚|🇹🇼|sm|射精|伪娘|偷看美女洗澡|爆菊|更衣|荤段子|汪洋大|勾引男|禽兽|纳粹|被操|淫|嫖|功高盖|野无遗|胖次|发春|互赞|呻吟|脱光|上吊|吊死|毛爷爷|陈独秀|自杀|宽衣|电倒|凹凸曼|床及|卖号|色群|妈逼|淫荡|亡政|阴道|线下约|大波妹|胎盘|丁字裤"; 
        //计算
        public static string Formula =>
            @"^\s*(?<formula>(?:[0-9]+(?:\.[0-9]+)?|\([^\)]+\)|（[^）]+）)(?:\s*[\+\-\*\/×÷＋－×÷／﹢﹣＊／]\s*(?:[0-9]+(?:\.[0-9]+)?|\([^\)]+\)|（[^）]+）))*?)\s*[=＝]\s*[?？]?\s*$";

        //coins
        public static string ExchangeCoins => @"^\s*(?<CmdName>兑换)\s*(?<cmdPara>金币|黑金币|紫币|游戏币|积分|本群分|本群积分)\s*(?<cmdPara2>\d*)\s*$";
        public static string AddMinus => @"^\s*(?<CmdName>充值|找)\s*(?<cmdPara>金币|黑金币|紫币|积分|群积分|本群分|本群积分|飞机票|禁言卡|解禁卡|解禁言)\s*(\[@:)?(?<cmdPara2>[1-9]\d{4,10})]?\s*(?<cmdPara3>\d*)\s*$";

        //fishing 
        public static string Fishing => @"^\s*(?<CmdName>钓鱼|抛竿|收竿)\s*$";
        public static string FishingBuy => @"^\s*(?<CmdName>购买)\s*(?<cmdPara>鱼(竿|钩|饵|线))\s*(?<cmdPara2>\d*)\s*$";
        public static string FishingSell => @"^\s*(?<CmdName>卖)\s*(?<cmdPara>水鬼|鲸鱼|章鱼|黄鱼|青蛙|贝壳|内衣|破鞋)\s*(?<cmdPara2>\d*)\s*$";

        //pet
        public static string BuyBet => @"^#?买[+＋ ]*(\[?@:?)?(?<UserId>[1-9]\d{4,10})(\])? *$";
        public static string PetPrice => @"^#?(查?身价|sj) *(\[?@:?)?(?<UserId>[1-9]\d{4,10})(\])?$";

        //music
        public static string SongId => @"[\s\S]*?songid=(\d*)[\s\S]*";
        public static string SongIdNetease => @"https://y.music.163.com/m/song[\s\S]*[\?&]id=(\d*)&[\s\S]*";
        public static string SongIdNetease2 => @"https://music.163.com/#/song\?id=(\d*)";
        public static string SongIdKugou => @"https://m.kugou.com/share/song.html\?chain=([\s\S]*)";
        public static string MusicIdZaomiao => @"https://sz84.com/music/([\s\S]*)";
        public static string MusicVideo => @"\*&vid=(?<vid>[\da-zA-Z]*)&ADTAG=qfshare";

        //dati
        public static string Dati => @"^\s*(?<CmdName>(三国|水浒|西游|红楼|诗词|历史|文学|百科|戏曲|音乐|美术|建筑|电影|金庸|古龙|人物|动物|植物|器物|药食|诗典|数字|地点|谜语|联想|话题|答案))\s*$";

        //block
        public static string BlockHash16 => @"^(^[\S\s]*(群|私)链[： :]*)?(?<block_hash>[0-9a-zA-Z]{16})$";
        public static string BlockHash => @"^上局HASH:(?<block_hash>\w{32}?)[\S\s]*$";
        public static string BlockCmd => @"^[#＃﹟]{0,1}(?<CmdName>[大小单双围四五六七八九十dxjswDXJSW]|(十[一二三四五六七]))[ \\/+]*(?<cmdPara>\d+)$";
        public static string BlockCmdMult => @"[#＃﹟]{0,1}(?<CmdName>[大小单双四五六七八九十dxjswDXJSW]|(十[一二三四五六七])|(押([1-6]+)|[围w]|(对[1-6])))[ \\/+]*(?<cmdPara>\d+)";
        public static string BlockPara => @"^(?<BlockNum>(\d{1,2}))[ \\/+]*(?<cmdPara>\d{1,9})";

        //game
        public static string Caiquan => @"^[#\s]*(?<CmdName>(jd|jiandao|剪刀|st|shitou|石头|bu|布))[\s+]*(?<cmdPara>\d{1,})\s*$";

        //credit
        public static string Transfer => @"^(\[?@:?)?(?<UserId>[1-9]\d{4,10})(\])?[ \]\\/+]+(?<balance>\d{0,8}(.\d{1,2})?$)";
        public static string Withdraw => @"^#?(?<CmdName>(提现|tx)) *(?<cmdPara>\d{1,10}(.\d{1,2})?) *$";
        public static string CreditParaAt => @"^[\s]*\[@:(?<UserId>[1-9]\d{4,10})][\s+]*(?<credit>\d{1,})[\s]*$";
        public static string CreditParaAt2 => @"^[\s]*(?<credit>\d{1,})[\s+]*\[@:(?<UserId>[1-9]\d{4,10})][\s]*$";
        public static string CreditPara => @"^(?<UserId>[1-9]\d{4,10})[\s+]+(?<credit>\d{1,}$)";
        public static string RewardCredit => @"^\s*(?<CmdName>打赏)\s*(\[@:)?(?<UserId>[1-9]\d{4,10})]?\s*(?<credit>\d+)\s*$";
        public static string CreditList => @"^\s*(?<CmdName>积分排行|积分排行榜|排行榜)\s*(?<top>\d*)\s*$";

        // group manager
        public static string KickCommandPrefixPattern => @"^[# ]*(踢|t|踢出|踢飞|tc|tf)\b";
        public static string QqNumberPattern => @"[1-9]\d{4,10}";
        public static string Mute => @"^#?((禁言)|(jy)) *(\[?@:?)?(?<UserId>[1-9]+\d{4,10})(\])? *(?<time>\d*)(?<unit>分|m|M|(分钟)|h|H|(小时)|时|d|D|日|天)? *$";
        public static string UnMute => @"^#?((取消禁言)|(解除禁言)|(qxjy)|(jcjy)) *(\[?@:?)?(?<UserId>[1-9]+\d{4,10})(\])?$";
        public static string Kick => @"^[# ]*(踢|t|踢出|踢飞|tc|tf)( *(\[?@:?)?(?<UserId>[1-9]\d{4,10})]?)+";
        public static string LeaveGroup => @"^[# ]*(退群|tq)( *(?<GroupId>[1-9]\d{4,10}))+";
        public static string SetTitle => @"^[# ]*(设置头衔|头衔|tx|sztx|touxian)\s*(?:\[@:(?<UserId>[1-9]+\d{4,10})\]|(?<UserId>[1-9]+\d{4,10}))\s*(?<title>.+)$";


        //command
        public static string BuyCredit => @"[#＃]?(买积?分|mj?f)[\s\W]*(\[?@:?)?(?<buy_qq>[1-9]\d{4,10})(\])? +(?<income_money>-?\d{1,4}) *(?<pay_method>wx|zfb|qq|支付宝|微信支付|QQ红包|微信)*$";
        public static string CreditUserId => @"^#?((查|查询)?积分|jf)[+＋ ]*(\[?@:?)?(?<UserId>[1-9]\d{4,10})(\])?$";
        public static string CreditUserId2 => @"^#?(\[?@:?)?(?<UserId>[1-9]\d{4,10})(\])?[+＋ ]*(查?积分|jf) *$";
        public static string CoinsUserId => @"^#?(c?jb|查?金币)[+＋ ]*(\[?@:?)?(?<UserId>[1-9]\d{4,10})(\])?$";
        public static string SaveCredit => @"^#?(?<CmdName>(存入?(积?分)?|取出?(积?分)?|cr?(j?f)?|qc?(j?f)?))[+＋ ]*(?<credit_value>\d+) *$";
        public static string WarnCmd => @"^[#\s]*(?<CmdName>(kq|kaiqi|开启|gb|guanbi|关闭|sz|shezhi|设置))\s*(?<cmdPara>([\s+]*(sp|shuaping|刷屏|tp|tupian|图片|wz|wangzhi|网址|zh|zhanghua|脏话|gg|guanggao|广告|tjq|推荐群|tjhy|推荐好友|hbzf|合并转发)\s*(ch|chehui|撤回|kf|koufen|扣分|jg|jinggao|警告|jy|jinyan|禁言|tc|tichu|踢出|jh|jiahei|加黑|lh|lahei|拉黑)+)+)";
        public static string WarnPara => @"(?<cmdPara>(sp|shuaping|刷屏|tp|tupian|图片|wz|wangzhi|网址|zh|zanghua|脏话|gg|guanggao|广告|tjq|推荐群|tjhy|推荐好友|hbzf|合并转发))\s*(?<cmdPara2>(ch|chehui|撤回|kf|koufen|扣分|jg|jinggao|警告|jy|jinyan|禁言|tc|tichu|踢出|jh|jiahei|加黑|lh|lahei|拉黑)+)";
        public static string WarnPara2 => "(?<cmdPara2>(ch|chehui|撤回|kf|koufen|扣分|jg|jinggao|警告|jy|jinyan|禁言|tc|tichu|踢出|jh|jiahei|加黑|lh|lahei|拉黑))";
        public static string BindToken => @"TOKEN:(?<token_type>MP|WX)(?<bind_token>[0-9a-zA-Z]{16})";
        public static string Cid => @"(?<cid>(?<=\D|\b)(\d{6}[12]+[90]+\d{2}[01]+\d[0123]+\d\d{3}[\d|x|X]+)(?=\D|\b))";
        public static string Todo => @"^#?(?<CmdName>td|todo)(?<cmd_oper>[\+\- ]*)(?<cmdPara>[\s\S]*)";


        public static string NewFace => "/(扯一扯|崇拜|菜汪|吃糖|生气|打脸|狂笑|口罩护体|摸锦鲤|吃瓜|啵啵|牛啊|摸鱼|汪汪|捂脸|问号脸|睁眼|哦|头秃|emm|元宝|哦哟|期待|胖三斤|牛气冲天|喵喵" +
            "|面无表情|沧桑|辣眼睛|暗中观察|生日快乐|无眼笑|无奈|请|呆|拍手|哼|呵呵哒|拍桌|暴击|喷脸|脑阔疼|我酸了|舔一舔|忙到飞起|魔鬼笑|原谅|撩一撩|偷看|甩头|拍头|颤抖|黑脸|糊脸|托腮|斜眼笑" +
            "|头撞击|拽炸天|原谅|啃头|恭喜|笑哭|羊驼|微笑|呲牙|药|鼓掌|得意|偷笑|抠鼻|尴尬|手枪|流泪|害羞|笑疯|撇嘴|坏笑|饥饿|糗大了|惊恐|闪电|大哭|色|发呆|你真棒棒|我想开了|狗狗笑哭|复兴号" +
            "|狗狗生气|敲敲|抛媚眼|续标识|玩火|花朵脸|超级赞|狗狗可怜|狗狗疑问|酸Q|大怨种|打call|仔细分析|比心|狼狗|庆祝|太赞了|你真棒棒|舔屏)";

        public static string PublicFace = "\\[Lol|LetDown|Duh|Terror|Flushed|Sick|Happy|Party|Fireworks|Worship|OMG|NoProb|MyBad|KeepFighting|Wow|Boring|Awesome|LetMeSee|Sigh|Hurt" +
            "|Broken|Packet|Rich|Blessing|GoForIt|Onlooker|Yeah!|Concerned|Smart|Smirk|Facepalm|Hey]";

        public static string PrivateToken => @"/login\?t=[0-9a-zA-Z]+";

        public static string Study => @"^[＃# ]*[问][：:;；，\s]*(?<question>[\s\S]*?)[#＃\s]*回?[答][：:;；，\s]*(?<answer>[\s\S]+)";
        public static string Study2 => @"^[＃# ]*[问][：:;；，\s]*(?<question>[\s\S]*?)[#＃\s]+回?[答][：:;；，\s]*(?<answer>[\s\S]+)";
        public static string QuestionRef => @"{{(?<question>[\s\S]*?)}}";

        public static string Url => @"(https?|ftp|file)://[-A-Za-z0-9+&@#/%?=~_*|!:,.;\()]+[-A-Za-z0-9+&@#/%=~_|\()]";
        public static string Url2 => @"(http://)*(www|[a-zA-Z0-9])*\.*[a-zA-Z0-9]+(\.(com|cn|hk|net|org|vc|cc))+";
        public static string Ip => @"^(?<ip>(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5]))(?=\D|\b)";
        public static string Mobile => @"^(?<mobile>(?<=\D|\b)[1][3568]{1}\d{9}(?=\D|\b))$";
        public static string Mobile2 => @"(?<mobile>(?<=\D|\b)[1][3568]{1}\d{9}(?=\D|\b))";
        public static string ShortNo => @"^(?<=\D|\b)(?<short_no>(6[1-9]{1}\d{2,4})(?=\D|\b)|(?<=\D|\b)(7[1-9]{1}\d{4})(?=\D|\b))$";
        public static string HaveUserId => @"(?<UserId>[1-9]\d{4,10})";
        /// <summary>
        /// 支持@ 完全匹配1个
        /// </summary>
        public static string User => @"^(\[?@:?)?(?<UserId>[1-9]\d{4,10})(\])?$";
        /// <summary>
        /// 支持@ 部分匹配1-N个
        /// </summary>
        public static string Users => @"(\[?@:?)?(?<UserId>[1-9]\d{4,10})(\])?";
        /// <summary>
        /// 必须@ 匹配1个
        /// </summary>
        public static string AtUser => @"^\s*\[@:(?<UserId>[1-9]\d{4,10})]\s*$";
        /// <summary>
        /// 必须@ 部分匹配1-N个
        /// </summary>
        public static string AtUsers => @"\[@:(?<UserId>[1-9]\d{4,10})]";
        public static string UserPara => @"^(\[?@:?)?(?<UserId>[1-9]\d{4,10})(\])?[ \\/+]*(?<cmdPara>[\s\S]*)$";
        public static string Bus => @"(?<=\D|\b)[mbkejnNMBKEJ]{0,1}(宝|康|阳光|区间|长途|地铁|高峰专线|高峰快巴|观光|机场|深莞|深惠|旅游|购物|海滨|福田保税区|凤凰山|假日|大运)*([①②③④⑤⑥⑦⑧⑨⑩罗宝蛇口环中龙岗龙华\d]{1,3}[-Mm]*\d{0,3})(区间|号线|快线|环线|专线|线|A|B|C|a|b|c){0,1}[①②③④⑤⑥⑦⑧⑨⑩\d]{0,1}(?=\D|\b)";
        public static string Key => @"^(?<para_1>[\w]*)[ ]*(?<para_2>[\w]*)";
        public static string BiaoDian => @"[/／，。？！【】（）：,.\?!<>《》:·“”""—、；;『』\[\]\(\)~～—…'!@￥%&*_+{}|=\\`^$‘’]";
        public static string Prefix => @"^[ ]*[#＃﹟]";
        public static string Province => "(河北|山西|内蒙古(自治区)?|辽宁|吉林|黑龙江|江苏|浙江|安徽|福建|江西|山东|河南|湖北|湖南|广东|广西(壮族自治区)?|海南|四川|贵州|云南|西藏(自治区)?|陕西|甘肃|青海" +
                                       "|宁夏(回族自治区)?|新疆(维吾尔自治区)?|台湾)省?";
        public static string DirtyWords => "草拟马|煞笔|傻屄|看看逼|有码|无码|卖淫|会所|双飞|3p|奸杀|自拍|偷拍|群交|暴你菊|暴菊|爆菊|阳痿|叫床|自慰|性交|fuck|肏|尻|屌|操烂|小逼|轮奸|鸡巴|基佬|做爱|搅基" +
                                        "|搞基|艹|你麻|泥煤|杂种|你妈的|他妈的|滚蛋|B好痒|煞笔|妈B|妈个B|日你|我操|卧槽|干你妈|操你|叼你|我操|草你|狗日|操逼|操B|泥马B|杂毛|傻逼|你妈逼|射了|屄|屌|脱衣舞" +
                                        "|是攻|是受|颠婆|菊花|撸|尼玛|裸聊|娇喘|jb|约炮|啪啪啪|便便|智障|傻×|人渣|车震|精子|猥琐|打飞机|强奸|粑粑|王八|gay|jj|鸡婆|怪胎|基友|淫荡|锤子|屁股|j8|屎|" +
                                        "畜生|废物|开房|充气娃娃|尿|屁|鸡婆|基友|淫荡|屎|激情|SB|人妖|充气娃娃|肉便器|rbq";
        public static string AdWords => @"搔B|我频繁了|荭|荭苞|红苞|[\+＋➕][我群裙qQｑＱ扣]|裙号|不收废|名片攒|免费送|免费翎|黄色软件|黄色视频|颜色视频|颜色软件|加好友咨询|搔B|动态裙|进群领取" +
                                        "|扫一扫进群|每人都有388|鎹你|V群|日入过百|兼zhi|时时彩|时彩|時時彩|時彩|兼职|红包雨|加盟|发红包群|本人QQ|骚女QQ|急聘兼职|急聘淘宝兼职|日赚150|工资按单现结|招聘兼职" +
                                        "|日赚百元|淘宝招兼职|夫妻大秀群|jumpqq|裙聊";
        public static string UrlWhite => @"[\w\d:/]*.(baidu.com|qq.com|sz84.com|sz84.net|pengguanghui.com｜windows.net)";
        public static string KeyGroup => "赞|攒|邀人|送赞|红包|名片赞|打字|拉人|互赞|信誉|资源|土豪|澳门|下注|激情|视频|成人|夫妻|机器人|免费|av|内部|天天有钱|天天来领钱|荭|福利|秒赞|苞|线报|利福|5中5";
        public static string BlackWords => "老虎机|六四事件|李洪志|法轮功|法轮大法|老虎机|胡锦涛|丁薛祥|胡春华|习近平|彭丽媛|江泽民|王沪宁|王岐山|李克强|李强|天安门";
        public static string ReplaceWords => "避孕套|制服|女优";
        public static string RejectWords => "爸|爹|爷|奶|丑|死|干|插|傻|笨|蠢|伞兵|智障|猪|狗"; 
    }            
}
