using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Common;
using BotWorker.Common.Extensions;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Repositories;

namespace BotWorker.Modules.Games
{
    public class QuestionInfoService : IQuestionInfoService
    {
        private readonly IQuestionInfoRepository _repository;

        public QuestionInfoService(IQuestionInfoRepository repository)
        {
            _repository = repository;
        }

        public async Task<bool> ExistsByQuestionAsync(string question)
        {
            return await _repository.ExistsByQuestionAsync(question);
        }

        public async Task<long> AddQuestionAsync(long botUin, long groupId, long userId, string question)
        {
            return await _repository.AddQuestionAsync(botUin, groupId, userId, question);
        }

        public async Task<int> IncrementUsedTimesAsync(long questionId)
        {
            return await _repository.IncrementUsedTimesAsync(questionId);
        }

        public async Task<bool> IsSystemAsync(long questionId)
        {
            return await _repository.IsSystemAsync(questionId);
        }

        public async Task<int> AuditAsync(long questionId, int audit2, int isSystem)
        {
            return await _repository.AuditAsync(questionId, audit2, isSystem);
        }

        public async Task<long> GetIdByQuestionAsync(string question)
        {
            return await _repository.GetIdByQuestionAsync(question);
        }

        public string GetNew(string text)
        {
            return QuestionStringUtil.GetNew(text);
        }
    }
}
