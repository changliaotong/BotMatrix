package bot

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"BotMatrix/common/config"
	"BotMatrix/common/database"
	"BotMatrix/common/models"
	"BotMatrix/common/types"
	"BotMatrix/common/utils"

	dclient "github.com/docker/docker/client"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/shirou/gopsutil/v3/process"
	"gorm.io/gorm"
)

// Manager holds the state of the BotMatrix system
type Manager struct {
	Config      *config.AppConfig
	Bots        map[string]*types.BotClient
	Subscribers map[*websocket.Conn]*types.Subscriber
	Workers     []*types.WorkerClient
	WorkerIndex int
	Mutex       sync.RWMutex
	Upgrader    websocket.Upgrader
	LogBuffer   []types.LogEntry
	LogMutex    sync.RWMutex

	PendingRequests   map[string]chan types.InternalMessage
	PendingTimestamps map[string]time.Time
	PendingMutex      sync.Mutex

	WorkerRequestTimes map[string]time.Time
	WorkerRequestMutex sync.Mutex

	Rdb *redis.Client

	DockerClient *dclient.Client

	RoutingRules map[string]string

	StatsMutex      sync.RWMutex
	StartTime       time.Time
	TotalMessages   int64
	SentMessages    int64
	UserStats       map[string]int64
	GroupStats      map[string]int64
	BotStats        map[string]int64
	BotStatsSent    map[string]int64
	UserStatsToday  map[string]int64
	GroupStatsToday map[string]int64
	BotStatsToday   map[string]int64
	LastResetDate   string

	BotDetailedStats map[string]*types.BotStatDetail

	HistoryMutex sync.RWMutex
	CPUTrend     []float64
	MemTrend     []uint64
	MsgTrend     []int64
	SentTrend    []int64
	RecvTrend    []int64
	NetSentTrend []uint64
	NetRecvTrend []uint64
	TrendLabels  []string
	TopProcesses []types.ProcInfo
	ProcMap      map[int32]*process.Process

	LastTrendTotal int64
	LastTrendSent  int64

	ConnectionStats types.ConnectionStats

	Users      map[string]*types.User
	UsersMutex sync.RWMutex
	DB         *sql.DB

	GORMDB      *gorm.DB
	GORMManager *database.GORMManager

	MessageCache []types.InternalMessage
	CacheMutex   sync.RWMutex
	GroupCache   map[string]types.GroupInfo
	MemberCache  map[string]types.MemberInfo
	FriendCache  map[string]types.FriendInfo
	BotCache     map[string]types.BotClient

	RulesMutex       sync.RWMutex
	LocalIdempotency sync.Map
	ConfigCache      map[string]string
	ConfigCacheMu    sync.RWMutex
	SessionCache     sync.Map
}

var GlobalManager = NewManager()

func (m *Manager) PrepareQuery(query string) string {
	if m.Config != nil && m.Config.PGHost != "" {
		n := 1
		for strings.Contains(query, "?") {
			query = strings.Replace(query, "?", fmt.Sprintf("$%d", n), 1)
			n++
		}
	}
	return query
}

// SaveGroupCache saves group cache to database
func (m *Manager) SaveGroupCache(cache *models.GroupCacheGORM) error {
	return database.SaveGroupCache(m.DB, m.PrepareQuery, cache)
}

// SaveFriendCache saves friend cache to database
func (m *Manager) SaveFriendCache(cache *models.FriendCacheGORM) error {
	return database.SaveFriendCache(m.DB, m.PrepareQuery, cache)
}

// SaveMemberCache saves member cache to database
func (m *Manager) SaveMemberCache(cache *models.MemberCacheGORM) error {
	return database.SaveMemberCache(m.DB, m.PrepareQuery, cache)
}

// LoadGroupCachesFromDB loads all group caches from database
func (m *Manager) LoadGroupCachesFromDB() ([]*models.GroupCacheGORM, error) {
	return database.LoadGroupCachesFromDB(m.DB, m.PrepareQuery)
}

// DeleteGroupCache deletes group cache from database
func (m *Manager) DeleteGroupCache(groupID string) error {
	return database.DeleteGroupCache(m.DB, m.PrepareQuery, groupID)
}

