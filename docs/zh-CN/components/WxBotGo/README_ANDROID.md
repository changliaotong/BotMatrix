# 如何在手机上运行 WxBotGo

由于 Go 语言具有强大的跨平台编译能力，我们可以直接编译出可以在 Android 手机上运行的二进制文件，并通过 **Termux** 终端应用来运行它。

## 1. 编译 (在电脑上执行)

双击运行 `build_android.bat` 脚本，或者在终端执行：

```bash
# 设置环境变量 (Windows CMD)
set GOOS=linux
set GOARCH=arm64
go build -o wxbot-android-arm64
```

这将生成一个名为 `wxbot-android-arm64` 的文件。

## 2. 准备手机环境

1.  下载并安装 **Termux** (建议从 F-Droid 下载，Play Store 版本已不再更新)。
2.  打开 Termux，执行以下命令获取存储权限（允许访问手机文件）：
    ```bash
    termux-setup-storage
    ```
    *(屏幕会弹出权限请求，点击“允许”)*

## 3. 传输文件

将电脑上生成的 `wxbot-android-arm64` 文件发送到手机（可以通过 USB、QQ、微信文件传输助手等）。
假设文件保存在手机的 `Download` (下载) 文件夹中。

## 4. 运行 Bot (自动安装)

为了简化操作，您可以同时将 `install_termux.sh` 文件发送到手机。

1.  在 Termux 中运行：
    ```bash
    cp /sdcard/Download/install_termux.sh ~/
    sh install_termux.sh
    ```

2.  以后每次启动只需要输入：
    ```bash
    ./start-wxbot.sh
    ```

## 5. 运行 Bot (手动方式)

如果不使用脚本，请执行以下命令：

1.  **将文件复制到 Termux 目录**：
    ```bash
    cp /sdcard/Download/wxbot-android-arm64 ~/
    ```

2.  **赋予执行权限**：
    ```bash
    chmod +x wxbot-android-arm64
    ```

3.  **运行**：
    *   **情况 A：BotNexus 管理端在电脑上** (假设电脑 IP 是 192.168.1.100)：
        ```bash
        export MANAGER_URL="ws://192.168.1.100:3001"
        ./wxbot-android-arm64
        ```
    *   **情况 B：BotNexus 管理端也在手机上** (直接运行)：
        ```bash
        ./wxbot-android-arm64
        ```

## 5. 扫码登录

程序启动后，会输出一个二维码的 URL 或者直接在终端显示二维码（视终端支持情况而定）。
如果是 URL，复制到浏览器打开即可扫描登录。

登录成功后，手机就变成了一个 24小时在线的微信机器人！

---

## 进阶：后台运行

如果希望 Termux 在后台持续运行不被系统杀掉：
1.  下拉通知栏，找到 Termux 通知，点击 "Acquire wakelock" (保持唤醒)。
2.  或者在 Termux 电池优化设置中选择“无限制”。
