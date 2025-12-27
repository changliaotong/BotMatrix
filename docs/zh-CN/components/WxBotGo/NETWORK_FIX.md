# 解决 Android SDK 下载失败/网络问题

**是的，在中国大陆，您必须配置代理或使用镜像才能下载 Android SDK 和 NDK。**

因为 Android 的下载服务器 (`dl.google.com`) 被屏蔽了。

## 方法一：配置代理 (如果您有梯子/VPN) - **推荐**

这是最稳妥的方法。

1.  打开 **Android Studio**。
2.  点击 **File** -> **Settings** (Windows) 或 **Android Studio** -> **Preferences** (Mac)。
3.  在左侧菜单找到 **Appearance & Behavior** -> **System Settings** -> **HTTP Proxy**。
4.  选择 **Manual proxy configuration** (手动配置代理)。
5.  选择 **HTTP**。
6.  填写您的代理信息：
    *   **Host name**: `127.0.0.1`
    *   **Port number**: `7890` (这是 Clash 等常见软件的默认端口，请根据您的软件实际设置填写，可能是 10809 等)。
7.  点击下方的 **Check connection**，输入 `https://dl.google.com` 测试一下。
    *   如果弹出 "Connection successful"，说明配置成功！
8.  点击 **Apply** 和 **OK**。

配置好后，再回到 **SDK Manager** 重新勾选 NDK 进行下载，应该就能成功了。

## 方法二：使用国内镜像 (如果您没有代理)

如果您没有代理软件，可以尝试修改 Hosts 文件强制指向国内镜像，或者在 SDK Manager 中添加国内源（但这种方法现在不如代理稳定）。

**更推荐直接下载离线包**：
如果 Android Studio 实在下载不动，您可以去国内的 Android 镜像站手动下载 NDK 压缩包，然后按照之前说的“手动解压”方式安装。

*   **推荐镜像站**: [https://www.androiddevtools.cn/](https://www.androiddevtools.cn/)
*   找到 **Android NDK** 栏目，下载最新的 Windows 64位版本。

---

## 解决 "SDK emulator directory is missing"

这个错误提示说明 SDK 的基础组件 `emulator` 没有下载完整。

**解决方法**：
配置好代理后 (方法一)，在 **SDK Manager** -> **SDK Tools** 列表中：
1.  找到 **Android Emulator**。
2.  如果前面没有勾选，请**勾选**它。
3.  如果前面是一个横杠 `-` (表示已安装但有更新) 或者已经是勾选状态，尝试**取消勾选** -> **Apply** (卸载)，然后再**重新勾选** -> **Apply** (重新安装)。