// DeleteFriendCache deletes friend cache from database
func (m *Manager) DeleteFriendCache(userID string) error {
	return database.DeleteFriendCache(m.DB, m.PrepareQuery, userID)
}

// DeleteMemberCache deletes member cache from database
func (m *Manager) DeleteMemberCache(groupID, userID string) error {
	return database.DeleteMemberCache(m.DB, m.PrepareQuery, groupID, userID)
}

// LoadFriendCachesFromDB loads all friend caches from database
func (m *Manager) LoadFriendCachesFromDB() ([]*models.FriendCacheGORM, error) {
	return database.LoadFriendCachesFromDB(m.DB, m.PrepareQuery)
}

func NewManager() *Manager {
	return &Manager{
		Bots:              make(map[string]*types.BotClient),
		Subscribers:       make(map[*websocket.Conn]*types.Subscriber),
		Workers:           make([]*types.WorkerClient, 0),
		PendingRequests:   make(map[string]chan types.InternalMessage),
		PendingTimestamps: make(map[string]time.Time),
		RoutingRules:      make(map[string]string),
		UserStats:         make(map[string]int64),
		GroupStats:        make(map[string]int64),
		BotStats:          make(map[string]int64),
		BotStatsSent:      make(map[string]int64),
		UserStatsToday:    make(map[string]int64),
		GroupStatsToday:   make(map[string]int64),
		BotStatsToday:     make(map[string]int64),
		LastResetDate:     time.Now().Format("2006-01-02"),
		StartTime:         time.Now(),
		ConnectionStats: types.ConnectionStats{
			BotConnectionDurations:    make(map[string]time.Duration),
			WorkerConnectionDurations: make(map[string]time.Duration),
			BotDisconnectReasons:      make(map[string]int64),
			WorkerDisconnectReasons:   make(map[string]int64),
			LastBotActivity:           make(map[string]time.Time),
			LastWorkerActivity:        make(map[string]time.Time),
		},
		GroupCache:         make(map[string]types.GroupInfo),
		MemberCache:        make(map[string]types.MemberInfo),
		FriendCache:        make(map[string]types.FriendInfo),
		BotCache:           make(map[string]types.BotClient),
		Users:              make(map[string]*types.User),
		ProcMap:            make(map[int32]*process.Process),
		WorkerRequestTimes: make(map[string]time.Time),
		LogBuffer:          make([]types.LogEntry, 0),
		ConfigCache:        make(map[string]string),
	}
}

// ValidateToken validates a JWT token
func (m *Manager) ValidateToken(tokenString string) (*types.UserClaims, error) {
	return utils.ValidateToken(tokenString, m.Config.JWTSecret)
}

// GenerateToken generates a JWT token for a user
func (m *Manager) GenerateToken(user *types.User) (string, error) {
	return utils.GenerateToken(user, m.Config.JWTSecret)
}

// LoadUsersFromDB 从数据库加载所有用户到内存
func (m *Manager) LoadUsersFromDB() error {
	if m.GORMDB == nil {
		return fmt.Errorf("GORMDB is not initialized")
	}

	var dbUsers []models.UserGORM
	if err := m.GORMDB.Find(&dbUsers).Error; err != nil {
		return fmt.Errorf("failed to load users from DB: %v", err)
	}

	m.UsersMutex.Lock()
	defer m.UsersMutex.Unlock()

	for _, u := range dbUsers {
		version := u.SessionVersion
		if version == 0 {
			version = 1
			// 异步更新数据库，不阻塞加载
			go func(id uint) {
				m.GORMDB.Model(&models.UserGORM{}).Where("id = ?", id).Update("session_version", 1)
			}(u.ID)
		}

		m.Users[u.Username] = &types.User{
			ID:             int64(u.ID),
			Username:       u.Username,
			PasswordHash:   u.PasswordHash,
			IsAdmin:        u.IsAdmin,
			Active:         u.Active,
			SessionVersion: version,
			CreatedAt:      u.CreatedAt,
			UpdatedAt:      u.UpdatedAt,
		}
	}

	log.Printf("Loaded %d users from database", len(dbUsers))
	return nil
}

