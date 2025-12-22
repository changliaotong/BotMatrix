package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"BotMatrix/common"
)

func main() {
	fmt.Println("🚀 开始测试GORM集成功能...")
	fmt.Println("=====================================")

	// 设置测试环境
	os.Setenv("DB_TYPE", "sqlite") // 使用SQLite进行测试
	os.Setenv("DB_PATH", "./test_gorm.db")
	os.Setenv("USE_GORM", "true") // 启用GORM

	// 创建管理器
	manager := &common.Manager{}

	fmt.Println("🔄 正在初始化数据库...")

	// 测试数据库初始化
	err := manager.InitDB()
	if err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
	}

	fmt.Println("✅ 数据库初始化成功")

	// 检查GORM是否启用
	if manager.IsGORMEnabled() {
		fmt.Println("✅ GORM已启用")
	} else {
		fmt.Println("❌ GORM未启用")
		return
	}

	fmt.Println("\n🔄 开始测试GORM基本操作...")
	fmt.Println("=====================================")

	// 测试用户操作
	testUserOperations(manager)

	// 测试路由规则操作
	testRoutingRuleOperations(manager)

	// 测试缓存操作
	testCacheOperations(manager)

	// 测试统计操作
	testStatsOperations(manager)

	fmt.Println("\n🔄 测试事务操作...")
	fmt.Println("=====================================")
	testTransactionOperations(manager)

	fmt.Println("\n🔄 测试批量操作...")
	fmt.Println("=====================================")
	testBatchOperations(manager)

	fmt.Println("\n🔄 测试查询性能...")
	fmt.Println("=====================================")
	testQueryPerformance(manager)

	fmt.Println("\n✅ 所有GORM测试完成！")

	// 清理测试数据
	cleanupTestData(manager)
}

