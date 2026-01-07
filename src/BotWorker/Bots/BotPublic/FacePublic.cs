using System.Text.RegularExpressions;
using BotWorker.Common.Exts;

namespace sz84.Bots.Public
{
    public class FaceInfo(int faceId, string faceWeixinName, string faceQQName)
    {
        public int FaceId { get; set; } = faceId;
        public string FaceWeixinName { get; set; } = faceWeixinName;
        public string FaceQQName { get; set; } = faceQQName;

    }

    public static class FacePublic
    {
        private static readonly Dictionary<string, FaceInfo> faceWeixinTable = [];
        private static readonly Dictionary<string, FaceInfo> faceQQTable = [];
        private static readonly FaceInfo[] faceInfos =
                                                  [
                                                        new(14,"/::)","/微笑"),
                                                        new(1,"/::~","/撇嘴"),
                                                        new(2,"/::B","/色"),
                                                        new(3,"/::|","/发呆"),
                                                        new(4,"/:8-)","/得意"),
                                                        new(5,"/::<","/流泪"),
                                                        new(6,"/::$","/害羞"),
                                                        new(7,"/::X","/闭嘴"),
                                                        new(8,"/::Z","/睡"),
                                                        new(9,"/::'(","/大哭"),
                                                        new(10,"/::-|","/尴尬"),
                                                        new(11,"/::@","/发怒"),
                                                        new(12,"/::P","/调皮"),
                                                        new(13,"/::D","/呲牙"),
                                                        new(0,"/::O","/惊讶"),
                                                        new(15,"/::(","/难过"),
                                                        new(16,"/::+","/酷"),
                                                        new(96,"/:--b","/冷汗"),
                                                        new(36,"/:,@!","/衰"),
                                                        new(37,"/:!!!","/骷髅"),
                                                        new(38,"/:xx","/敲打"),
                                                        new(39,"/:bye","/再见"),
                                                        new(97,"/:wipe","/擦汗"),
                                                        new(98,"/:dig","/扣鼻"),
                                                        new(99,"/:handclap","/鼓掌"),
                                                        new(100,"/:&-(","/糗大了"),
                                                        new(101,"/:B-)","/坏笑"),
                                                        new(102,"/:<@","/左哼哼"),
                                                        new(103,"/:@>","/右哼哼"),
                                                        new(104,"/::-O","/哈欠"),
                                                        new(105,"/:>-|","/鄙视"),
                                                        new(106,"/:P-(","/委屈"),
                                                        new(107,"/::'|","/快哭了"),
                                                        new(108,"/:X-)","/阴险"),
                                                        new(109,"/::*","/亲亲"),
                                                        new(110,"/:@x","/吓"),
                                                        new(111,"/:8*","/可怜"),
                                                        new(112,"/:pd","/菜刀"),
                                                        new(89,"/:<W>","/西瓜"),
                                                        new(113,"/:beer","/啤酒"),
                                                        new(77,"/:weak","/弱"),
                                                        new(78,"/:share","/握手"),
                                                        new(79,"/:v","/胜利"),
                                                        new(118,"/:@)","/抱拳"),
                                                        new(119,"/:jj","/勾引"),
                                                        new(120,"/:@@","/拳头"),
                                                        new(121,"/:bad","/差劲"),
                                                        new(122,"/:lvu","/爱你"),
                                                        new(123,"/:no","/NO"),
                                                        new(124,"/:ok","/OK"),
                                                        new(42,"/:love","/爱情"),
                                                        new(85,"/:<L>","/飞吻"),
                                                        new(43,"/:jump","/跳跳"),
                                                        new(18,"/::Q","/抓狂"),
                                                        new(19,"/::T","/吐"),
                                                        new(20,"/:,@P","/偷笑"),
                                                        new(21,"/:,@-D","/愉快"),
                                                        new(22,"/::d","/白眼"),
                                                        new(23,"/:,@o","/傲慢"),
                                                        new(24,"/::g","/饥饿"),
                                                        new(25,"/:|-)","/困"),
                                                        new(26,"/::!","/惊恐"),
                                                        new(27,"/::L","/流汗"),
                                                        new(28,"/::>","/憨笑"),
                                                        new(29,"/::,@","/悠闲"),
                                                        new(30,"/:,@f","/奋斗"),
                                                        new(31,"/::-S","/咒骂"),
                                                        new(32,"/:?","/疑问"),
                                                        new(33,"/:,@x","/嘘"),
                                                        new(34,"/:,@@","/晕"),
                                                        new(35,"/::8","/疯了"),
                                                        new(114,"/:basketb","/篮球"),
                                                        new(115,"/:oo","/乒乓"),
                                                        new(60,"/:coffee","/咖啡"),
                                                        new(61,"/:eat","/饭"),
                                                        new(46,"/:pig","/猪头"),
                                                        new(63,"/:rose","/玫瑰"),
                                                        new(64,"/:fade","/凋谢"),
                                                        new(116,"/:showlove","/嘴唇"),
                                                        new(66,"/:heart","/爱心"),
                                                        new(67,"/:break","/心碎"),
                                                        new(53,"/:cake","/蛋糕"),
                                                        new(54,"/:li","/闪电"),
                                                        new(55,"/:bome","/炸弹"),
                                                        new(56,"/:kn","/刀"),
                                                        new(57,"/:footb","/足球"),
                                                        new(117,"/:ladybug","/瓢虫"),
                                                        new(59,"/:shit","/便便"),
                                                        new(75,"/:moon","/月亮"),
                                                        new(74,"/:sun","/太阳"),
                                                        new(69,"/:gift","/礼物"),
                                                        new(49,"/:hug","/拥抱"),
                                                        new(76,"/:strong","/强"),
                                                        new(41,"/:shake","/发抖"),
                                                        new(86,"/:<O>","/怄火"),
                                                        new(125,"/:circle","/转圈"),
                                                        new(126,"/:kotow","/磕头"),
                                                        new(127,"/:turn","/回头"),
                                                        new(128,"/:skip","/跳绳"),
                                                        new(129,"/:oY","/投降"),
                                                        new(130,"/:#-0","/激动"),
                                                        new(131,"/:hiphot","/乱舞"),
                                                        new(132,"/:kiss","/献吻"),
                                                        new(133,"/:<&","/左太极"),
                                                        new(134,"/:&>","/右太极")
                                                  ];
        static FacePublic()
        {
            foreach (FaceInfo faceInfo in faceInfos)
            {
                faceWeixinTable.Add(faceInfo.FaceWeixinName, faceInfo);
                faceQQTable.Add(faceInfo.FaceQQName, faceInfo);
            }
        }

