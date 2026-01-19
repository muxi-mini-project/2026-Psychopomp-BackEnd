package engine

type InteractionType string

const (
	Click InteractionType = "CLICK"
	Drag  InteractionType = "DRAG"
	Input InteractionType = "INPUT"
)

// Quest 任务逻辑定义
type Quest struct {
	ID             string          `json:"id"`
	Type           InteractionType `json:"type"`
	TargetID       string          `json:"target_id"`       // 完成任务交互目标物体ID
	RequiredItem   string          `json:"required_item"`   // 完成任务物品需求
	CorrectContent string          `json:"correct_content"` // 输入类任务的正确内容
	IsOptional     bool            `json:"is_optional"`     // 是否为可选任务
	DependsOn      map[string]bool `json:"depends_on"`      // 任务依赖(需满足前置任务条件，且不可做过期任务)

	Description  []string `json:"description"`   // 剧情文字:0-失败, 1-成功
	TriggerEvent []string `json:"trigger_event"` // 动画指令、状态变更:0-失败, 1-成功
}

// GameState 玩家进度存档
type GameState struct {
	SceneID         string          `json:"scene_id"`         // 当前场景ID
	Inventory       []string        `json:"inventory"`        // 玩家物品栏
	ActiveQuestID   []string        `json:"active_quest_id"`  // 当前任务池
	CompletedQuests map[string]bool `json:"completed_quests"` // 已完成任务记录
	WorldFlags      map[string]int  `json:"world_flags"`      // 场景物体状态记录
	LastSaveTime    int64           `json:"last_save_time"`
}

// Interaction 静态交互（观察、环境反馈）
type Interaction struct {
	TargetID    string   `json:"target_id"`   // 交互对象ID
	Description string   `json:"description"` // 交互描述文字
	ItemReward  []string `json:"item_reward"` // 触发后获得的物品
	Actions     []string `json:"actions"`     // 触发的动作
	Flag        string   `json:"flag"`        // 交互后改变的状态名
	FlagValue   int      `json:"flag_value"`  // 交互后改变的状态值
}

// WorldConfig 游戏静态数据配置
type WorldConfig struct {
	StaticInteractions map[string]Interaction `json:"static_interactions"` // 物体ID到静态交互定义的映射
	Quests             map[string]Quest       `json:"quests"`              // 任务ID到任务定义的映射
	Scenes             map[string]string      `json:"scenes"`              // 场景ID到场景文件的映射
}

// Engine 核心驱动
type Engine struct {
	WorldConfig *WorldConfig
	GameState   *GameState
}
