# coding: utf-8
from SQLConn import *
from decimal import *

class MetaData(object):
    
    database = ""
    table_name = ""
    key_field = ""
    key_field2 = ""
    key_value = ""
    key_value2 = ""        
    
    @classmethod
    def get_value(cls, field_name, key_value, key_value2=None):
        database = cls.database
        table_name = cls.table_name
        key_field = cls.key_field
        key_field2 = cls.key_field2

        sql = f"SELECT {field_name} FROM {database}.dbo.{table_name} WHERE {key_field} = %s"
        params = (key_value,)

        if key_field2:
            sql += f" AND {key_field2} = %s"
            params += (key_value2,)

        return cls.Query(sql, params)

    @classmethod
    def get_int(cls, field_name, key_value, key_value2=None):
        res = cls.get_value(field_name, key_value, key_value2)
        return int(res) if res else 0

    @classmethod
    def get_float(cls, field_name, key_value, key_value2=None):
        res = cls.get_value(field_name, key_value, key_value2)
        return float(res) if res else 0.0

    @classmethod
    def exists(cls, key_value, key_value2=None):
        return cls.get_int("1", key_value, key_value2) == 1

    @classmethod
    def not_exists(cls, key_value, key_value2=None):
        return not cls.exists(key_value, key_value2)

    def sql_delete(self):
        sql = f"DELETE FROM {self.database}.dbo.{self.table_name} WHERE {self.key_field} = %s"
        params = (self.key_value,)

        if self.key_field2:
            sql += f" AND {self.key_field2} = %s"
            params += (self.key_value2,)

        return sql, params

    def delete_self(self):
        sql, params = self.sql_delete()
        return SQLConn.Exec(sql, params)
    
    @classmethod
    def delete(cls, key_value, key_value2=None):
        _cls = cls(key_value, key_value2) if key_value2 else cls(key_value)
        sql, params = _cls.sql_delete()
        return SQLConn.Exec(sql, params)
    
    def get_sql_plus_self(self, field_name, plus_value):
        if not self.key_field2:
            sql = f"update {self.database}.dbo.{self.table_name} set {field_name} = isnull({field_name}, 0) + %s where {self.key_field} = %s"
            return sql, (plus_value, self.key_value)
        else:
            sql = f"update {self.database}.dbo.{self.table_name} set {field_name} = isnull({field_name}, 0) + %s where {self.key_field} = %s and {self.key_field2} = %s"
            return sql, (plus_value, self.key_value, self.key_value2)

    def plus_value_self(self, field_name, plus_value):
        sql, params = self.get_sql_plus_self(field_name, plus_value)
        return SQLConn.Exec(sql, params)

    @classmethod
    def get_sql_plus(cls, field_name, plus_value, key_value, key_value2=None):
        _cls = cls(key_value, key_value2) if key_value2 else cls(key_value)
        return _cls.get_sql_plus_self(field_name, plus_value)

    @classmethod
    def set_value_plus(cls, field_name, plus_value, key_value, key_value2=None):
        return cls.Exec(cls.get_sql_plus(field_name, plus_value, key_value, key_value2)) 
    
    #update sql
    def get_sql_upd_self(self, field_name, field_value):
        field_value = str(field_value).replace("'", "''")
        if (self.key_field2 == ""):
            return str.format("update {0}.dbo.{1} set {2} = '{3}' where {4} = '{5}'", self.database, self.table_name, field_name, field_value, self.key_field, self.key_value)
        else: 
            return str.format("update {0}.dbo.{1} set {2} = '{3}' where {4} = '{5}' and {6} = '{7}'", self.database, self.table_name, field_name, field_value, self.key_field, self.key_value, self.key_field2, self.key_value2)  
    
    def set_value_self(self, field_name, field_value):
        sql = self.get_sql_upd_self(field_name, field_value)
        return self.Exec(sql)
    
    @classmethod
    def get_sql_upd(cls, field_name, field_value, key_value, key_value2=None):
        _cls = cls(key_value, key_value2) if key_value2 else cls(key_value)
        return _cls.get_sql_upd_self(field_name, field_value)

    @classmethod
    def set_value(cls, field_name, field_value, key_value, key_value2=None):
        return cls.Exec(cls.get_sql_upd(field_name, field_value, key_value, key_value2)) 
    
    #update other sql
    def get_sql_upd_other_self(self, field_name, other_value):
        if (self.key_field2 == ""):
            return str.format("update {0}.dbo.{1} set {2} = {3} where {4} = {5}", self.database, self.table_name, field_name, other_value, self.key_field, self.key_value)
        else: 
            return str.format("update {0}.dbo.{1} set {2} = {3} where {4} = {5} and {6} = {7}", self.database, self.table_name, field_name, other_value, self.key_field, self.key_value, self.key_field2, self.key_value2)       

    #other
    def set_value_other_self(self, field_name, other_value):
        sql = self.get_sql_upd_other_self(field_name, other_value)
        return self.Exec(sql)

    @classmethod
    def get_sql_upd_other(cls, field_name, field_value, key_value, key_value2=None):
        _cls = cls(key_value, key_value2) if key_value2 else cls(key_value)
        return _cls.get_sql_upd_other_self(field_name, field_value)

    @classmethod
    def set_value_other(cls, field_name, field_value, key_value, key_value2=None):
        _cls = cls(key_value, key_value2) if key_value2 else cls(key_value)
        return _cls.set_value_other_self(field_name, field_value)

    #date
    def set_value_date_self(self, field_name):
        return self.set_value_other_self(field_name, "getdate()")

    @classmethod
    def set_value_date(cls, field_name, key_value, key_value2=None):
        _cls = cls(key_value, key_value2) if key_value2 else cls(key_value)
        return _cls.set_value_date_self(field_name)  

    #exec query trans res
    @staticmethod
    def Exec(sql, params=None):
        return SQLConn.Exec(sql, params)

    @staticmethod
    def Query(sql, params=None):
        return SQLConn.Query(sql, params)

    @staticmethod
    def getQueryRes(sql, format):
        return SQLConn.QueryRes(sql, format)

    @staticmethod
    def ExecTrans(*sqls):
        return SQLConn.ExecTrans(*sqls)

    @staticmethod
    def QueryRes(sql, format):
        return SQLConn.QueryRes(sql, format)