// LoadBotsFromDB 从数据库加载所有 Online 平台机器人到内存
func (m *Manager) LoadBotsFromDB() error {
	if m.GORMDB == nil {
		return fmt.Errorf("GORMDB is not initialized")
	}

	var dbBots []models.BotEntityGORM
	// 我们目前只自动加载 Online 平台的机器人，其他平台的由其自身的 client 连接
	if err := m.GORMDB.Where("platform = ?", "Online").Find(&dbBots).Error; err != nil {
		return fmt.Errorf("failed to load bots from DB: %v", err)
	}

	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	if m.Bots == nil {
		m.Bots = make(map[string]*types.BotClient)
	}

	for _, b := range dbBots {
		// 如果内存中已经存在（可能已经连接），则跳过
		if _, exists := m.Bots[b.SelfID]; exists {
			continue
		}

		m.Bots[b.SelfID] = &types.BotClient{
			SelfID:    b.SelfID,
			Nickname:  b.Nickname,
			Platform:  b.Platform,
			Protocol:  "v11", // 默认为 v11
			Connected: time.Now(),
		}

		// 为 Online 机器人初始化模拟联系人
		m.CacheMutex.Lock()
		if m.GroupCache == nil {
			m.GroupCache = make(map[string]types.GroupInfo)
		}
		if m.FriendCache == nil {
			m.FriendCache = make(map[string]types.FriendInfo)
		}

		// 添加一个模拟群组
		mockGroupID := "10001"
		m.GroupCache[mockGroupID] = types.GroupInfo{
			BotID:     b.SelfID,
			GroupID:   mockGroupID,
			GroupName: "模拟群聊 (Online)",
		}

		// 添加一个模拟好友
		mockFriendID := "admin"
		m.FriendCache[mockFriendID] = types.FriendInfo{
			BotID:    b.SelfID,
			UserID:   mockFriendID,
			Nickname: "管理员 (Mock)",
		}
		m.CacheMutex.Unlock()
	}

	log.Printf("Loaded %d online bots from database", len(dbBots))
	return nil
}

// TrackBotDisconnection tracks bot disconnection events
func (m *Manager) TrackBotDisconnection(botID string, reason string, duration time.Duration) {
	m.ConnectionStats.Mutex.Lock()
	defer m.ConnectionStats.Mutex.Unlock()

	m.ConnectionStats.BotDisconnectReasons[reason]++
	if duration > 0 {
		m.ConnectionStats.BotConnectionDurations[botID] = duration
	}
}

// TrackWorkerDisconnection tracks worker disconnection events
func (m *Manager) TrackWorkerDisconnection(workerID string, reason string, duration time.Duration) {
	m.ConnectionStats.Mutex.Lock()
	defer m.ConnectionStats.Mutex.Unlock()

	m.ConnectionStats.WorkerDisconnectReasons[reason]++
	if duration > 0 {
		m.ConnectionStats.WorkerConnectionDurations[workerID] = duration
	}
}

// RefreshConfigCache 刷新本地配置缓存 (从 Redis 同步)
func (m *Manager) RefreshConfigCache() {
	if m.Rdb == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	newCache := make(map[string]string)

	// 1. 加载频率限制配置
	if data, err := m.Rdb.HGetAll(ctx, config.REDIS_KEY_CONFIG_RATELIMIT).Result(); err == nil {
		for k, v := range data {
			newCache["ratelimit:"+k] = v
		}
	}

	// 2. 加载 TTL 配置
	if data, err := m.Rdb.HGetAll(ctx, config.REDIS_KEY_CONFIG_TTL).Result(); err == nil {
		for k, v := range data {
			newCache["ttl:"+k] = v
		}
	}

	m.ConfigCacheMu.Lock()
	m.ConfigCache = newCache
	m.ConfigCacheMu.Unlock()
}

// StartConfigCacheRefresh 启动定期刷新任务
func (m *Manager) StartConfigCacheRefresh() {
	// 立即刷新一次
	m.RefreshConfigCache()

	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			m.RefreshConfigCache()
		}
	}()
}

