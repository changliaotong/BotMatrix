using System.Text;
using System.Security.Cryptography;
using Org.BouncyCastle.Crypto;
using Org.BouncyCastle.Security;
using BotWorker.Common;

namespace BotWorker.Infrastructure.Utils
{
    public static class Encrypt
    {
        public const string RsaPublicKey = $"<RSAKeyValue><Modulus>v3YlM8/BZ/nC+Ix3W0CtWUoOhkkN2XTbrXTwYlppxbTHVtWkZ9Mm+E4ZIbaaxja18LxmcOjvo0rHRZbD++/XK98fwTtfJPIKMKSaJR8WsrDyntQUB2rdfCRNmx3O17ds6PGVnjefHWUc4Yichdl/E//ITyJ6XXUqPLO8IWCT86E=</Modulus><Exponent>AQAB</Exponent></RSAKeyValue>";


        /// <summary>
        /// 机器人加密解密
        /// </summary>
        /// <param name="guid"></param>
        /// <param name="cmdName"></param>
        /// <param name="cmdPara"></param>
        /// <returns></returns>
        public static string GetEncryptRes(string guid, string cmdName, string cmdPara)
        {
            string key = guid.MD5().Substring(1, 24);
            if (cmdName == "加密")
                return cmdPara.Encrypt3DES(key);
            else
                return cmdPara.Decrypt3DES(key);
        }

        #region DES

        /// <summary>
        /// DES加密
        /// </summary>
        /// <param name="data">加密数据</param>
        /// <param name="key">8位字符的密钥字符串</param>
        /// <param name="iv">8位字符的初始化向量字符串</param>
        /// <returns></returns>
        public static string EncryptDes(this string data, string key, string iv)
        {
            byte[] byKey = Encoding.ASCII.GetBytes(key);
            byte[] byIV = Encoding.ASCII.GetBytes(iv);

            var cryptoProvider = TripleDES.Create();
            //_ = cryptoProvider.KeySize;
            MemoryStream ms = new();
            CryptoStream cst = new(ms, cryptoProvider.CreateEncryptor(byKey, byIV), CryptoStreamMode.Write);

            StreamWriter sw = new(cst);
            sw.Write(data);
            sw.Flush();
            cst.FlushFinalBlock();
            sw.Flush();
            return Convert.ToBase64String(ms.GetBuffer(), 0, (int)ms.Length);
        }

        /// <summary>
        /// DES解密
        /// </summary>
        /// <param name="data">解密数据</param>
        /// <param name="key">8位字符的密钥字符串(需要和加密时相同)</param>
        /// <param name="iv">8位字符的初始化向量字符串(需要和加密时相同)</param>
        /// <returns></returns>
        public static string? DecryptDes(this string data, string key, string iv)
        {
            byte[] byKey = Encoding.ASCII.GetBytes(key);
            byte[] byIV = Encoding.ASCII.GetBytes(iv);

            byte[] byEnc;
            try
            {
                byEnc = Convert.FromBase64String(data);
            }
            catch
            {
                return null;
            }

            var cryptoProvider = TripleDES.Create();
            MemoryStream ms = new(byEnc);
            CryptoStream cst = new(ms, cryptoProvider.CreateDecryptor(byKey, byIV), CryptoStreamMode.Read);
            StreamReader sr = new(cst);
            return sr.ReadToEnd();
        }

        /// <summary>
            /// 加密
            /// </summary>
            /// <param name="strString"></param>
            /// <param name="strKey"></param>
            /// <param name="encoding"></param>
            /// <returns></returns>
        public static string Encrypt3DES(this string strString, string key)
        {
            var DES = TripleDES.Create();


            DES.Key = Encoding.UTF8.GetBytes(key);
            DES.Mode = CipherMode.ECB;

            ICryptoTransform DESEncrypt = DES.CreateEncryptor();

            byte[] Buffer = Encoding.UTF8.GetBytes(strString);

            return Convert.ToBase64String(DESEncrypt.TransformFinalBlock(Buffer, 0, Buffer.Length));
        }

