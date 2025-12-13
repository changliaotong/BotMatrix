#encoding:utf-8
import re

from SQLConn import *
from common import *

class msg():
    #消息管理

    qq_face_value = [(14,"/::)","/微笑"),
                (1,"/::~","/撇嘴"),
                (2,"/::B","/色"),
                (3,"/::|","/发呆"),
                (4,"/:8-)","/得意"),
                (5,"/::<","/流泪"),
                (6,"/::$","/害羞"),
                (7,"/::X","/闭嘴"),
                (8,"/::Z","/睡"),
                (9,"/::'(","/大哭"),
                (10,"/::-|","/尴尬"),
                (11,"/::@","/发怒"),
                (12,"/::P","/调皮"),
                (13,"/::D","/呲牙"),
                (0,"/::O","/惊讶"),
                (15,"/::(","/难过"),
                (16,"/::+","/酷"),
                (96,"/:--b","/冷汗"),
                (36,"/:,@!","/衰"),
                (37,"/:!!!","/骷髅"),
                (38,"/:xx","/敲打"),
                (39,"/:bye","/再见"),
                (97,"/:wipe","/擦汗"),
                (98,"/:dig","/扣鼻"),
                (99,"/:handclap","/鼓掌"),
                (100,"/:&-(","/糗大了"),
                (101,"/:B-)","/坏笑"),
                (102,"/:<@","/左哼哼"),
                (103,"/:@>","/右哼哼"),
                (104,"/::-O","/哈欠"),
                (105,"/:>-|","/鄙视"),
                (106,"/:P-(","/委屈"),
                (107,"/::'|","/快哭了"),
                (108,"/:X-)","/阴险"),
                (109,"/::*","/亲亲"),
                (110,"/:@x","/吓"),
                (111,"/:8*","/可怜"),
                (112,"/:pd","/菜刀"),
                (89,"/:<W>","/西瓜"),
                (113,"/:beer","/啤酒"),
                (77,"/:weak","/弱"),
                (78,"/:share","/握手"),
                (79,"/:v","/胜利"),
                (118,"/:@)","/抱拳"),
                (119,"/:jj","/勾引"),
                (120,"/:@@","/拳头"),
                (121,"/:bad","/差劲"),
                (122,"/:lvu","/爱你"),
                (123,"/:no","/NO"),
                (124,"/:ok","/OK"),
                (42,"/:love","/爱情"),
                (85,"/:<L>","/飞吻"),
                (43,"/:jump","/跳跳"),
                (18,"/::Q","/抓狂"),
                (19,"/::T","/吐"),
                (20,"/:,@P","/偷笑"),
                (21,"/:,@-D","/愉快"),
                (22,"/::d","/白眼"),
                (23,"/:,@o","/傲慢"),
                (24,"/::g","/饥饿"),
                (25,"/:|-)","/困"),
                (26,"/::!","/惊恐"),
                (27,"/::L","/流汗"),
                (28,"/::>","/憨笑"),
                (29,"/::,@","/悠闲"),
                (30,"/:,@f","/奋斗"),
                (31,"/::-S","/咒骂"),
                (32,"/:?","/疑问"),
                (33,"/:,@x","/嘘"),
                (34,"/:,@@","/晕"),
                (35,"/::8","/疯了"),
                (114,"/:basketb","/篮球"),
                (115,"/:oo","/乒乓"),
                (60,"/:coffee","/咖啡"),
                (61,"/:eat","/饭"),
                (46,"/:pig","/猪头"),
                (63,"/:rose","/玫瑰"),
                (64,"/:fade","/凋谢"),
                (116,"/:showlove","/嘴唇"),
                (66,"/:heart","/爱心"),
                (67,"/:break","/心碎"),
                (53,"/:cake","/蛋糕"),
                (54,"/:li","/闪电"),
                (55,"/:bome","/炸弹"),
                (56,"/:kn","/刀"),
                (57,"/:footb","/足球"),
                (117,"/:ladybug","/瓢虫"),
                (59,"/:shit","/便便"),
                (75,"/:moon","/月亮"),
                (74,"/:sun","/太阳"),
                (69,"/:gift","/礼物"),
                (49,"/:hug","/拥抱"),
                (76,"/:strong","/强"),
                (41,"/:shake","/发抖"),
                (86,"/:<O>","/怄火"),
                (125,"/:circle","/转圈"),
                (126,"/:kotow","/磕头"),
                (127,"/:turn","/回头"),
                (128,"/:skip","/跳绳"),
                (129,"/:oY","/投降"),
                (130,"/:#-0","/激动"),
                (131,"/:hiphot","/乱舞"),
                (132,"/:kiss","/献吻"),
                (133,"/:<&","/左太极"),
                (134,"/:&>","/右太极")]

    @staticmethod
    def get_face_qq(s):
        for f in msg.qq_face_value:
            if f[2] == "/" + s[1:-1]:
                return "[Face" + str(f[0]) + ".gif]"  

    @staticmethod
    def get_face_wx(s):
        for f in msg.qq_face_value:
            if str(f[0]) == s[5:-5]:
                return "[" + f[2][1:] + "]"  

    @staticmethod
    def replace_face_qq(m):
        return msg.get_face_wx(m.group(1))
    
    @staticmethod
    def cq_to_wx_face(text: str) -> str:
        # 构建编号 -> 微信名称 映射
        face_map = {f[0]: f[2][1:] for f in msg.qq_face_value}
        pattern = re.compile(r"\[CQ:face,id=(\d+)\]")
        def repl(m):
            num = int(m.group(1))
            if num in face_map:
                return f"[{face_map[num]}]"
            return ""   # 未匹配到则删除
        return pattern.sub(repl, text)

    @staticmethod
    def get_wx_msg(s):
        p = re.compile(r"(?P<face_qq>\[Face\d*.gif\])")
        return p.sub(msg.replace_face_qq, s)

    @staticmethod
    def replace_face_wx(m):
        return msg.get_face_qq(m.group(1))    

    @staticmethod
    def get_qq_msg(s):
        p = re.compile(r"(?P<face_wx>\[[\S\s]*?\])")
        return p.sub(msg.replace_face_wx, s)                

    #QQ表情
    @staticmethod
    def remove_qq_face(message):
        return re.sub(r"\[Face\d{1,3}.gif\]", "", message)

    #微信表情
    @staticmethod
    def remove_wx_face(message):
        return re.sub(r"\[[\S\s]*?\]", "", message)

    #emoji
    @staticmethod
    def remove_Emoji(message):
        return common.removeEmoji(message)
