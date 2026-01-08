using System.Text.Json.Nodes;
using sz84.Bots.Entries;
using BotWorker.Common.Exts;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Models.Messages.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        public string HandleOtherMessage()
        {
            if (IsImage)
                return HandleImageMessage();

            if (IsFile)
                return HandleFileMessage();

            if (IsVideo)
                return HandleVideoMessage();

            if (IsXml)
                return HandleXmlMessage();

            if (IsJson)
                return HandleJsonMessage();

            if (IsKeyboard)
                return HandleKeyboardMessage();

            if (IsLightApp)
                return HandleLightAppMessage();

            if (IsLongMsg)
                return HandleLongMsgMessage();

            if (IsMarkdown)
                return HandleMarkdownMessage();

            if (IsStream)
                return HandleStreamMessage();

            if (IsVoice)
                return HandleVoiceMessage();

            if (IsMusic)
                return HandleMusicMessage();

            if (IsPoke)
                return HandlePokeMessage();

            return $"";
        }

        public static string HandleImageMessage()
        {
            return $"";
        }

        public static string HandleVideoMessage()
        {
            return $"";
        }

        public static string HandleXmlMessage()
        {
            return $"";
        }

        public static string HandleJsonMessage()
        {
            return $"";
        }

        public static string HandleKeyboardMessage()
        {
            return $"";
        }

        public static string? ExtractJumpUrl(string payloadJson)
        {
            try
            {
                var root = JsonNode.Parse(payloadJson);
                return root?["meta"]?["music"]?["jumpUrl"]?.ToString();
            }
            catch (Exception ex)
            {
                Console.WriteLine(ex.Message);
                return null;
            }
        }

        public string HandleLightAppMessage()
        {
            InfoMessage($"[LightApp] {AppName} {Payload} \n{ExtractJumpUrl(Payload)}");
            if (AppName == "com.tencent.music.lua")
            {
                var msm = Music.ParseMusicPayload(Payload);
                if (msm != null)
                {
                    if (msm.JumpUrl == null) return "";
                    var song = Music.GetSong(msm.JumpUrl, msm.MusicUrl);
                    if (song.MusicId == 0)
                        HandleMusic(msm, Payload);
                    else
                        Music.Update($"Payload={Payload.Quotes()}, IsPayload=1, PayloadDate=GETDATE()", song.MusicId);
                }
            }
            return $"";
        }

        public static string HandleLongMsgMessage()
        {
            return $"";
        }

        public static string HandleMarkdownMessage()
        {
            return $"";
        }

        public static string HandleStreamMessage()
        {
            return $"";
        }

        public static string HandleVoiceMessage()
        {
            return $"";
        }

        public static string HandleMusicMessage()
        {
            return $"";
        }

        public static string HandlePokeMessage()
        {
            return $"";
        }
    }
}
