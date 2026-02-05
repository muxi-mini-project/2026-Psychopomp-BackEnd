package engine

import (
	"encoding/json"
	"time"
)

type Engine struct {
	WorldConfig *WorldConfig
	GameState   *GameState
}

// HandleAction 处理前端传来的所有交互请求
func (e *Engine) HandleAction(input InteractionInput) InteractionResult {
	targetID := input.Target

	for i, qID := range e.GameState.ActiveQuestID {
		quest := e.WorldConfig.Quests[qID]

		if quest.TargetObj == targetID && e.checkCondition(quest.RequiredFlag, quest.RequiredValue) {
			if quest.Type == QuestTypeCustom && input.Action == "CLICK" {
				return InteractionResult{
					Status:    StatusOpenSubView,
					SubViewID: quest.SubViewID,
					Message:   quest.Description[1],
				}
			}

			if e.validateQuestInput(input, quest) {
				return e.processQuestSuccess(qID, i)
			}

			if input.Action == "DRAG" || input.Action == "INPUT" || input.Action == "CUSTOM" {
				return InteractionResult{Status: StatusFail, Message: "这似乎不起作用。"}
			}
		}
	}

	if interactions, exists := e.WorldConfig.StaticInteractions[targetID]; exists {
		for _, inter := range interactions {
			if e.checkCondition(inter.RequiredFlag, inter.RequiredValue) {

				localUpdates := e.applyEffects(inter.Effects)

				result := InteractionResult{
					Status:       StatusSuccess,
					Message:      inter.Description,
					Dialogue:     inter.Dialogue,
					Speaker:      inter.Speaker,
					TriggerEvent: inter.TriggerEvent,
					SubViewID:    inter.SubViewID,
					UpdatedFlags: localUpdates,
					NewItems:     inter.ItemReward,
				}

				if len(inter.ItemReward) > 0 {
					e.GameState.Inventory = append(e.GameState.Inventory, inter.ItemReward...)
					e.GameState.IsDirty = true
				}

				if len(inter.Dialogue) > 0 {
					result.Status = StatusDialogue
				} else if inter.SubViewID != "" {
					result.Status = StatusOpenSubView
				}

				if inter.TargetScene != "" {
					e.GameState.SceneID = inter.TargetScene
					result.Status = StatusChangeScene
					result.NextSceneID = inter.TargetScene
					result.UpdatedFlags = nil // 切场景时不需要返回旧场景的 Flag 更新，前端会重新拉取新场景
					e.GameState.IsDirty = true
				}

				// 如果发生了重要变化 (切场景/拿物品)，标记自动存档
				if e.GameState.IsDirty {
					result.AutoSave = true
				}

				return result
			}
		}
	}

	return InteractionResult{Status: "NONE", Message: "没什么特别的。"}
}

// validateQuestInput 校验输入是否满足任务要求
func (e *Engine) validateQuestInput(input InteractionInput, quest Quest) bool {
	if quest.Type == QuestTypeCustom {
		if input.Action != "CUSTOM" {
			return false
		}
		if input.Code != quest.CorrectCode {
			return false
		}
		if quest.RequiredItem != "" && input.Item != quest.RequiredItem {
			return false
		}
		return true
	}

	if input.Action != quest.Type {
		return false
	}

	switch quest.Type {
	case QuestTypeDrag:
		return input.Item == quest.RequiredItem
	case QuestTypeInput:
		return input.Code == quest.CorrectCode
	case QuestTypeClick:
		return true
	}
	return false
}