        /// <summary>
        /// 解密
        /// </summary>
        /// <param name="strString"></param>
        /// <param name="strKey"></param>
        /// <returns></returns>
        public static string Decrypt3DES(this string strString, string key)
        {
            try
            {
                var DES = TripleDES.Create();

                DES.Key = Encoding.UTF8.GetBytes(key);
                DES.Mode = CipherMode.ECB;
                ICryptoTransform DESDecrypt = DES.CreateDecryptor();

                byte[] Buffer = Convert.FromBase64String(strString);
                return Encoding.UTF8.GetString(DESDecrypt.TransformFinalBlock(Buffer, 0, Buffer.Length));
            }
            catch (Exception e)
            {
                return "错误：" + e.Message;
            }
        }

        #endregion DES


        #region Base64

        /// <summary>
        /// Base64加密
        /// </summary>
        /// <param name="input">需要加密的字符串</param>
        /// <returns></returns>
        public static string EncryptBase64(this string input)
        {
            return input.EncryptBase64(new UTF8Encoding());
        }

        /// <summary>
        /// Base64加密
        /// </summary>
        /// <param name="input">需要加密的字符串</param>
        /// <param name="encode">字符编码</param>
        /// <returns></returns>
        public static string EncryptBase64(this string input, Encoding encode)
        {
            return Convert.ToBase64String(encode.GetBytes(input));
        }

        /// <summary>
        /// Base64解密
        /// </summary>
        /// <param name="input">需要解密的字符串</param>
        /// <returns></returns>
        public static string DecryptBase64(this string input)
        {
            return input.DecryptBase64(new UTF8Encoding());
        }

        /// <summary>
        /// Base64解密
        /// </summary>
        /// <param name="input">需要解密的字符串</param>
        /// <param name="encode">字符的编码</param>
        /// <returns></returns>
        public static string DecryptBase64(this string input, Encoding encode)
        {
            return encode.GetString(Convert.FromBase64String(input));
        }

        #endregion base64


        #region SHA

        /// <summary>
        /// MD5
        /// </summary>
        /// <param name="text"></param>
        /// <returns></returns>
        public static string MD5(this string text)
        {
            var md = System.Security.Cryptography.MD5.Create();
            byte[] value, hash;
            value = Encoding.UTF8.GetBytes(text);
            hash = md.ComputeHash(value);
            md.Clear();
            string temp = "";
            for (int i = 0, len = hash.Length; i < len; i++)
            {
                temp += hash[i].ToString("x").PadLeft(2, '0');
            }
            return temp.ToUpper();
        }

        /// <summary>
        /// MD5
        /// </summary>
        /// <param name="text"></param>
        /// <returns></returns>
        public static string MD5Lower(this string text)
        {
            var md = System.Security.Cryptography.MD5.Create();
            byte[] value, hash;
            value = Encoding.UTF8.GetBytes(text);
            hash = md.ComputeHash(value);
            md.Clear();
            string temp = "";
            for (int i = 0, len = hash.Length; i < len; i++)
            {
                temp += hash[i].ToString("x").PadLeft(2, '0');
            }
            return temp;
        }

        /// <summary>
        /// SHA1
        /// </summary>
        /// <param name="text"></param>
        /// <returns></returns>
        /// <exception cref="Exception"></exception>
        public static string Sha1(this string text)
        {
            byte[] bytValue = Encoding.UTF8.GetBytes(text);
            try
            {
                var sha1 = SHA1.Create();
                byte[] retVal = sha1.ComputeHash(bytValue);
                StringBuilder sb = new();
                for (int i = 0; i < retVal.Length; i++)
                {
                    sb.Append(retVal[i].ToString("x2"));
                }
                return sb.ToString();
            }
            catch (Exception ex)
            {
                throw new Exception("GetSHA1HashFromString() fail,error:" + ex.Message);
            }
        }

