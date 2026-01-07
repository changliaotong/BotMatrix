using System.Text.RegularExpressions;

namespace sz84.Bots.Models.PollSystem
{
    public class PollService
    {
        private readonly Func<long, string, Task> SendGroupMessage;
        private readonly Func<long, long, bool> IsGroupAdmin;

        public PollService(
            Func<long, string, Task> sendGroupMessage,
            Func<long, long, bool> isGroupAdmin)
        {
            SendGroupMessage = sendGroupMessage;
            IsGroupAdmin = isGroupAdmin;
        }

        public async Task HandleGroupMessage(long groupId, long userId, string message)
        {
            if (!message.StartsWith("#投票")) return;

            var content = message[3..].Trim();

            if (content.StartsWith("投="))
            {
                await HandleVoteAsync(groupId, userId, content);
            }
            else if (content.StartsWith("结果"))
            {
                await HandleShowResultAsync(groupId);
            }
            else
            {
                await HandleCreatePollAsync(groupId, userId, content);
            }
        }

        private async Task HandleCreatePollAsync(long groupId, long userId, string content)
        {
            if (!IsGroupAdmin(groupId, userId))
            {
                await SendGroupMessage(groupId, "❌ 只有群主或管理员可以发起投票。");
                return;
            }

            string title = Extract(content, "标题");
            string optionStr = Extract(content, "选项");
            string type = Extract(content, "类型");
            string deadline = Extract(content, "截止");

            if (string.IsNullOrEmpty(title) || string.IsNullOrEmpty(optionStr))
            {
                await SendGroupMessage(groupId, "❌ 投票内容格式错误，请提供标题和选项。");
                return;
            }

            var options = optionStr.Split('|', StringSplitOptions.RemoveEmptyEntries)
                .Select(o => new PollOption { Text = o.Trim() })
                .ToList();

            if (options.Count < 2)
            {
                await SendGroupMessage(groupId, "❌ 至少需要两个选项。");
                return;
            }

            var poll = new Poll
            {
                GroupId = groupId,
                CreatorId = userId,
                Title = title,
                IsMultiple = type.Contains('多'),
                ExpireAt = ParseDeadline(deadline),
                Options = options
            };

            foreach (var opt in poll.Options)
            {
                opt.PollId = poll.PollId;
            }

            PollStorage.Polls.Add(poll);

            string reply = $"📢 新投票：{poll.Title}\n";
            for (int i = 0; i < poll.Options.Count; i++)
            {
                reply += $"{i + 1}. {poll.Options[i].Text}\n";
            }

            reply += "\n📝 回复：#投票 投=编号";
            await SendGroupMessage(groupId, reply);
        }

        private async Task HandleVoteAsync(long groupId, long userId, string content)
        {
            var poll = PollStorage.Polls.LastOrDefault(p => p.GroupId == groupId && !p.IsClosed);
            if (poll == null)
            {
                await SendGroupMessage(groupId, "❌ 当前没有活跃投票。");
                return;
            }

            if (poll.ExpireAt.HasValue && poll.ExpireAt < DateTime.Now)
            {
                poll.IsClosed = true;
                await SendGroupMessage(groupId, "🕒 投票已结束。");
                return;
            }

            if (PollStorage.Votes.Any(v => v.PollId == poll.PollId && v.VoterId == userId))
            {
                await SendGroupMessage(groupId, "❌ 你已经投过票了！");
                return;
            }

            var indexStr = content.Substring(3).Trim();
            if (!int.TryParse(indexStr, out int index) || index < 1 || index > poll.Options.Count)
            {
                await SendGroupMessage(groupId, "❌ 无效选项编号。");
                return;
            }

            var selected = poll.Options[index - 1];
            PollStorage.Votes.Add(new PollVote
            {
                PollId = poll.PollId,
                OptionId = selected.OptionId,
                VoterId = userId
            });

            await SendGroupMessage(groupId, $"✅ 投票成功，你选择了：{selected.Text}");
        }

        private async Task HandleShowResultAsync(long groupId)
        {
            var poll = PollStorage.Polls.LastOrDefault(p => p.GroupId == groupId);
            if (poll == null)
            {
                await SendGroupMessage(groupId, "❌ 没有找到相关投票。");
                return;
            }

            var result = poll.Options
                .Select(opt => new
                {
                    Text = opt.Text,
                    Count = PollStorage.Votes.Count(v => v.OptionId == opt.OptionId)
                })
                .ToList();

            int total = result.Sum(r => r.Count);
            string msg = $"📊 投票结果：{poll.Title}\n";

            foreach (var r in result)
            {
                int percent = total > 0 ? (int)(r.Count * 100.0 / total) : 0;
                msg += $"{r.Text}：{r.Count}票（{percent}%）\n";
            }

            await SendGroupMessage(groupId, msg);
        }

        private static string Extract(string input, string key)
        {
            var match = Regex.Match(input, $@"{key}=([^\s]+)");
            return match.Success ? match.Groups[1].Value : "";
        }

        private static DateTime? ParseDeadline(string input)
        {
            if (string.IsNullOrWhiteSpace(input)) return null;
            if (input.EndsWith("小时") && int.TryParse(input[..^2], out int h))
                return DateTime.Now.AddHours(h);
            if (input.EndsWith("分钟") && int.TryParse(input[..^2], out int m))
                return DateTime.Now.AddMinutes(m);
            return null;
        }
    }

}
