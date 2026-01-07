using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace sz84.Bots.Models.PollSystem
{
    public class Poll
    {
        public Guid PollId { get; set; } = Guid.NewGuid();
        public long GroupId { get; set; }
        public long CreatorId { get; set; }
        public string Title { get; set; } = "";
        public bool IsMultiple { get; set; }
        public DateTime CreatedAt { get; set; } = DateTime.Now;
        public DateTime? ExpireAt { get; set; }
        public bool IsClosed { get; set; } = false;
        public List<PollOption> Options { get; set; } = new();
    }

    public class PollOption
    {
        public Guid OptionId { get; set; } = Guid.NewGuid();
        public Guid PollId { get; set; }
        public string Text { get; set; } = "";
    }

    public class PollVote
    {
        public Guid VoteId { get; set; } = Guid.NewGuid();
        public Guid PollId { get; set; }
        public Guid OptionId { get; set; }
        public long VoterId { get; set; }
        public DateTime VotedAt { get; set; } = DateTime.Now;
    }

    public static class PollStorage
    {
        public static List<Poll> Polls = new();
        public static List<PollVote> Votes = new();
    }
}