        /// <summary>
        /// sha256
        /// </summary>
        /// <param name="text"></param>
        /// <returns></returns>
        /// <exception cref="Exception"></exception>
        public static string Sha256(this string text)
        {
            byte[] bytValue = Encoding.UTF8.GetBytes(text);
            try
            {
                var sha256 = SHA256.Create();
                byte[] retVal = sha256.ComputeHash(bytValue);
                StringBuilder sb = new();
                for (int i = 0; i < retVal.Length; i++)
                {
                    sb.Append(retVal[i].ToString("x2"));
                }
                return sb.ToString();
            }
            catch (Exception ex)
            {
                throw new Exception("GetSHA256HashFromString() fail,error:" + ex.Message);
            }
        }


        /// <summary>
        /// hmacsha256 加密
        /// </summary>
        /// <param name="secret">加密秘钥</param>
        /// <param name="message">待加密的内容</param>
        /// <returns></returns>
        public static string CreateToken(this string secret, string message)
        {
            var encoding = Encoding.UTF8;
            byte[] keyByte = encoding.GetBytes(secret);
            byte[] messageBytes = encoding.GetBytes(message);
            using var hmacsha256 = new HMACSHA256(keyByte);
            byte[] hashmessage = hmacsha256.ComputeHash(messageBytes);
            StringBuilder sb = new();
            for (int i = 0; i < hashmessage.Length; i++)
            {
                sb.Append(hashmessage[i].ToString("x2"));
            }
            return sb.ToString();
        }

        /// <summary>
        /// sha384
        /// </summary>
        /// <param name="text"></param>
        /// <returns></returns>
        /// <exception cref="Exception"></exception>
        public static string Sha384(this string text)
        {
            byte[] bytValue = Encoding.UTF8.GetBytes(text);
            try
            {
                var sha384 = SHA384.Create();
                byte[] retVal = sha384.ComputeHash(bytValue);
                StringBuilder sb = new();
                for (int i = 0; i < retVal.Length; i++)
                {
                    sb.Append(retVal[i].ToString("x2"));
                }
                return sb.ToString();
            }
            catch (Exception ex)
            {
                throw new Exception("GetSHA384HashFromString() fail,error:" + ex.Message);
            }
        }

        /// <summary>
        /// sha512
        /// </summary>
        /// <param name="text"></param>
        /// <returns></returns>
        /// <exception cref="Exception"></exception>
        public static string Sha512(this string text)
        {
            byte[] bytValue = Encoding.UTF8.GetBytes(text);
            try
            {
                var sha512 = SHA512.Create();
                byte[] retVal = sha512.ComputeHash(bytValue);
                StringBuilder sb = new();
                for (int i = 0; i < retVal.Length; i++)
                {
                    sb.Append(retVal[i].ToString("x2"));
                }
                return sb.ToString();
            }
            catch (Exception ex)
            {
                throw new Exception("GetSHA512HashFromString() fail,error:" + ex.Message);
            }
        }

        #endregion SHA


        #region RSA

        /// <summary>
        /// 使用私钥解密后再用私钥加密,结果需要用公钥解密
        /// </summary>
        /// <param name="text"></param>
        /// <param name="key"></param>
        /// <returns></returns>
        public static string EnCryptPrivateDecryptRsa(this string text, string key = "")
        {
            return text.DecryptRsa(key).EncryptRsaPrivate(key);
        }

