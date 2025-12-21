USE [sz84_robot]
GO

IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[ChatLog]') AND type in (N'U'))
BEGIN
CREATE TABLE [dbo].[ChatLog](
	[id] [int] IDENTITY(1,1) NOT NULL,
	[msg_type] [nvarchar](50) NULL,
	[user_id] [nvarchar](100) NULL,
	[user_name] [nvarchar](255) NULL,
	[group_id] [nvarchar](100) NULL,
	[group_name] [nvarchar](255) NULL,
	[content] [nvarchar](max) NULL,
	[create_time] [datetime] DEFAULT (getdate()) NULL,
 CONSTRAINT [PK_ChatLog] PRIMARY KEY CLUSTERED 
(
	[id] ASC
)WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, IGNORE_DUP_KEY = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON) ON [PRIMARY]
) ON [PRIMARY] TEXTIMAGE_ON [PRIMARY]
END
GO