// SaveBotToDB persists bot info to database
func (m *Manager) SaveBotToDB(selfID, nickname, platform, protocol string) error {
	if m.GORMDB == nil {
		return fmt.Errorf("GORM database not initialized")
	}
	var bot models.BotEntityGORM
	result := m.GORMDB.Where("self_id = ?", selfID).First(&bot)
	if result.Error != nil {
		bot = models.BotEntityGORM{
			SelfID:    selfID,
			Nickname:  nickname,
			Platform:  platform,
			Status:    "online",
			Connected: true,
			LastSeen:  time.Now(),
		}
		return m.GORMDB.Create(&bot).Error
	}

	bot.Nickname = nickname
	bot.Platform = platform
	bot.Status = "online"
	bot.Connected = true
	bot.LastSeen = time.Now()
	return m.GORMDB.Save(&bot).Error
}

// SaveGroupToDB persists group info to database
func (m *Manager) SaveGroupToDB(groupID, groupName, botID string) error {
	if m.GORMDB == nil {
		return fmt.Errorf("GORM database not initialized")
	}
	cache := &models.GroupCacheGORM{
		GroupID:   groupID,
		GroupName: groupName,
		BotID:     botID,
		LastSeen:  time.Now(),
	}
	return m.SaveGroupCache(cache)
}

// SaveMemberToDB persists member info to database
func (m *Manager) SaveMemberToDB(groupID, userID, nickname, card, role string) error {
	if m.GORMDB == nil {
		return fmt.Errorf("GORM database not initialized")
	}
	cache := &models.MemberCacheGORM{
		GroupID:  groupID,
		UserID:   userID,
		Nickname: nickname,
		Card:     card,
		Role:     role,
		LastSeen: time.Now(),
	}
	return m.SaveMemberCache(cache)
}

// SaveFriendToDB persists friend info to database
func (m *Manager) SaveFriendToDB(userID, nickname, botID string) error {
	if m.GORMDB == nil {
		return fmt.Errorf("GORM database not initialized")
	}
	cache := &models.FriendCacheGORM{
		UserID:   userID,
		Nickname: nickname,
		BotID:    botID,
		LastSeen: time.Now(),
	}
	return m.SaveFriendCache(cache)
}

// SaveMessageToDB persists message log to database
func (m *Manager) SaveMessageToDB(msgID, botID, userID, groupID, msgType, content, rawData string) error {
	if m.GORMDB == nil {
		return fmt.Errorf("GORM database not initialized")
	}
	log := &models.MessageLogGORM{
		BotID:     botID,
		UserID:    userID,
		GroupID:   groupID,
		Content:   content,
		RawData:   rawData,
		Direction: "incoming",
		CreatedAt: time.Now(),
	}
	return m.GORMDB.Create(log).Error
}

// InitDockerClient initializes the Docker client
func (m *Manager) InitDockerClient() error {
	cli, err := utils.InitDockerClient()
	if err != nil {
		return err
	}
	m.DockerClient = cli
	return nil
}

// UpdateBotSentStats updates bot sent message stats
func (m *Manager) UpdateBotSentStats(botID string) {
	m.StatsMutex.Lock()
	defer m.StatsMutex.Unlock()

	m.SentMessages++
	m.BotStatsSent[botID]++
	m.BotStatsToday[botID]++
}

// SaveUserToDB persists user info to database
func (m *Manager) SaveUserToDB(u any) error {
	var userGORM *models.UserGORM

	switch user := u.(type) {
	case *models.UserGORM:
		userGORM = user
	case *types.User:
		userGORM = &models.UserGORM{
			ID:             uint(user.ID),
			Username:       user.Username,
			PasswordHash:   user.PasswordHash,
			IsAdmin:        user.IsAdmin,
			Active:         user.Active,
			SessionVersion: user.SessionVersion,
			CreatedAt:      user.CreatedAt,
			UpdatedAt:      user.UpdatedAt,
		}
	default:
		return fmt.Errorf("unsupported user type: %T", u)
	}

	return m.GORMDB.Save(userGORM).Error
}

// SaveConfig persists configuration to disk
func (m *Manager) SaveConfig() error {
	return config.SaveConfig(m.Config)
}