        /// <summary>
        /// 使用公钥(默认官方的)加密
        /// </summary>
        /// <param name="text"></param>
        /// <param name="key"></param>
        /// <returns></returns>
        public static string EncryptRsa(this string text, string key = RsaPublicKey)
        {
            using var rsaProvider = new RSACryptoServiceProvider();
            var inputBytes = Encoding.UTF8.GetBytes(text);
            rsaProvider.FromXmlString(key);
            int bufferSize = rsaProvider.KeySize / 8 - 11;
            var buffer = new byte[bufferSize];
            using MemoryStream inputStream = new(inputBytes),
                 outputStream = new();
            while (true)
            {
                int readSize = inputStream.Read(buffer, 0, bufferSize);
                if (readSize <= 0)
                {
                    break;
                }

                var temp = new byte[readSize];
                Array.Copy(buffer, 0, temp, 0, readSize);
                var encryptedBytes = rsaProvider.Encrypt(temp, false);
                outputStream.Write(encryptedBytes, 0, encryptedBytes.Length);
            }
            return Convert.ToBase64String(outputStream.ToArray());
        }

        /// <summary>
        /// 使用私钥(默认官方的)解密
        /// </summary>
        /// <param name="text"></param>
        /// <param name="key"></param>
        /// <returns></returns>
        public static string DecryptRsa(this string text, string key = "")
        {
            try
            {
                using var rsaProvider = new RSACryptoServiceProvider();
                var inputBytes = Convert.FromBase64String(text);
                rsaProvider.FromXmlString(key);
                int bufferSize = rsaProvider.KeySize / 8;
                var buffer = new byte[bufferSize];
                using MemoryStream inputStream = new(inputBytes),
                     outputStream = new();
                while (true)
                {
                    int readSize = inputStream.Read(buffer, 0, bufferSize);
                    if (readSize <= 0)
                    {
                        break;
                    }
                    var temp = new byte[readSize];
                    Array.Copy(buffer, 0, temp, 0, readSize);
                    var rawBytes = rsaProvider.Decrypt(temp, false);
                    outputStream.Write(rawBytes, 0, rawBytes.Length);
                }
                return Encoding.UTF8.GetString(outputStream.ToArray());
            }
            catch (Exception ex)
            {
                ErrorMessage($"RSA解密出错:{ex.Message}");
                return text;
            }
        }

        /// <summary>
        /// 使用私钥(默认官方的)加密，加密内容可以使用公钥解密
        /// </summary>
        /// <param name="xmlPrivateKey"> 私钥(XML格式字符串)</param>
        /// <param name="strEncryptString"> 要加密的数据 </param>
        /// <returns> 加密后的数据 </returns>
        public static string EncryptRsaPrivate(this string text, string key = "")
        {
            using var rsaProvider = new RSACryptoServiceProvider();
            var inputBytes = Encoding.UTF8.GetBytes(text);
            rsaProvider.FromXmlString(key);
            int bufferSize = rsaProvider.KeySize / 8 - 11;
            var buffer = new byte[bufferSize];
            using MemoryStream inputStream = new(inputBytes),
                 outputStream = new();
            AsymmetricCipherKeyPair keyPair = DotNetUtilities.GetKeyPair(rsaProvider);
            IBufferedCipher c = CipherUtilities.GetCipher("RSA/ECB/PKCS1Padding");
            c.Init(true, keyPair.Private);
            while (true)
            {
                int readSize = inputStream.Read(buffer, 0, bufferSize);
                if (readSize <= 0)
                {
                    break;
                }

                var DataToEncrypt = new byte[readSize];
                Array.Copy(buffer, 0, DataToEncrypt, 0, readSize);
                byte[] outBytes = c.DoFinal(DataToEncrypt);//加密
                outputStream.Write(outBytes, 0, outBytes.Length);
            }
            return Convert.ToBase64String(outputStream.ToArray());

        }


