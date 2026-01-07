using BotWorker.Common;
using sz84.Core.MetaDatas;

namespace sz84.Bots.Games
{
    internal class SimpleGame : MetaData<SimpleGame>
    {
        public override string TableName => throw new NotImplementedException();

        public override string KeyField => throw new NotImplementedException();

        /// <summary>
        /// 抢楼游戏      
        /// </summary>
        /// <param name="qq"></param>
        /// <returns></returns>
        public static string RobBuilding(long qq)
        {
            int i = RandomInt(1, 19);
            return i switch
            {
                1 => $"恭喜你，抽到了xxx积分",
                2 => $"恭喜你，抽到了禁言卡",
                4 => $"恭喜你，抽到了xxx经验",
                5 => $"恭喜你，抽到了xxx经验",
                16 => $"恭喜你，抽到解禁卡",
                18 => $"恭喜你，抽到紫币卡",
                19 => $"恭喜你，抽到xxx金币",
                _ => $"很遗憾，抢楼失败"
            };
        }

        /*
         * 打飞机
         */

        public static string DaFeiji(long qq)
        {
            int i = RandomInt(1, 6);
            return i switch
            {
                1 => $"拿着大炮打飞机，飞机跑掉了，损失xxx金币",
                2 => $"拿着连环炮打飞机，一连打了好几架，获得xxx金币",
                3 => $"左手只是辅助，右手才是关键，打飞机成功达到最高境界，舒服的享受，扣除纸巾费用xxx金币，获得社会威望50，享受期间不得离开，禁止游戏1分钟。",
                4 => $"拿着手枪打飞机，飞机跑掉了，损失xxx金币",
                5 => $"拿着火箭打飞机，一连下了好几架，获得金币，积分",
                6 => $"左手只是辅助，右手才是关键，打飞机成功达到最高境界，舒服的享受，扣除纸巾费用xxx金币，获得社会威望50，享受期间不得离开，禁止游戏1分钟。",
                _ => $""
            };
        }


        //打地鼠
        public static string DaDishu(long qq)
        {
            int i = RandomInt(1, 5);
            return i switch
            {
                1 => $"拿着锤子打地鼠，本次打死0个地鼠，损失xx金币",
                2 => $"赤手空拳打地鼠，反而被地鼠打死，损失金币xx复活",
                3 => $"拿着锤子打地鼠，锤子都被地鼠抢了，损失金币xx",
                4 => $"拿着锤子打地鼠，你大喊我操你妈，顿时血槽全满，地鼠都被打死了，获得金币xxx",
                5 => $"拿着锤子打地鼠，你大喊春哥我爱你，顿时血槽全满，地鼠都被打死了，获得xxx金币",
                _ => $""
            };
        }

        //打群主
        public static string DaQunzhu(long qq)
        {
            int i = RandomInt(1, 5);
            return i switch
            {
                1 => $"✅ 打群主成功",
                2 => $"你今天已经打过群主了",
                3 => $"✅ 恭喜你获得[取:随机数, 1] 张刮卡",
                4 => $"✅ 恭喜你获得500金币",
                _ => $""
            };
        }

        //抢救群主
        public static string QiangjiuQunzhu(long qq)
        {
            int i = RandomInt(1, 5);
            return i switch
            {
                1 => $"✅ 抢救群主成功",
                2 => $"✅ 你今天已经抢救过群主了",
                3 => $"✅ 恭喜你获得[取:随机数, 1] 张刮卡",
                4 => $"✅ 恭喜你获得500金币",
                _ => $""
            };
        }

        //爱群主
        public static string AiQunzhu(long qq)
        {
            int i = RandomInt(1, 5);
            return i switch
            {
                1 => $"✅ 群主非常高兴",
                2 => $"✅ 你今天已经爱过了",
                3 => $"✅ 恭喜你获得[取:随机数, 1] 张刮卡",
                4 => $"✅ 恭喜你获得500金币",
                _ => $""
            };
        }

        /*
        送群主进棺材
        成功将群主送入棺材，大家默哀群主一分钟
        你今天已经送过群主进棺材了
        恭喜你获得[取:随机数,1]张刮卡
        恭喜你获得500金币
        
        砍群主
        砍群主成功
        你今天已经砍过群主了
        恭喜你获得[取:随机数,1]张刮卡
        恭喜你获得500金币
        
        送群主去土里
        送群主去土里成功
        你今天已经送过群主去土里了
        恭喜你获得[取:随机数,1]张刮卡
        恭喜你获得500金币

        (群主最伟大了|群主最伟大|我爱群主)
        本群：[group]
        QQ：[qq]
        获得经验
        快看看多了还是少了:金币，积分，刮刮卡
        群主最伟大惹窝啦


         */

    }

}
