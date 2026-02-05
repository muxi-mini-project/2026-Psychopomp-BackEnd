//go:build js && wasm
package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"WASM/engine" 
)

var gameEngine *engine.Engine

func main() {
	c := make(chan struct{})

	fmt.Println("WASM: Go Runtime Initializing...")

	js.Global().Set("initEngine", js.FuncOf(initEngine))

	js.Global().Set("handleAction", js.FuncOf(handleAction))

	js.Global().Set("getGameStateJSON", js.FuncOf(getGameStateJSON))

	js.Global().Set("loadGameStateJSON", js.FuncOf(loadGameStateJSON))

	fmt.Println("WASM: Ready to accept commands.")
	
	<-c
}

// initEngine 初始化游戏
// 参数 args[0]: string - WorldConfig 的 JSON 字符串
// 参数 args[1]: string - (可选) GameState 的 JSON 字符串，如果是新游戏传空字符串 ""
func initEngine(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		fmt.Println("WASM Error: initEngine requires config data")
		return false
	}

	configJSON := args[0].String()
	saveJSON := ""
	if len(args) > 1 {
		saveJSON = args[1].String()
	}

	// 解析世界配置
	var worldConfig engine.WorldConfig
	err := json.Unmarshal([]byte(configJSON), &worldConfig)
	if err != nil {
		fmt.Printf("WASM Error: Failed to parse WorldConfig: %s\n", err)
		return false
	}

	// 准备游戏状态
	var gameState *engine.GameState

	if saveJSON != "" && saveJSON != "null" && saveJSON != "undefined" {
		var wrapper engine.GameSaveData
		if err := json.Unmarshal([]byte(saveJSON), &wrapper); err == nil {
			gameState = &wrapper.State
			fmt.Println("WASM: Game loaded from save data.")
		} else {
			fmt.Printf("WASM Error: Failed to parse SaveData: %s\n", err)
			return false
		}
	} else {
		gameState = &engine.GameState{
			SceneID:         "start_room", // 重要字段，必须有初始场景
			WorldFlags:      make(map[string]int),
			Inventory:       make([]string, 0),
			ActiveQuestID:   make([]string, 0),
			CompletedQuests: make(map[string]bool),
		}
		fmt.Println("WASM: New game started.")
	}

	gameEngine = &engine.Engine{
		WorldConfig: &worldConfig,
		GameState:   gameState,
	}

    gameEngine.RefreshQuests()
	
	return true
}

// handleAction 处理前端操作
// 参数 args[0]: string - InteractionInput 的 JSON 字符串
// 返回: string - InteractionResult 的 JSON 字符串
func handleAction(this js.Value, args []js.Value) any {
	if gameEngine == nil {
		return errorResponse("Engine not initialized")
	}

	if len(args) < 1 {
		return errorResponse("No input provided")
	}

	inputJSON := args[0].String()
	var input engine.InteractionInput
	err := json.Unmarshal([]byte(inputJSON), &input)
	if err != nil {
		return errorResponse("Invalid input JSON")
	}

	result := gameEngine.HandleAction(input)

	resBytes, err := json.Marshal(result)
	if err != nil {
		return errorResponse("Result marshal failed")
	}

	return string(resBytes)
}

// getGameStateJSON 获取当前存档
// 返回: string - 完整存档 JSON
func getGameStateJSON(this js.Value, args []js.Value) any {
	if gameEngine == nil {
		return ""
	}

	jsonStr, err := gameEngine.GetSerializedState()
	if err != nil {
		fmt.Printf("WASM Error: Serialize failed: %s\n", err)
		return ""
	}
	return jsonStr
}

// loadGameStateJSON 运行时动态读档
// 参数 args[0]: string - 存档 JSON
func loadGameStateJSON(this js.Value, args []js.Value) any {
	if gameEngine == nil {
		return false
	}
	if len(args) < 1 {
		return false
	}

	jsonStr := args[0].String()
	err := gameEngine.LoadFromSerializedState(jsonStr)
	if err != nil {
		fmt.Printf("WASM Error: Load failed: %s\n", err)
		return false
	}
	
	fmt.Println("WASM: State loaded successfully.")
	return true
}

// errorResponse 返回一个表示错误信息的 JSON 
func errorResponse(msg string) string {
	return fmt.Sprintf(`{"status":"FAIL", "message":"WASM Error: %s"}`, msg)
}