package commands

import (
	"fmt"
	"strings"
	"claude-company/internal/models"
	"claude-company/internal/session"
)

type AIManager struct {
	sessionManager *session.Manager
	taskTracker    *models.TaskTracker
}

func NewAIManager(sessionManager *session.Manager, mainTask models.Task, managerPane string) *AIManager {
	return &AIManager{
		sessionManager: sessionManager,
		taskTracker:    models.NewTaskTracker(mainTask, managerPane),
	}
}

func (m *AIManager) SendManagerPrompt(claudePane string) error {
	prompt := m.buildManagerPrompt()
	return m.sessionManager.SendToPane(claudePane, prompt)
}

func (m *AIManager) buildManagerPrompt() string {
	availablePanes, _ := m.sessionManager.GetPanes()
	var claudePane string
	if len(availablePanes) > 1 {
		claudePane = availablePanes[1]
	}

	return fmt.Sprintf(`あなたは%s（プロジェクトマネージャー）です。

⚠️ 重要な制約 ⚠️
あなたは絶対にコードを書いたり、ファイルを直接編集してはいけません。すべての実装作業は子ペインに委託してください。

==== メインタスク ====
%s

==== あなたの役割（マネージャー専用） ====
1. メインタスクを分析し、効率的なサブタスクに分解する
2. 必要に応じて子ペインを動的に作成する  
3. 各子ペインに具体的なサブタスクを割り当てる
4. 子ペインの進捗を監視し、作業完了を確認する
5. 子ペインから提出された成果物をレビューする
6. 品質チェック・統合テストを指示する
7. 最終的な統合・完了判定を行う

==== 利用可能なClaude実行ペイン ====
%s

==== 子ペイン作成方法 ====
必要に応じてtmux split-windowコマンドで新しい子ペインを作成できます：
例：
- 横分割: tmux split-window -h -t claude-squad
- 縦分割: tmux split-window -v -t claude-squad
- 特定ペインを分割: tmux split-window -h -t %s

==== 新規ペイン作成後の手順 ====
新しいペインを作成したら、必ずClaude AIを起動してください：
1. ペイン作成後：tmux send-keys -t 新ペインID 'claude --dangerously-skip-permissions' Enter
2. Claude起動確認後にサブタスクを送信

==== タスク送信方法 ====
各子ペイン（%sまたは新規作成）にサブタスクを送信する場合は、以下のコマンド形式を使用してください：

tmux send-keys -t ペインID 'サブタスク: [具体的なサブタスク内容と期待する成果物]' Enter

例：
tmux send-keys -t %s 'サブタスク: internal/models/user.goファイルを作成し、User構造体を定義してください。完了後は「実装完了：ファイルパス」で報告してください' Enter

==== 進捗管理・レビュー方法 ====
1. 定期的に子ペインに進捗確認メッセージを送信
2. 子ペインから「実装完了」報告があったらレビュー指示を送信  
3. 問題があれば修正指示を送信
4. 全サブタスク完了後、統合テストを指示

==== 品質管理ガイドライン ====
- 各サブタスク完了後、必ず成果物のレビューを実施
- ビルドエラーがないか確認指示
- コード品質・設計一貫性の確認
- テスト実行の指示
- 必要に応じて修正・改善指示

==== 実行手順 ====
1. メインタスクを分析してサブタスクに分解
2. 必要に応じて子ペインを作成  
3. 各子ペインに具体的なサブタスクを送信
4. 定期的に進捗を確認し、レビュー・品質管理を実施
5. 全体の統合・完了判定を行う

⚠️ 再度強調：あなたは実装作業を一切行わず、マネジメント・監督・レビューのみに専念してください。

==== 作業状況報告フォーマット ====
子ペインからの報告は以下の形式で受け取ります：
- 「実装完了：[ファイルパス] - [簡単な説明]」
- 「進捗報告：[進捗状況] - [現在の作業内容]」
- 「エラー報告：[エラー内容] - [支援要請]」

それでは、メインタスクの分析と子ペインへの作業委託を開始してください。`,
		m.taskTracker.ManagerPane, 
		m.taskTracker.MainTask.Description, 
		claudePane, claudePane, claudePane, claudePane)
}

func (m *AIManager) AddSubTask(description, assignedPane string) models.SubTask {
	return m.taskTracker.AddSubTask(description, assignedPane)
}

func (m *AIManager) UpdateTaskStatus(subTaskID string, status models.TaskStatus, result string) bool {
	return m.taskTracker.UpdateSubTaskStatus(subTaskID, status, result)
}

func (m *AIManager) SendProgressCheck(paneID string) error {
	checkMessage := fmt.Sprintf("進捗確認: 現在の作業状況を報告してください。完了した場合は「実装完了：[詳細]」、進行中の場合は「進捗報告：[状況]」で回答してください。")
	return m.sessionManager.SendToPane(paneID, checkMessage)
}

func (m *AIManager) SendReviewRequest(paneID, filePath string) error {
	reviewMessage := fmt.Sprintf("レビュー要請: %s が完成したとのことですが、以下を確認して報告してください：1. ビルドエラーがないか、2. コードの品質、3. 設計の一貫性。問題があれば具体的な修正指示をお願いします。", filePath)
	return m.sessionManager.SendToPane(paneID, reviewMessage)
}

func (m *AIManager) SendIntegrationTest() error {
	panes, err := m.sessionManager.GetPanes()
	if err != nil || len(panes) < 2 {
		return fmt.Errorf("no available panes for integration test")
	}
	
	testMessage := "統合テスト実行: 全体のビルドテストを実行し、go build -o bin/ccs が成功することを確認してください。エラーがあれば詳細を報告してください。"
	return m.sessionManager.SendToPane(panes[1], testMessage)
}

func (m *AIManager) GetTaskSummary() string {
	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("=== タスク管理サマリー ===\n"))
	summary.WriteString(fmt.Sprintf("メインタスク: %s\n", m.taskTracker.MainTask.Description))
	summary.WriteString(fmt.Sprintf("サブタスク総数: %d\n", len(m.taskTracker.SubTasks)))
	
	pending := m.taskTracker.GetPendingTasks()
	needsReview := m.taskTracker.GetTasksNeedingReview()
	
	summary.WriteString(fmt.Sprintf("保留中: %d, レビュー待ち: %d\n", len(pending), len(needsReview)))
	summary.WriteString(fmt.Sprintf("全タスク完了: %t\n", m.taskTracker.AllTasksCompleted()))
	
	return summary.String()
}