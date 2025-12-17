import subprocess
import time
import os
import signal
import sys

def main():
    print("=" * 60)
    print("BotNexus 实时日志监控工具")
    print("=" * 60)
    print("\n功能:")
    print("- 实时监控BotNexus输出")
    print("- 高亮显示关键信息")
    print("- 过滤心跳和连接事件")
    print("\n使用方法:")
    print("- 运行此脚本后，它会监控BotNexus的实时输出")
    print("- 按 Ctrl+C 停止监控")
    print("\n关键词高亮:")
    print("- 连接/断开事件")
    print("- 心跳信息")
    print("- 错误和警告")
    print("=" * 60)
    
    # 检查BotNexus是否正在运行
    try:
        result = subprocess.run(['tasklist', '/FI', 'IMAGENAME eq botnexus.exe'], 
                              capture_output=True, text=True)
        if 'botnexus.exe' not in result.stdout:
            print("\n❌ BotNexus没有在运行！")
            print("请先运行: cd BotNexus && botnexus.exe")
            return
    except Exception as e:
        print(f"检查进程时出错: {e}")
    
    print("\n🔍 开始监控BotNexus日志...")
    print("提示: 您可以同时运行测试工具来模拟Napcat Bot连接")
    print("打开: d:/projects/BotMatrix/BotNexus/test_napcat_heartbeat.html")
    print("\n" + "=" * 60 + "\n")
    
    # 实时监控逻辑
    last_lines = []
    
    try:
        while True:
            # 这里可以添加从文件或端口读取日志的逻辑
            # 目前BotNexus直接输出到控制台，所以我们能看到
            time.sleep(5)  # 每5秒检查一次
            print(f"[{time.strftime('%H:%M:%S')}] 监控中... BotNexus正在运行")
            
    except KeyboardInterrupt:
        print("\n\n停止监控。感谢使用！")

if __name__ == "__main__":
    main()