// SaveRoutingRuleToDB persists routing rule to database
func (m *Manager) SaveRoutingRuleToDB(pattern, targetWorkerID string) error {
	rule := &models.RoutingRuleGORM{
		Pattern:        pattern,
		TargetWorkerID: targetWorkerID,
	}
	// Use Upsert logic
	var existing models.RoutingRuleGORM
	result := m.GORMDB.Where("pattern = ?", pattern).First(&existing)
	if result.Error == nil {
		existing.TargetWorkerID = targetWorkerID
		return m.GORMDB.Save(&existing).Error
	}
	return m.GORMDB.Create(rule).Error
}

// DeleteRoutingRuleFromDB deletes routing rule from database
func (m *Manager) DeleteRoutingRuleFromDB(pattern string) error {
	return m.GORMDB.Where("pattern = ?", pattern).Delete(&models.RoutingRuleGORM{}).Error
}

// LoadRoutingRulesFromDB loads all routing rules from database
func (m *Manager) LoadRoutingRulesFromDB() error {
	var rules []models.RoutingRuleGORM
	if err := m.GORMDB.Find(&rules).Error; err != nil {
		return err
	}

	m.RulesMutex.Lock()
	defer m.RulesMutex.Unlock()

	m.RoutingRules = make(map[string]string)
	for _, rule := range rules {
		m.RoutingRules[rule.Pattern] = rule.TargetWorkerID
	}
	return nil
}

// LoadCachesFromDB loads all group/member/friend caches from database
func (m *Manager) LoadCachesFromDB() error {
	m.CacheMutex.Lock()
	defer m.CacheMutex.Unlock()

	// Load groups
	var groups []models.GroupCacheGORM
	if err := m.GORMDB.Find(&groups).Error; err == nil {
		for _, g := range groups {
			m.GroupCache[g.GroupID] = types.GroupInfo{
				GroupID:   g.GroupID,
				GroupName: g.GroupName,
				BotID:     g.BotID,
				LastSeen:  g.LastSeen,
			}
		}
	}

	// Load members
	var members []models.MemberCacheGORM
	if err := m.GORMDB.Find(&members).Error; err == nil {
		for _, mem := range members {
			m.MemberCache[fmt.Sprintf("%s:%s", mem.GroupID, mem.UserID)] = types.MemberInfo{
				GroupID:  mem.GroupID,
				UserID:   mem.UserID,
				Nickname: mem.Nickname,
				LastSeen: mem.LastSeen,
			}
		}
	}

	// Load friends
	var friends []models.FriendCacheGORM
	if err := m.GORMDB.Find(&friends).Error; err == nil {
		for _, f := range friends {
			m.FriendCache[f.UserID] = types.FriendInfo{
				UserID:   f.UserID,
				Nickname: f.Nickname,
				BotID:    f.BotID,
				LastSeen: f.LastSeen,
			}
		}
	}

	return nil
}

// InitDB initializes the database connection
func (m *Manager) InitDB() error {
	db, err := database.InitDB(m.Config)
	if err != nil {
		return err
	}
	m.DB = db

	if m.GORMManager == nil {
		m.GORMManager = &database.GORMManager{}
	}
	err = m.GORMManager.InitGORM(m.Config)
	if err == nil {
		m.GORMDB = m.GORMManager.DB
		// 确保管理员账号存在
		if adminErr := m.EnsureAdminUser(); adminErr != nil {
			log.Printf("Warning: Failed to ensure admin user: %v\n", adminErr)
		}
	}
	return err
}

// GetOrLoadUser retrieves a user from memory or loads it from the database if not present
func (m *Manager) GetOrLoadUser(username string) (*types.User, bool) {
	m.UsersMutex.RLock()
	user, exists := m.Users[username]
	m.UsersMutex.RUnlock()

	if exists {
		return user, true
	}

	if m.GORMDB == nil {
		return nil, false
	}

	var u models.UserGORM
	result := m.GORMDB.Where("username = ?", username).First(&u)
	if result.Error != nil {
		return nil, false
	}

	// Convert to types.User
	version := u.SessionVersion
	if version == 0 {
		version = 1
		// 异步更新数据库，不阻塞加载
		go func(id uint) {
			m.GORMDB.Model(&models.UserGORM{}).Where("id = ?", id).Update("session_version", 1)
		}(u.ID)
	}

	user = &types.User{
		ID:             int64(u.ID),
		Username:       u.Username,
		PasswordHash:   u.PasswordHash,
		IsAdmin:        u.IsAdmin,
		Active:         u.Active,
		SessionVersion: version,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
	}

	m.UsersMutex.Lock()
	m.Users[user.Username] = user
	m.UsersMutex.Unlock()

	return user, true
}

