# 资产更新与缓存同步重构方案 (Refactoring Guide)

## 1. 核心目标
解决高频资产（积分、余额、金币、算力）在多线程/多实例环境下的缓存一致性问题，全面对齐 `Infrastructure/Persistence` 长路径重构版本，弃用旧版 `Core` 目录。

## 2. 目录架构说明
| 模块 | 旧版目录 (待删除/弃用) | 新版目录 (唯一基准) |
| :--- | :--- | :--- |
| **元数据 ORM** | `Core/MetaDatas` | `Infrastructure/Persistence/ORM` |
| **业务实体** | 引用 `Core.MetaDatas` 命名空间 | 引用 `Infrastructure.Persistence.ORM` 命名空间 |

## 3. 缓存同步机制
新版架构（长路径版）更注重事务与缓存的精准配合。

### A. 自动同步说明
在新版 `Infrastructure/Persistence/ORM/MetaDataUpdate.cs` 中，底层不强制在 `UpdateAsync` 后自动删除全量缓存，而是要求业务层结合事务进行精准同步。

### B. 手动同步 (核心规范)
在执行 `ExecTrans`（多表事务）或自定义 SQL 资产更新（如 `SqlPlus` 相对增减）后，必须显式调用 `SyncCacheField`。
- **核心方法**：`SyncCacheField(long id, string field, object value)`
- **同步要求**：
  1. 计算变动后的最终值。
  2. 确保 `ExecTrans` 返回 `0`（成功）后再同步。
  3. 若涉及转账，需同步发起者和接收者双方。

## 4. 资产操作重构范式 (Example)
```csharp
public static int AddAsset(long qq, long addValue)
{
    var finalValue = GetAsset(qq) + addValue; // 1. 先计算最终值
    var sql1 = ...; // 资产更新SQL
    var sql2 = ...; // 日志SQL
    
    int result = ExecTrans(sql1, sql2); // 2. 执行事务
    if (result == 0) 
    {
        SyncCacheField(qq, "AssetField", finalValue); // 3. 成功后同步缓存
    }
    return result;
}
```

## 5. 重构红线
- **禁止批量替换**：必须逐个文件审计引用和逻辑，防止命名空间冲突。
- **命名空间锁定**：所有重构必须指向 `BotWorker.Infrastructure.Persistence.ORM`。
- **清理原则**：在确认业务逻辑平移完成后，视情况清理 `Core` 目录。
