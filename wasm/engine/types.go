package engine

// Quest 主线任务
type Quest struct {
	ID             string          `json:"id"`
	Type           InteractionType `json:"type"`            // 交互类型
	TargetObj      string          `json:"target_obj"`      // 目标物体ID
	RequiredItem   string          `json:"required_item"`   // 如果是DRAG，需要的道具ID
	CorrectContent string          `json:"correct_content"` // 如果是INPUT，正确的密码

	Description  string `json:"description"`   // 成功后的剧情文字
	TriggerEvent string `json:"trigger_event"` // 成功后的动画指令
}

// GameState 游戏状态
type GameState struct {
	ActiveQuestID []string        `json:"active_quest_id"` // 存储当前任务id
	Inventory     []string        `json:"inventory"`       // 物品栏
	Unlocked      map[string]bool `json:"unlocked"`        // 记录哪些支线或开关已打开
}

// Interaction 交互点
type Interaction struct {
	TargetID    string   `json:"target_id"`   // 任务id
	Description string   `json:"description"` // 剧情文本
	ItemReward  []string `json:"item_reward"` // 触发后获得的道具
	Actions     []string `json:"actions"`     // 触发的动作 (如 PLAY_SOUND)
}

// 基本类型
type InteractionType string

const (
	Click InteractionType = "CLICK"
	Drag  InteractionType = "DRAG"
	Input InteractionType = "INPUT"
)

// WorldConfig 互动内容
type WorldConfig struct {
	StaticInteractions map[string]Interaction `json:"static_interactions"` // 互动点
	Quests             map[string]Quest       `json:"quests"`              // 主线
}

// Engine 游戏引擎、核心结构体
type Engine struct {
	worldConfig *WorldConfig
	gameState   *GameState
}