using System;
using BotWorker.Domain.Interfaces;

namespace BotWorker.Application.Services
{
    public class SimpleGameService : ISimpleGameService
    {
        private readonly Random _random = new();

        private int RandomInt(int min, int max) => _random.Next(min, max + 1);

        public string RobBuilding(long userId)
        {
            int i = RandomInt(1, 19);
            return i switch
            {
                1 => "恭喜你，抽到了xxx积分",
                2 => "恭喜你，抽到了禁言卡",
                4 => "恭喜你，抽到了xxx经验",
                5 => "恭喜你，抽到了xxx经验",
                16 => "恭喜你，抽到解禁卡",
                18 => "恭喜你，抽到紫币卡",
                19 => "恭喜你，抽到xxx金币",
                _ => "很遗憾，抢楼失败"
            };
        }

        public string DaFeiji(long userId)
        {
            int i = RandomInt(1, 6);
            return i switch
            {
                1 => "拿着大炮打飞机，飞机跑掉了，损失xxx金币",
                2 => "拿着连环炮打飞机，一连打了好几架，获得xxx金币",
                3 => "左手只是辅助，右手才是关键，打飞机成功达到最高境界，舒服的享受，扣除纸巾费用xxx金币，获得社会威望50，享受期间不得离开，禁止游戏1分钟。",
                4 => "拿着手枪打飞机，飞机跑掉了，损失xxx金币",
                5 => "拿着火箭打飞机，一连下了好几架，获得金币，积分",
                _ => "打飞机失败"
            };
        }

        public string DaDishu(long userId)
        {
            int i = RandomInt(1, 4);
            return i switch
            {
                1 => "打地鼠成功，获得xxx金币",
                2 => "打地鼠失败，损失xxx金币",
                3 => "打地鼠平手",
                _ => "打地鼠失败"
            };
        }

        public string DaQunzhu(long userId)
        {
            int i = RandomInt(1, 4);
            return i switch
            {
                1 => "打群主成功，获得xxx金币",
                2 => "打群主失败，被群主反杀，损失xxx金币",
                3 => "打群主平手",
                _ => "打群主失败"
            };
        }

        public string QiangjiuQunzhu(long userId)
        {
            int i = RandomInt(1, 4);
            return i switch
            {
                1 => "抢救群主成功，获得xxx金币",
                2 => "抢救群主失败，群主驾鹤西去，损失xxx金币",
                3 => "抢救群主平手",
                _ => "抢救群主失败"
            };
        }

        public string AiQunzhu(long userId)
        {
            int i = RandomInt(1, 4);
            return i switch
            {
                1 => "爱群主成功，获得xxx金币",
                2 => "爱群主失败，群主不爱你，损失xxx金币",
                3 => "爱群主平手",
                _ => "爱群主失败"
            };
        }
    }
}
