-- ############################################################
-- Object Name: sp_TransferCoins
-- Object Type: SQL_STORED_PROCEDURE
-- ############################################################


 
CREATE procedure "dbo"."sp_TransferCoins"(@robot_qq bigint, @group_id bigint, @group_name varchar(50), @from_qq bigint, @to_qq bigint, @amount money, @transfer_info varchar(200))
as
begin
	--金币转帐系统 4步 减去/记账/加上/记账
	declare @amount2 money
	select @amount2 = -1 * @amount
	exec sp_AddCoins @robot_qq, @group_id, @group_name, @from_qq,'',@amount2,@transfer_info
	exec sp_AddCoins @robot_qq, @group_id, @group_name, @to_qq,'',@amount,@transfer_info	
end



GO

-- ############################################################
-- Object Name: sp_GetLists
-- Object Type: SQL_STORED_PROCEDURE
-- ############################################################


CREATE  PROCEDURE "dbo"."sp_GetLists"
@tblName varchar(200), 
@idName varchar(200), 
@OrderField varchar(200),
@PageNo int, 
@PageSize int, 
@sWhere varchar(4000)
AS
set nocount on 
declare @sql nvarchar(4000)
set @sql = N'select top ' + convert(varchar(12),@pagesize) + 
' * from ' + @tblName + ' where ' + @idName + 
' not in (select top ' + convert(varchar(12),(@PageNo-1)*@PageSize) + ' ' + @idName + ' from ' + @tblName + 
' where 1=1 ' + @sWhere + ' order by ' + @OrderField + ' ) ' + @sWhere + ' order by ' + @OrderField
print @sql
exec sp_executesql @sql



GO

-- ############################################################
-- Object Name: sp_UpdateClientQQ
-- Object Type: SQL_STORED_PROCEDURE
-- ############################################################

CREATE OR REPLACE PROCEDURE "dbo"."sp_UpdateClientQQ"
    @client_qq BIGINT,
    @target_qq BIGINT
AS
BEGIN
    -- 启动事务
    BEGIN TRANSACTION;

    BEGIN TRY
		-- 1. 获取 @client_qq 的 credit 值
		DECLARE @client_credit BIGINT;
		DECLARE @save_credit BIGINT;
		DECLARE @tokens BIGINT;
		SELECT @client_credit = isnull(credit, 0), @save_credit = isnull(savecredit, 0), @tokens = isnull(tokens, 0)
		FROM "User"
		WHERE Id = @client_qq;

		-- 2. 更新 @target_qq 的 credit 值
		UPDATE "User" SET credit = credit + @client_credit, SaveCredit = SaveCredit + @save_credit, tokens = tokens + @tokens
		WHERE Id = @target_qq;

		-- 3. 将 @client_qq 的 credit 值重置为 0 (or another appropriate value)
        UPDATE "User" SET TargetUserId = @target_qq, credit = 0, SaveCredit = 0, tokens = 0
        WHERE Id = @client_qq;

		UPDATE credit SET userid = @target_qq WHERE userid = @client_qq

		UPDATE tokens SET userid = @target_qq WHERE userid = @client_qq

        -- 更新 robot_send_message 表
        UPDATE SendMessage SET UserId = @target_qq WHERE UserId = @client_qq;

		-- 更新 robot_group 表
        UPDATE "group" SET GroupOwner = @target_qq WHERE GroupOwner = @client_qq;

		UPDATE "group" SET RobotOwner = @target_qq WHERE RobotOwner = @client_qq;

        -- 更新 robot_answer 表
        UPDATE answer SET userid = @target_qq WHERE userid = @client_qq;

        -- 提交事务
        COMMIT TRANSACTION;
    END TRY
    BEGIN CATCH
        -- 回滚事务
        ROLLBACK TRANSACTION;

        -- 返回错误信息
        THROW;
    END CATCH;
END;


GO

-- ############################################################
-- Object Name: sp_UpdateGroupInfo
-- Object Type: SQL_STORED_PROCEDURE
-- ############################################################

CREATE OR REPLACE PROCEDURE "dbo"."sp_UpdateGroupInfo"
    @group_id BIGINT,
    @target_group BIGINT
