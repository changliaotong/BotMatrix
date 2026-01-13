import re
import os

def fix_tw_terms(file_path):
    if not os.path.exists(file_path):
        print(f"File not found: {file_path}")
        return

    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()

    # Define replacements (Old Pattern, New String)
    # This covers both Simplified and Traditional source terms
    replacements = [
        (r'操作系统|操作系統', '作業系統'),
        (r'配置', '設定'),
        (r'数据|數據', '資料'),
        (r'网络|網絡', '網路'),
        (r'智能', '智慧'),
        (r'数字|數字', '數位'),
        (r'文档|文檔', '文件'),
        (r'文件', '檔案'), # Usually 文件 in CN is 檔案 in TW
        (r'支持', '支援'),
        (r'服务器|服務器', '伺服器'),
        (r'数据库|數據庫', '資料庫'),
        (r'内存|內存', '記憶體'),
        (r'带宽|帶寬', '頻寬'),
        (r'性能', '效能'),
        (r'算法', '演算法'),
        (r'集群', '叢集'),
        (r'负载均衡|負載均衡', '負載平衡'),
        (r'项目|項目', '專案'),
        (r'软件|軟件', '軟體'),
        (r'程序', '程式'),
        (r'加载|加載', '載入'),
        (r'响应|響應', '回應'),
        (r'激活', '啟用'),
        (r'登录|登錄', '登入'),
        (r'用户|用戶', '使用者'),
        (r'默认|默認', '預設'),
        (r'网关|網關', '閘道器'),
        (r'质量|質量', '品質'),
        (r'视频|視頻', '影片'),
        (r'音频|音頻', '音訊'),
        (r'雇员|雇員', '員工'),
        (r'接口', '介面'),
        (r'退出', '登出'),
        (r'保存', '儲存'),
        (r'设置|設置', '設定'),
        (r'参数|參數', '參數'),
        (r'变量|變數', '變數'),
        (r'点击|點擊', '點擊'),
        (r'联系|聯繫', '聯絡'),
        (r'社区|社區', '社群'),
        (r'通过|通過', '透過'),
        (r'实时|實時', '即時'),
        (r'客户端|客戶端', '用戶端'),
        (r'控制台', '主控台'),
        (r'信息', '資訊'), # Default to 資訊
        (r'發送', '傳送'), # TW uses 傳送 more for messages
        (r'發佈', '發布'),
        (r'產品', '產品'), # Ensure Traditional
        (r'更新', '更新'), # Ensure Traditional (though same)
        (r'批量', '批次'),
        (r'過程', '過程'), # Ensure Traditional
        (r'發生', '發生'), # Ensure Traditional
        (r'錯誤', '錯誤'), # Ensure Traditional
        (r'成功', '成功'), # Ensure Traditional
        (r'失敗', '失敗'), # Ensure Traditional
        (r'動態', '動態'), # Ensure Traditional
        (r'檢測', '檢測'), # Ensure Traditional
        (r'遠端', '遠端'), # Ensure Traditional
        (r'異常', '異常'), # Ensure Traditional
        (r'自動', '自動'), # Ensure Traditional
        (r'觸發', '觸發'), # Ensure Traditional
        (r'系統', '系統'), # Ensure Traditional
        (r'級', '級'),     # Ensure Traditional
        (r'效應', '效應'), # Ensure Traditional
        (r'完成', '完成'), # Ensure Traditional
        (r'檔案', '檔案'), # Ensure Traditional
    ]

    # Add general character conversion for common Simplified Chinese characters
    # This is a bit brute force but works for the current scope
    char_map = {
        '发': '發', '过': '過', '处': '處', '时': '時', '检测': '檢測',
        '异常': '異常', '触发': '觸發', '系统': '系統', '级': '級',
        '效应': '效應', '失败': '失敗', '错误': '錯誤', '产品': '產品',
        '动态': '動態', '过程': '過程', '发生': '發生', '监控': '監控',
        '代代码': '代碼', '代码': '程式碼', '企业': '企業', '实现': '實現',
        '协作': '協作', '安全': '安全', '自主': '自主', '学习': '學習',
        '记忆': '記憶', '体系': '體系', '赋予': '賦予', '能力': '能力',
        '工号': '工號', '权限': '權限', '审计': '審計', '建立': '建立',
        '连接': '連接', '签名': '簽名', '验证': '驗證', '可信': '可信',
        '发现': '發現', '隐患': '隱患', '立即': '立即', '专项': '專項',
        '奖励': '獎勵', '密钥': '金鑰', '加密': '加密', '存储': '儲存',
        '物理': '物理', '硬盘': '硬碟', '丢失': '遺失', '泄露': '洩漏',
        '流量': '流量', '熔断': '熔斷', '雪崩': '雪崩', '核心': '核心',
        '业务': '業務', '流转': '流轉', '打造': '打造', '专属': '專屬',
        '离': '離', '防火墙': '防火牆', '失控': '失控', '唯一': '唯一',
        '财务': '財務', '专员': '專員', '细节': '細節', '第一批': '第一批',
        '架构': '架構', '指令': '指令', '传输': '傳輸', '搜索': '搜尋',
        '批量': '批次', '应用': '應用', '体验': '體驗', '矩阵': '矩陣',
        '控制台': '主控台', '数字员工': '數位員工', '响应式': '回應式',
        '人工智能': '人工智慧', '机器学习': '機器學習', '深度学习': '深度學習',
    }
    for old, new in char_map.items():
        content = content.replace(old, new)

    # Special case: '信息' meaning 'message' in some keys
    # But for simplicity, we use 資訊 for now and manually check critical ones.
    
    # Special case: '人工智能' should stay or become '人工智慧'
    # The script will change it to '人工智慧' because of '智能' -> '智慧'

    for old, new in replacements:
        content = re.sub(old, new, content)

    # Manual fixes for common over-replacements if any
    # e.g. if '信息' should be '訊息' in some places
    # content = content.replace('收到資訊', '收到訊息')
    
    # Specific common ones in TW:
    content = content.replace('收到資訊', '收到訊息')
    content = content.replace('資訊記錄', '訊息記錄')
    content = content.replace('發送資訊', '發送訊息')
    content = content.replace('輸入資訊', '輸入訊息')
    content = content.replace('暫無資訊', '暫無訊息')

    with open(file_path, 'w', encoding='utf-8') as f:
        f.write(content)
    print(f"Fixed terms in {file_path}")

if __name__ == "__main__":
    target = r'c:\Users\彭光辉\projects\BotMatrix\src\WebUI\src\locales\zh-TW.ts'
    fix_tw_terms(target)
