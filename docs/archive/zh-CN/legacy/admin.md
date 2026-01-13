# Admin Plugin Documentation

## Overview
The Admin plugin provides administrative features for the bot, including:
- 教学 (Teaching)
- 本群 (Group Information)
- 话唠 (Chatty Mode)
- 终极 (Ultimate Mode)
- 智能体 (Agent Mode)
- 后台 (Admin Panel)
- 设置 (Settings)
- 开启 (Enable)
- 关闭 (Disable)
- 敏感词 (Sensitive Words)
 - 语音回复 (AI Voice Reply)
 - 阅后即焚 (Burn-after-reading Auto Recall)

## Commands

### !教学
View the bot's usage tutorial.

**Examples:**
```
!教学
```

### !本群
View group information.

**Examples:**
```
!本群
```

### !话唠
Enable chatty mode.

**Examples:**
```
!话唠
```

### !终极
Enable ultimate mode.

**Examples:**
```
!终极
```

### !智能体
Enable agent mode.

**Examples:**
```
!智能体
```

### !后台
Open the admin panel.

**Examples:**
```
!后台
```

### !设置 <parameter> <value>
Set a parameter.

**Parameters:**
- `<parameter>`: The parameter to set
- `<value>`: The value to set

**Examples:**
```
!设置 greeting "欢迎加入本群！"
!设置 max_points 1000
```

### !开启 <feature>
Enable a feature.

**Parameters:**
- `<feature>`: The feature to enable

**Examples:**
```
!开启 weather
!开启 translation
!开启 语音回复
!开启 阅后即焚
```

### !关闭 <feature>
Disable a feature.

**Parameters:**
- `<feature>`: The feature to disable

**Examples:**
```
!关闭 weather
!关闭 translation
!关闭 语音回复
!关闭 阅后即焚
```

## Configuration

The Admin plugin requires the following configuration:

```json
{
  "admin": {
    "admins": ["admin1", "admin2", "admin3"],
    "feature_switches": {
      "weather": true,
      "translation": true,
      "music": true
    }
  }
}
```

## Notes
- Admin commands can only be used by bot administrators
- Feature switches can be used to enable or disable specific features
- The admin panel provides a centralized interface for managing the bot
- Sensitive words can be customized in the configuration file
