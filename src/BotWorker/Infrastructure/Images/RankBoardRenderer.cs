using SixLabors.Fonts;
using SixLabors.ImageSharp;
using SixLabors.ImageSharp.Drawing;
using SixLabors.ImageSharp.Drawing.Processing;
using SixLabors.ImageSharp.PixelFormats;
using SixLabors.ImageSharp.Processing;

namespace BotWorker.Infrastructure.Images;

public class RankBoardRenderer
{
    private readonly Font titleFont;
    private readonly Font nameFont;
    private readonly Font scoreFont;
    private readonly HttpClient http = new HttpClient();

    public RankBoardRenderer()
    {
        var collection = new FontCollection();
        var family = collection.AddSystemFonts();

        titleFont = SystemFonts.CreateFont("Microsoft YaHei", 78, FontStyle.Bold);
        nameFont = SystemFonts.CreateFont("Microsoft YaHei", 46, FontStyle.Bold);
        scoreFont = SystemFonts.CreateFont("Microsoft YaHei", 42, FontStyle.Regular);
    }

    public async Task GetRankBoardImageAsync()
    {
        var ranks = new List<RankItem>
            {
                new RankItem{ Rank=1, Name="笑看人生", QQ=82793358, Score=110801527 },
                new RankItem{ Rank=2, Name="⚡勿私密💃拒闲聊💋", QQ=136908071, Score=13568211 },
                new RankItem{ Rank=3, Name="涼城〃暖不ㄋ人心", QQ=1912873768, Score=5306622 },
                new RankItem{ Rank=4, Name="🐾醉话见心", QQ=917216383, Score=4808687 },
                new RankItem{ Rank=5, Name="ヅ醉挽清风&", QQ=1464829292, Score=4042946 },
                new RankItem{ Rank=6, Name="WS→霸气♚90後℡丿東♨灿", QQ=597550585, Score=3672858 },
                new RankItem{ Rank=7, Name="Ｐr!ncess°", QQ=363546336, Score=3316584 },
                new RankItem{ Rank=8, Name="0度感慨与ωēη柔乄", QQ=1422079582, Score=2879570 },
                new RankItem{ Rank=9, Name="明月何皎皎", QQ=1046312610, Score=2497405 },
                new RankItem{ Rank=10, Name="光辉岁月", QQ=1653346663, Score=1898900 }
            };

        var renderer = new RankBoardRenderer();
        var img = await renderer.RenderAsync(ranks);


        await img.SaveAsPngAsync("rank.png");
        Console.WriteLine("生成成功 rank.png");
    }

    public Image CropCircle(Image<Rgba32> img, int size)
    {
        img.Mutate(x => x.Resize(size, size));
        //var mask = new Image<Rgba32>(size, size);
        //mask.Mutate(x => x.Fill(Color.White, new EllipsePolygon(size / 2, size / 2, size / 2)));

        //img.Mutate(g => g.DrawImage(mask, new Point(0, 0), 1f));
        return img;
    }

    public async Task<Image> RenderAsync(List<RankItem> ranks)
    {
        int width = 1080;
        int height = 1920;

        var img = new Image<Rgba32>(width, height);
        img.Mutate(g => g.Fill(Color.Parse("#0A0A0F")));

        // 背景渐变
        img.Mutate(g => g.Fill(new LinearGradientBrush(
            new PointF(0, 0), new PointF(0, height),
            GradientRepetitionMode.None,
            new ColorStop(0f, Color.Parse("#1A1D2E")),
            new ColorStop(1f, Color.Parse("#090A13"))
        )));

        // 标题
        img.Mutate(g => g.DrawText(
            new RichTextOptions(titleFont)
            {
                Origin = new PointF(width / 2, 80),
                HorizontalAlignment = HorizontalAlignment.Center
            },
            "积分排行榜",
            Color.Parse("#EBC875")
        ));

        int startY = 260;
        int itemHeight = 150;

        for (int i = 0; i < ranks.Count; i++)
        {
            var r = ranks[i];
            int y = startY + i * itemHeight;

            Color rankColor = r.Rank switch
            {
                1 => Color.Parse("#FFD700"),
                2 => Color.Parse("#C0C0C0"),
                3 => Color.Parse("#CD7F32"),
                _ => Color.Parse("#FFFFFF")
            };

            // 卡片背景
            img.Mutate(g => g.Fill(
                Color.Parse(i % 2 == 0 ? "#141621" : "#10121A"),
                new Rectangle(60, y - 20, 960, 130)
            ));

            // 头像加载（安全）
            Image<Rgba32>? avatar = null;
            try
            {
                var stream = await http.GetStreamAsync(
                    $"https://q1.qlogo.cn/g?b=qq&nk={r.QQ}&s=640"
                );
                avatar = await Image.LoadAsync<Rgba32>(stream);
            }
            catch
            {
                avatar = new Image<Rgba32>(110, 110);
            }

            // 裁圆
            avatar = (Image<Rgba32>)CropCircle(avatar, 110);

            img.Mutate(g => g.DrawImage(avatar, new Point(90, y), 1f));

            // 排名
            img.Mutate(g => g.DrawText(
                new RichTextOptions(nameFont)
                {
                    Origin = new PointF(240, y),
                    HorizontalAlignment = HorizontalAlignment.Left
                },
                r.Rank.ToString(),
                rankColor
            ));

            // 名字
            img.Mutate(g => g.DrawText(
                new RichTextOptions(nameFont)
                {
                    Origin = new PointF(330, y),
                    HorizontalAlignment = HorizontalAlignment.Left
                },
                r.Name,
                Color.White
            ));

            // 分数
            img.Mutate(g => g.DrawText(
                new RichTextOptions(scoreFont)
                {
                    Origin = new PointF(330, y + 60),
                    HorizontalAlignment = HorizontalAlignment.Left
                },
                $"{r.Score:N0}",
                Color.Parse("#8BC7FF")
            ));
        }

        return img;
    }


    public class RankItem
    {
        public int Rank { get; set; }
        public string Name { get; set; } = "";
        public long Score { get; set; }
        public long QQ { get; set; }
    }
}