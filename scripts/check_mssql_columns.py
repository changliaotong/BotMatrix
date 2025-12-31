import pyodbc
MSSQL_CONFIG = {
    "DRIVER": "{ODBC Driver 17 for SQL Server}",
    "SERVER": "192.168.0.114,1433",
    "DATABASE": "sz84_robot",
    "UID": "derlin",
    "PWD": "fkueiqiq461686"
}
try:
    conn = pyodbc.connect(
        f"DRIVER={MSSQL_CONFIG['DRIVER']};SERVER={MSSQL_CONFIG['SERVER']};DATABASE={MSSQL_CONFIG['DATABASE']};UID={MSSQL_CONFIG['UID']};PWD={MSSQL_CONFIG['PWD']}"
    )
    cursor = conn.cursor()
    cursor.execute("SELECT TOP 1 * FROM AgentLog")
    columns = [column[0] for column in cursor.description]
    print(columns)
except Exception as e:
    print(e)
