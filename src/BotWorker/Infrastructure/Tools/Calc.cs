using Microsoft.CodeAnalysis.CSharp.Scripting;
using Microsoft.CodeAnalysis.Scripting;
using sz84.Core.MetaDatas;
using sz84.Core.Database;

namespace sz84.Infrastructure.Tools
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
                if (key.Contains('/'))
                    key = key.Replace("/", "*1.0/");
                return SQLConn.Query("select " + key + " as res");
            }
            catch
            {
                return "不正确的表达式";
            }
        }

        public static string GetJsRes(string key)
        {
            try
            {                
                //错误输入纠正 + - * / 等
                key = key.Replace("＋", "+").Replace("－", "-").Replace("×", "*").Replace("／", "/").Replace("[", "(").Replace("]", ")").Replace("（", "(").Replace("）", ")").Replace("÷", "/");
                key = key.Replace(";", "").Replace("ｘ", "*").Replace("＊", "*");
                if (key.Contains('/'))
                    key = key.Replace("/", "*1.0/");

                var result = CSharpScript.EvaluateAsync<double>(key).Result;
                return result.ToString();
            }
            catch (CompilationErrorException ex)
            {
                Debug(string.Join(", ", ex.Diagnostics), "计算");
                //return "表达式编译错误";// +  );                
                return "";
            }
            catch (Exception ex)
            {
                Debug(ex.Message, "计算");
                return "不正确的表达式"; 
            }
        }
    }
    
}
