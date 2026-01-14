import pymssql

def create_table():
    try:
        conn = pymssql.connect(
            server='192.168.0.114',
            user='derlin',
            password='fkueiqiq461686',
            database='sz84_robot'
        )
        cursor = conn.cursor()
        
        sql = """
        IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[ToolAuditLogs]') AND type in (N'U'))
        BEGIN
            CREATE TABLE [dbo].[ToolAuditLogs](
                [Id] [bigint] IDENTITY(1,1) NOT NULL,
                [Guid] [uniqueidentifier] NOT NULL DEFAULT (newid()),
                [TaskId] [nvarchar](100) NULL,
                [StaffId] [nvarchar](100) NULL,
                [ToolName] [nvarchar](100) NULL,
                [InputArgs] [nvarchar](max) NULL,
                [OutputResult] [nvarchar](max) NULL,
                [RiskLevel] [int] NOT NULL DEFAULT ((0)),
                [Status] [nvarchar](50) NULL DEFAULT ('InProgress'),
                [ApprovedBy] [nvarchar](100) NULL,
                [RejectionReason] [nvarchar](max) NULL,
                [ApprovedAt] [datetime] NULL,
                [CreateTime] [datetime] NOT NULL DEFAULT (getdate()),
                CONSTRAINT [PK_ToolAuditLogs] PRIMARY KEY CLUSTERED ([Id] ASC)
            )
            PRINT 'Table ToolAuditLogs created.'
        END
        ELSE
        BEGIN
            PRINT 'Table ToolAuditLogs already exists.'
        END
        """
        
        cursor.execute(sql)
        conn.commit()
        print("SQL execution finished.")
    except Exception as e:
        print(f"Error: {e}")
    finally:
        if 'conn' in locals():
            conn.close()

if __name__ == "__main__":
    create_table()
