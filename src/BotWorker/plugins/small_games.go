package plugins

import (
	"BotMatrix/common"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
	"log"
	"math/rand"
	"strings"
)

// SmallGamesPlugin 小型游戏插件
type SmallGamesPlugin struct {
	cmdParser     *CommandParser
	gameStatus    map[string]bool     // 游戏状态，userID: 是否开启
	smallGamesMap map[string][]string // 小游戏命令到结果的映射
}

// NewSmallGamesPlugin 创建小型游戏插件实例
func NewSmallGamesPlugin() *SmallGamesPlugin {
	plugin := &SmallGamesPlugin{
		cmdParser:     NewCommandParser(),
		gameStatus:    make(map[string]bool),
		smallGamesMap: make(map[string][]string),
	}

	// 初始化小游戏命令和结果
	plugin.initSmallGames()

	return plugin
}

func (p *SmallGamesPlugin) Name() string {
	return "smallgames"
}

func (p *SmallGamesPlugin) Description() string {
	return common.T("", "small_games_plugin_desc|小型游戏插件，支持跳高楼、来抱抱、吃豆豆等多种趣味小游戏")
}

func (p *SmallGamesPlugin) Version() string {
	return "1.0.0"
}

// initSmallGames 初始化所有小游戏命令和结果
func (p *SmallGamesPlugin) initSmallGames() {
	// 跳高楼
	p.smallGamesMap["跳高楼"] = []string{
		common.T("", "sg_jump_res_1|你勇敢地跳过了一层楼，获得了10分！"),
		common.T("", "sg_jump_res_2|你轻松地跳过了两层楼，获得了20分！"),
		common.T("", "sg_jump_res_3|你挑战了三层楼，成功了！获得了30分！"),
		common.T("", "sg_jump_res_4|你尝试跳四层楼，失败了，摔了个狗啃泥！"),
		common.T("", "sg_jump_res_5|你极限挑战五层楼，成功了！大家都为你鼓掌！"),
	}

	// 来抱抱
	p.smallGamesMap["来抱抱"] = []string{
		common.T("", "sg_hug_res_1|你给了群主一个温暖的拥抱，群主很开心！"),
		common.T("", "sg_hug_res_2|你想要拥抱管理，管理害羞地躲开了！"),
		common.T("", "sg_hug_res_3|你和群员互相拥抱，氛围很温馨！"),
		common.T("", "sg_hug_res_4|你想拥抱空气，结果摔了一跤！"),
		common.T("", "sg_hug_res_5|你给了自己一个大大的拥抱，心情好多了！"),
	}

	// 吃豆豆
	p.smallGamesMap["吃豆豆"] = []string{
		common.T("", "sg_bean_res_1|你吃了一颗豆豆，味道不错！获得了5分！"),
		common.T("", "sg_bean_res_2|你吃了三颗豆豆，肚子有点饱了！获得了15分！"),
		common.T("", "sg_bean_res_3|你吃了五颗豆豆，变成了大胃王！获得了25分！"),
		common.T("", "sg_bean_res_4|你吃到了一颗坏豆豆，肚子疼了！"),
		common.T("", "sg_bean_res_5|你吃了十颗豆豆，成为了豆豆达人！"),
	}

	// 打鬼子
	p.smallGamesMap["打鬼子"] = []string{
		common.T("", "sg_devil_res_1|你勇敢地打死了一个鬼子，获得了100分！"),
		common.T("", "sg_devil_res_2|你打死了两个鬼子，被授予抗日小英雄称号！"),
		common.T("", "sg_devil_res_3|你打死了三个鬼子，大家都敬佩你！"),
		common.T("", "sg_devil_res_4|你没打中鬼子，反而被鬼子打了一枪！"),
		common.T("", "sg_devil_res_5|你使用了手榴弹，炸死了一群鬼子！"),
	}

	// 爱管理
	p.smallGamesMap["爱管理"] = []string{
		common.T("", "sg_love_admin_res_1|你给了管理一个爱心，管理很感动！"),
		common.T("", "sg_love_admin_res_2|你夸管理很漂亮/帅，管理开心地笑了！"),
		common.T("", "sg_love_admin_res_3|你帮管理整理了群公告，管理很感谢你！"),
		common.T("", "sg_love_admin_res_4|你想送管理礼物，但是管理拒绝了！"),
		common.T("", "sg_love_admin_res_5|你和管理成为了好朋友！"),
	}

	// 去约会
	p.smallGamesMap["去约会"] = []string{
		common.T("", "sg_date_res_1|你和约会对象度过了美好的一天！"),
		common.T("", "sg_date_res_2|你约会迟到了，对象有点生气！"),
		common.T("", "sg_date_res_3|你给对象买了花，对象很开心！"),
		common.T("", "sg_date_res_4|你和对象去看了电影，电影很好看！"),
		common.T("", "sg_date_res_5|你和对象吃了烛光晚餐，氛围浪漫极了！"),
	}

	// 抢包包
	p.smallGamesMap["抢包包"] = []string{
		common.T("", "sg_grab_bag_res_1|你成功抢到了一个包包，获得了50分！"),
		common.T("", "sg_grab_bag_res_2|你抢到了一个名牌包包，价值连城！"),
		common.T("", "sg_grab_bag_res_3|你抢包包时被发现了，赶紧跑！"),
		common.T("", "sg_grab_bag_res_4|你没抢到包包，反而被别人抢了！"),
		common.T("", "sg_grab_bag_res_5|你抢到了一个空包包，什么都没有！"),
	}

	// 数猪头
	p.smallGamesMap["数猪头"] = []string{
		common.T("", "sg_count_pig_res_1|你数了一个猪头，发现是群主！"),
		common.T("", "sg_count_pig_res_2|你数了两个猪头，一个是管理，一个是群员！"),
		common.T("", "sg_count_pig_res_3|你数了三个猪头，其中一个是你自己！"),
		common.T("", "sg_count_pig_res_4|你数错了猪头数量，被大家笑话了！"),
		common.T("", "sg_count_pig_res_5|你数了十个猪头，成为了猪头计数专家！"),
	}

	// 打地鼠
	p.smallGamesMap["打地鼠"] = []string{
		common.T("", "sg_whack_mole_res_1|你打中了一只地鼠，获得了10分！"),
		common.T("", "sg_whack_mole_res_2|你打中了三只地鼠，速度真快！获得了30分！"),
		common.T("", "sg_whack_mole_res_3|你打中了五只地鼠，成为了打地鼠高手！"),
		common.T("", "sg_whack_mole_res_4|你没打中地鼠，地鼠嘲笑你！"),
		common.T("", "sg_whack_mole_res_5|你打中了十只地鼠，打破了记录！"),
	}

	// 过马路
	p.smallGamesMap["过马路"] = []string{
		common.T("", "sg_cross_road_res_1|你安全地过了马路，获得了5分！"),
		common.T("", "sg_cross_road_res_2|你闯红灯过马路，被交警叔叔批评了！"),
		common.T("", "sg_cross_road_res_3|你帮助老奶奶过马路，获得了好人卡！"),
		common.T("", "sg_cross_road_res_4|你过马路时差点被车撞到，吓死了！"),
		common.T("", "sg_cross_road_res_5|你和朋友们一起过马路，说说笑笑！"),
	}

	// 吃面条
	p.smallGamesMap["吃面条"] = []string{
		common.T("", "sg_eat_noodle_res_1|你吃了一碗面条，味道好极了！"),
		common.T("", "sg_eat_noodle_res_2|你吃了两碗面条，肚子饱饱的！"),
		common.T("", "sg_eat_noodle_res_3|你吃面条时噎到了，赶紧喝水！"),
		common.T("", "sg_eat_noodle_res_4|你吃了一碗酸辣粉，辣得直冒汗！"),
		common.T("", "sg_eat_noodle_res_5|你吃了一碗阳春面，清淡又美味！"),
	}

	// 打小人
	p.smallGamesMap["打小人"] = []string{
		common.T("", "sg_beat_villain_res_1|你打了一个小人，心情好多了！"),
		common.T("", "sg_beat_villain_res_2|你打了两个小人，小人不敢再欺负你了！"),
		common.T("", "sg_beat_villain_res_3|你打小人时不小心打到了自己！"),
		common.T("", "sg_beat_villain_res_4|你打了三个小人，成为了打小人专家！"),
		common.T("", "sg_beat_villain_res_5|你打小人时被小人发现了，赶紧跑！"),
	}

	// 下象棋
	p.smallGamesMap["下象棋"] = []string{
		common.T("", "sg_chess_res_1|你和机器人下象棋，赢了！获得了50分！"),
		common.T("", "sg_chess_res_2|你和机器人下象棋，输了！继续努力！"),
		common.T("", "sg_chess_res_3|你和机器人下象棋，平局！棋艺不错！"),
		common.T("", "sg_chess_res_4|你下棋时走错了一步，导致满盘皆输！"),
		common.T("", "sg_chess_res_5|你下了一盘精彩的棋局，大家都围观！"),
	}

	// 发传单
	p.smallGamesMap["发传单"] = []string{
		common.T("", "sg_leaflet_res_1|你发了十张传单，获得了10分！"),
		common.T("", "sg_leaflet_res_2|你发了二十张传单，获得了20分！"),
		common.T("", "sg_leaflet_res_3|你发传单时被城管叔叔赶走了！"),
		common.T("", "sg_leaflet_res_4|你发了五十张传单，成为了发传单达人！"),
		common.T("", "sg_leaflet_res_5|你发传单时遇到了熟人，聊了一会儿！"),
	}

	// 打麻将
	p.smallGamesMap["打麻将"] = []string{
		common.T("", "sg_mahjong_res_1|你打麻将赢了一百块！"),
		common.T("", "sg_mahjong_res_2|你打麻将输了五十块！"),
		common.T("", "sg_mahjong_res_3|你摸到了一副好牌，杠上开花！"),
		common.T("", "sg_mahjong_res_4|你打麻将时点炮了，输掉了这局！"),
		common.T("", "sg_mahjong_res_5|你打麻将时糊了一把大的，赢得盆满钵满！"),
	}

	// 打色狼
	p.smallGamesMap["打色狼"] = []string{
		common.T("", "sg_pervert_res_1|你打了一个色狼，获得了正义使者称号！"),
		common.T("", "sg_pervert_res_2|你打了两个色狼，色狼们都害怕你了！"),
		common.T("", "sg_pervert_res_3|你打色狼时不小心打到了好人，赶紧道歉！"),
		common.T("", "sg_pervert_res_4|你打了三个色狼，成为了色狼克星！"),
		common.T("", "sg_pervert_res_5|你打色狼时被色狼反击了！"),
	}

	// 打土豪
	p.smallGamesMap["打土豪"] = []string{
		common.T("", "sg_tycoon_res_1|你打了一个土豪，获得了100分！"),
		common.T("", "sg_tycoon_res_2|你打土豪时抢到了一些金币！"),
		common.T("", "sg_tycoon_res_3|你打土豪时被土豪的保镖发现了！"),
		common.T("", "sg_tycoon_res_4|你打了两个土豪，成为了打土豪专家！"),
		common.T("", "sg_tycoon_res_5|你打土豪时得到了大家的支持！"),
	}

	// 开宝箱
	p.smallGamesMap["开宝箱"] = []string{
		common.T("", "sg_chest_res_1|你打开了一个宝箱，获得了50分！"),
		common.T("", "sg_chest_res_2|你打开了一个宝箱，获得了一件宝物！"),
		common.T("", "sg_chest_res_3|你打开了一个宝箱，里面是空的！"),
		common.T("", "sg_chest_res_4|你打开了一个宝箱，获得了100分！"),
		common.T("", "sg_chest_res_5|你打开了一个宝箱，获得了神秘礼物！"),
	}

	// 送外卖
	p.smallGamesMap["送外卖"] = []string{
		common.T("", "sg_delivery_res_1|你送了一份外卖，获得了10分！"),
		common.T("", "sg_delivery_res_2|你送外卖时迟到了，顾客有点生气！"),
		common.T("", "sg_delivery_res_3|你送了五份外卖，获得了50分！"),
		common.T("", "sg_delivery_res_4|你送外卖时遇到了暴雨，淋成了落汤鸡！"),
		common.T("", "sg_delivery_res_5|你送了十份外卖，成为了送外卖达人！"),
	}

	// 洗厕所
	p.smallGamesMap["洗厕所"] = []string{
		common.T("", "sg_toilet_res_1|你洗了一个厕所，获得了5分！"),
		common.T("", "sg_toilet_res_2|你洗了三个厕所，获得了15分！"),
		common.T("", "sg_toilet_res_3|你洗厕所时不小心滑倒了！"),
		common.T("", "sg_toilet_res_4|你洗了五个厕所，获得了25分！"),
		common.T("", "sg_toilet_res_5|你洗厕所洗得很干净，得到了表扬！"),
	}

	// 扫大街
	p.smallGamesMap["扫大街"] = []string{
		common.T("", "sg_street_res_1|你扫了一段大街，获得了5分！"),
		common.T("", "sg_street_res_2|你扫了整条大街，获得了20分！"),
		common.T("", "sg_street_res_3|你扫大街时捡到了一块钱！"),
		common.T("", "sg_street_res_4|你扫大街时遇到了熟人！"),
		common.T("", "sg_street_res_5|你扫大街扫得很干净，得到了清洁工阿姨的表扬！"),
	}

	// 晒太阳
	p.smallGamesMap["晒太阳"] = []string{
		common.T("", "sg_sun_res_1|你晒了一会儿太阳，感觉很舒服！"),
		common.T("", "sg_sun_res_2|你晒了太久太阳，皮肤有点晒伤了！"),
		common.T("", "sg_sun_res_3|你晒太阳时睡着了，做了一个好梦！"),
		common.T("", "sg_sun_res_4|你晒太阳时遇到了一只小猫！"),
		common.T("", "sg_sun_res_5|你晒太阳时看了一本书，很惬意！"),
	}

	// 看AV
	p.smallGamesMap["看AV"] = []string{
		common.T("", "sg_av_res_1|你看了一会儿AV，感觉有点不好意思！"),
		common.T("", "sg_av_res_2|你看AV时被家人发现了，赶紧关掉！"),
		common.T("", "sg_av_res_3|你看了一部好看的AV，回味无穷！"),
		common.T("", "sg_av_res_4|你看AV时网络断了，很扫兴！"),
		common.T("", "sg_av_res_5|你决定不看AV了，去做更有意义的事情！"),
	}

	// 揍群主
	p.smallGamesMap["揍群主"] = []string{
		common.T("", "sg_beat_owner_res_1|你揍了群主一拳，群主很生气！"),
		common.T("", "sg_beat_owner_res_2|你揍群主时被管理发现了，管理要踢你出群！"),
		common.T("", "sg_beat_owner_res_3|你轻轻揍了群主一下，群主没在意！"),
		common.T("", "sg_beat_owner_res_4|你揍群主时群主反击了，你被揍得很惨！"),
		common.T("", "sg_beat_owner_res_5|你和群主开玩笑揍了他一下，大家都笑了！"),
	}

	// 打管理
	p.smallGamesMap["打管理"] = []string{
		common.T("", "sg_beat_admin_res_1|你打了管理一下，管理很生气！"),
		common.T("", "sg_beat_admin_res_2|你打管理时被群主发现了，群主批评了你！"),
		common.T("", "sg_beat_admin_res_3|你轻轻打了管理一下，管理没在意！"),
		common.T("", "sg_beat_admin_res_4|你打管理时管理反击了，你被揍得很惨！"),
		common.T("", "sg_beat_admin_res_5|你和管理开玩笑打了他一下，大家都笑了！"),
	}

	// 揍群员
	p.smallGamesMap["揍群员"] = []string{
		common.T("", "sg_beat_member_res_1|你揍了群员一下，群员很生气！"),
		common.T("", "sg_beat_member_res_2|你揍群员时被管理发现了，管理警告了你！"),
		common.T("", "sg_beat_member_res_3|你轻轻揍了群员一下，群员没在意！"),
		common.T("", "sg_beat_member_res_4|你揍群员时群员反击了，你被揍得很惨！"),
		common.T("", "sg_beat_member_res_5|你和群员开玩笑揍了他一下，大家都笑了！"),
	}

	// 抢老公
	p.smallGamesMap["抢老公"] = []string{
		common.T("", "sg_grab_husband_res_1|你成功抢到了一个老公！"),
		common.T("", "sg_grab_husband_res_2|你抢老公时被原配发现了，原配很生气！"),
		common.T("", "sg_grab_husband_res_3|你没抢到老公，反而被老公拒绝了！"),
		common.T("", "sg_grab_husband_res_4|你抢到了一个好老公，幸福极了！"),
		common.T("", "sg_grab_husband_res_5|你抢老公时遇到了竞争对手！"),
	}

	// 抢老婆
	p.smallGamesMap["抢老婆"] = []string{
		common.T("", "sg_grab_wife_res_1|你成功抢到了一个老婆！"),
		common.T("", "sg_grab_wife_res_2|你抢老婆时被原配发现了，原配很生气！"),
		common.T("", "sg_grab_wife_res_3|你没抢到老婆，反而被老婆拒绝了！"),
		common.T("", "sg_grab_wife_res_4|你抢到了一个好老婆，幸福极了！"),
		common.T("", "sg_grab_wife_res_5|你抢老婆时遇到了竞争对手！"),
	}

	// 抢情人
	p.smallGamesMap["抢情人"] = []string{
		common.T("", "sg_grab_lover_res_1|你成功抢到了一个情人！"),
		common.T("", "sg_grab_lover_res_2|你抢情人时被发现了，赶紧跑！"),
		common.T("", "sg_grab_lover_res_3|你没抢到情人，反而被情人拒绝了！"),
		common.T("", "sg_grab_lover_res_4|你抢到了一个好情人，开心极了！"),
		common.T("", "sg_grab_lover_res_5|你抢情人时遇到了竞争对手！"),
	}

	// 打老公
	p.smallGamesMap["打老公"] = []string{
		common.T("", "sg_beat_husband_res_1|你打了老公一下，老公很生气！"),
		common.T("", "sg_beat_husband_res_2|你打老公时老公反击了，你被揍得很惨！"),
		common.T("", "sg_beat_husband_res_3|你轻轻打了老公一下，老公没在意！"),
		common.T("", "sg_beat_husband_res_4|你打老公时老公道歉了，你们和好了！"),
		common.T("", "sg_beat_husband_res_5|你和老公开玩笑打了他一下，大家都笑了！"),
	}

	// 打老婆
	p.smallGamesMap["打老婆"] = []string{
		common.T("", "sg_beat_wife_res_1|你打了老婆一下，老婆很生气！"),
		common.T("", "sg_beat_wife_res_2|你打老婆时老婆反击了，你被揍得很惨！"),
		common.T("", "sg_beat_wife_res_3|你轻轻打了老婆一下，老婆没在意！"),
		common.T("", "sg_beat_wife_res_4|你打老婆时老婆哭了，你赶紧道歉！"),
		common.T("", "sg_beat_wife_res_5|你和老婆开玩笑打了她一下，大家都笑了！"),
	}

	// 打小三
	p.smallGamesMap["打小三"] = []string{
		common.T("", "sg_beat_mistress_res_1|你打了小三一下，小三很生气！"),
		common.T("", "sg_beat_mistress_res_2|你打小三时小三反击了，你被揍得很惨！"),
		common.T("", "sg_beat_mistress_res_3|你打了小三一顿，小三不敢再来了！"),
		common.T("", "sg_beat_mistress_res_4|你打小三时被老公/老婆发现了，大家都很尴尬！"),
		common.T("", "sg_beat_mistress_res_5|你成功赶走了小三，保卫了家庭！"),
	}

	// 打飞机
	p.smallGamesMap["打飞机"] = []string{
		common.T("", "sg_plane_res_1|你打了一架飞机，获得了50分！"),
		common.T("", "sg_plane_res_2|你打了两架飞机，获得了100分！"),
		common.T("", "sg_plane_res_3|你打飞机时没打中，飞机飞走了！"),
		common.T("", "sg_plane_res_4|你打了三架飞机，获得了150分！"),
		common.T("", "sg_plane_res_5|你打飞机时被发现了，赶紧跑！"),
	}

	// 爬围墙
	p.smallGamesMap["爬围墙"] = []string{
		common.T("", "sg_climb_wall_res_1|你成功爬上了围墙，获得了20分！"),
		common.T("", "sg_climb_wall_res_2|你爬围墙时不小心摔了下来，疼死了！"),
		common.T("", "sg_climb_wall_res_3|你爬围墙时被保安发现了，赶紧跑！"),
		common.T("", "sg_climb_wall_res_4|你轻松爬上了围墙，身手敏捷！"),
		common.T("", "sg_climb_wall_res_5|你爬围墙时遇到了一只猫！"),
	}

	// 去跳舞
	p.smallGamesMap["去跳舞"] = []string{
		common.T("", "sg_dance_res_1|你去跳了一会儿舞，感觉很开心！"),
		common.T("", "sg_dance_res_2|你跳舞时不小心踩到了别人的脚！"),
		common.T("", "sg_dance_res_3|你跳了一首很流行的舞，大家都为你鼓掌！"),
		common.T("", "sg_dance_res_4|你跳舞时遇到了一个舞伴！"),
		common.T("", "sg_dance_res_5|你跳舞跳得很累，但是很开心！"),
	}

	// 做好事
	p.smallGamesMap["做好事"] = []string{
		common.T("", "sg_good_deed_res_1|你做了一件好事，获得了10分！"),
		common.T("", "sg_good_deed_res_2|你做了三件好事，获得了30分！"),
		common.T("", "sg_good_deed_res_3|你帮助了一位老奶奶过马路，老奶奶很感谢你！"),
		common.T("", "sg_good_deed_res_4|你捡到了一块钱，交给了警察叔叔！"),
		common.T("", "sg_good_deed_res_5|你做了五件好事，获得了50分！"),
	}

	// 逛公园
	p.smallGamesMap["逛公园"] = []string{
		common.T("", "sg_park_res_1|你逛了一会儿公园，感觉很舒服！"),
		common.T("", "sg_park_res_2|你逛公园时遇到了一只小狗！"),
		common.T("", "sg_park_res_3|你逛公园时看到了美丽的风景！"),
		common.T("", "sg_park_res_4|你逛公园时遇到了熟人！"),
		common.T("", "sg_park_res_5|你逛公园时买了一根冰淇淋，很好吃！"),
	}

	// 打土匪
	p.smallGamesMap["打土匪"] = []string{
		common.T("", "sg_bandit_res_1|你打了一个土匪，获得了50分！"),
		common.T("", "sg_bandit_res_2|你打了两个土匪，获得了100分！"),
		common.T("", "sg_bandit_res_3|你打土匪时被土匪发现了，赶紧跑！"),
		common.T("", "sg_bandit_res_4|你打了三个土匪，获得了150分！"),
		common.T("", "sg_bandit_res_5|你成功消灭了一群土匪，获得了200分！"),
	}

	// 斗地主
	p.smallGamesMap["斗地主"] = []string{
		common.T("", "sg_landlord_res_1|你斗地主赢了，获得了50分！"),
		common.T("", "sg_landlord_res_2|你斗地主输了，失去了20分！"),
		common.T("", "sg_landlord_res_3|你斗地主时摸到了一手好牌，轻松赢了！"),
		common.T("", "sg_landlord_res_4|你斗地主时被对手赢了，有点不甘心！"),
		common.T("", "sg_landlord_res_5|你斗地主时成为了地主，但是输了！"),
	}

	// 逛商场
	p.smallGamesMap["逛商场"] = []string{
		common.T("", "sg_mall_res_1|你逛了一会儿商场，什么都没买！"),
		common.T("", "sg_mall_res_2|你逛商场时买了一件衣服，很喜欢！"),
		common.T("", "sg_mall_res_3|你逛商场时遇到了促销活动，买了很多东西！"),
		common.T("", "sg_mall_res_4|你逛商场时看到了一件喜欢的东西，但是太贵了！"),
		common.T("", "sg_mall_res_5|你逛商场时累了，找了个地方休息！"),
	}

	// 打汉奸
	p.smallGamesMap["打汉奸"] = []string{
		common.T("", "sg_traitor_res_1|你打了一个汉奸，获得了100分！"),
		common.T("", "sg_traitor_res_2|你打了两个汉奸，获得了200分！"),
		common.T("", "sg_traitor_res_3|你打汉奸时被汉奸发现了，赶紧跑！"),
		common.T("", "sg_traitor_res_4|你成功消灭了一群汉奸，获得了300分！"),
		common.T("", "sg_traitor_res_5|你打汉奸时得到了大家的支持！"),
	}

	// 挑大粪
	p.smallGamesMap["挑大粪"] = []string{
		common.T("", "sg_manure_res_1|你挑了一担大粪，获得了5分！"),
		common.T("", "sg_manure_res_2|你挑大粪时不小心洒了出来，臭死了！"),
		common.T("", "sg_manure_res_3|你挑了三担大粪，获得了15分！"),
		common.T("", "sg_manure_res_4|你挑大粪时累得满头大汗！"),
		common.T("", "sg_manure_res_5|你挑了五担大粪，获得了25分！"),
	}

	// 挖宝藏
	p.smallGamesMap["挖宝藏"] = []string{
		common.T("", "sg_treasure_res_1|你挖了一会儿，没找到宝藏！"),
		common.T("", "sg_treasure_res_2|你挖了一会儿，找到了一些金币！"),
		common.T("", "sg_treasure_res_3|你挖了很久，终于找到了宝藏！获得了100分！"),
		common.T("", "sg_treasure_res_4|你挖宝藏时不小心挖到了石头，手很疼！"),
		common.T("", "sg_treasure_res_5|你找到了一个大宝藏，发财了！"),
	}

	// 挖宝
	p.smallGamesMap["挖宝"] = []string{
		common.T("", "sg_dig_res_1|你挖了一会儿，没找到宝物！"),
		common.T("", "sg_dig_res_2|你挖了一会儿，找到了一件宝物！获得了50分！"),
		common.T("", "sg_dig_res_3|你挖了很久，终于找到了一件珍贵的宝物！"),
		common.T("", "sg_dig_res_4|你挖宝时不小心挖到了别人的东西！"),
		common.T("", "sg_dig_res_5|你找到了一件价值连城的宝物，发财了！"),
	}

	// 午安
	p.smallGamesMap["午安"] = []string{
		common.T("", "sg_noon_res_1|午安！祝你有个美好的下午！"),
		common.T("", "sg_noon_res_2|午安！记得午休哦！"),
		common.T("", "sg_noon_res_3|午安！下午也要加油！"),
		common.T("", "sg_noon_res_4|午安！希望你下午一切顺利！"),
		common.T("", "sg_noon_res_5|午安！好好休息，下午才有精神！"),
	}

	// 晚安
	p.smallGamesMap["晚安"] = []string{
		common.T("", "sg_night_res_1|晚安！祝你做个好梦！"),
		common.T("", "sg_night_res_2|晚安！早点睡哦！"),
		common.T("", "sg_night_res_3|晚安！明天见！"),
		common.T("", "sg_night_res_4|晚安！好好休息！"),
		common.T("", "sg_night_res_5|晚安！希望你明天有个好心情！"),
	}

	// 早安
	p.smallGamesMap["早安"] = []string{
		common.T("", "sg_morning_res_1|早安！新的一天开始了！"),
		common.T("", "sg_morning_res_2|早安！祝你今天一切顺利！"),
		common.T("", "sg_morning_res_3|早安！记得吃早餐哦！"),
		common.T("", "sg_morning_res_4|早安！今天也要加油！"),
		common.T("", "sg_morning_res_5|早安！美好的一天从现在开始！"),
	}

	// 打怪
	p.smallGamesMap["打怪"] = []string{
		common.T("", "sg_monster_res_1|你打了一个怪，获得了30分！"),
		common.T("", "sg_monster_res_2|你打了两个怪，获得了60分！"),
		common.T("", "sg_monster_res_3|你打了三个怪，获得了90分！"),
		common.T("", "sg_monster_res_4|你打怪时没打中，怪跑了！"),
		common.T("", "sg_monster_res_5|你成功消灭了一群怪，获得了150分！"),
	}

	// 大家好
	p.smallGamesMap["大家好"] = []string{
		common.T("", "sg_hello_res_1|大家好！很高兴见到你们！"),
		common.T("", "sg_hello_res_2|大家好！今天天气真好！"),
		common.T("", "sg_hello_res_3|大家好！有什么好玩的事情吗？"),
		common.T("", "sg_hello_res_4|大家好！我是新来的，请多多关照！"),
		common.T("", "sg_hello_res_5|大家好！今天心情不错！"),
	}

	// 打篮球
	p.smallGamesMap["打篮球"] = []string{
		common.T("", "sg_basketball_res_1|你打了一会儿篮球，感觉很舒服！"),
		common.T("", "sg_basketball_res_2|你打篮球时投进了一个三分球，大家都为你鼓掌！"),
		common.T("", "sg_basketball_res_3|你打篮球时不小心摔了一跤！"),
		common.T("", "sg_basketball_res_4|你打篮球时和朋友们玩得很开心！"),
		common.T("", "sg_basketball_res_5|你打篮球时累得满头大汗！"),
	}

	// 我来了
	p.smallGamesMap["我来了"] = []string{
		common.T("", "sg_come_res_1|我来了！大家想我了吗？"),
		common.T("", "sg_come_res_2|我来了！有什么好玩的事情吗？"),
		common.T("", "sg_come_res_3|我来了！今天天气真好！"),
		common.T("", "sg_come_res_4|我来了！大家好！"),
		common.T("", "sg_come_res_5|我来了！准备好和我一起玩了吗？"),
	}

	// 收网
	p.smallGamesMap["收网"] = []string{
		common.T("", "sg_collect_net_res_1|你收网时捕到了一条鱼，获得了20分！"),
		common.T("", "sg_collect_net_res_2|你收网时捕到了两条鱼，获得了40分！"),
		common.T("", "sg_collect_net_res_3|你收网时什么都没捕到，有点失望！"),
		common.T("", "sg_collect_net_res_4|你收网时捕到了一条大鱼，获得了50分！"),
		common.T("", "sg_collect_net_res_5|你收网时捕到了很多鱼，获得了100分！"),
	}

	// 游泳
	p.smallGamesMap["游泳"] = []string{
		common.T("", "sg_swim_res_1|你游了一会儿泳，感觉很舒服！"),
		common.T("", "sg_swim_res_2|你游泳时不小心呛了水！"),
		common.T("", "sg_swim_res_3|你游了一公里，获得了50分！"),
		common.T("", "sg_swim_res_4|你游泳时遇到了一条鱼！"),
		common.T("", "sg_swim_res_5|你游泳时累了，上岸休息！"),
	}

	// 撒网
	p.smallGamesMap["撒网"] = []string{
		common.T("", "sg_cast_net_res_1|你撒网撒得很好，准备捕鱼！"),
		common.T("", "sg_cast_net_res_2|你撒网时不小心把网弄坏了！"),
		common.T("", "sg_cast_net_res_3|你撒了一张大网，希望能捕到很多鱼！"),
		common.T("", "sg_cast_net_res_4|你撒网时遇到了大风，网被吹跑了！"),
		common.T("", "sg_cast_net_res_5|你撒网撒得很完美，就等收网了！"),
	}

	// 爬山
	p.smallGamesMap["爬山"] = []string{
		common.T("", "sg_climb_mountain_res_1|你爬了一会儿山，感觉很舒服！"),
		common.T("", "sg_climb_mountain_res_2|你爬山时累了，找了个地方休息！"),
		common.T("", "sg_climb_mountain_res_3|你成功爬到了山顶，获得了100分！"),
		common.T("", "sg_climb_mountain_res_4|你爬山时看到了美丽的风景！"),
		common.T("", "sg_climb_mountain_res_5|你爬山时不小心摔了一跤！"),
	}
}

