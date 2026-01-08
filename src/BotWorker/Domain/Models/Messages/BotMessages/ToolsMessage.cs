using BotWorker.Common;
using BotWorker.Common.Exts;
using BotWorker.Infrastructure.Persistence.ORM;
using sz84.Infrastructure.Tools;
using sz84.Bots.Users;

namespace BotWorker.Domain.Models.Messages.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        public async Task<string> GetWeatherResAsync(string cityName)
        {
            cityName = cityName.Replace("预报", "");
            if (cityName.IsNull())
            {
                cityName = User.CityName ?? "";
                return cityName == ""
                    ? "命令格式：天气 + 城市名 例如：\n天气 深圳\n设置城市 北京"
                    : await GetWeatherResAsync(cityName);
            }

            string? _city_name = cityName.RegexReplace(Regexs.Province, "");
            cityName = _city_name == "" ? cityName : _city_name;

            string res = await Weather.GetWeatherAsync(cityName);
            if (res != Weather.GetWhere("WeatherInfo", $"cityName = {cityName.Quotes()}", "Id desc"))
                Weather.Append(cityName, res);
            return res;
        }

        // 翻译
        public async Task<string> GetTranslateAsync()
        {
            string res = string.Empty;
            CmdPara = CmdPara.RemoveQqAds().Trim();

            if (CmdPara == "结束")
            {
                return UserInfo.SetState(UserInfo.States.Chat, UserId) == -1
                    ? RetryMsg
                    : "✅ 翻译服务结束！";
            }

            if (CmdPara.IsNull())
            {
                if (RealGroupId == 0 || IsPublic)
                {
                    int i = UserInfo.SetState(UserInfo.States.Translate, UserId);
                    return i == -1
                        ? RetryMsg
                        : "✅ 我已变身翻译，支持英日韩法俄西->中文 中文->英语";
                }
                return "命令格式：翻译 + 内容";
            }

            if (res.IsNull())
                res = await Translate.GetAzureResAsync(CmdPara);

            res = res.ReplaceInvalid();

            return res == ""
                ? "翻译服务暂时不能使用"
                : res + GetHintInfo(); 
        }
    }
}