AS
BEGIN
    -- 启动事务
    BEGIN TRANSACTION;

    BEGIN TRY
		declare @group_owner bigint
		select @group_owner = GroupOwner from "group" where Id = ISNULL((select top 1 GroupId from vip where groupid = @target_group), @group_id)

        -- 更新 robot_group 表
		MERGE INTO "Group" AS rg
		USING (SELECT @group_id AS group_id, @target_group AS target_group, @group_owner AS group_owner) AS src
		ON rg.Id = src.group_id
		WHEN MATCHED THEN
			UPDATE SET rg.TargetGroup = src.target_group, rg.GroupOwner = src.group_owner
		WHEN NOT MATCHED THEN
			INSERT (Id, TargetGroup, GroupOwner) VALUES (src.group_id, src.target_group, src.group_owner);

        -- 更新 robot_send_message 表
        UPDATE SendMessage SET GroupId = @target_group WHERE GroupId = @group_id;

        -- 更新 robot_answer 表的 group_id 字段
        UPDATE answer SET groupid = @target_group WHERE groupid = @group_id;

        -- 更新 robot_answer 表的 robot_id 字段
        UPDATE answer SET robotid = @target_group, groupid = @target_group WHERE robotid = @group_id;

        -- 提交事务
        COMMIT TRANSACTION;
    END TRY
    BEGIN CATCH
        -- 回滚事务
        ROLLBACK TRANSACTION;

        -- 抛出错误信息
        THROW;
    END CATCH;
END;


GO

-- ############################################################
-- Object Name: sp_UpdateUserIds
-- Object Type: SQL_STORED_PROCEDURE
-- ############################################################

CREATE OR REPLACE PROCEDURE "dbo"."sp_UpdateUserIds"
AS
BEGIN
	DECLARE @sql NVARCHAR(MAX) = N'';

	SELECT @sql = @sql + 
		'EXEC sz84_robot.DBO.sp_UpdateClientQQ ' 
		+ CAST(sourceQQ AS VARCHAR) + ', ' 
		+ CAST(targetQQ AS VARCHAR) + ';' + CHAR(13) + CHAR(10)
	FROM (
		SELECT DISTINCT 
			a.UserId AS sourceQQ, 
			b.UserId AS targetQQ
		FROM SendMessage a
		INNER JOIN SendMessage b
			ON a.GroupId = b.GroupId
			AND a.UserId > 980000000000
			AND a.GroupId < 990000000000
			AND '"@:3889418604"' + a.Question = b.Question
			AND ABS(DATEDIFF(SECOND, a.InsertDate, b.InsertDate)) < 10
	) t;

	-- 打印调试（可选）
	--PRINT @sql;

	-- 执行动态 SQL
	EXEC (@sql);
END;


GO

-- ############################################################
-- Object Name: sp_UpdateWeixinQQ
-- Object Type: SQL_STORED_PROCEDURE
-- ############################################################


CREATE procedure "dbo"."sp_UpdateWeixinQQ"(@client_qq bigint, @client_qq2 bigint)
as
begin
	update wx_client set client_qq = @client_qq2 where client_qq = @client_qq
	delete from groupmember where userId = @client_qq2
	update groupmember set userId = @client_qq2 where userId = @client_qq
	update sendmessage set userId = @client_qq2 where userId = @client_qq
	update answer set userId = @client_qq2 where userId = @client_qq
end



GO

-- ############################################################
-- Object Name: sp_BuyVIP
-- Object Type: SQL_STORED_PROCEDURE
-- ############################################################

