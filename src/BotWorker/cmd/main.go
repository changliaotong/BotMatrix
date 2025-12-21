package main

import (
	"botworker/internal/config"
	"botworker/internal/db"
	"botworker/internal/redis"
	"botworker/internal/server"
	"botworker/plugins"
	"log"
)

func main() {
	// 加载配置
	cfg, _, err := config.LoadFromCLI()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 测试数据库连接
	log.Printf("数据库配置: Host=%s, Port=%d, User=%s, DBName=%s, SSLMode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.DBName, cfg.Database.SSLMode)

	database, err := db.NewDBConnection(&cfg.Database)
	if err != nil {
		log.Printf("警告: 无法连接到数据库: %v", err)
		log.Println("服务器将继续运行，但数据库功能将不可用")
	} else {
		log.Println("成功连接到数据库")

		// 初始化数据库表
		if err := db.InitDatabase(database); err != nil {
			log.Printf("警告: 初始化数据库表失败: %v", err)
		} else {
			log.Println("成功初始化数据库表")

			// 测试用户创建
			testUser := &db.User{
				UserID:   "test_user_123",
				Nickname: "测试用户",
				Avatar:   "http://example.com/avatar.jpg",
				Gender:   "male",
			}
			if err := db.CreateUser(database, testUser); err != nil {
				log.Printf("警告: 创建测试用户失败: %v", err)
			} else {
				log.Println("成功创建测试用户")

				// 测试获取用户
				retrievedUser, err := db.GetUserByUserID(database, "test_user_123")
				if err != nil {
					log.Printf("警告: 获取测试用户失败: %v", err)
				} else if retrievedUser != nil {
					log.Printf("成功获取测试用户: ID=%s, Nickname=%s", retrievedUser.UserID, retrievedUser.Nickname)
				}
			}
		}

		// 不关闭数据库连接，由插件管理器管理连接生命周期
	}

	// 测试Redis连接
	log.Printf("Redis配置: Host=%s, Port=%d, DB=%d",
		cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.DB)

	redisClient, err := redis.NewClient(&cfg.Redis)
	if err != nil {
		log.Printf("警告: 无法连接到Redis服务器: %v", err)
		log.Println("服务器将继续运行，但Redis功能将不可用")
	} else {
		log.Println("成功连接到Redis服务器")
		// 不关闭Redis连接，由插件管理器管理连接生命周期
	}

	// 创建组合服务器，同时支持WebSocket和HTTP
	combinedServer := server.NewCombinedServer(cfg)

	// 获取插件管理器
	pluginManager := combinedServer.GetPluginManager()

	// 加载示例插件
	echoPlugin := &plugins.EchoPlugin{}
	if err := pluginManager.LoadPlugin(echoPlugin); err != nil {
		log.Fatalf("加载插件失败: %v", err)
	}

	// 加载欢迎语插件
	welcomePlugin := &plugins.WelcomePlugin{}
	if err := pluginManager.LoadPlugin(welcomePlugin); err != nil {
		log.Fatalf("加载欢迎语插件失败: %v", err)
	}

	// 加载群管插件
	groupManagerPlugin := plugins.NewGroupManagerPlugin(database, redisClient)
	if err := pluginManager.LoadPlugin(groupManagerPlugin); err != nil {
		log.Fatalf("加载群管插件失败: %v", err)
	}

	// 加载天气插件
	weatherPlugin := plugins.NewWeatherPlugin(&cfg.Weather)
	if err := pluginManager.LoadPlugin(weatherPlugin); err != nil {
		log.Fatalf("加载天气插件失败: %v", err)
	}

	// 加载积分系统插件
	pointsPlugin := plugins.NewPointsPlugin()
	if err := pluginManager.LoadPlugin(pointsPlugin); err != nil {
		log.Fatalf("加载积分系统插件失败: %v", err)
	}

	// 加载签到系统插件（传递积分插件引用）
	signInPlugin := plugins.NewSignInPlugin(pointsPlugin)
	if err := pluginManager.LoadPlugin(signInPlugin); err != nil {
		log.Fatalf("加载签到系统插件失败: %v", err)
	}

	// 加载抽签插件
	lotteryPlugin := plugins.NewLotteryPlugin()
	if err := pluginManager.LoadPlugin(lotteryPlugin); err != nil {
		log.Fatalf("加载抽签插件失败: %v", err)
	}

	// 加载菜单插件
	menuPlugin := plugins.NewMenuPlugin()
	if err := pluginManager.LoadPlugin(menuPlugin); err != nil {
		log.Fatalf("加载菜单插件失败: %v", err)
	}

	// 加载翻译插件
	translatePlugin := plugins.NewTranslatePlugin()
	if err := pluginManager.LoadPlugin(translatePlugin); err != nil {
		log.Fatalf("加载翻译插件失败: %v", err)
	}

	// 加载点歌插件
	musicPlugin := plugins.NewMusicPlugin()
	if err := pluginManager.LoadPlugin(musicPlugin); err != nil {
		log.Fatalf("加载点歌插件失败: %v", err)
	}

	// 打印已加载的插件
	log.Println("已加载的插件:")
	for _, plugin := range pluginManager.GetPlugins() {
		log.Printf("- %s v%s: %s", plugin.Name(), plugin.Version(), plugin.Description())
	}

	// 启动服务器
	log.Println("启动OneBot协议机器人服务器...")
	if err := combinedServer.Run(); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
