import pyodbc
import pymssql
import datetime
import os
import json
from common import *

class SQLConn():
    @staticmethod
    def get_config():
        config_path = os.path.join(os.path.dirname(__file__), 'config.json')
        if os.path.exists(config_path):
            try:
                with open(config_path, 'r', encoding='utf-8') as f:
                    return json.load(f)
            except:
                pass
        return {}

    @staticmethod
    def conn():
        config = SQLConn.get_config().get('database', {})
        
        server = os.environ.get('DB_SERVER', config.get('server', ''))
        database = os.environ.get('DB_NAME', config.get('name', ''))
        username = os.environ.get('DB_USER', config.get('user', ''))
        password = os.environ.get('DB_PASSWORD', config.get('password', ''))

        if not all([server, database, username, password]):
            # print("Database configuration missing.")
            return None

        # 使用 pyodbc 连接
        conn_str = f"DRIVER={{SQL Server}};SERVER={server};DATABASE={database};UID={username};PWD={password}"
        try:
            return pyodbc.connect(conn_str)
        except pyodbc.Error:
            pass

        # 使用 pymssql 连接
        conn = pymssql.connect(server=server, database=database, user=username, password=password)
        return conn
   
    @staticmethod
    def Query(sql, params=None):
        try:
            _conn = SQLConn.conn()
            if not _conn: return ""
            cursor = _conn.cursor()
            if common.is_debug_sql:
                print("sql:\n", sql, params)
            if params:
                cursor.execute(sql, params)
            else:
                cursor.execute(sql)
            row = cursor.fetchone()
            _conn.close()
            if row:
                res = row[0]
                res = str(res) if res else ""
            else:
                res = ""
            if common.is_debug_sql:
                print("sql result:", res)
            return res
        except Exception as e:
            if common.is_debug_sql:
                print("sql result:", "")
            return ""

    @staticmethod
    def QueryDict(sql, params=None):
        try:
            _conn = SQLConn.conn()
            if not _conn: return []
            cursor = _conn.cursor(as_dict=True)
            if common.is_debug_sql:
                print("query sql:", sql, params)
            if params:
                cursor.execute(sql, params)
            else:
                cursor.execute(sql)
            res = cursor.fetchall()
            _conn.close()
            if common.is_debug_sql:
                print("sql result:", res)
            return res
        except Exception as e:
            if common.is_debug_sql:
                print("sql result:", "")
            return []
    
    @staticmethod
    def Exec(sql, params=None):
        _conn = None
        try:
            _conn = SQLConn.conn()
            if not _conn: return False
            cursor = _conn.cursor()
            if common.is_debug_sql:
                print("exec sql:", sql, params)
            if params:
                cursor.execute(sql, params)
            else:
                cursor.execute(sql)
            _conn.commit()
            if common.is_debug_sql:
                print("exec result: True")
            return True
        except Exception as e:
            if _conn: _conn.rollback()
            # ALWAYS print error for critical DB failures
            print(f"[SQLConn] Exec Failed: {e} SQL={sql} Params={params}")
            if common.is_debug_sql:
                print("exec result: False")
            return False
        finally:
            if _conn: _conn.close()

    @staticmethod
    def ExecTrans(*sqls):
        _conn = None
        try:
            _conn = SQLConn.conn()
            if not _conn: return False
            cursor = _conn.cursor()
            for sql in sqls:
                cursor.execute(sql)
            _conn.commit()
            return True
        except Exception as e:
            if _conn: _conn.rollback()
            return False
        finally:
            if _conn: _conn.close()
