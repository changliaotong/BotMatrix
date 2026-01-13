using Microsoft.CodeAnalysis.CSharp.Scripting;
using Microsoft.CodeAnalysis.Scripting;

namespace BotWorker.Infrastructure.Tools
{
    public class Calc : MetaData<Calc>
    {
        public override string TableName => throw new NotImplementedException();
        public override string KeyField => throw new NotImplementedException();


        //表达式计算器
        public static string GetJsRes2(string key)
        {
            if (key.IsNull()) return "";
            try
            {
                //错误输入纠正 + - * / 等
                key = key.Replace("＋", "+").Replace("－", "-").Replace("×", "*").Replace("／", "/").Replace("[", "(").Replace("]", ")").Replace("（", "(").Replace("）", ")").Replace("÷", "/");
                key = key.Replace(";", "").Replace("ｘ", "*").Replace("＊", "*");
                key = key.Replace("=", "").Replace("＝", "").Replace("?", "").Replace("？", "");
                if (key.Contains('/'))
                    key = key.Replace("/", "*1.0/");
                return SQLConn.Query("select " + key + " as res");
            }
            catch
            {
                return "不正确的表达式";
            }
        }

        public static async Task<string> GetJsRes(string key)
        {
            try
            {                
                //错误输入纠正 + - * / 等
                key = key.Replace("＋", "+").Replace("－", "-").Replace("×", "*").Replace("／", "/").Replace("[", "(").Replace("]", ")").Replace("（", "(").Replace("）", ")").Replace("÷", "/");
                key = key.Replace(";", "").Replace("ｘ", "*").Replace("＊", "*");
                key = key.Replace("=", "").Replace("＝", "").Replace("?", "").Replace("？", "");
                if (key.Contains('/'))
                    key = key.Replace("/", "*1.0/");

                var result = await CSharpScript.EvaluateAsync<double>(key);
                return result.ToString();
            }
            catch (CompilationErrorException ex)
            {
                Logger.Error(string.Join(", ", ex.Diagnostics));
                //return "表达式编译错误";// +  );                
                return "";
            }
            catch (Exception ex)
            {
                Logger.Error(ex.Message);
                return "不正确的表达式"; 
            }
        }
    }
    
}