// processQuestSuccess 任务成功后的核心处理
func (e *Engine) processQuestSuccess(qID string, idx int) InteractionResult {
	quest := e.WorldConfig.Quests[qID]

	if quest.RequiredItem != "" {
		e.removeItemFromInventory(quest.RequiredItem)
	}

	e.GameState.ActiveQuestID = append(e.GameState.ActiveQuestID[:idx], e.GameState.ActiveQuestID[idx+1:]...)
	e.GameState.CompletedQuests[qID] = true

	localUpdates := e.applyEffects(quest.Effects)

	e.RefreshQuests()

	e.GameState.IsDirty = true

	res := InteractionResult{
		Status:       StatusSuccess,
		Message:      quest.Description[1],
		Dialogue:     quest.Dialogue,
		Speaker:      quest.Speaker,
		TriggerEvent: quest.TriggerEvent,
		UpdatedFlags: localUpdates,
		RemoveItem:   quest.RequiredItem, // 前端视觉上移除该物品
		AutoSave:     true,               // 强制触发前端存档
	}

	if len(quest.Dialogue) > 0 {
		res.Status = StatusDialogue
	}

	return res
}

// refreshQuests 处理解锁与冲突（剧情分支核心逻辑）
func (e *Engine) RefreshQuests() {
	validActive := []string{}
	for _, aqID := range e.GameState.ActiveQuestID {
		quest := e.WorldConfig.Quests[aqID]
		keep := true
		for _, dep := range quest.DependsOn {
			if dep.MustBe == false && e.GameState.CompletedQuests[dep.QuestID] {
				keep = false
				break
			}
		}
		if keep {
			validActive = append(validActive, aqID)
		}
	}
	e.GameState.ActiveQuestID = validActive

	for id, q := range e.WorldConfig.Quests {
		if e.GameState.CompletedQuests[id] || contains(e.GameState.ActiveQuestID, id) {
			continue
		}

		allMet := true
		for _, dep := range q.DependsOn {
			isCompleted := e.GameState.CompletedQuests[dep.QuestID]
			if isCompleted != dep.MustBe {
				allMet = false
				break
			}
		}

		if allMet {
			e.GameState.ActiveQuestID = append(e.GameState.ActiveQuestID, id)
		}
	}
}

// applyEffects 更改 Flag 并返回当前场景关心的更改
func (e *Engine) applyEffects(effects []Effect) map[string]int {
	updates := make(map[string]int)
	if len(effects) == 0 {
		return updates
	}

	currentSceneConf := e.WorldConfig.Scenes[e.GameState.SceneID]
	relevantSet := make(map[string]bool)

	for _, f := range currentSceneConf.RelevantFlags {
		relevantSet[f] = true
	}

	for _, eff := range effects {
		e.GameState.WorldFlags[eff.FlagName] = eff.Value

		// 如果当前场景关心这个 Flag，加入返回列表
		if relevantSet[eff.FlagName] {
			updates[eff.FlagName] = eff.Value
		}
	}

	if len(effects) > 0 {
		e.GameState.IsDirty = true
	}

	return updates
}

// 我去我怎么连删东西都要遍历
func (e *Engine) removeItemFromInventory(itemID string) {
	if itemID == "" {
		return
	}

	inventory := e.GameState.Inventory
	found := false

	j := 0
	for i := 0; i < len(inventory); i++ {
		if inventory[i] == itemID && !found {
			found = true
			continue
		}
		inventory[j] = inventory[i]
		j++
	}

	if found {
		e.GameState.Inventory = inventory[:j]
		e.GameState.IsDirty = true
	}
}

// checkCondition 检查 Flag 是否满足要求
func (e *Engine) checkCondition(flag string, val int) bool {
	if flag == "" {
		return true // 无flag要求直接过
	}
	return e.GameState.WorldFlags[flag] == val
}

func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// GetSerializedState 导出存档
func (e *Engine) GetSerializedState() (string, error) {
	saveData := GameSaveData{
		State:     *e.GameState,
		Timestamp: time.Now().Unix(),
		Version:   "1.0.0", // 1.0.0 版本重磅上线！！！
	}

	bytes, err := json.Marshal(saveData)
	if err != nil {
		return "", err
	}

	e.GameState.IsDirty = false

	return string(bytes), nil
}

// LoadFromSerializedState 导入存档
func (e *Engine) LoadFromSerializedState(jsonStr string) error {
	var wrapper GameSaveData
	if err := json.Unmarshal([]byte(jsonStr), &wrapper); err != nil {
		return err
	}

	e.GameState = &wrapper.State

	e.RefreshQuests()

	return nil
}