// EnsureAdminUser 确保数据库中存在默认管理员账号
func (m *Manager) EnsureAdminUser() error {
	if m.GORMDB == nil {
		return fmt.Errorf("GORM database not initialized")
	}

	var existingAdmin models.UserGORM
	result := m.GORMDB.Where("username = ?", "admin").First(&existingAdmin)
	if result.Error == nil {
		// 如果已存在，确保其为激活状态且是管理员，并更新密码以匹配配置
		password := m.Config.DefaultAdminPassword
		if password == "" {
			password = "admin"
		}
		hash, _ := utils.HashPassword(password)

		updates := map[string]interface{}{
			"active":   true,
			"is_admin": true,
		}

		// 只有当密码不匹配时才更新密码，避免不必要的哈希计算
		if !utils.CheckPassword(password, existingAdmin.PasswordHash) {
			updates["password_hash"] = hash
			log.Printf("Admin user 'admin' password updated to match config")
		}

		if existingAdmin.SessionVersion == 0 {
			updates["session_version"] = 1
			existingAdmin.SessionVersion = 1
		}

		m.GORMDB.Model(&existingAdmin).Updates(updates)
		log.Printf("Admin user 'admin' verified and is active")

		// 同步更新内存缓存
		m.UsersMutex.Lock()
		user, ok := m.Users["admin"]
		if !ok {
			// 如果内存中没有，则创建一个并加入
			user = &types.User{
				ID:             int64(existingAdmin.ID),
				Username:       existingAdmin.Username,
				IsAdmin:        existingAdmin.IsAdmin,
				Active:         existingAdmin.Active,
				SessionVersion: existingAdmin.SessionVersion,
				CreatedAt:      existingAdmin.CreatedAt,
				UpdatedAt:      existingAdmin.UpdatedAt,
			}
			m.Users[existingAdmin.Username] = user
		}

		user.Active = true
		user.IsAdmin = true
		user.SessionVersion = existingAdmin.SessionVersion
		if h, ok := updates["password_hash"].(string); ok {
			user.PasswordHash = h
		} else {
			user.PasswordHash = existingAdmin.PasswordHash
		}
		m.UsersMutex.Unlock()

		return nil
	}

	password := m.Config.DefaultAdminPassword
	if password == "" {
		password = "admin" // 默认回退
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash default admin password: %v", err)
	}

	admin := &models.UserGORM{
		Username:       "admin",
		PasswordHash:   hash,
		IsAdmin:        true,
		Active:         true,
		SessionVersion: 1, // 明确设置版本
	}

	if err := m.GORMDB.Create(admin).Error; err != nil {
		return fmt.Errorf("failed to create default admin user: %v", err)
	}

	log.Printf("Default admin user 'admin' created successfully")

	// 同步到内存
	m.UsersMutex.Lock()
	m.Users[admin.Username] = &types.User{
		ID:             int64(admin.ID),
		Username:       admin.Username,
		PasswordHash:   admin.PasswordHash,
		IsAdmin:        admin.IsAdmin,
		Active:         admin.Active,
		SessionVersion: admin.SessionVersion,
		CreatedAt:      admin.CreatedAt,
		UpdatedAt:      admin.UpdatedAt,
	}
	m.UsersMutex.Unlock()

	return nil
}

// StartTrendCollection starts the background trend collection task
func (m *Manager) StartTrendCollection() {
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			m.SaveAllStatsToDB()
		}
	}()
}

// LoadStatsFromDB loads statistics from database (stub)
func (m *Manager) LoadStatsFromDB() error {
	// Implementation depends on how stats are stored in DB
	return nil
}
