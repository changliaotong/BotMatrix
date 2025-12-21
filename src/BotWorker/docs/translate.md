# Translate 插件文档

## 功能描述
Translate插件是一个翻译插件，支持中英文互译，使用Azure Translator API提供翻译服务。

## 配置说明
在配置文件中需要配置以下参数：
- `api_key`: Azure Translator API密钥
- `endpoint`: Azure Translator API端点（默认：https://api.cognitive.microsofttranslator.com/translate）
- `timeout`: API请求超时时间（默认：10s）
- `region`: Azure服务区域（默认：eastus）

示例配置：
```json
{
  "translate": {
    "api_key": "your-azure-api-key",
    "endpoint": "https://api.cognitive.microsofttranslator.com/translate",
    "timeout": "10s",
    "region": "eastus"
  }
}
```

## 命令格式
- `!翻译 <文本>`: 翻译指定文本
- `!translate <文本>`: 翻译指定文本

## 示例
```
用户: !translate Hello world
机器人: 翻译结果：
原文：Hello world
译文：你好，世界
```

```
用户: !翻译 你好，世界
机器人: 翻译结果：
原文：你好，世界
译文：Hello, world
```

## 版本
1.0.0

## 作者
未知