func testUserOperations(manager *common.Manager) {
	fmt.Println("📋 测试用户操作...")

	// 创建测试用户
	testUser := &common.User{
		Username:       "testuser",
		PasswordHash:   "hashed_password_123",
		IsAdmin:        false,
		SessionVersion: 1,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 保存用户
	err := manager.SaveUserWithGORM(testUser)
	if err != nil {
		log.Printf("❌ 保存用户失败: %v", err)
		return
	}
	fmt.Println("✅ 保存用户成功")

	// 加载用户
	users, err := manager.LoadUsersWithGORM()
	if err != nil {
		log.Printf("❌ 加载用户失败: %v", err)
		return
	}

	found := false
	for _, user := range users {
		if user.Username == "testuser" {
			found = true
			break
		}
	}

	if found {
		fmt.Println("✅ 加载用户成功，找到测试用户")
	} else {
		fmt.Println("❌ 加载用户成功，但未找到测试用户")
	}

	// 更新用户
	testUser.IsAdmin = true
	err = manager.SaveUserWithGORM(testUser)
	if err != nil {
		log.Printf("❌ 更新用户失败: %v", err)
		return
	}
	fmt.Println("✅ 更新用户成功")

	// 删除用户
	err = manager.DeleteUserWithGORM("testuser")
	if err != nil {
		log.Printf("❌ 删除用户失败: %v", err)
		return
	}
	fmt.Println("✅ 删除用户成功")
}

func testRoutingRuleOperations(manager *common.Manager) {
	fmt.Println("📋 测试路由规则操作...")

	// 创建测试路由规则
	testRule := &common.RoutingRule{
		Pattern:        "test_pattern_*",
		TargetWorkerID: "worker_123",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 保存路由规则
	err := manager.SaveRoutingRuleWithGORM(testRule)
	if err != nil {
		log.Printf("❌ 保存路由规则失败: %v", err)
		return
	}
	fmt.Println("✅ 保存路由规则成功")

	// 加载路由规则
	rules, err := manager.LoadRoutingRulesWithGORM()
	if err != nil {
		log.Printf("❌ 加载路由规则失败: %v", err)
		return
	}

	found := false
	for _, rule := range rules {
		if rule.Pattern == "test_pattern_*" {
			found = true
			break
		}
	}

	if found {
		fmt.Println("✅ 加载路由规则成功，找到测试规则")
	} else {
		fmt.Println("❌ 加载路由规则成功，但未找到测试规则")
	}

	// 删除路由规则
	err = manager.DeleteRoutingRuleWithGORM("test_pattern_*")
	if err != nil {
		log.Printf("❌ 删除路由规则失败: %v", err)
		return
	}
	fmt.Println("✅ 删除路由规则成功")
}

func testCacheOperations(manager *common.Manager) {
	fmt.Println("📋 测试缓存操作...")

	// 测试群组缓存
	groupCache := &common.GroupCache{
		GroupID:   "group_123",
		GroupName: "测试群组",
		BotID:     "bot_456",
		LastSeen:  time.Now(),
	}

	err := manager.SaveGroupCacheWithGORM(groupCache)
	if err != nil {
		log.Printf("❌ 保存群组缓存失败: %v", err)
		return
	}
	fmt.Println("✅ 保存群组缓存成功")

	// 测试好友缓存
	friendCache := &common.FriendCache{
		UserID:   "user_789",
		Nickname: "测试好友",
		LastSeen: time.Now(),
	}

	err = manager.SaveFriendCacheWithGORM(friendCache)
	if err != nil {
		log.Printf("❌ 保存好友缓存失败: %v", err)
		return
	}
	fmt.Println("✅ 保存好友缓存成功")

	// 测试群成员缓存
	memberCache := &common.MemberCache{
		GroupID:  "group_123",
		UserID:   "user_789",
		Nickname: "测试成员",
		Card:     "管理员",
		LastSeen: time.Now(),
	}

	err = manager.SaveMemberCacheWithGORM(memberCache)
	if err != nil {
		log.Printf("❌ 保存群成员缓存失败: %v", err)
		return
	}
	fmt.Println("✅ 保存群成员缓存成功")

	// 加载缓存
	groups, err := manager.LoadGroupCachesWithGORM()
	if err != nil {
		log.Printf("❌ 加载群组缓存失败: %v", err)
		return
	}
	fmt.Printf("✅ 加载群组缓存成功，共%d个群组\n", len(groups))

	friends, err := manager.LoadFriendCachesWithGORM()
	if err != nil {
		log.Printf("❌ 加载好友缓存失败: %v", err)
		return
	}
	fmt.Printf("✅ 加载好友缓存成功，共%d个好友\n", len(friends))

	members, err := manager.LoadMemberCachesWithGORM()
	if err != nil {
		log.Printf("❌ 加载群成员缓存失败: %v", err)
		return
	}
	fmt.Printf("✅ 加载群成员缓存成功，共%d个成员\n", len(members))
}

func testStatsOperations(manager *common.Manager) {
	fmt.Println("📋 测试统计操作...")

	// 测试系统统计
	err := manager.SaveSystemStatWithGORM("test_key", "test_value")
	if err != nil {
		log.Printf("❌ 保存系统统计失败: %v", err)
		return
	}
	fmt.Println("✅ 保存系统统计成功")

	// 加载系统统计
	value, err := manager.LoadSystemStatWithGORM("test_key")
	if err != nil {
		log.Printf("❌ 加载系统统计失败: %v", err)
		return
	}

	if value == "test_value" {
		fmt.Println("✅ 加载系统统计成功，值正确")
	} else {
		fmt.Printf("❌ 加载系统统计成功，但值不匹配: %v\n", value)
	}

	// 测试群组统计
	err = manager.SaveGroupStatsWithGORM("group_123", 100)
	if err != nil {
		log.Printf("❌ 保存群组统计失败: %v", err)
		return
	}
	fmt.Println("✅ 保存群组统计成功")

	count, err := manager.LoadGroupStatsWithGORM("group_123")
	if err != nil {
		log.Printf("❌ 加载群组统计失败: %v", err)
		return
	}

	if count == 100 {
		fmt.Println("✅ 加载群组统计成功，值正确")
	} else {
		fmt.Printf("❌ 加载群组统计成功，但值不匹配: %d\n", count)
	}

	// 测试用户统计
	err = manager.SaveUserStatsWithGORM("user_789", 50)
	if err != nil {
		log.Printf("❌ 保存用户统计失败: %v", err)
		return
	}
	fmt.Println("✅ 保存用户统计成功")

	// 测试每日统计
	today := time.Now().Format("2006-01-02")
	err = manager.SaveGroupStatsTodayWithGORM("group_123", today, 25)
	if err != nil {
		log.Printf("❌ 保存群组每日统计失败: %v", err)
		return
	}
	fmt.Println("✅ 保存群组每日统计成功")

	count, err = manager.LoadGroupStatsTodayWithGORM("group_123", today)
	if err != nil {
		log.Printf("❌ 加载群组每日统计失败: %v", err)
		return
	}

	if count == 25 {
		fmt.Println("✅ 加载群组每日统计成功，值正确")
	} else {
		fmt.Printf("❌ 加载群组每日统计成功，但值不匹配: %d\n", count)
	}
}

func testTransactionOperations(manager *common.Manager) {
	fmt.Println("📋 测试事务操作...")

	// 测试事务
	err := manager.TransactionWithGORM(func(tx *common.Manager) error {
		// 在事务中执行多个操作
		user := &common.User{
			Username:       "tx_user",
			PasswordHash:   "tx_password",
			IsAdmin:        false,
			SessionVersion: 1,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		// 在事务中使用 tx 执行操作
		return tx.SaveUserWithGORM(user)
	})

	if err != nil {
		log.Printf("❌ 事务操作失败: %v", err)
		return
	}
	fmt.Println("✅ 事务操作成功")

	// 清理测试用户
	manager.DeleteUserWithGORM("tx_user")
}

func testBatchOperations(manager *common.Manager) {
	fmt.Println("📋 测试批量操作...")

	// 批量创建用户
	start := time.Now()
	for i := 0; i < 10; i++ {
		user := &common.User{
			Username:       fmt.Sprintf("batch_user_%d", i),
			PasswordHash:   fmt.Sprintf("password_%d", i),
			IsAdmin:        false,
			SessionVersion: 1,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		err := manager.SaveUserWithGORM(user)
		if err != nil {
			log.Printf("❌ 批量创建用户失败: %v", err)
			return
		}
	}
	elapsed := time.Since(start)
	fmt.Printf("✅ 批量创建10个用户成功，耗时: %v\n", elapsed)

	// 批量加载用户
	start = time.Now()
	users, err := manager.LoadUsersWithGORM()
	if err != nil {
		log.Printf("❌ 批量加载用户失败: %v", err)
		return
	}
	elapsed = time.Since(start)
	fmt.Printf("✅ 批量加载%d个用户成功，耗时: %v\n", len(users), elapsed)

	// 清理批量用户
	for i := 0; i < 10; i++ {
		manager.DeleteUserWithGORM(fmt.Sprintf("batch_user_%d", i))
	}
}

func testQueryPerformance(manager *common.Manager) {
	fmt.Println("📋 测试查询性能...")

	// 创建一些测试数据
	for i := 0; i < 100; i++ {
		cache := &common.GroupCache{
			GroupID:   fmt.Sprintf("perf_group_%d", i),
			GroupName: fmt.Sprintf("性能测试群组%d", i),
			BotID:     fmt.Sprintf("bot_%d", i%10),
			LastSeen:  time.Now(),
		}

		err := manager.SaveGroupCacheWithGORM(cache)
		if err != nil {
			log.Printf("❌ 创建性能测试数据失败: %v", err)
			return
		}
	}

	// 测试查询性能
	start := time.Now()
	caches, err := manager.LoadGroupCachesWithGORM()
	if err != nil {
		log.Printf("❌ 性能测试查询失败: %v", err)
		return
	}
	elapsed := time.Since(start)

	fmt.Printf("✅ 查询%d个群组缓存成功，耗时: %v\n", len(caches), elapsed)

	// 清理性能测试数据
	for i := 0; i < 100; i++ {
		manager.DeleteGroupCacheWithGORM(fmt.Sprintf("perf_group_%d", i))
	}
}

func cleanupTestData(manager *common.Manager) {
	fmt.Println("\n🧹 清理测试数据...")

	// 删除所有测试数据
	manager.DeleteUserWithGORM("testuser")
	manager.DeleteRoutingRuleWithGORM("test_pattern_*")
	manager.DeleteGroupCacheWithGORM("group_123")
	manager.DeleteFriendCacheWithGORM("user_789")
	manager.DeleteMemberCacheWithGORM("group_123", "user_789")
	manager.DeleteSystemStatWithGORM("test_key")
	manager.DeleteGroupStatsWithGORM("group_123")
	manager.DeleteUserStatsWithGORM("user_789")

	fmt.Println("✅ 测试数据清理完成")
}