        /// <summary>
        /// 使用公钥(默认官方)解密私钥加密的内容
        /// </summary>
        /// <param name="key"> 公钥(XML格式字符串) </param>
        /// <param name="text"> 要解密数据 </param>
        /// <returns> 解密后的数据 </returns>
        public static string DecryptPublicRsa(this string text, string key = RsaPublicKey)
        {
            try
            {
                using var rsaProvider = new RSACryptoServiceProvider();
                var inputBytes = Convert.FromBase64String(text);
                rsaProvider.FromXmlString(key);
                int bufferSize = rsaProvider.KeySize / 8;
                var buffer = new byte[bufferSize];
                using MemoryStream inputStream = new(inputBytes),
                     outputStream = new();
                RSAParameters rp = rsaProvider.ExportParameters(false);//转换密钥
                AsymmetricKeyParameter pbk = DotNetUtilities.GetRsaPublicKey(rp);

                IBufferedCipher c = CipherUtilities.GetCipher("RSA/ECB/PKCS1Padding"); //第一个参数为true表示加密，为false表示解密；第二个参数表示密钥
                c.Init(false, pbk);
                while (true)
                {
                    int readSize = inputStream.Read(buffer, 0, bufferSize);
                    if (readSize <= 0)
                    {
                        break;
                    }
                    var DataToDecrypt = new byte[readSize];
                    Array.Copy(buffer, 0, DataToDecrypt, 0, readSize);
                    byte[] outBytes = c.DoFinal(DataToDecrypt);//解密
                    outputStream.Write(outBytes, 0, outBytes.Length);
                }
                return Encoding.UTF8.GetString(outputStream.ToArray());
            }
            catch (Exception ex)
            {
                ErrorMessage($"RSA解密2出错:{ex.Message}");
                return text;
            }
        }

        /// <summary>
        /// 用私钥进行RSA加密，原函数，存在不能加密长消息的问题
        /// </summary>
        /// <param name="xmlPrivateKey"> 私钥(XML格式字符串)</param>
        /// <param name="strEncryptString"> 要加密的数据 </param>
        /// <returns> 加密后的数据 </returns>
        public static string PrivateKeyEncrypt(this string text, string xmlPrivateKey)
        {    //加载私钥
            RSACryptoServiceProvider privateRsa = new();
            privateRsa.FromXmlString(xmlPrivateKey);
            //转换密钥
            AsymmetricCipherKeyPair keyPair = DotNetUtilities.GetKeyPair(privateRsa);
            IBufferedCipher c = CipherUtilities.GetCipher("RSA/ECB/PKCS1Padding"); //使用RSA/ECB/PKCS1Padding格式
                                                                                   //第一个参数为true表示加密，为false表示解密；第二个参数表示密钥

            c.Init(true, keyPair.Private);
            byte[] DataToEncrypt = Encoding.UTF8.GetBytes(text);
            byte[] outBytes = c.DoFinal(DataToEncrypt);//加密
            string strBase64 = Convert.ToBase64String(outBytes);
            return strBase64;
        }


        /// <summary>
        /// 用公钥进行RSA解密 
        /// </summary>
        /// <param name="xmlPublicKey"> 公钥(XML格式字符串) </param>
        /// <param name="strDecryptString"> 要解密数据 </param>
        /// <returns> 解密后的数据 </returns>
        public static string PublicKeyDecrypt(this string text, string xmlPublicKey)
        {
            //加载公钥
            RSACryptoServiceProvider publicRsa = new();
            publicRsa.FromXmlString(xmlPublicKey);
            RSAParameters rp = publicRsa.ExportParameters(false);//转换密钥
            AsymmetricKeyParameter pbk = DotNetUtilities.GetRsaPublicKey(rp);

            IBufferedCipher c = CipherUtilities.GetCipher("RSA/ECB/PKCS1Padding"); //第一个参数为true表示加密，为false表示解密；第二个参数表示密钥
            c.Init(false, pbk);
            byte[] DataToDecrypt = Convert.FromBase64String(text);
            byte[] outBytes = c.DoFinal(DataToDecrypt);//解密

            string strDec = Encoding.UTF8.GetString(outBytes);
            return strDec;
        }

        #endregion RSA
    }

}
