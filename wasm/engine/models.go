package engine

// Effect 定义了一个作用效果：当事件发生时，哪个 Flag 将被修改
type Effect struct {
	FlagName string `json:"flag_name"` 
	Value    int    `json:"value"`     
}

// Dependency 结构化的任务依赖，可实现正向依赖（满足时加入列表）、负向依赖（冲突或过期时踢出列表）
type Dependency struct {
	QuestID string 
	MustBe  bool
}

// SceneConfig 场景配置
// 定义了后端感知的场景逻辑，布局信息(Layout)由前端 JSON 独立维护
type SceneConfig struct {
	ID            string   `json:"id"`
	RelevantFlags []string `json:"relevant_flags"` // 该场景关注哪些全局 Flag
}

// Interaction 静态交互逻辑
// 对应场景中点击普通物体（非当前任务目标）时的反应
type Interaction struct {
	TargetID    string `json:"target_id"`
	Description string `json:"description"` // 返回给前端的文本

	// 触发条件 
	RequiredFlag  string `json:"required_flag"`  // 只有当此 Flag 等于...
	RequiredValue int    `json:"required_value"` // ...这个值时，才触发此交互

	// 执行后果
	Effects      []Effect `json:"effects"`       // 修改一组全局 Flag
	ItemReward   []string `json:"item_reward"`   // 获得道具 ID 列表
	TargetScene  string   `json:"target_scene"`  // 如果不为空，触发后切换场景
	TriggerEvent string   `json:"trigger_event"` // 触发前端动画/音效的指令 (如: "PLAY_CRACK_SOUND")
}

// QuestType 定义任务类型
type QuestType string

const (
	Click QuestType = "CLICK"
	Drag  QuestType = "DRAG"
	Input QuestType = "INPUT"
)

// Quest 任务/解谜逻辑
type Quest struct {
	ID        string `json:"id"`
	TargetObj string `json:"target_obj"`

	// 触发条件
	RequiredFlag  string          `json:"required_flag"`
	RequiredValue int             `json:"required_value"`
	DependsOn     []Dependency `json:"depends_on"` // 前置任务 ID 列表

	// 解谜逻辑
	Type           QuestType `json:"type"`
	RequiredItem   string    `json:"required_item"`   // DRAG 模式下需要的物品 ID
	CorrectContent string    `json:"correct_content"` // INPUT 模式下需要的内容

	// 触发后效果
	Description  []string `json:"description"`   // 0-失败，1-成功
	Effects      []Effect `json:"effects"`       // 任务完成后修改的世界状态
	TriggerEvent []string `json:"trigger_event"` // 0-失败， 1-成功
}

// WorldConfig 总配置 (从 JSON 加载)
type WorldConfig struct {
	Scenes             map[string]SceneConfig   `json:"scenes"`
	StaticInteractions map[string][]Interaction `json:"static_interactions"` // 注意：一个 ID 对应一组交互(不同状态)
	Quests             map[string]Quest         `json:"quests"`
}

// GameState 玩家存档数据
// 只有这个结构体的数据会被序列化存入数据库
type GameState struct {
	SceneID         string          `json:"scene_id"`         // 当前所在场景
	Inventory       []string        `json:"inventory"`        // 背包物品 ID 列表
	WorldFlags      map[string]int  `json:"world_flags"`      // 全局状态开关 (核心事实数据库)
	ActiveQuestID   []string        `json:"active_quest_id"`  // 当前激活的任务列表
	CompletedQuests map[string]bool `json:"completed_quests"` // 已完成的任务
	Mode            string          `json:"mode"`             // 游戏模式 ("NORMAL", "",) 
}

// InteractionInput 前端 -> WASM 的输入
type InteractionInput struct {
	Action  string `json:"action"`  // "CLICK", "DRAG", "INPUT"
	Target  string `json:"target"`  // 交互目标的 ID
	Item    string `json:"item"`    // 拖拽的物品 ID (可选)
	Content string `json:"content"` // 输入的密码 (可选)
}

// InteractionResult WASM -> 前端的输出
type InteractionResult struct {
	Status       string `json:"status"`                  // "SUCCESS", "FAIL", "CHANGE_SCENE", "HINT", "NONE"
	Message      string `json:"message"`                 // 显示给玩家的文字
	TriggerEvent string `json:"trigger_event,omitempty"` // 前端需要执行的动画

	// 状态变更
	NextSceneID  string         `json:"next_scene_id,omitempty"` // 如果发生场景切换
	UpdatedFlags map[string]int `json:"updated_flags,omitempty"` // 仅包含当前场景感知到的 Flag 变化

	// 物品变更
	RemoveItem string   `json:"remove_item,omitempty"` // 需要从背包移除的物品
	NewItems   []string `json:"new_items,omitempty"`   // 获得的物品
}

// Engine 核心结构
type Engine struct {
	gameState   *GameState
	worldConfig *WorldConfig
}
