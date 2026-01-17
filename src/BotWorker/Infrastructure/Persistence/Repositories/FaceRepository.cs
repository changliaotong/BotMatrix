using System;
using BotWorker.Domain.Repositories;
using BotWorker.Infrastructure.Communication.Platforms.BotPublic;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class FaceRepository : IFaceRepository
    {
        public string ConvertFacesBack(string text)
        {
            if (string.IsNullOrEmpty(text)) return text;
            return FacePublic.ConvertFacesBack(text);
        }
    }
}