func (p *SmallGamesPlugin) Init(robot plugin.Robot) {
	log.Println(common.T("", "sg_plugin_loaded|加载小型游戏插件"))

	robot.OnMessage(func(event *onebot.Event) error {
		msg := event.RawMessage
		userID := fmt.Sprintf("%d", event.UserID)

		// 开启小游戏命令
		if match, _ := p.cmdParser.MatchCommand("开启小游戏", msg); match {
			p.gameStatus[userID] = true
			p.sendMessage(robot, event, common.T("", "sg_game_enabled|✅ 小游戏功能已开启！你可以输入“小游戏”查看可用游戏列表。"))
			return nil
		}

		// 关闭小游戏命令
		if match, _ := p.cmdParser.MatchCommand("关闭小游戏", msg); match {
			p.gameStatus[userID] = false
			p.sendMessage(robot, event, common.T("", "sg_game_disabled|❌ 小游戏功能已关闭！"))
			return nil
		}

		// 查看小游戏列表命令
		if match, _ := p.cmdParser.MatchCommand("小游戏", msg); match {
			if !p.gameStatus[userID] {
				p.sendMessage(robot, event, common.T("", "sg_game_not_enabled|⚠️ 你还没有开启小游戏功能，请输入“开启小游戏”来开启。"))
				return nil
			}

			var sb strings.Builder
			sb.WriteString(common.T("", "sg_list_header|🎮 可用小游戏列表：\n"))
			for cmd := range p.smallGamesMap {
				sb.WriteString(fmt.Sprintf("- %s\n", cmd))
			}
			sb.WriteString(common.T("", "sg_list_footer|\n输入游戏名称即可开始游戏！"))
			p.sendMessage(robot, event, sb.String())
			return nil
		}

		// 处理所有小游戏命令
		if p.gameStatus[userID] {
			for cmd, results := range p.smallGamesMap {
				if match, _ := p.cmdParser.MatchCommand(cmd, msg); match {
					res := results[rand.Intn(len(results))]
					p.sendMessage(robot, event, res)
					return nil
				}
			}
		}

		return nil
	})
}

// sendMessage 发送消息
func (p *SmallGamesPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}
