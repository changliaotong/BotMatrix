# Social Plugin Documentation

## Overview
The Social plugin provides social features for the bot, including:
- 爱群主 (Love Group Owner)
- 变身 (Transform)
- 头衔 (Title)

## Commands

### !爱群主
Express love for the group owner.

**Examples:**
```
!爱群主
```

### !变身
Transform into a different character.

**Examples:**
```
!变身
```

### !头衔
Get a random title.

**Examples:**
```
!头衔
```

## Configuration

The Social plugin requires the following configuration:

```json
{
  "social": {
    "titles": ["群主大大", "管理员", "超级会员", "VIP", "普通用户", "萌新", "大佬", "学霸", "学渣"]
  }
}
```

## Notes
- Titles are randomly selected from the pre-defined list
- Transform command will change the bot's nickname temporarily
- Love group owner command will send a message expressing love for the group owner