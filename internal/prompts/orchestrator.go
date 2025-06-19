package prompts

import (
	"fmt"
	"strings"
)

// OrchestratorPrompts manages templates for orchestrator prompts
type OrchestratorPrompts struct {
	*TemplateManager
}

// NewOrchestratorPrompts creates a new orchestrator prompt manager
func NewOrchestratorPrompts() *OrchestratorPrompts {
	op := &OrchestratorPrompts{
		TemplateManager: NewTemplateManager(),
	}
	
	if err := op.registerTemplates(); err != nil {
		panic(fmt.Sprintf("Failed to register orchestrator templates: %v", err))
	}
	
	return op
}

// OrchestratorData represents data for orchestrator prompts
type OrchestratorData struct {
	PaneID      string
	MainTask    string
	Context     string
	PaneList    []string
	ChildPanes  []string
	ReportFormat string
}

func (op *OrchestratorPrompts) registerTemplates() error {
	// Main manager prompt template
	managerTemplate := `ultrathink

プロジェクトマネージャー({{.PaneID}})として機能してください。

## 制限事項
禁止: コード編集、ファイル操作、ビルド、テスト、デプロイ、技術実装
許可: コード解析、タスク分析・分解、割り当て、進捗管理、品質管理、統合判定

## メインタスク
{{.MainTask}}

{{if .Context}}
## 追加コンテキスト
{{.Context}}
{{end}}

## 管理フロー
1. コードの理解
2. タスク分析→サブタスク分解
3. 子ペイン作成(並行可能なら複数)
4. サブタスク割り当て
5. 子ペインに依頼したサブタスクの進捗監視・成果物レビュー
6. 統合テスト指示・完了判定

## ペイン操作
**作成**: tmux split-window -v -t claude-squad
**起動**: tmux send-keys -t 新ペインID 'claude --dangerously-skip-permissions' Enter
**送信**: tmux send-keys -t 新ペインID Enter

## サブタスク送信
**重要**: 子ペインのみに送信、親ペイン({{.PaneID}})は管理専用

テンプレート:
` + "`" + `
サブタスク: [タスク名]
目的: [達成目標]
成果物: [具体的な成果物]
完了条件: [完了基準]
報告方法: tmux send-keys -t {{.PaneID}} '[報告内容]' Enter; sleep 1; tmux send-keys -t {{.PaneID}} '' Enter
送信方法: tmux send-keys -t %s Enter
` + "`" + `

## 進捗管理
- 定期進捗確認
- 完了報告時のレビュー指示
- 問題発生時の修正指示
- 全体統合テスト指示

## 報告フォーマット
{{if .ReportFormat}}{{.ReportFormat}}{{else}}- 実装完了: [ファイルパス] - [説明]
- 進捗報告: [状況] - [作業内容]
- エラー報告: [内容] - [支援要請]{{end}}

メインタスクの分析とサブタスク委託を開始してください。`

	if err := op.RegisterTemplate("manager", managerTemplate); err != nil {
		return err
	}

	// Task assignment template
	taskTemplate := `サブタスク: {{.TaskDesc}}
目的: {{.Context}}
成果物: {{.AdditionalData.deliverables}}
完了条件: {{.AdditionalData.completion_criteria}}
報告方法: tmux send-keys -t {{.AdditionalData.report_pane}} '{{.AdditionalData.report_message}}' Enter; sleep 1; tmux send-keys -t {{.AdditionalData.report_pane}} '' Enter`

	if err := op.RegisterTemplate("task_assignment", taskTemplate); err != nil {
		return err
	}

	// Progress check template
	progressTemplate := `進捗確認: {{.TaskDesc}}

現在の状況と次のステップを報告してください:
- 完了した作業
- 進行中の作業
- 遭遇した問題
- 必要な支援

{{if .AdditionalData.deadline}}期限: {{.AdditionalData.deadline}}{{end}}`

	if err := op.RegisterTemplate("progress_check", progressTemplate); err != nil {
		return err
	}

	// Review request template
	reviewTemplate := `レビュー依頼: {{.TaskDesc}}

以下の成果物をレビューし、品質確認してください:
{{.AdditionalData.deliverables}}

確認項目:
- 仕様への適合性
- コード品質
- テスト充足度
- ドキュメント完成度

{{if .AdditionalData.approval_criteria}}承認基準:
{{.AdditionalData.approval_criteria}}{{end}}`

	if err := op.RegisterTemplate("review_request", reviewTemplate); err != nil {
		return err
	}

	return nil
}

// BuildManagerPrompt builds the main manager prompt
func (op *OrchestratorPrompts) BuildManagerPrompt(data OrchestratorData) (string, error) {
	return op.ExecuteTemplate("manager", data)
}

// BuildTaskAssignment builds a task assignment prompt
func (op *OrchestratorPrompts) BuildTaskAssignment(taskDesc, purpose string, data map[string]interface{}) (string, error) {
	templateData := TemplateData{
		TaskDesc: taskDesc,
		Context:  purpose,
		AdditionalData: data,
	}
	return op.ExecuteTemplate("task_assignment", templateData)
}

// BuildProgressCheck builds a progress check prompt
func (op *OrchestratorPrompts) BuildProgressCheck(taskDesc string, data map[string]interface{}) (string, error) {
	templateData := TemplateData{
		TaskDesc: taskDesc,
		AdditionalData: data,
	}
	return op.ExecuteTemplate("progress_check", templateData)
}

// BuildReviewRequest builds a review request prompt
func (op *OrchestratorPrompts) BuildReviewRequest(taskDesc string, data map[string]interface{}) (string, error) {
	templateData := TemplateData{
		TaskDesc: taskDesc,
		AdditionalData: data,
	}
	return op.ExecuteTemplate("review_request", templateData)
}

// BuildCustomPrompt builds a custom prompt with variables
func (op *OrchestratorPrompts) BuildCustomPrompt(templateName string, variables map[string]interface{}) (string, error) {
	return op.ExecuteTemplate(templateName, variables)
}

// GetAvailableTemplates returns list of available template names
func (op *OrchestratorPrompts) GetAvailableTemplates() []string {
	templates := []string{}
	for name := range op.templates {
		templates = append(templates, name)
	}
	return templates
}

// ValidatePromptVariables validates that required variables are present
func (op *OrchestratorPrompts) ValidatePromptVariables(templateName string, variables map[string]interface{}) error {
	requiredVars := map[string][]string{
		"manager": {"PaneID", "MainTask"},
		"task_assignment": {"TaskDesc", "Context"},
		"progress_check": {"TaskDesc"},
		"review_request": {"TaskDesc"},
	}
	
	required, exists := requiredVars[templateName]
	if !exists {
		return nil // No validation for unknown templates
	}
	
	missing := []string{}
	for _, reqVar := range required {
		if _, exists := variables[reqVar]; !exists {
			missing = append(missing, reqVar)
		}
	}
	
	if len(missing) > 0 {
		return fmt.Errorf("missing required variables for template %s: %s", templateName, strings.Join(missing, ", "))
	}
	
	return nil
}

// FormatPaneList formats a list of pane IDs for display
func (op *OrchestratorPrompts) FormatPaneList(panes []string, includeIndices bool) string {
	if len(panes) == 0 {
		return "なし"
	}
	
	if includeIndices {
		formatted := make([]string, len(panes))
		for i, pane := range panes {
			formatted[i] = fmt.Sprintf("%d. %s", i+1, pane)
		}
		return strings.Join(formatted, "\n")
	}
	
	return strings.Join(panes, ", ")
}