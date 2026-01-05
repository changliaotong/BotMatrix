# WxBot App 编译环境搭建指南

要将 WxBot 打包成 Android App，您需要安装 **Android Studio** 以及配置相应的开发工具包 (SDK/NDK)。

## 1. 下载并安装 Android Studio (必须)

这是 Google 官方的 Android 开发工具，它包含了我们所需的所有编译器和工具链。

*   **下载地址**: [https://developer.android.com/studio](https://developer.android.com/studio)
*   **安装步骤**:
    1.  下载 Windows 版本安装包 (.exe)。
    2.  运行安装程序，**一路点击 Next (下一步)** 保持默认选项即可。
    3.  安装完成后，启动 Android Studio。

## 2. 初始化配置 (第一次启动)

1.  启动 Android Studio 后，会弹出 "Welcome" 向导。
2.  选择 **Standard** (标准) 安装类型，点击 Next。
3.  它会下载并安装 **Android SDK** (约 1-2 GB)，请耐心等待下载完成。
4.  下载完成后，点击 Finish 进入欢迎主界面。

## 3. 安装 NDK 和 CMake (关键步骤)

Go 语言绑定到 Android 需要 NDK (Native Development Kit)。

1.  在 Android Studio 欢迎界面，点击 **More Actions** (或三个点图标) -> **SDK Manager**。
2.  在弹出的窗口中，切换到中间的 **SDK Tools** 标签页 (默认是 SDK Platforms)。
3.  勾选以下三项：
    *   [x] **Android SDK Command-line Tools (latest)**
    *   [x] **NDK (Side by side)**
    *   [x] **CMake**
4.  点击右下角的 **Apply**，然后点击 **OK** 开始下载安装。

## 4. 验证安装

安装完成后，请在终端 (PowerShell) 中运行以下命令来检查是否配置成功：

```powershell
flutter doctor
```

如果看到 `[√] Android toolchain` 前面变成了绿色的勾，说明环境已经准备好了！

---

## 5. (可选) 设置环境变量

如果安装后 `gomobile` 仍然报错找不到 SDK，您可能需要手动设置一下环境变量：

1.  打开 Windows 搜索，输入 "环境变量"，选择 "编辑系统环境变量"。
2.  点击 "环境变量" 按钮。
3.  在 "用户变量" 中，点击 "新建"：
    *   变量名: `ANDROID_HOME`
    *   变量值: `C:\Users\Administrator\AppData\Local\Android\Sdk` (这是默认路径，如果您修改了安装位置，请填入实际位置)
4.  点击确定保存。

完成以上步骤后，请告诉我，我就可以为您一键打包 APK 了！