CREATE procedure "dbo"."sp_BuyVIP"(@group_id bigint,@client_qq bigint,@c_month int,@pay_method varchar(50),@pay_money money, @vip_trade varchar(50), @vip_info varchar(200), @insert_by int)
as
begin
	declare @group_name varchar(50)
	declare @is_cloud_answer int
	select @group_name = GroupName,@is_cloud_answer=IsCloudAnswer from "group" where id = @group_id
	declare @client_name varchar(50)
	
	select @client_name = name from "user" where id = @client_qq
	
	--判断是否已经开通vip
	declare @vip_id int
	declare @end_date datetime
	declare @is_year_vip int
	select @vip_id=id,@end_date=enddate,@is_year_vip=DATEDIFF(MONTH,GETDATE(),enddate) from vip where GroupId = @group_id
	select @is_year_vip = isnull(@is_year_vip,0) + @c_month
	if (@is_year_vip >= 11) 
		select @is_year_vip = 1
	else
		select @is_year_vip = 0 
	
	begin tran 
	
	--新开VIP
	if (@vip_id is null)
	begin
		insert into vip(groupid,groupname,firstpay,startdate,enddate,vipinfo,UserId,incomeday,isyearvip,insertby) 
		values(@group_id,@group_name,@pay_money,getdate(),dateadd(month,@c_month,getdate()),@vip_info,@client_qq,@pay_money/(@c_month*1.0),@is_year_vip,@insert_by)	
	end else
	begin
		--已经开过vip的群更新有效期
		update vip set enddate = dateadd(month, @c_month, enddate),vipinfo=@vip_info, UserId = @client_qq,
		incomeday = (incomeday * DATEDIFF(MONTH,GETDATE(),enddate) + @pay_money) /((DATEDIFF(MONTH,GETDATE(),enddate) + @c_month + 0.0000001)*1.0),
		isyearvip = @is_year_vip,insertby = @insert_by,isgoon = null
		where groupid = @group_id		
	end
	
	insert into income(groupid, goodscount, goodsname, paymethod, incomemoney, incometrade, incomeinfo,UserId, insertby) 
	values(@group_id, @c_month, '机器人', @pay_method, @pay_money, @vip_trade, @vip_info, @client_qq, @insert_by)
	
	--如果群信息不存在则添加 存在则更新机器人主人	
	declare @group_oid bigint
	select @group_oid = oid from "group" where id = @group_id
	declare @group_owner bigint
	select @group_owner = groupowner from "group" where id = @group_id
	if (@pay_method <> '余额')
		select @group_owner = @client_qq
	if (@group_oid is null)
		insert into "group"(Id, GroupOwner, RobotOwner, IsExitHint, IsChangeHint, IsSz84, IsWhite) values(@group_id, @client_qq, @client_qq, 1, 1, 0, 1)
	else	
		update "group" set IsValid=1, GroupOwner=@client_qq, RobotOwner=@client_qq, isSz84=0, IsExitHint = 1, IsChangeHint = 1, IsWhite=1 where Id = @group_id	
		
	delete from BlackList where groupid = 86433316 and BlackId = @client_qq	
	
	if @is_cloud_answer = 3  	--非年费版不能用话痨模式
		update "group" set iscloudanswer = 2 where iscloudanswer = 3 and id = @group_id and @is_year_vip = 0 	
	else                        --升级成年费的自动将话痨模式打开
		update "group" set iscloudanswer = 3 where iscloudanswer = 2 and id = @group_id and @is_year_vip = 1
				
	commit tran
end



GO

-- ############################################################
-- Object Name: getMinRobot
-- Object Type: SQL_SCALAR_FUNCTION
-- ############################################################


CREATE OR REPLACE FUNCTION "dbo"."getMinRobot"() 
RETURNS bigint
as

begin
	declare @MinRobot bigint
	select top 1 @MinRobot = robot_qq from sz84_robot.dbo.v_robot_member_vip 
	return @MinRobot
end




GO

-- ############################################################
-- Object Name: get_sell_price
-- Object Type: SQL_SCALAR_FUNCTION
-- ############################################################

CREATE OR REPLACE FUNCTION "dbo"."get_sell_price"(@sell_price money, @buy_date datetime) 
RETURNS bigint
as
begin
	declare @c_hour int
	select @c_hour = DATEDIFF(hour,@buy_date, getdate()) 
	--if (@c_hour > 8)
	return @sell_price * 12 / 20
	--return @sell_price * (20 - @c_hour) / 20
end

GO

-- ############################################################
-- Object Name: getMinRobotFree
-- Object Type: SQL_SCALAR_FUNCTION
-- ############################################################


CREATE OR REPLACE FUNCTION "dbo"."getMinRobotFree"() 
RETURNS bigint
as

begin
	declare @MinRobot bigint
	select top 1 @MinRobot = robot_qq from sz84_robot..v_robot_member_free
	return @MinRobot
end




GO

-- ############################################################
-- Object Name: StringAggregateMultipleColumns
-- Object Type: SQL_SCALAR_FUNCTION
-- ############################################################

