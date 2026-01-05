import pyodbc
import psycopg2
from psycopg2.extras import execute_values
import uuid
import json
import os
import sys
import re

# 迁移配置
DATABASES_TO_MIGRATE = [
    "sz84_robot",
    # "other_db_1",
    # "other_db_2",
]

MSSQL_CONFIG = {
    "DRIVER": "{ODBC Driver 17 for SQL Server}",
    "SERVER": "192.168.0.114,1433",
    "UID": "derlin",
    "PWD": "fkueiqiq461686"
}

PG_CONFIG = {
    "host": "192.168.0.114",
    "port": 5432,
    "user": "derlin",
    "password": "fkueiqiq461686"
}

TEST_TABLE = None  # 设置要测试的表名，设为 None 则迁移所有表
FORCE_RECREATE = True # 如果设置为 True，则每次运行都会先删除 PG 中已存在的表（小心使用）
DROP_PUBLIC_SCHEMA = False # 如果设置为 True，运行前会清空整个 PostgreSQL 的 public 模式（危险：相当于重置数据库）
LIMIT_ROWS = None # 测试阶段限制每个表迁移的行数，设为 None 则迁移所有数据
INCREMENTAL_MODE = False # 如果设置为 True，将根据主键进行增量迁移（仅支持单列数值主键）

# 类型映射函数
def map_type(sqlserver_type, value):
    if value is None:
        return None
    sqlserver_type = sqlserver_type.lower()
    
    if "uniqueidentifier" in sqlserver_type:
        return str(value)
    elif "bit" in sqlserver_type:
        return bool(value)
    elif "bigint" in sqlserver_type or "int" in sqlserver_type:
        return int(value)
    elif "decimal" in sqlserver_type or "numeric" in sqlserver_type or "money" in sqlserver_type:
        return value  # psycopg2 handles Decimal
    elif "float" in sqlserver_type or "real" in sqlserver_type:
        return float(value)
    elif "datetime" in sqlserver_type or "smalldatetime" in sqlserver_type:
        return value
    elif "binary" in sqlserver_type or "varbinary" in sqlserver_type or "image" in sqlserver_type:
        return psycopg2.Binary(value)
    else:
        # PostgreSQL 不支持在 text/varchar 中存储 NUL (0x00) 字符
        val_str = str(value)
        if '\x00' in val_str:
            return val_str.replace('\x00', '')
        return val_str

def get_pg_type(sql_type):
    sql_type = sql_type.lower()
    if "uniqueidentifier" in sql_type:
        return "uuid"
    elif "bit" in sql_type:
        return "boolean"
    elif "bigint" in sql_type:
        return "bigint"
    elif "int" in sql_type:
        return "bigint"
    elif "decimal" in sql_type or "numeric" in sql_type or "money" in sql_type:
        return "numeric"
    elif "float" in sql_type or "real" in sql_type:
        return "double precision"
    elif "datetime" in sql_type or "smalldatetime" in sql_type:
        return "timestamptz"
    elif "date" == sql_type:
        return "date"
    elif "xml" in sql_type:
        return "xml"
    elif "binary" in sql_type or "varbinary" in sql_type or "image" in sql_type:
        return "bytea"
    else:
        return "text"

def translate_sql(definition, obj_type):
    """
    极简 T-SQL 到 PL/pgSQL 转换启发式算法
    注意：复杂逻辑仍需人工微调
    """
    sql = definition
    
    # 1. 处理方括号
    sql = sql.replace("[", "\"").replace("]", "\"")
    
    # 2. 基础语法替换
    sql = re.sub(r'CREATE\s+PROCEDURE', 'CREATE OR REPLACE PROCEDURE', sql, flags=re.IGNORECASE)
    sql = re.sub(r'CREATE\s+VIEW', 'CREATE OR REPLACE VIEW', sql, flags=re.IGNORECASE)
    sql = re.sub(r'CREATE\s+FUNCTION', 'CREATE OR REPLACE FUNCTION', sql, flags=re.IGNORECASE)
    
    # 3. 函数常用替换
    sql = re.sub(r'GETDATE\(\)', 'NOW()', sql, flags=re.IGNORECASE)
    sql = re.sub(r'GETUTCDATE\(\)', "NOW() AT TIME ZONE 'UTC'", sql, flags=re.IGNORECASE)
    sql = re.sub(r'ISNULL\(', 'COALESCE(', sql, flags=re.IGNORECASE)
    sql = re.sub(r'LEN\(', 'LENGTH(', sql, flags=re.IGNORECASE)
    
    # 4. 字符串处理
    # N'string' -> 'string'
    sql = re.sub(r"N'(.*?)'", r"'\1'", sql)
    
    # 5. 存储过程和函数结构转换 (非常粗略)
    if obj_type in ('P', 'FN', 'IF', 'TF'):
        # 尝试寻找 AS BEGIN ... END 结构
        # T-SQL: AS BEGIN ... END
        # PL/pgSQL: AS $$ BEGIN ... END; $$ LANGUAGE plpgsql;
        if "BEGIN" in sql.upper() and "END" in sql.upper():
            # 这是一个非常简单的替换，实际可能更复杂
            sql = re.sub(r'\bAS\b\s+BEGIN', 'AS $$\nBEGIN', sql, flags=re.IGNORECASE)
            # 在最后一个 END 后添加结束符
            last_end_idx = sql.upper().rfind("END")
            if last_end_idx != -1:
                sql = sql[:last_end_idx] + "END;\n$$ LANGUAGE plpgsql;" + sql[last_end_idx+3:]

    return sql

