package services

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"time"
)

type DailyFortune struct {
	Date        time.Time
	Overall     int
	Love        int
	Wealth      int
	Career      int
	Health      int
	Color       string
	LuckyNumber int
	Direction   string
	Taboo       string
	Comment     string
}

type FortuneService struct {
	colors       []string
	luckyNumbers []int
	directions   []string
	taboos       []string
}

func NewFortuneService() *FortuneService {
	return &FortuneService{
		colors:       []string{"珊瑚红", "天空蓝", "墨绿色", "靛青", "浅紫", "鹅黄", "藏青", "象牙白", "奶油色", "玫瑰金"},
		luckyNumbers: []int{1, 3, 5, 6, 7, 8, 9},
		directions:   []string{"正东", "正西", "正南", "正北", "东南", "西北", "东北", "西南"},
		taboos: []string{
			"避免与上级争论", "避免久坐久看手机", "切忌冲动消费", "勿轻信他人承诺", "忌讳外出远行",
			"今日不宜开始新计划", "避免熬夜", "小心交通安全", "远离是非之地", "少说多做",
		},
	}
}

func (s *FortuneService) GenerateFortune(qq string) *DailyFortune {
	// 复刻 C# 的 seed 生成逻辑
	dateStr := time.Now().Format("20060102")
	h := fnv.New32a()
	h.Write([]byte(qq + dateStr))
	seed := int64(h.Sum32())

	rng := rand.New(rand.NewSource(seed))

	fortune := &DailyFortune{
		Date:        time.Now(),
		Love:        rng.Intn(56) + 44, // 44-99
		Wealth:      rng.Intn(56) + 44,
		Career:      rng.Intn(56) + 44,
		Health:      rng.Intn(56) + 44,
		Color:       s.colors[rng.Intn(len(s.colors))],
		LuckyNumber: s.luckyNumbers[rng.Intn(len(s.luckyNumbers))],
		Direction:   s.directions[rng.Intn(len(s.directions))],
		Taboo:       s.taboos[rng.Intn(len(s.taboos))],
	}

	fortune.Overall = (fortune.Love + fortune.Wealth + fortune.Career + fortune.Health) / 4
	fortune.Comment = s.getComment(fortune.Overall)

	return fortune
}

func (s *FortuneService) getComment(score int) string {
	if score >= 90 {
		return "鸿运当头，万事大吉"
	}
	if score >= 70 {
		return "顺风顺水，小有收获"
	}
	if score >= 50 {
		return "平平稳稳，按部就班"
	}
	if score >= 30 {
		return "小心应对，略有波折"
	}
	return "事与愿违，宜静不宜动"
}

func (s *FortuneService) Format(fortune *DailyFortune) string {
	// 注意：这里的占位符 {农历月} 和 {农历日} 会由 PlaceholderService 进行第二轮解析
	return fmt.Sprintf("🔮 今日运势（{农历月}月{农历日}）\n"+
		"🌟 综合运势：%d / 100\n"+
		"✨ 福运评价：%s\n"+
		"❤️ 爱情运势：%d\n"+
		"💰 财富运势：%d\n"+
		"📚 事业运势：%d\n"+
		"💪 健康运势：%d\n"+
		"🎨 幸运颜色：%s\n"+
		"🔢 幸运数字：%d\n"+
		"🧭 幸运方向：%s\n"+
		"🙅‍♂️ 禁忌事项：%s\n",
		fortune.Overall, fortune.Comment, fortune.Love, fortune.Wealth,
		fortune.Career, fortune.Health, fortune.Color, fortune.LuckyNumber,
		fortune.Direction, fortune.Taboo)
}