CREATE OR REPLACE FUNCTION dbo.StringAggregateMultipleColumns
(
    @tableName NVARCHAR(255),
    @columnNames NVARCHAR(MAX),
    @groupColumnName NVARCHAR(255),
    @groupColumnValue NVARCHAR(255)
)
RETURNS NVARCHAR(MAX)
AS
BEGIN
    DECLARE @result NVARCHAR(MAX)
    SET @result = ''

    DECLARE @sql NVARCHAR(MAX)
    SET @sql = 'SELECT ' + @columnNames + ' FROM ' + @tableName + ' WHERE ' + @groupColumnName + ' = ''' + @groupColumnValue + ''''

    DECLARE @value NVARCHAR(MAX)
    DECLARE @delimiter NVARCHAR(5)
    SET @delimiter = ', '

    DECLARE @execSql NVARCHAR(MAX)
    SET @execSql = 'DECLARE cur CURSOR LOCAL FAST_FORWARD FOR ' + @sql

    EXEC sp_executesql @execSql

    OPEN cur
    FETCH NEXT FROM cur INTO @value
    WHILE @@FETCH_STATUS = 0
    BEGIN
        SET @result = @result + @value + @delimiter
        FETCH NEXT FROM cur INTO @value
    END

    CLOSE cur
    DEALLOCATE cur

    IF LEN(@result) > 0
        SET @result = STUFF(@result, LEN(@result) - LEN(@delimiter) + 1, LEN(@delimiter), '')

    RETURN @result
END

GO

-- ############################################################
-- Object Name: getMinRobotEnt
-- Object Type: SQL_SCALAR_FUNCTION
-- ############################################################

 
CREATE  FUNCTION "dbo"."getMinRobotEnt"() 
RETURNS bigint
as

begin
	declare @MinRobot bigint
	select top 1 @MinRobot = robot_qq from sz84_robot..robot_member	
	where valid in(1,5) and is_freeze = 0 and is_blocked = 0 and is_vip = 1 and c_hour >= 24
	and robot_name like '%企业版%'
	order by c_msg 
	return @MinRobot
end




GO

-- ############################################################
-- Object Name: get_bots
-- Object Type: SQL_INLINE_TABLE_VALUED_FUNCTION
-- ############################################################

CREATE OR REPLACE FUNCTION "dbo"."get_bots"
(	
	@client_qq bigint
)
RETURNS TABLE 
AS
RETURN 
(select a.groupid, ISNULL(a.groupname, '') as group_name,trialstartdate, trialenddate, c_audit, c_all,
startdate, enddate, ISNULL(is_vip, 0) as is_vip, isvalid, isnull(c_msg, '') as c_msg from 

(select a.id as GroupId, a.groupname, trialstartdate, trialenddate, 
(select count(id) from sz84_robot..answer where robotid = a.id and audit<0 and audit2<>-4) as c_audit,
 null as c_all, b.startdate, b.enddate, 1 as is_vip, a.IsValid , null as c_msg from sz84_robot.."group" a
inner join sz84_robot..vip b on a.id = b.groupid where groupowner = @client_qq

union

select id, groupname, trialstartdate, trialenddate, 
(select count(id) from sz84_robot..answer where robotid = sz84_robot.."group".id and audit<0 and audit2<>-4) as c_audit, null as c_all, 
null as startdate, null as end_date, 0 as is_vip, IsValid, null as c_msg from sz84_robot.."group" 
where id not in (select groupid from sz84_robot..vip) and groupowner = @client_qq) a)

GO

-- ############################################################
-- Object Name: getChengyuGameCount
-- Object Type: SQL_SCALAR_FUNCTION
-- ############################################################

 

CREATE OR REPLACE FUNCTION "dbo"."getChengyuGameCount"(@group_id bigint, @client_qq bigint) 
RETURNS varchar(10)
as
begin
	declare @max_game_id int
	declare @c_all int
	declare @c_client int
	select @max_game_id = max(id) from chengyu where groupid = @group_id and gameno = 1 
	select @c_all = isnull(COUNT(id),0) from chengyu where groupid = @group_id and id >= @max_game_id
	select @c_client = isnull(COUNT(id),0) from chengyu where groupid = @group_id and userid = @client_qq and id >= @max_game_id
	if (@group_id = 0)
		return (@c_all - 1)/2	
	return convert(varchar(8), @c_client) + '/' + convert(varchar(8), @c_all)
end




GO

-- ############################################################
-- Object Name: sp_GetPage
-- Object Type: SQL_STORED_PROCEDURE
-- ############################################################

CREATE OR REPLACE PROCEDURE "dbo"."sp_GetPage"
    @tblName varchar(500), 
    @idName varchar(500), 
    @OrderField varchar(500),
    @PageNo int, 
    @PageSize int, 
    @sWhere varchar(3000),
    @sSelect varchar(3000)
AS
BEGIN
    SET NOCOUNT ON;

    IF (@PageNo <= 0)
        SET @PageNo = 1;

    DECLARE @sql NVARCHAR(4000);

    SET @sql = N'SELECT TOP ' + CONVERT(VARCHAR(12), @PageSize) + ' ' + @sSelect + 
        ' FROM ' + @tblName + ' WHERE ' + @idName + 
        ' NOT IN (SELECT TOP ' + CONVERT(VARCHAR(12), (@PageNo - 1) * @PageSize) + ' ' + @idName + 
        ' FROM ' + @tblName + 
        ' WHERE 1=1 ' + @sWhere + ' ORDER BY ' + @OrderField + ' ) ' + 
        @sWhere + ' ORDER BY ' + @OrderField;

    PRINT @sql;
    EXEC sp_executesql @sql;
END


GO

-- ############################################################
-- Object Name: v_group_members
-- Object Type: VIEW
-- ############################################################

CREATE OR REPLACE VIEW dbo.v_group_members
AS
SELECT   a.Oid, a.Id, a.Guid, a.GroupOpenid, a.TargetGroup, a.GroupName, a.IsValid, a.IsProxy, a.GroupMemo, a.GroupOwner, 
                a.GroupOwnerName, a.GroupOwnerNickName, a.GroupType, a.RobotOwner, a.RobotOwnerName, a.WelcomeMessage, 
                a.InsertDate, a.GroupState, a.BotUin, a.BotName, a.LastDate, a.IsInGame, a.IsOpen, a.UseRight, a.TeachRight, 
                a.AdminRight, a.QuietTime, a.IsCloseManager, a.IsAcceptNewMember, a.CloseRegex, a.RegexRequestJoin, 
                a.RejectMessage, a.IsWelcomeHint, a.IsExitHint, a.IsKickHint, a.IsChangeHint, a.IsRightHint, a.IsCloudBlack, 
                a.IsCloudAnswer, a.IsRequirePrefix, a.IsSz84, a.IsWarn, a.IsBlackExit, a.IsBlackKick, a.IsBlackShare, a.IsChangeEnter, 
                a.IsMuteEnter, a.IsChangeMessage, a.IsSaveRecord, a.IsPause, a.RecallKeyword, a.WarnKeyword, a.MuteKeyword, 
                a.KickKeyword, a.BlackKeyword, a.MuteEnterCount, a.MuteKeywordCount, a.KickCount, a.BlackCount, a.ParentGroup, 
                a.CardNamePrefixBoy, a.CardNamePrefixGirl, a.CardNamePrefixManager, a.LastAnswerId, a.LastAnswer, 
                a.LastAnswerDate, a.LastChengyu, a.LastChengyuDate, a.TrialStartDate, a.TrialEndDate, a.LastExitHintDate, 
                a.BlockRes, a.BlockType, a.BlockMin, a.BlockFee, a.GroupGuid, a.IsBlock, a.IsWhite, a.CityName, a.IsMuteRefresh, 
                a.MuteRefreshCount, a.IsProp, a.IsPet, a.IsBlackRefresh, a.FansName, a.IsConfirmNew, a.WhiteKeyword, 
                a.CreditKeyword, a.IsCredit, a.IsPowerOn, a.IsHintClose, a.IsRecall, a.RecallTime, a.IsInvite, a.InviteStartDate, 
                a.InviteCredit, a.ReplyMode, a.IsReplyImage, a.IsReplyRecall, a.IsAI, a.SystemPrompt, a.IsOwnerPay, a.ContextCount, 
                a.IsMultAI, a.IsAutoSignin, a.IsUseKnowledgebase, a.IsSendHelpInfo, b.Id AS Expr1, b.GroupId, b.UserId, 
                b.Guid AS Expr2, b.UserName, b.DisplayName, b.NickName, b.RemarkName, b.IsCreator, b.IsAdmin, b.Age, b.Gender, 
                b.LastSpeak, b.InsertDate AS Expr3, b.ExitDate, b.Status, b.SignTimes, b.SignDate, b.SignLevel, b.SignTimesAll, 
                b.Credit, b.IsVip, b.VipStart, b.VipEnd, b.VipLevel, b.ConfirmCode, b.IsConfirm, b.ConfirmDate, b.FansLevel, b.FansValue, 
                b.IsFans, b.FansDate, b.LampDate, b.CountMessageToday, b.GroupCredit, b.SaveCredit, b.GoldCoins, b.PurpleCoins, 
                b.BlackCoins, b.GameCoins, b.InvitorUserId, b.InviteCount, b.InviteExitCount, b.SystemPrompt AS Expr4, b.AiTime, 
                b.LastQuestion, b.LastTime, b.LastMsgId, b.AppId
FROM      dbo."Group" AS a INNER JOIN
                dbo.GroupMember AS b ON a.Id = b.GroupId


GO

-- ############################################################
-- Object Name: getMinRobotEntFree
-- Object Type: SQL_SCALAR_FUNCTION
-- ############################################################


create FUNCTION "dbo"."getMinRobotEntFree"() 
RETURNS bigint
as

begin
	declare @MinRobot bigint
	select top 1 @MinRobot = robot_qq from sz84_robot..robot_member	
	where valid = 1  and is_freeze = 0 and is_blocked = 0 and is_vip = 0 and c_hour >= 24
    and robot_name like '%企业版%'
	order by group_count_free
	return @MinRobot
end




GO

-- ############################################################
-- Object Name: income_sum
-- Object Type: SQL_INLINE_TABLE_VALUED_FUNCTION
-- ############################################################

-- =============================================
-- Author:		<Author,,Name>
-- Create date: <Create Date,,>
-- Description:	<Description,,>
-- =============================================
CREATE OR REPLACE FUNCTION "dbo"."income_sum"
(	
)
RETURNS TABLE 
AS
RETURN 
select convert(varchar(50), year(incomedate-5/24.0) * 100 + month(incomedate-5/24.0)) as income_time, 
count(distinct groupid) as group_count, 
sum(incomemoney) as income_sum,
 sum(incomemoney)/(datediff(day,min(incomedate),max(incomedate)) +1) as income_avg, 
(sum(incomemoney)/(datediff(day,min(incomedate-5/24.0),max(incomedate-5/24.0)) +1))*30.0 as income_about_month, 
(sum(incomemoney)/(datediff(day,min(incomedate-5/24.0),max(incomedate-5/24.0)) +1))*365.0 as income_about_year 
from sz84_robot..income group by convert(varchar(50), year(incomedate-5/24.0) * 100 + month(incomedate-5/24.0)) 

union select convert(varchar(50), 111111) as income_time, 
count(distinct groupid) as group_count, 
SUM(incomemoney) as income_sum,
sum(incomemoney)/(datediff(day,min(incomedate),max(incomedate))) as income_avg,
(sum(incomemoney)/(datediff(day,min(incomedate),max(incomedate))))*30.0 as income_about_month, 
(sum(incomemoney)/(datediff(day,min(incomedate),max(incomedate))))*365.0 as income_about_year
from sz84_robot..income  

GO

-- ############################################################
-- Object Name: income_sum_year
-- Object Type: SQL_INLINE_TABLE_VALUED_FUNCTION
-- ############################################################

-- =============================================
-- Author:		<Author,,Name>
-- Create date: <Create Date,,>
-- Description:	<Description,,>
-- =============================================
CREATE OR REPLACE FUNCTION "dbo"."income_sum_year"
(	
)
RETURNS TABLE 
AS
RETURN 
select convert(varchar(50), year(incomedate-5/24.0)) as income_time,
count(distinct groupid) as group_count,
sum(incomemoney) as income_sum, 
sum(incomemoney)/(datediff(month, min(incomedate),max(incomedate)) +1) as income_avg, 
(sum(incomemoney)/(datediff(month,min(incomedate-5/24.0),max(incomedate-5/24.0)) +1))  as income_about_month, 
(sum(incomemoney)/(datediff(month,min(incomedate-5/24.0),max(incomedate-5/24.0)) +1)) * 12.0 as income_about_year 
from sz84_robot..income 
group by convert(varchar(50), year(incomedate-5/24.0)) 

union 

select convert(varchar(50), 1111) as income_time,  
count(distinct groupid) as group_count, 
SUM(incomemoney) as income_sum,
sum(incomemoney)/(datediff(month,min(incomedate),max(incomedate))) as income_avg, 
(sum(incomemoney)/(datediff(month,min(incomedate),max(incomedate)))) as income_about_month, 
(sum(incomemoney)/(datediff(day,min(incomedate),max(incomedate)))) * 12.0 as income_about_year 
from sz84_robot..income

GO

-- ############################################################
-- Object Name: sp_UpdatePublicQQ
-- Object Type: SQL_STORED_PROCEDURE
-- ############################################################


CREATE procedure "dbo"."sp_UpdatePublicQQ"(@client_qq bigint, @client_qq2 bigint)
as
begin
	update "user" set useropenid = (select useropenid from "user" where Id = @client_qq2) where Id = @client_qq
	delete from "user" where Id = @client_qq2
	delete from GroupMember where userid = @client_qq2
	update GroupMember set userid = @client_qq2 where userid = @client_qq
	update SendMessage set userid = @client_qq2 where userid = @client_qq
	update Answer set userid = @client_qq2 where userid = @client_qq
end



GO

-- ############################################################
-- Object Name: v_robot_member_vip
-- Object Type: VIEW
-- ############################################################

CREATE OR REPLACE VIEW dbo.v_robot_member_vip
AS
SELECT     TOP (8) robot_qq
FROM         (SELECT     TOP (8) robot_qq
                       FROM          dbo.robot_member
                       WHERE      (valid = 1) AND (is_freeze = 0) AND (is_vip = 1) AND (c_hour >= 24)
                       ORDER BY c_msg) AS a
ORDER BY NEWID()


GO

-- ############################################################
-- Object Name: v_robot_member_free
-- Object Type: VIEW
-- ############################################################

CREATE OR REPLACE VIEW dbo.v_robot_member_free
AS
SELECT     TOP (18) robot_qq
FROM         (SELECT     TOP (18) robot_qq
                       FROM          dbo.robot_member
                       WHERE      (valid = 1) AND (is_freeze = 0) AND (is_blocked = 0) AND (is_vip = 0) AND (c_hour >= 24)
                       ORDER BY c_msg) AS a
ORDER BY NEWID()


GO

-- ############################################################
-- Object Name: sp_AddCoins
-- Object Type: SQL_STORED_PROCEDURE
-- ############################################################


CREATE procedure "dbo"."sp_AddCoins"(@robot_qq bigint, @group_id bigint, @group_name varchar(50), @client_qq bigint, @client_name varchar(50), @coins_add money, @coins_info varchar(200))
as
begin
	if @group_name = ''
	begin
		select @group_name = groupname from "group" where id = @group_id
	end
	if @client_name = ''
	begin		
		select @client_name = name from "User" where Id = @client_qq
	end
	
	declare @coins_value int
	select @coins_value = coins from "User" where Id = @client_qq
	select @coins_value = @coins_value + @coins_add
	
	insert into Coins(BotUin, UserId, UserName, GroupId, GroupName, CoinsAdd, CoinsValue, CoinsInfo) 
	values(@robot_qq, @client_qq, @client_name, @group_id, @group_name, @coins_add, @coins_value, @coins_info)
	
	update "User" set Coins = Coins + @coins_add where Id = @client_qq
end



GO

-- ############################################################
-- Object Name: sp_BuyCredit
-- Object Type: SQL_STORED_PROCEDURE
-- ############################################################

CREATE procedure "dbo"."sp_BuyCredit"(@robot_qq bigint, @client_qq bigint,@credit_add int,@pay_method varchar(50),@pay_money money, @credit_trade varchar(50), @credit_info varchar(200), @insert_by int)
as
begin
	declare @client_name varchar(50)
	declare @ref_qq bigint
	
	select @client_name = name, @ref_qq = isnull(RefUserId, 0) from "user" where Id = @client_qq
	
	insert into income(groupid, goodscount, goodsname, paymethod, incomemoney,incometrade, incomeinfo,UserId, InsertBy) 
	values(0, @credit_add , '积分', @pay_method, @pay_money, @credit_trade, @credit_info, @client_qq, @insert_by)

    exec sp_AddCredit @robot_qq, @client_qq, @credit_add, '购买积分'  
end


GO

-- ############################################################
-- Object Name: sp_AddCredit
-- Object Type: SQL_STORED_PROCEDURE
-- ############################################################

CREATE procedure "dbo"."sp_AddCredit"(@robot_qq bigint, @client_qq bigint,@credit_add int, @credit_info varchar(200))
as
begin
	declare @client_name varchar(50)
	select @client_name = name from "User" where Id = @client_qq
		
	declare @oid bigint
	select @oid = oid from "User" where Id = @client_qq
	if (@oid is null)
		insert into "User"(Id) values(@client_qq);			
	
	declare @credit_value int
	select @credit_value = credit from "User" where Id = @client_qq
	select @credit_value = @credit_value + @credit_add

	insert into Credit(BotUin, UserId, UserName, GroupId, GroupName, CreditAdd, CreditValue, CreditInfo)
	values(@robot_qq, @client_qq, @client_name, 0, '', @credit_add, @credit_value, @credit_info)
	
	update "User" set Credit=isnull(credit,0) + @credit_add where Id = @client_qq	
end


GO

-- ############################################################
-- Object Name: sp_ChangeVip
-- Object Type: SQL_STORED_PROCEDURE
-- ############################################################

 
CREATE procedure "dbo"."sp_ChangeVip"(@group_id bigint, @new_group_id bigint, @new_client_qq bigint, @change_by int)
as
begin
	declare @client_qq bigint
	select @client_qq = groupowner from "group" where id = @group_id	
	update "group" set isvalid = 0 where id = @group_id	
	declare @new_group_oid bigint
	select @new_group_oid = id from "group" where id = @new_group_id
	if (@new_group_oid is null)
		insert into "group"(id, isvalid, groupowner,issz84) values(@new_group_id, 1, @new_client_qq,0)
	else
		update "group" set isvalid = 1, groupowner = @new_client_qq,issz84=0,isexithint=1,ischangehint=1 where id = @new_group_id	
	update vip set groupid = @new_group_id where groupid = @group_id	
	update income set groupid = @new_group_id where groupid = @group_id 	
	insert into change(groupid,newgroupid,userid, NewUserId, ChangeBy) values(@group_id, @new_group_id, @client_qq, @new_client_qq, @change_by)
end




GO

-- ############################################################
-- Object Name: get_fans_level
-- Object Type: SQL_SCALAR_FUNCTION
-- ############################################################

CREATE OR REPLACE FUNCTION "dbo"."get_fans_level"(@fans_value int) 
RETURNS int
as

begin
	declare @level_id int 
	select top 1 @level_id = id from sz84_robot..fanslevel
	where @fans_value <= fansvalue
	order by fansvalue
	return isnull(@level_id,0)
end

GO

-- ############################################################
-- Object Name: sp_BuyCoins
-- Object Type: SQL_STORED_PROCEDURE
-- ############################################################

CREATE procedure "dbo"."sp_BuyCoins"(@robot_qq bigint, @client_qq bigint, @coins_add int, @pay_method varchar(50),@pay_money money, @coins_trade varchar(50), @coins_info varchar(200), @insert_by int)
as
begin
	declare @client_name varchar(50)
	select @client_name = name from "User" where Id = @client_qq		

	insert into income(groupid, goodscount, goodsname, paymethod, incomemoney, incometrade, incomeinfo, UserId, InsertBy) 
	values(0, @coins_add, '金币', @pay_method, @pay_money, @coins_trade, @coins_info, @client_qq, @insert_by)
    exec sp_AddCoins @robot_qq,0,'',@client_qq,@client_name,@coins_add,'购买金币'
end


GO

-- ############################################################
-- Object Name: get_client_level
-- Object Type: SQL_SCALAR_FUNCTION
-- ############################################################

CREATE OR REPLACE FUNCTION "dbo"."get_client_level"(@income_money money) 
RETURNS int
as

begin
	declare @level_id int 
	select top 1 @level_id = id from sz84_robot..level
	where @income_money <= levelvalue
	order by levelvalue
	return @level_id
end

GO

-- ############################################################
-- Object Name: remove_biaodian
-- Object Type: SQL_SCALAR_FUNCTION
-- ############################################################


CREATE OR REPLACE FUNCTION "dbo"."remove_biaodian"(@key varchar(5000)) 
RETURNS varchar(5000)
as

begin
	declare @res varchar(5000)
	select @res = @key
	select @res = REPLACE(@res, ' ', '')
	select @res = REPLACE(@res, '　', '')
	select @res = REPLACE(@res, ',', '')
	select @res = REPLACE(@res, '，', '')
	select @res = REPLACE(@res, '.', '')
	select @res = REPLACE(@res, '。', '')
	select @res = REPLACE(@res, '?', '')
	select @res = REPLACE(@res, '？', '')
	select @res = REPLACE(@res, '!', '')
	select @res = REPLACE(@res, '！', '')
	select @res = REPLACE(@res, ';', '')
	select @res = REPLACE(@res, '；', '')
	select @res = REPLACE(@res, '~', '')
	select @res = REPLACE(@res, '～', '')
	select @res = REPLACE(@res, '"', '')
	select @res = REPLACE(@res, '【', '')
	select @res = REPLACE(@res, '"', '')
	select @res = REPLACE(@res, '】', '')
	select @res = REPLACE(@res, '{', '')
	select @res = REPLACE(@res, '『', '')
	select @res = REPLACE(@res, '}', '')
	select @res = REPLACE(@res, '』', '')
	select @res = REPLACE(@res, '\', '')
	select @res = REPLACE(@res, '、', '')
	select @res = REPLACE(@res, '…', '')
	select @res = REPLACE(@res, '/', '')
	select @res = REPLACE(@res, '／', '')
	select @res = REPLACE(@res, '(', '')
	select @res = REPLACE(@res, '（', '')
	select @res = REPLACE(@res, ')', '')
	select @res = REPLACE(@res, '）', '')
	select @res = REPLACE(@res, ':', '')
	select @res = REPLACE(@res, '：', '')
	select @res = REPLACE(@res, '<', '')
	select @res = REPLACE(@res, '《', '')
	select @res = REPLACE(@res, '>', '')
	select @res = REPLACE(@res, '》', '')
	select @res = REPLACE(@res, '"', '')
	select @res = REPLACE(@res, '“', '')
	select @res = REPLACE(@res, '”', '')
	select @res = REPLACE(@res, '''', '')
	select @res = REPLACE(@res, '‘', '')
	select @res = REPLACE(@res, '’', '')
	select @res = REPLACE(@res, '`', '')
	select @res = REPLACE(@res, '｀', '')
	select @res = REPLACE(@res, '	', '')
	select @res = REPLACE(@res, '·', '')	
	return @res
end

GO

