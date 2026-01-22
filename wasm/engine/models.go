package engine

const (
	StatusSuccess     = "SUCCESS"
	StatusFail        = "FAIL"
	StatusChangeScene = "CHANGE_SCENE"
	StatusDialogue    = "DIALOGUE"
	StatusOpenSubView = "OPEN_SUBVIEW"

	QuestTypeClick  = "CLICK"
	QuestTypeDrag   = "DRAG"
	QuestTypeInput  = "INPUT"
	QuestTypeCustom = "CUSTOM"
)

type Effect struct {
	FlagName string `json:"flag_name"`
	Value    int    `json:"value"`
}

type Dependency struct {
	QuestID string `json:"quest_id"`
	MustBe  bool   `json:"must_be"` // true: 必须已完成; false: 必须未完成(用于任务互斥/过期)
}

type SceneConfig struct {
	ID            string   `json:"id"`
	RelevantFlags []string `json:"relevant_flags"` // 该场景关心的 Flag，用于前端渲染过滤
}

type Interaction struct {
	TargetID string `json:"target_id"`

	// 触发条件
	RequiredFlag  string `json:"required_flag"`
	RequiredValue int    `json:"required_value"`

	// 文本与对话
	Description string   `json:"description"`
	Dialogue    []string `json:"dialogue"` // 不为空则触发 AVG 对话模式
	Speaker     string   `json:"speaker"`

	Effects      []Effect `json:"effects"`
	ItemReward   []string `json:"item_reward"`
	TargetScene  string   `json:"target_scene"`  // 触发场景切换
	TriggerEvent string   `json:"trigger_event"` // 触发前端特效/动画 (如 "play_sound_click")

	// 特殊 UI
	SubViewID string `json:"sub_view_id"` // 打开观察界面 (如看一封信的特写)
}

// Quest 任务配置
type Quest struct {
	ID        string `json:"id"`
	TargetObj string `json:"target_obj"`

	// 激活条件
	DependsOn     []Dependency `json:"depends_on"` // 任务依赖链
	RequiredFlag  string       `json:"required_flag"`
	RequiredValue int          `json:"required_value"`

	// 交互逻辑
	Type         string `json:"type"`          // CLICK, DRAG, INPUT, CUSTOM
	RequiredItem string `json:"required_item"` // 需要的物品 ID (成功后通常会消耗)
	CorrectCode  string `json:"correct_code"`  // 密码或自定义逻辑的校验码
	SubViewID    string `json:"sub_view_id"`   // CUSTOM 模式下需弹出的界面 ID

	// 触发反馈
	Description  []string `json:"description"`
	Dialogue     []string `json:"dialogue"`
	Speaker      string   `json:"speaker"`
	Effects      []Effect `json:"effects"`
	TriggerEvent string   `json:"trigger_event"`
}

// 总配置 (从 JSON 加载)
type WorldConfig struct {
	Scenes             map[string]SceneConfig   `json:"scenes"`
	StaticInteractions map[string][]Interaction `json:"static_interactions"`
	Quests             map[string]Quest         `json:"quests"`
}

type GameState struct {
	SceneID         string          `json:"scene_id"`
	Inventory       []string        `json:"inventory"`
	WorldFlags      map[string]int  `json:"world_flags"`
	ActiveQuestID   []string        `json:"active_quest_id"`
	CompletedQuests map[string]bool `json:"completed_quests"`

	// 仅用于判断是否需要存档
	IsDirty bool `json:"-"`
}

type GameSaveData struct {
	State     GameState `json:"state"`
	Timestamp int64     `json:"timestamp"`
	Version   string    `json:"version"` // 存档版本号（万一我写的这玩意还会更新呢）
}

type InteractionInput struct {
	Action string `json:"action"`
	Target string `json:"target"`
	Item   string `json:"item"`
	Code   string `json:"code"`
}

type InteractionResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`

	TriggerEvent string `json:"trigger_event,omitempty"`
	NextSceneID  string `json:"next_scene_id,omitempty"`

	UpdatedFlags map[string]int `json:"updated_flags,omitempty"` // 仅包含当前场景关心的 Flag

	RemoveItem string   `json:"remove_item,omitempty"`
	NewItems   []string `json:"new_items,omitempty"`

	Dialogue  []string `json:"dialogue,omitempty"`
	Speaker   string   `json:"speaker,omitempty"`
	SubViewID string   `json:"sub_view_id,omitempty"` // 需打开的小游戏/UI Prefab ID

	// 指示前端立即保存
	AutoSave bool `json:"auto_save,omitempty"`
}
