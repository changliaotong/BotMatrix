# Android Studio NDK 安装指南

如果您已经下载了 Android Studio，但是不知道如何安装 NDK，请按照以下步骤操作。

## 方法一：通过 Android Studio 自动安装 (推荐)

这是最简单的方法，它会自动配置所有路径。

1.  打开 Android Studio。
2.  在欢迎界面点击 **More Actions** (或三个点图标) -> **SDK Manager**。
    *(如果已经进入了项目界面，请点击菜单栏 Tools -> SDK Manager)*
3.  在弹出的窗口中：
    *   点击中间的 **SDK Tools** 选项卡 (注意不是默认的 SDK Platforms)。
    *   在列表中找到 **NDK (Side by side)**，勾选它。
    *   同时勾选 **CMake**。
4.  点击右下角的 **Apply** (应用)。
5.  系统会弹出一个确认窗口，点击 **OK**。
6.  **关键步骤**：此时会自动开始下载并安装 NDK。等待进度条走完。
7.  安装完成后，点击 Finish。

## 方法二：手动安装 (如果您下载的是 ZIP 包)

如果您是从官网手动下载了 `android-ndk-r26b-windows-x86_64.zip` 之类的压缩包：

1.  **解压**：
    将压缩包解压到一个固定目录，例如：`D:\Android\ndk-bundle` (路径中最好不要有空格和中文)。

2.  **配置环境变量**：
    *   打开 Windows 搜索，输入 "环境变量"，选择 "编辑系统环境变量"。
    *   点击 "环境变量"。
    *   在 "用户变量" 区域，点击 "新建"：
        *   变量名: `ANDROID_NDK_HOME`
        *   变量值: `D:\Android\ndk-bundle` (填您实际解压的路径)
    *   找到 Path 变量，点击编辑，添加 `%ANDROID_NDK_HOME%`。

## 验证是否成功

打开终端 (PowerShell)，输入：

```powershell
flutter doctor
```

如果 `Android toolchain` 这一项变成了绿色对勾 `[√]`，说明安装成功！
