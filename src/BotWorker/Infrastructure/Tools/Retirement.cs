using System.Text;

namespace BotWorker.Infrastructure.Tools
{
    public class Retirement
    {
        public enum Gender
        {
            男,
            女,
            male,
            female,
        }

        public enum Cadre
        {  
            未知,
            职工,
            干部,
        }

        public static string CalculateRetirement(DateTime birthday, Gender gender, Cadre isCadre = 0)
        {
            int baseRetirementAge = 60; // 男性基础退休年龄
            int maxMaleRetirementAge = 63;
            int femaleCadreBaseRetirementAge = 55; // 女干部基础退休年龄
            int femaleWorkerBaseRetirementAge = 50; // 女工基础退休年龄
            int maxCadreRetirementAge = 58; // 女干部最大退休年龄
            int maxWorkerRetirementAge = 55; // 女工最大退休年龄
            DateTime baseDate = new(1965, 1, 1);
            int delayMonths = 0;

            // 计算出生月份与基准月份的差值
            int monthsSinceBase = (birthday.Year - baseDate.Year) * 12 + (birthday.Month - baseDate.Month);

            // 计算延迟的月数，最多为60个月
            if (monthsSinceBase >= 0)
            {
                delayMonths = Math.Min(monthsSinceBase / 4, 60); // 每4个月延迟1个月，最多60个月
            }

            // 计算退休年龄和预计退休时间
            StringBuilder result = new();
            result.AppendLine($"出生年月为：{birthday:yyyy年MM月}。\n");

            if (gender == Gender.男 || gender == Gender.male)
            {
                int retirementAge = baseRetirementAge + delayMonths;
                if (retirementAge > maxMaleRetirementAge) retirementAge = maxMaleRetirementAge;

                DateTime retirementDate = birthday.AddYears(retirementAge);
                result.AppendLine($"男性退休年龄为：{retirementAge}岁。");
                result.AppendLine($"预计退休时间为：{retirementDate:yyyy年MM月}。");
                result.AppendLine($"延迟了 {delayMonths} 个月退休。\n");
            }
            else if (gender == Gender.女 || gender == Gender.female)
            {
                // 计算女干部的退休年龄
                int cadreRetirementAge = femaleCadreBaseRetirementAge + delayMonths;
                if (cadreRetirementAge > maxCadreRetirementAge) cadreRetirementAge = maxCadreRetirementAge;

                DateTime cadreRetirementDate = birthday.AddYears(cadreRetirementAge);

                if (isCadre != Cadre.职工)
                {
                    result.AppendLine($"女干部退休年龄：{cadreRetirementAge}岁。");
                    result.AppendLine($"预计退休时间为：{cadreRetirementDate:yyyy年MM月}。");
                    result.AppendLine($"延迟了 {delayMonths} 个月退休。\n");
                }

                // 计算女职工的退休年龄
                int workerRetirementAge = femaleWorkerBaseRetirementAge + delayMonths;
                if (workerRetirementAge > maxWorkerRetirementAge) workerRetirementAge = maxWorkerRetirementAge;

                DateTime workerRetirementDate = birthday.AddYears(workerRetirementAge);

                if (isCadre != Cadre.干部)
                {
                    result.AppendLine($"女职工退休年龄：{workerRetirementAge}岁。");
                    result.AppendLine($"预计退休时间为：{workerRetirementDate:yyyy年MM月}。");
                    result.AppendLine($"延迟了 {delayMonths} 个月退休。\n");
                }
            }
            else
            {
                return "性别输入有误，请输入“男”或“女”。";
            }

            return result.ToString();
        }


    }
}
