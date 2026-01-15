import pyodbc
import psycopg2
from psycopg2.extras import execute_values
import uuid
import json
import os
import sys
import re
import logging
from datetime import datetime
from psycopg2 import sql

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [%(levelname)s] %(message)s',
    handlers=[
        logging.FileHandler(f"migration_baseinfo_{datetime.now().strftime('%Y%m%d_%H%M%S')}.log"),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)

# 迁移配置
DATABASES_TO_MIGRATE = [
    "baseinfo",
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

BATCH_SIZE = 1000
TEST_TABLE = None  # 设置要测试的表名，设为 None 则迁移所有表
FORCE_RECREATE = True # 如果设置为 True，则每次运行都会先删除 PG 中已存在的表
DROP_PUBLIC_SCHEMA = False # 如果设置为 True，运行前会清空整个 PostgreSQL 的 public 模式
LIMIT_ROWS = None # 测试阶段限制每个表迁移的行数
INCREMENTAL_MODE = False # 如果设置为 True，将根据主键进行增量迁移

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
        return value
    elif "float" in sqlserver_type or "real" in sqlserver_type:
        return float(row) if isinstance(value, (int, float, str)) else value # fallback
    elif "datetime" in sqlserver_type or "smalldatetime" in sqlserver_type:
        return value
    elif "binary" in sqlserver_type or "varbinary" in sqlserver_type or "image" in sqlserver_type:
        return psycopg2.Binary(value)
    else:
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

def get_pg_default(mssql_default, pg_type):
    if mssql_default is None:
        return None
    d = mssql_default.strip()
    while d.startswith('(') and d.endswith(')'):
        d = d[1:-1]
    
    d_lower = d.lower()
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
        val = d[2:-1] if d.startswith("N'") else d[1:-1]
        if pg_type == 'timestamptz' or pg_type == 'date':
            val = val.replace(')-(', '-').replace('(', '').replace(')', '')
        return f"'{val}'"
    
    if (pg_type == 'timestamptz' or pg_type == 'date') and '-' in d:
        val = d.replace(')-(', '-').replace('(', '').replace(')', '')
        return f"'{val}'"

    return d

def ensure_pg_database_exists(db_name):
    try:
        conn = psycopg2.connect(
            host=PG_CONFIG['host'],
            port=PG_CONFIG['port'],
            dbname='postgres',
            user=PG_CONFIG['user'],
            password=PG_CONFIG['password']
        )
        conn.autocommit = True
        cursor = conn.cursor()
        
        cursor.execute("SELECT 1 FROM pg_database WHERE datname = %s", (db_name,))
        if not cursor.fetchone():
            logger.info(f"Creating database {db_name} in PostgreSQL...")
            cursor.execute(sql.SQL("CREATE DATABASE {}").format(sql.Identifier(db_name)))
        else:
            logger.info(f"Database {db_name} already exists in PostgreSQL.")
        
        cursor.close()
        conn.close()
    except Exception as e:
        logger.error(f"Error ensuring database exists: {e}")
        raise

def migrate_database(db_name):
    logger.info(f"{'='*50}")
    logger.info(f"STARTING MIGRATION FOR DATABASE: {db_name}")
    logger.info(f"{'='*50}")

    ensure_pg_database_exists(db_name)

    try:
        mssql_conn_str = f"DRIVER={MSSQL_CONFIG['DRIVER']};SERVER={MSSQL_CONFIG['SERVER']};DATABASE={db_name};UID={MSSQL_CONFIG['UID']};PWD={MSSQL_CONFIG['PWD']}"
        mssql_conn = pyodbc.connect(mssql_conn_str)
        mssql_cursor = mssql_conn.cursor()

        pg_conn = psycopg2.connect(
            host=PG_CONFIG['host'],
            port=PG_CONFIG['port'],
            dbname=db_name,
            user=PG_CONFIG['user'],
            password=PG_CONFIG['password']
        )
        pg_cursor = pg_conn.cursor()

        pg_cursor.execute("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")
        pg_cursor.execute("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\";")
        pg_conn.commit()

        if DROP_PUBLIC_SCHEMA:
            logger.warning(f"DROP_PUBLIC_SCHEMA is True for {db_name}. Cleaning up target schema...")
            pg_cursor.execute("DROP SCHEMA IF EXISTS public CASCADE;")
            pg_cursor.execute("CREATE SCHEMA public;")
            pg_cursor.execute("GRANT ALL ON SCHEMA public TO public;")
            pg_cursor.execute(f"GRANT ALL ON SCHEMA public TO {PG_CONFIG['user']};")
            pg_conn.commit()

        query = """
        SELECT 
            s.name AS TABLE_SCHEMA, 
            t.name AS TABLE_NAME,
            p.rows AS ROW_COUNT
        FROM sys.tables t
        JOIN sys.schemas s ON t.schema_id = s.schema_id
        JOIN sys.partitions p ON t.object_id = p.object_id
        WHERE p.index_id IN (0,1)
        """
        if TEST_TABLE:
            query += f" AND t.name = '{TEST_TABLE}'"

        mssql_cursor.execute(query)
        all_tables = mssql_cursor.fetchall()
        all_tables.sort(key=lambda x: x[2])
        
        if not all_tables:
            logger.info(f"No tables found in {db_name} to migrate.")
            return

        logger.info(f"Total tables to migrate: {len(all_tables)}")

        for i, (schema, table, row_count) in enumerate(all_tables):
            table_full_name = f"{schema}.{table}"
            try:
                logger.info(f"[{i+1}/{len(all_tables)}] Migrating {table_full_name} ({row_count} rows)...")

                mssql_cursor.execute(f"""
                    SELECT c.name, t.name, d.definition, 
                           CASE WHEN c.is_nullable = 1 THEN 'YES' ELSE 'NO' END,
                           c.is_identity
                    FROM sys.columns c
                    JOIN sys.types t ON c.user_type_id = t.user_type_id
                    JOIN sys.tables st ON c.object_id = st.object_id
                    LEFT JOIN sys.default_constraints d ON c.default_object_id = d.object_id
                    WHERE st.name = '{table}' AND SCHEMA_NAME(st.schema_id) = '{schema}'
                    ORDER BY c.column_id
                """)
                columns_info = mssql_cursor.fetchall()
                column_names = [col[0] for col in columns_info]

                mssql_cursor.execute(f"""
                    SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE KU
                    JOIN INFORMATION_SCHEMA.TABLE_CONSTRAINTS TC ON KU.CONSTRAINT_NAME = TC.CONSTRAINT_NAME
                    WHERE TC.CONSTRAINT_TYPE = 'PRIMARY KEY' AND KU.TABLE_SCHEMA='{schema}' AND KU.TABLE_NAME = '{table}'
                """)
                pk_columns = [row[0] for row in mssql_cursor.fetchall()]

                pg_cols = []
                for col_name, col_type, col_default, is_nullable, is_identity in columns_info:
                    pg_type = get_pg_type(col_type)
                    col_def = f"\"{col_name}\" {pg_type}"
                    if is_identity:
                        col_def += " GENERATED BY DEFAULT AS IDENTITY"
                    elif col_default:
                        pg_default = get_pg_default(col_default, pg_type)
                        if pg_default: col_def += f" DEFAULT {pg_default}"
                    if is_nullable == 'NO' and col_name not in pk_columns:
                        col_def += " NOT NULL"
                    if col_name in pk_columns and len(pk_columns) == 1:
                        col_def += " PRIMARY KEY"
                    pg_cols.append(col_def)

                if FORCE_RECREATE:
                    pg_cursor.execute(f"DROP TABLE IF EXISTS \"{table}\" CASCADE;")
                
                pg_cursor.execute(f"CREATE TABLE IF NOT EXISTS \"{table}\" ({', '.join(pg_cols)});")
                pg_conn.commit()

                # 数据迁移
                select_query = f"SELECT {', '.join([f'[{c}]' for c in column_names])} FROM {table_full_name}"
                if LIMIT_ROWS:
                    select_query = f"SELECT TOP {LIMIT_ROWS} {', '.join([f'[{c}]' for c in column_names])} FROM {table_full_name}"
                
                mssql_cursor.execute(select_query)
                
                total_migrated = 0
                while True:
                    rows = mssql_cursor.fetchmany(BATCH_SIZE)
                    if not rows:
                        break
                    
                    data = []
                    for row in rows:
                        data.append(tuple(map_type(columns_info[j][1], row[j]) for j in range(len(column_names))))
                    
                    insert_query = f"INSERT INTO \"{table}\" ({', '.join([f'\"{c}\"' for c in column_names])}) VALUES %s"
                    execute_values(pg_cursor, insert_query, data)
                    pg_conn.commit()
                    total_migrated += len(rows)
                    if row_count > 0:
                        logger.info(f"  Progress: {total_migrated}/{row_count} ({(total_migrated/row_count)*100:.1f}%)")

                logger.info(f"  Successfully migrated {total_migrated} rows.")

            except Exception as e:
                logger.error(f"  Failed to migrate table {table_full_name}: {e}")
                pg_conn.rollback()

    except Exception as e:
        logger.error(f"Database migration failed: {e}")
    finally:
        if 'mssql_conn' in locals(): mssql_conn.close()
        if 'pg_conn' in locals(): pg_conn.close()

if __name__ == "__main__":
    for db in DATABASES_TO_MIGRATE:
        migrate_database(db)
    logger.info("All migrations completed.")