def get_pg_default(mssql_default, pg_type):
    if mssql_default is None:
        return None
    d = mssql_default.strip()
    # 移除多余的括号
    while d.startswith('(') and d.endswith(')'):
        d = d[1:-1]
    
    d_lower = d.lower()
    
    # 处理布尔值默认值
    if pg_type == 'boolean':
        if d_lower in ('0', '((0))', 'false'):
            return 'false'
        if d_lower in ('1', '((1))', 'true'):
            return 'true'

    if 'getdate()' in d_lower or 'getutcdate()' in d_lower:
        return 'CURRENT_TIMESTAMP'
    elif 'newid()' in d_lower or 'newsequentialid()' in d_lower:
        return 'gen_random_uuid()'
    elif d_lower == '0' or d_lower == '((0))':
        return '0'
    elif d_lower == '1' or d_lower == '((1))':
        return '1'
    elif d.startswith("N'") or d.startswith("'"):
        # 处理 SQL Server 的 N'string' 语法
        val = d[2:-1] if d.startswith("N'") else d[1:-1]
        # 简单处理日期格式的默认值 (1970)-(1)-(1)
        if pg_type == 'timestamptz' or pg_type == 'date':
            val = val.replace(')-(', '-').replace('(', '').replace(')', '')
        return f"'{val}'"
    
    # 处理类似 ((1970)-(1)-(1)) 的奇怪格式
    if (pg_type == 'timestamptz' or pg_type == 'date') and '-' in d:
        val = d.replace(')-(', '-').replace('(', '').replace(')', '')
        return f"'{val}'"

    return d

