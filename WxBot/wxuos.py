import requests, re, time

wechatHeaders = {
    "extspam": "Go8FCIkFEokFCggwMDAwMDAwMRAGGvAESySibk50w5Wb3uTl2c2h64jVVrV7gNs06GFlWplHQbY/5FfiO++1yH4ykCyNPWKXmco+wfQzK5R98D3so7rJ5LmGFvBLjGceleySrc3SOf2Pc1gVehzJgODeS0lDL3/I/0S2SSE98YgKleq6Uqx6ndTy9yaL9qFxJL7eiA/R3SEfTaW1SBoSITIu+EEkXff+Pv8NHOk7N57rcGk1w0ZzRrQDkXTOXFN2iHYIzAAZPIOY45Lsh+A4slpgnDiaOvRtlQYCt97nmPLuTipOJ8Qc5pM7ZsOsAPPrCQL7nK0I7aPrFDF0q4ziUUKettzW8MrAaiVfmbD1/VkmLNVqqZVvBCtRblXb5FHmtS8FxnqCzYP4WFvz3T0TcrOqwLX1M/DQvcHaGGw0B0y4bZMs7lVScGBFxMj3vbFi2SRKbKhaitxHfYHAOAa0X7/MSS0RNAjdwoyGHeOepXOKY+h3iHeqCvgOH6LOifdHf/1aaZNwSkGotYnYScW8Yx63LnSwba7+hESrtPa/huRmB9KWvMCKbDThL/nne14hnL277EDCSocPu3rOSYjuB9gKSOdVmWsj9Dxb/iZIe+S6AiG29Esm+/eUacSba0k8wn5HhHg9d4tIcixrxveflc8vi2/wNQGVFNsGO6tB5WF0xf/plngOvQ1/ivGV/C1Qpdhzznh0ExAVJ6dwzNg7qIEBaw+BzTJTUuRcPk92Sn6QDn2Pu3mpONaEumacjW4w6ipPnPw+g2TfywJjeEcpSZaP4Q3YV5HG8D6UjWA4GSkBKculWpdCMadx0usMomsSS/74QgpYqcPkmamB4nVv1JxczYITIqItIKjD35IGKAUwAA==",
    "client-version": "2.0.0",
    "User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.121 Safari/537.36",
    "Referer": "https://wx.qq.com/",
}

def wechat_uos_login():
    s = requests.Session()

    # 1️⃣ 获取 UUID
    r = s.get(
        "https://login.wx.qq.com/jslogin",
        params={"appid": "wx782c26e4c19acffb", "fun": "new", "lang": "zh_CN", "_": int(time.time() * 1000)},
        headers=wechatHeaders
    )
    uuid = re.search(r'uuid\s*=\s*"([^"]+)"', r.text).group(1)
    print(f"[+] UUID = {uuid}")
    print(f"[+] 请扫码登录: https://login.weixin.qq.com/qrcode/{uuid}")
    print(f"可能封号，谨慎试用")
    
    # 2️⃣ 轮询扫码状态
    while True:
        login_url = f"https://login.wx.qq.com/cgi-bin/mmwebwx-bin/login?loginicon=true&uuid={uuid}&tip=0&_={int(time.time()*1000)}"
        r = s.get(login_url, headers=wechatHeaders)
        code_match = re.search(r"window.code=(\d+);", r.text)
        if not code_match:
            continue
        code = code_match.group(1)

        if code == "201":
            print("[√] 已扫码，请在手机上确认登录")
        elif code == "200":
            redirect_uri = re.search(r'window.redirect_uri="([^"]+)"', r.text).group(1)
            print(f"[√] 登录确认，跳转链接: {redirect_uri}")
            break
        elif code == "408":
            print("[...] 等待扫码中...")
        else:
            print("[×] 登录失败或超时:", r.text)
            return

        time.sleep(2)

    # 3️⃣ 模拟请求 webwxnewloginpage（带UOS头）
    r2 = s.get(redirect_uri + "&fun=new", headers=wechatHeaders)
    print("[+] 登录响应：", r2.text[:300])

if __name__ == "__main__":
    wechat_uos_login()
