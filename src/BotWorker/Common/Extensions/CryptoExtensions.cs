using System.Security.Cryptography;
using System.Text;

namespace BotWorker.Common.Extensions
{
    public static class CryptoExtensions
    {
        // �����ַ���MD5��ϣ��32λСд��
        public static string ToMd5Hash(this string input)
        {
            var bytes = Encoding.UTF8.GetBytes(input);
            var hash = MD5.HashData(bytes);
            return BitConverter.ToString(hash).Replace("-", "").ToLowerInvariant();
        }

        // AES���ܣ����Base64
        public static string AesEncrypt(this string plainText, byte[] key, byte[] iv)
        {
            using var aes = Aes.Create();
            aes.Key = key;
            aes.IV = iv;
            var encryptor = aes.CreateEncryptor(aes.Key, aes.IV);
            using var ms = new System.IO.MemoryStream();
            using (var cs = new CryptoStream(ms, encryptor, CryptoStreamMode.Write))
            {
                var bytes = Encoding.UTF8.GetBytes(plainText);
                cs.Write(bytes, 0, bytes.Length);
            }
            return Convert.ToBase64String(ms.ToArray());
        }

        // AES���ܣ�Base64����
        public static string AesDecrypt(this string cipherTextBase64, byte[] key, byte[] iv)
        {
            var cipherBytes = Convert.FromBase64String(cipherTextBase64);
            using var aes = Aes.Create();
            aes.Key = key;
            aes.IV = iv;
            var decryptor = aes.CreateDecryptor(aes.Key, aes.IV);
            using var ms = new System.IO.MemoryStream(cipherBytes);
            using var cs = new CryptoStream(ms, decryptor, CryptoStreamMode.Read);
            using var sr = new System.IO.StreamReader(cs, Encoding.UTF8);
            return sr.ReadToEnd();
        }
    }
}


