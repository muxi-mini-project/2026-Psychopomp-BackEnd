//go:build js && wasm

package engine

// HandleAction 处理前端传来的所有交互
func (e *Engine) HandleAction(input InteractionInput) InteractionResult {
	targetID := input.Target

	for idx, qID := range e.gameState.ActiveQuestID {
		quest := e.worldConfig.Quests[qID]

		if quest.TargetObj == targetID {
			if e.validateQuestInput(input, quest) && e.checkCondition(quest.RequiredFlag, quest.RequiredValue) {
				return e.processQuestSuccess(qID, idx)
			} else {
				return InteractionResult{
					Status:       "FAIL",
					Message:      quest.Description[0],
					TriggerEvent: quest.TriggerEvent[0],
				}
			}
		}
	}

	if interactions, exists := e.worldConfig.StaticInteractions[targetID]; exists {
		// 遍历该物体定义的所有可能交互 (按顺序匹配)
		for _, inter := range interactions {
			if e.checkCondition(inter.RequiredFlag, inter.RequiredValue) {

				localUpdates := e.applyEffects(inter.Effects)

				result := InteractionResult{
					Status:       "SUCCESS",
					Message:      inter.Description,
					TriggerEvent: inter.TriggerEvent,
					UpdatedFlags: localUpdates,
					NewItems:     inter.ItemReward,
				}

				if len(inter.ItemReward) > 0 {
					e.gameState.Inventory = append(e.gameState.Inventory, inter.ItemReward...)
				}

				if inter.TargetScene != "" {
					e.gameState.SceneID = inter.TargetScene
					result.Status = "CHANGE_SCENE"
					result.NextSceneID = inter.TargetScene
					// 切换场景时，清空 updated_flags，前端将重绘场景
					result.UpdatedFlags = nil
				}

				return result
			}
		}
	}

	// 兜底逻辑
	return InteractionResult{Status: "NONE", Message: "没什么特别的。"}
}

// GetCurrentSceneState 用于获取新场景的初始状态
func (e *Engine) GetCurrentSceneState() map[string]any {
	sceneID := e.gameState.SceneID
	sceneConf, exists := e.worldConfig.Scenes[sceneID]

	localFlags := make(map[string]int)

	if exists {
		// 只提取该场景关心的 Flag
		for _, flagName := range sceneConf.RelevantFlags {
			localFlags[flagName] = e.gameState.WorldFlags[flagName]
		}
	}

	return map[string]any{
		"scene_id":     sceneID,
		"active_flags": localFlags,
		"inventory":    e.gameState.Inventory,
	}
}

// applyEffects 修改全局，返回局部
func (e *Engine) applyEffects(effects []Effect) map[string]int {
	updates := make(map[string]int)
	if len(effects) == 0 {
		return updates
	}

	currentSceneID := e.gameState.SceneID
	currentSceneConf := e.worldConfig.Scenes[currentSceneID]

	relevantSet := make(map[string]bool)
	for _, f := range currentSceneConf.RelevantFlags {
		relevantSet[f] = true
	}

	for _, eff := range effects {
		e.gameState.WorldFlags[eff.FlagName] = eff.Value

		if relevantSet[eff.FlagName] {
			updates[eff.FlagName] = eff.Value
		}
	}

	return updates
}

// processQuestSuccess 处理任务成功完成的逻辑
func (e *Engine) processQuestSuccess(qID string, idxInActive int) InteractionResult {
	quest := e.worldConfig.Quests[qID]

	e.gameState.ActiveQuestID = append(e.gameState.ActiveQuestID[:idxInActive], e.gameState.ActiveQuestID[idxInActive+1:]...)
	e.gameState.CompletedQuests[qID] = true

	e.updateQuestList()

	localUpdates := e.applyEffects(quest.Effects)

	return InteractionResult{
		Status:       "SUCCESS",
		Message:      quest.Description[1],
		TriggerEvent: quest.TriggerEvent[1],
		UpdatedFlags: localUpdates,
		RemoveItem:   quest.RequiredItem, // 只有 Quest 会消耗物品
	}
}

// validateQuestInput 校验玩家的操作是否符合任务要求
func (e *Engine) validateQuestInput(input InteractionInput, quest Quest) bool {
	if input.Action != string(quest.Type) {
		return false
	}

	switch quest.Type {
	case Drag:
		return input.Item == quest.RequiredItem
	case Input:
		return input.Content == quest.CorrectContent
	case Click:
		return true
	default:
		return false
	}
}

// checkCondition 辅助检查条件
func (e *Engine) checkCondition(flag string, val int) bool {
	if flag == "" {
		return true // 没有条件即为无条件满足
	}
	
	// 检查 Map 中是否存在 key，如果不存在默认为 0
	currentVal := e.gameState.WorldFlags[flag]
	return currentVal == val
}

// unlockNewQuests 将满足正向依赖的任务加入 Active 列表, 将冲突或过期的任务踢出
func (e *Engine) updateQuestList() {
	newActiveList := []string{}

	// 删除过期冲突任务
	for _, qID := range e.gameState.ActiveQuestID {
		quest := e.worldConfig.Quests[qID]
		keep := true

		for _, dep := range quest.DependsOn {
			isCompleted := e.gameState.CompletedQuests[qID]
			if dep.MustBe == false && isCompleted {
				keep = false
				break
			}
		}

		if keep {
			newActiveList = append(newActiveList, qID)
		}
	}

	e.gameState.ActiveQuestID = newActiveList

	// 添加新任务
	for qID, quest := range e.worldConfig.Quests {
		if len(quest.DependsOn) == 0 {
			continue
		}

		if e.gameState.CompletedQuests[qID] || contains(e.gameState.ActiveQuestID, qID) {
			continue
		}

		satisfiesAll := true
		for _, dep := range quest.DependsOn {
			isCompleted := e.gameState.CompletedQuests[dep.QuestID]
			if isCompleted && dep.MustBe {
				satisfiesAll = false
				break
			}
		}

		if satisfiesAll {
			e.gameState.ActiveQuestID = append(e.gameState.ActiveQuestID, qID)
		}

	}
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
