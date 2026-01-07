package utils

// Regex Patterns migrated from sz84 Regexs.cs

// 官机敏感词 (Official Reject Words)
const OfficalRejectWords = `啪啪啪|黄爆|跑骚|🇹🇼|sm|射精|伪娘|偷看美女洗澡|爆菊|更衣|荤段子|汪洋大|勾引男|禽兽|纳粹|被操|淫|嫖|功高盖|野无遗|胖次|发春|互赞|呻吟|脱光|上吊|吊死|毛爷爷|陈独秀|自杀|宽衣|电倒|凹凸曼|床及|卖号|色群|妈逼|淫荡|亡政|阴道|线下约|大波妹|胎盘|丁字裤`

// 脏话 (Dirty Words)
const DirtyWords = `草拟马|煞笔|傻屄|看看逼|有码|无码|卖淫|会所|双飞|3p|奸杀|自拍|偷拍|群交|暴你菊|暴菊|爆菊|阳痿|叫床|自慰|性交|fuck|肏|尻|屌|操烂|小逼|轮奸|鸡巴|基佬|做爱|搅基|搞基|艹|你麻|泥煤|杂种|你妈的|他妈的|滚蛋|B好痒|煞笔|妈B|妈个B|日你|我操|卧槽|干你妈|操你|叼你|我操|草你|狗日|操逼|操B|泥马B|杂毛|傻逼|你妈逼|射了|屄|屌|脱衣舞|是攻|是受|颠婆|菊花|撸|尼玛|裸聊|娇喘|jb|约炮|啪啪啪|便便|智障|傻×|人渣|车震|精子|猥琐|打飞机|强奸|粑粑|王八|gay|jj|鸡婆|怪胎|基友|淫荡|锤子|屁股|j8|屎|畜生|废物|开房|充气娃娃|尿|屁|鸡婆|基友|淫荡|屎|激情|SB|人妖|充气娃娃|肉便器|rbq`

// 广告词 (Ad Words)
const AdWords = `搔B|我频繁了|荭|荭苞|红苞|[\+＋➕][我群裙qQｑＱ扣]|裙号|不收废|名片攒|免费送|免费翎|黄色软件|黄色视频|颜色视频|颜色软件|加好友咨询|搔B|动态裙|进群领取|扫一扫进群|每人都有388|鎹你|V群|日入过百|兼zhi|时时彩|时彩|時時彩|時彩|兼职|红包雨|加盟|发红包群|本人QQ|骚女QQ|急聘兼职|急聘淘宝兼职|日赚150|工资按单现结|招聘兼职|日赚百元|淘宝招兼职|夫妻大秀群|jumpqq|裙聊`

// 网址白名单 (Url White List)
const UrlWhite = `[\w\d:/]*.(baidu.com|qq.com|sz84.com|sz84.net|pengguanghui.com｜windows.net)`

// 敏感词 (Black Words)
const BlackWords = `老虎机|六四事件|李洪志|法轮功|法轮大法|老虎机|胡锦涛|丁薛祥|胡春华|习近平|彭丽媛|江泽民|王沪宁|王岐山|李克强|李强|天安门`

// 拒绝词 (Reject Words)
const RejectWords = `爸|爹|爷|奶|丑|死|干|插|傻|笨|蠢|伞兵|智障|猪|狗`

// 标点符号 (Punctuation)
const BiaoDian = `[/／，。？！【】（）：,.\?!<>《》:·“”""—、；;『』\[\]\(\)~～—…'!@￥%&*_+{}|\\` + "`" + `^$‘’]`

// QQ表情 (QQ Faces)
const NewFace = `/(扯一扯|崇拜|菜汪|吃糖|生气|打脸|狂笑|口罩护体|摸锦鲤|吃瓜|啵啵|牛啊|摸鱼|汪汪|捂脸|问号脸|睁眼|哦|头秃|emm|元宝|哦哟|期待|胖三斤|牛气冲天|喵喵|面无表情|沧桑|辣眼睛|暗中观察|生日快乐|无眼笑|无奈|请|呆|拍手|哼|呵呵哒|拍桌|暴击|喷脸|脑阔疼|我酸了|舔一舔|忙到飞起|魔鬼笑|原谅|撩一撩|偷看|甩头|拍头|颤抖|黑脸|糊脸|托腮|斜眼笑|头撞击|拽炸天|原谅|啃头|恭喜|笑哭|羊驼|微笑|呲牙|药|鼓掌|得意|偷笑|抠鼻|尴尬|手枪|流泪|害羞|笑疯|撇嘴|坏笑|饥饿|糗大了|惊恐|闪电|大哭|色|发呆|你真棒棒|我想开了|狗狗笑哭|复兴号|狗狗生气|敲敲|抛媚眼|续标识|玩火|花朵脸|超级赞|狗狗可怜|狗狗疑问|酸Q|大怨种|打call|仔细分析|比心|狼狗|庆祝|太赞了|你真棒棒|舔屏)`
const PublicFace = `\[Lol|LetDown|Duh|Terror|Flushed|Sick|Happy|Party|Fireworks|Worship|OMG|NoProb|MyBad|KeepFighting|Wow|Boring|Awesome|LetMeSee|Sigh|Hurt|Broken|Packet|Rich|Blessing|GoForIt|Onlooker|Yeah!|Concerned|Smart|Smirk|Facepalm|Hey]`

// Emoji
const EmojiPattern = `[\x{1F300}-\x{1F64F}\x{1F680}-\x{1F6FF}\x{2600}-\x{26FF}\x{2700}-\x{27BF}\x{1F900}-\x{1F9FF}\x{1F1E6}-\x{1F1FF}]`

// Regex for detection (Combined)
// Note: We might want to compile these if used frequently.
