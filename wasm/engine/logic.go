//go:build js && wasm

package engine

// InteractionInput 交互数据模型
type InteractionInput struct {
	Action  string `json:"action"`  // CLICK, DRAG, INPUT
	Target  string `json:"target"`  // 物体ID
	Item    string `json:"item"`    // 拖拽的物品ID
	Content string `json:"content"` // 输入的密码
}

// HandleAction 操作分发
func (e *Engine) HandleAction(input InteractionInput) any {
	targetID := input.Target

	// 检查是否已经完成过 防止剧情回溯
	for qID := range e.GameState.CompletedQuests {
		quest := e.WorldConfig.Quests[qID]
		if quest.TargetID == targetID {
			return map[string]any{
				"status":  "ALREADY_DONE",
				"msg": "这里已经处理过了。",
			}
		}
	}

	// 遍历当前任务池
	for _, qID := range e.GameState.ActiveQuestID {
		quest := e.WorldConfig.Quests[qID]

		if quest.TargetID == targetID {
			// 校验交互行为是否符合要求
			if e.validateInteraction(input, quest) {
				// TODO:任务完成后的逻辑
			} else {
				return map[string]any{
					"status": "HINT",
					"msg":    quest.Description[0],
					"evt":    quest.TriggerEvent[0],
				}
			}
		}
	}

	// 匹配静态交互
	if interaction, exists := e.WorldConfig.StaticInteractions[targetID]; exists {
		if len(interaction.ItemReward) > 0 {
			e.GameState.Inventory = append(e.GameState.Inventory, interaction.ItemReward...)
			// TODO: 领取后清空奖励或记录到WorldFlags，考虑中
		}

		return map[string]any{
			"status":  "EXAMINE",
			"msg":     interaction.Description,
			"actions": interaction.Actions,
			"flag" : interaction.Flag,
			"value_of_flag" : interaction.FlagValue,
		}
	}

	// 兜底逻辑（正常不走这个）
	return map[string]any{
		"status": "NONE",
		"msg":    "只是一个普通的物体。",
	}
}

// validateInteraction 检验交互是否符合任务要求
func (e *Engine) validateInteraction(input InteractionInput, quest Quest) bool {
	switch quest.Type {
	case Click:
		return true // 点击类目标匹配即成功
	case Drag:
		return input.Item == quest.RequiredItem
	case Input:
		return input.Content == quest.CorrectContent
	default:
		return false
	}
}

//TODO: 任务成功后的逻辑
//  func (e *Engine) success 

//TODO: 更新任务池

//