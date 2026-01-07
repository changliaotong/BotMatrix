using System.Net;
using System.Net.Mail;

namespace sz84.Infrastructure.Utils
{
    public class Mail
    {
        public static void SendMail(string smtpUrl, string userName, string password, string fromAddress, string toAddress, string subject, string body)
        {
            try
            {
                // 设置 SMTP 服务器的详细信息
                using SmtpClient smtpClient = new(smtpUrl)
                {
                    Port = 587, // SMTP 服务器的端口号
                    Credentials = new NetworkCredential(userName, password),
                    EnableSsl = true // 如果使用 SSL 连接，请设置为 true
                };

                // 创建邮件对象
                using MailMessage mailMessage = new(fromAddress, toAddress, subject, body);

                // 发送邮件
                smtpClient.Send(mailMessage);
                Console.WriteLine("邮件发送成功！");
            }
            catch (SmtpException ex)
            {
                Console.WriteLine("SMTP 错误： " + ex.Message);
            }
            catch (Exception ex)
            {
                Console.WriteLine("邮件发送失败： " + ex.Message);
            }
        }

        public static void SendMail(string toAddress, string subject, string body)
        {
            SendMail("smtp.exmail.qq.com", "penggh@sz84.com", "fkueiqiq461686", "penggh@sz84.com", toAddress, subject, body);
        }
    }
}
