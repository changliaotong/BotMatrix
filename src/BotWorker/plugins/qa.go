package plugins

import (
	"BotMatrix/common"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"log"
	"strings"
)

// QAPlugin 问答系统插件
type QAPlugin struct {
	cmdParser *CommandParser
}

func (p *QAPlugin) Name() string {
	return "qa"
}

func (p *QAPlugin) Description() string {
	return common.T("", "qa_plugin_desc|智能问答系统，提供常见问题解答和知识库查询")
}

func (p *QAPlugin) Version() string {
	return "1.0.0"
}

// NewQAPlugin 创建问答插件实例
func NewQAPlugin() *QAPlugin {
	return &QAPlugin{
		cmdParser: NewCommandParser(),
	}
}

func (p *QAPlugin) Init(robot plugin.Robot) {
	log.Println(common.T("", "qa_plugin_loaded|加载问答系统插件"))

	// 处理问答命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为问答命令
		if match, _ := p.cmdParser.MatchCommand(common.T("", "qa_cmd_match|问答|qa|知识库"), event.RawMessage); match {
			// 显示问答菜单
			menu := p.GetQAMenu()
			p.sendMessage(robot, event, menu)
			return nil
		}

		// 处理常见问题查询
		if answer := p.SearchQA(event.RawMessage); answer != "" {
			p.sendMessage(robot, event, answer)
			return nil
		}

		return nil
	})
}

// GetQAMenu 获取问答菜单
func (p *QAPlugin) GetQAMenu() string {
	menu := common.T("", "qa_menu_header|🤖 智能问答系统\n")
	menu += common.T("", "qa_menu_sep|====================\n\n")
	menu += common.T("", "qa_menu_faq_title| 常见问题:\n")
	menu += common.T("", "qa_menu_faq_1|1. 如何签到？\n")
	menu += common.T("", "qa_menu_faq_2|2. 如何查询积分？\n")
	menu += common.T("", "qa_menu_faq_3|3. 如何使用翻译功能？\n")
	menu += common.T("", "qa_menu_faq_4|4. 如何点歌？\n")
	menu += common.T("", "qa_menu_faq_5|5. 如何领养宠物？\n")
	menu += common.T("", "qa_menu_faq_6|6. 如何查询天气？\n")
	menu += common.T("", "qa_menu_faq_7|7. 如何使用抽签功能？\n")
	menu += common.T("", "qa_menu_faq_8|8. 如何查看排行榜？\n\n")
	menu += common.T("", "qa_menu_usage|💡 使用方法: 直接发送问题关键词，例如'如何签到'\n")
	menu += common.T("", "qa_menu_footer|ℹ️ 输入'问答'或'qa'查看此菜单")

	return menu
}

// SearchQA 搜索问答知识库
func (p *QAPlugin) SearchQA(question string) string {
	question = strings.ToLower(question)

	qaPairs := map[string]string{
		"如何签到":     common.T("", "qa_answer_signin|使用命令 /签到 或 /signin 进行每日签到，签到后可获得积分奖励。"),
		"签到":       common.T("", "qa_answer_signin|使用命令 /签到 或 /signin 进行每日签到，签到后可获得积分奖励。"),
		"如何查询积分":   common.T("", "qa_answer_points|使用命令 /积分 查询 或 /points 查看当前积分。"),
		"积分查询":     common.T("", "qa_answer_points|使用命令 /积分 查询 或 /points 查看当前积分。"),
		"如何使用翻译功能": common.T("", "qa_answer_translate|使用命令 /翻译 <文本> 或 /translate <文本> 进行中英文互译。"),
		"翻译":       common.T("", "qa_answer_translate|使用命令 /翻译 <文本> 或 /translate <文本> 进行中英文互译。"),
		"如何点歌":     common.T("", "qa_answer_music|使用命令 /点歌 <歌曲名称> 或 /music <歌曲名称> 搜索并播放指定歌曲。"),
		"点歌":       common.T("", "qa_answer_music|使用命令 /点歌 <歌曲名称> 或 /music <歌曲名称> 搜索并播放指定歌曲。"),
		"如何领养宠物":   common.T("", "qa_answer_adopt|使用命令 /领养 领养一只新宠物，领养后可以喂食、玩耍、洗澡。"),
		"领养宠物":     common.T("", "qa_answer_adopt|使用命令 /领养 领养一只新宠物，领养后可以喂食、玩耍、洗澡。"),
		"如何查询天气":   common.T("", "qa_answer_weather|使用命令 /天气 <城市名> 或 /weather <城市名> 查询指定城市的天气信息。"),
		"天气查询":     common.T("", "qa_answer_weather|使用命令 /天气 <城市名> 或 /weather <城市名> 查询指定城市的天气信息。"),
		"如何使用抽签功能": common.T("", "qa_answer_tarot|使用命令 /抽签 进行一次抽签，使用 /解签 <签文> 解析签文含义。"),
		"抽签":       common.T("", "qa_answer_tarot|使用命令 /抽签 进行一次抽签，使用 /解签 <签文> 解析签文含义。"),
		"如何查看排行榜":  common.T("", "qa_answer_rank|使用命令 /积分排行 或 /rank 查看积分排行榜。"),
		"排行榜":      common.T("", "qa_answer_rank|使用命令 /积分排行 或 /rank 查看积分排行榜。"),
		"帮助":       common.T("", "qa_answer_help|输入'菜单'或'help'查看所有命令，输入'问答'或'qa'查看问答系统菜单。"),
	}

	for q, a := range qaPairs {
		if strings.Contains(question, strings.ToLower(q)) {
			return a
		}
	}

	return ""
}

// sendMessage 发送消息
func (p *QAPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf(common.T("", "qa_send_failed|发送消息失败: %v\n"), err)
	}
}