def migrate_database(db_name):
    print(f"\n{'='*50}")
    print(f"STARTING MIGRATION FOR DATABASE: {db_name}")
    print(f"{'='*50}")

    # SQL Server 连接
    mssql_conn_str = f"DRIVER={MSSQL_CONFIG['DRIVER']};SERVER={MSSQL_CONFIG['SERVER']};DATABASE={db_name};UID={MSSQL_CONFIG['UID']};PWD={MSSQL_CONFIG['PWD']}"
    mssql_conn = pyodbc.connect(mssql_conn_str)
    mssql_cursor = mssql_conn.cursor()

    # PostgreSQL 连接
    pg_conn = psycopg2.connect(
        host=PG_CONFIG['host'],
        port=PG_CONFIG['port'],
        dbname=db_name, # 假设 PG 中已经创建了同名数据库
        user=PG_CONFIG['user'],
        password=PG_CONFIG['password']
    )
    pg_cursor = pg_conn.cursor()

    # 启用必要扩展
    pg_cursor.execute("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")
    pg_cursor.execute("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\";")
    pg_conn.commit()

    # 执行数据库级重置逻辑
    if DROP_PUBLIC_SCHEMA:
        print(f"Warning: DROP_PUBLIC_SCHEMA is True for {db_name}. Cleaning up target schema...")
        pg_cursor.execute("DROP SCHEMA IF EXISTS public CASCADE;")
        pg_cursor.execute("CREATE SCHEMA public;")
        pg_cursor.execute("GRANT ALL ON SCHEMA public TO public;")
        pg_cursor.execute(f"GRANT ALL ON SCHEMA public TO {PG_CONFIG['user']};")
        pg_conn.commit()
        print(f"Target schema reset completed for {db_name}.")

    # 获取 SQL Server 所有表及其数据量
    query = """
    SELECT 
        s.name AS TABLE_SCHEMA, 
        t.name AS TABLE_NAME,
        p.rows AS ROW_COUNT
    FROM sys.tables t
    JOIN sys.schemas s ON t.schema_id = s.schema_id
    JOIN sys.partitions p ON t.object_id = p.object_id
    WHERE p.index_id IN (0,1) -- 0:堆, 1:聚簇索引
    """
    if TEST_TABLE:
        query += f" AND t.name = '{TEST_TABLE}'"

    mssql_cursor.execute(query)
    all_tables = mssql_cursor.fetchall()

    # 按行数升序排序 (从小表到大表)
    all_tables.sort(key=lambda x: x[2])
    tables_to_migrate = all_tables

    if not tables_to_migrate:
        print(f"No tables found in {db_name} to migrate.")
    else:
        print(f"Total tables to migrate in {db_name}: {len(tables_to_migrate)}.")

    # 遍历每个表
    for i, (schema, table, row_count) in enumerate(tables_to_migrate):
        table_full_name = f"{schema}.{table}"
        try:
            print(f"[{i+1}/{len(tables_to_migrate)}] Start migrating table {table_full_name} ({row_count} rows)")

            # 获取列信息
            mssql_cursor.execute(f"""
                SELECT 
                    c.name AS COLUMN_NAME, 
                    t.name AS DATA_TYPE, 
                    d.definition AS COLUMN_DEFAULT, 
                    CASE WHEN c.is_nullable = 1 THEN 'YES' ELSE 'NO' END AS IS_NULLABLE,
                    c.is_identity AS IS_IDENTITY,
                    ep.value AS COLUMN_COMMENT
                FROM sys.columns c
                JOIN sys.types t ON c.user_type_id = t.user_type_id
                JOIN sys.tables st ON c.object_id = st.object_id
                LEFT JOIN sys.default_constraints d ON c.default_object_id = d.object_id
                LEFT JOIN sys.extended_properties ep ON ep.major_id = st.object_id 
                    AND ep.minor_id = c.column_id 
                    AND ep.name = 'MS_Description'
                WHERE st.name = '{table}' AND SCHEMA_NAME(st.schema_id) = '{schema}'
                ORDER BY c.column_id
            """)
            columns_info = mssql_cursor.fetchall()
            column_names = [col[0] for col in columns_info]

            # 获取主键信息
            mssql_cursor.execute(f"""
                SELECT COLUMN_NAME
                FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE KU
                JOIN INFORMATION_SCHEMA.TABLE_CONSTRAINTS TC ON KU.CONSTRAINT_NAME = TC.CONSTRAINT_NAME
                WHERE TC.CONSTRAINT_TYPE = 'PRIMARY KEY' AND KU.TABLE_SCHEMA='{schema}' AND KU.TABLE_NAME = '{table}'
            """)
            pk_columns = [row[0] for row in mssql_cursor.fetchall()]

            # 创建 PostgreSQL 表
            pg_cols = []
            column_comments = []
            for col_name, col_type, col_default, is_nullable, is_identity, col_comment in columns_info:
                pg_type = get_pg_type(col_type)
                col_def = f"\"{col_name}\" {pg_type}"
                if is_identity:
                    col_def += " GENERATED BY DEFAULT AS IDENTITY"
                elif col_default:
                    pg_default = get_pg_default(col_default, pg_type)
                    if pg_default:
                        col_def += f" DEFAULT {pg_default}"
                if is_nullable == 'NO' and col_name not in pk_columns:
                    col_def += " NOT NULL"
                if col_name in pk_columns and len(pk_columns) == 1:
                     col_def += " PRIMARY KEY"
                pg_cols.append(col_def)
                if col_comment:
                    column_comments.append(f"COMMENT ON COLUMN \"{table}\".\"{col_name}\" IS %s;")

            pk_def = ""
            if len(pk_columns) > 1:
                pk_def = f", PRIMARY KEY ({', '.join([f'\"{c}\"' for c in pk_columns])})"

            should_recreate = FORCE_RECREATE
            if INCREMENTAL_MODE:
                pg_cursor.execute("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = %s);", (table.lower(),))
                if pg_cursor.fetchone()[0]:
                    should_recreate = False
                    print(f"  Incremental mode: Table {table} exists, skipping recreation.")

            if should_recreate:
                pg_cursor.execute(f"DROP TABLE IF EXISTS \"{table}\" CASCADE;")
                pg_conn.commit()

            create_sql = f"CREATE TABLE IF NOT EXISTS \"{table}\" ({', '.join(pg_cols)}{pk_def});"
            pg_cursor.execute(create_sql)
            for comment_sql in column_comments:
                col_comment = next(c[5] for c in columns_info if f"COMMENT ON COLUMN \"{table}\".\"{c[0]}\" IS %s;" == comment_sql)
                pg_cursor.execute(comment_sql, (col_comment,))
            pg_conn.commit()

            max_id = None
            if INCREMENTAL_MODE and len(pk_columns) == 1:
                pk_col = pk_columns[0]
                pk_info = next(c for c in columns_info if c[0] == pk_col)
                if "int" in pk_info[1].lower() or "decimal" in pk_info[1].lower() or "numeric" in pk_info[1].lower():
                    pg_cursor.execute(f"SELECT MAX(\"{pk_col}\") FROM \"{table}\"")
                    max_id = pg_cursor.fetchone()[0]
                    if max_id is not None:
                        print(f"  Incremental mode: Starting from {pk_col} > {max_id}")

            # 迁移索引
            mssql_cursor.execute(f"""
                SELECT i.name, i.is_unique, c.name
                FROM sys.indexes i
                INNER JOIN sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
                INNER JOIN sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
                INNER JOIN sys.tables t ON i.object_id = t.object_id
                WHERE t.name = '{table}' AND SCHEMA_NAME(t.schema_id) = '{schema}'
                AND i.is_primary_key = 0
                ORDER BY i.name, ic.key_ordinal
            """)
            indexes = {}
            for idx_name, is_unique, col_name in mssql_cursor.fetchall():
                if idx_name not in indexes:
                    indexes[idx_name] = {'unique': is_unique, 'columns': []}
                indexes[idx_name]['columns'].append(f"\"{col_name}\"")

            for idx_name, idx_info in indexes.items():
                unique_str = "UNIQUE" if idx_info['unique'] else ""
                cols_str = ", ".join(idx_info['columns'])
                safe_idx_name = f"idx_{table}_{idx_name}"[:63] 
                create_idx_sql = f"CREATE {unique_str} INDEX IF NOT EXISTS \"{safe_idx_name}\" ON \"{table}\" ({cols_str});"
                try:
                    pg_cursor.execute(create_idx_sql)
                    pg_conn.commit()
                except:
                    pg_conn.rollback()

            # 数据拉取
            fetch_sql = f"SELECT * FROM [{schema}].[{table}]"
            if max_id is not None:
                fetch_sql += f" WHERE [{pk_col}] > {max_id}"
            mssql_cursor.execute(fetch_sql)
            
            total_rows = 0
            while True:
                rows = mssql_cursor.fetchmany(500)
                if not rows: break
                values = [[map_type(columns_info[i][1], row[i]) for i in range(len(columns_info))] for row in rows]
                quoted_col_names = [f"\"{name}\"" for name in column_names]
                conflict_clause = "ON CONFLICT DO NOTHING" if pk_columns else ""
                insert_sql = f"INSERT INTO \"{table}\" ({', '.join(quoted_col_names)}) VALUES %s {conflict_clause}"
                try:
                    execute_values(pg_cursor, insert_sql, values)
                    pg_conn.commit()
                except Exception as e:
                    pg_conn.rollback()
                    for row_val in values:
                        try:
                            execute_values(pg_cursor, f"INSERT INTO \"{table}\" ({', '.join(quoted_col_names)}) VALUES %s {conflict_clause}", [row_val])
                            pg_conn.commit()
                        except:
                            pg_conn.rollback()
                total_rows += len(rows)
                if total_rows % 500 == 0: print(f"  Migrated {total_rows} rows...")

            # 同步序列
            for col_name, col_type, col_default, is_nullable, is_identity, col_comment in columns_info:
                if is_identity:
                    sync_seq_sql = f"SELECT setval(pg_get_serial_sequence('\"{table}\"', '{col_name}'), COALESCE(MAX(\"{col_name}\"), 0) + 1, false) FROM \"{table}\";"
                    try:
                        pg_cursor.execute(sync_seq_sql)
                        pg_conn.commit()
                    except:
                        pg_conn.rollback()

        except Exception as e:
            print(f"Error migrating table {table_full_name}: {e}")
            pg_conn.rollback()

    # 迁移对象
    print(f"Starting object migration for {db_name}...")
    mssql_cursor.execute("""
        SELECT o.name, o.type, o.type_desc, m.definition
        FROM sys.sql_modules m JOIN sys.objects o ON m.object_id = o.object_id
        WHERE o.type IN ('V', 'P', 'FN', 'IF', 'TF') AND o.is_ms_shipped = 0
    """)
    modules = mssql_cursor.fetchall()
    obj_file = f"migration_objects_{db_name}.sql"
    with open(obj_file, "w", encoding="utf-8") as f:
        for name, obj_type, type_desc, definition in modules:
            pg_sql = translate_sql(definition, obj_type.strip())
            f.write(f"-- Object: {name}\n{pg_sql}\n\nGO\n\n")
            try:
                pg_cursor.execute(pg_sql)
                pg_conn.commit()
            except:
                pg_conn.rollback()
    
    print(f"Migration for {db_name} finished!")
    mssql_conn.close()
    pg_cursor.close()
    pg_conn.close()

if __name__ == "__main__":
    for db in DATABASES_TO_MIGRATE:
        try:
            migrate_database(db)
        except Exception as e:
            print(f"FAILED to migrate database {db}: {e}")
    print("\nAll database migration tasks finished!")