        //根据表情id返回表情名
        public static string GetFaceByName(string weixin)
        {
            if (faceWeixinTable.TryGetValue(weixin, out var fi) && fi is FaceInfo faceInfo)
            {
                return $"[Face{faceInfo.FaceId}.gif]";
            }
            return string.Empty;
        }

        //根据表情id返回表情名
        public static string GetFaceInfo(int face_id)
        {
            foreach (FaceInfo faceInfo in faceInfos)
            {
                if (faceInfo.FaceId == face_id)
                    return faceInfo.FaceWeixinName;
            }
            return string.Empty;
        }

        public const string RegexFaceWeixin = @"/::\)|/::~|/::B|/::\||/:8-\)|/::<|/::\$|/::X|/::Z|/::'\(|/::-\||/::@|/::P|/::D|/::O|/::\(|/::\+|/:--b|/::Q|/::T|/:,@P|/:,@-D|/::d|/:,@o|/::g|/:\|-\)|/::!|/::L|/::>|/::,@|/:,@f|/::-S|/:\?|/:,@x|/:,@@|/::8|/:,@!|/:!!!|/:xx|/:bye|/:wipe|/:dig|/:handclap|/:&-\(|/:B-\)|/:<@|/:@>|/::-O|/:>-\||/:P-\(|/::'\||/:X-\)|/::\*|/:@x|/:8\*|/:pd|/:<W>|/:beer|/:basketb|/:oo|/:coffee|/:eat|/:pig|/:rose|/:fade|/:showlove|/:heart|/:break|/:cake|/:li|/:bome|/:kn|/:footb|/:ladybug|/:shit|/:moon|/:sun|/:gift|/:hug|/:strong|/:weak|/:share|/:v|/:@\)|/:jj|/:@@|/:bad|/:lvu|/:no|/:ok|/:love|/:<L>|/:jump|/:shake|/:<O>|/:circle|/:kotow|/:turn|/:skip|/:oY|/:#-0|/:hiphot|/:kiss|/:<&|/:&>";

        public static string ReplaceFace(string Message)
        {
            //微信表情处理 将微信表情转换为QQ表情
            var mc = Message.Matches(RegexFaceWeixin);
            foreach (var match in mc.Cast<Match>())
            {
                if (match.Value.StartsWith("/", StringComparison.OrdinalIgnoreCase))
                {
                    Message = Message.Replace(match.Value, GetFaceByName(match.Value));
                    continue;
                }
            }
            return Message;
        }


        public static string ConvertFacesBack(string message)
        {
            //结果中表情的处理 将QQ表情转换为微信表情
            var mc = message.Matches(@"\[Face(?<face_id>\d*).gif\]");
            foreach (Match match in mc)
            {
                int face_id = int.Parse(match.Groups["face_id"].Value);
                message = message.Replace(match.Value, GetFaceInfo(face_id));
            }
            return message;
        }

    }
}
