using System.Text;

namespace sz84.Agents.Entries;

public record VoiceItem(string Id, string Name, string PreviewUrl);

public record VoiceCategory(string Name, List<VoiceItem> Items);

public static class VoiceMap
{
    public static readonly List<VoiceCategory> Categories = new()
    {
        new("推荐", new()
        {
            new("lucy-voice-laibixiaoxin", "小新", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-laibixiaoxin.wav"),
            new("lucy-voice-houge", "猴哥", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-houge.wav"),
            new("lucy-voice-silang", "四郎", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-silang.wav"),
            new("lucy-voice-guangdong-f1", "东北老妹儿", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-guangdong-f1.wav"),
            new("lucy-voice-guangxi-m1", "广西大表哥", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-guangxi-m1.wav"),
            new("lucy-voice-daji", "妲己", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-daji.wav"),
            new("lucy-voice-lizeyan", "霸道总裁", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-lizeyan-2.wav"),
            new("lucy-voice-suxinjiejie", "酥心御姐", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-suxinjiejie.wav")
        }),

        new("搞怪", new()
        {
            new("lucy-voice-laibixiaoxin", "小新", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-laibixiaoxin.wav"),
            new("lucy-voice-houge", "猴哥", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-houge.wav"),
            new("lucy-voice-guangdong-f1", "东北老妹儿", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-guangdong-f1.wav"),
            new("lucy-voice-guangxi-m1", "广西大表哥", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-guangxi-m1.wav"),
            new("lucy-voice-m8", "说书先生", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-m8.wav"),
            new("lucy-voice-male1", "憨憨小弟", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-male1.wav"),
            new("lucy-voice-male3", "憨厚老哥", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-male3.wav")
        }),

        new("古风", new()
        {
            new("lucy-voice-daji", "妲己", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-daji.wav"),
            new("lucy-voice-silang", "四郎", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-silang.wav"),
            new("lucy-voice-lvbu", "吕布", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-lvbu.wav")
        }),

        new("现代", new()
        {
            new("lucy-voice-lizeyan", "霸道总裁", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-lizeyan-2.wav"),
            new("lucy-voice-suxinjiejie", "酥心御姐", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-suxinjiejie.wav"),
            new("lucy-voice-xueling", "元气少女", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-xueling.wav"),
            new("lucy-voice-f37", "文艺少女", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-f37.wav"),
            new("lucy-voice-male2", "磁性大叔", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-male2.wav"),
            new("lucy-voice-female1", "邻家小妹", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-female1.wav"),
            new("lucy-voice-m14", "低沉男声", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-m14.wav"),
            new("lucy-voice-f38", "傲娇少女", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-f38.wav"),
            new("lucy-voice-m101", "爹系男友", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-m101.wav"),
            new("lucy-voice-female2", "暖心姐姐", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-female2.wav"),
            new("lucy-voice-f36", "温柔妹妹", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-f36.wav"),
            new("lucy-voice-f34", "书香少女", "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-f34.wav")
        })
    };

    /// <summary>
    /// 所有语音（根据 Id 去重）
    /// </summary>
    public static readonly List<VoiceItem> All =
        [.. Categories
            .SelectMany(c => c.Items)
            .GroupBy(v => v.Id)
            .Select(g => g.First())];
}

public static class VoiceMapUtil
{
    /// <summary>
    /// 全部语音列表（顺序按 Categories + Items 保持稳定）
    /// </summary>
    private static readonly Lazy<List<VoiceItem>> _all = new(() =>
        VoiceMap.Categories
            .SelectMany(c => c.Items)
            .ToList()
    );

    public static List<VoiceItem> All => _all.Value;

    /// <summary>
    /// 名称 → ID 映射（精确匹配）
    /// </summary>
    private static readonly Lazy<Dictionary<string, string>> _nameToId =
        new(() => All
            .GroupBy(v => v.Name)
            .ToDictionary(g => g.Key, g => g.First().Id)
        );

    public static Dictionary<string, string> NameToId => _nameToId.Value;

    /// <summary>
    /// 构建编号列表（分组显示 + 稳定顺序）
    /// </summary>
    public static string BuildVoiceList(string? currentVoiceId = null)
    {
        var sb = new StringBuilder();
        sb.Append("🎙 可用语音列表：");

        int index = 1;
        foreach (var cat in VoiceMap.Categories)
        {
            sb.Append($"\n【{cat.Name}】\n");

            foreach (var v in cat.Items)
            {
                sb.Append($"{index}.{(!string.IsNullOrEmpty(currentVoiceId) && v.Id == currentVoiceId ? "✅" : "")}{v.Name} ");
                index++;
            }
        }
        return sb.ToString();
    }

    /// <summary>
    /// 根据编号获取语音
    /// </summary>
    public static (string Name, string Id)? FindByIndex(int index)
    {
        int idx = 1;
        foreach (var cat in VoiceMap.Categories)
        {
            foreach (var v in cat.Items)
            {
                if (idx == index)
                    return (v.Name, v.Id);
                idx++;
            }
        }
        return null;
    }

    /// <summary>
    /// 根据名称获取语音名
    /// </summary>
    public static string GetVoiceName(string id)
    {
        return All.FirstOrDefault(v => v.Id == id)?.Name ?? id;
    }
}

