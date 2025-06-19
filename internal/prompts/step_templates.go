package prompts

import (
	"fmt"
	"strings"
)

// StepTemplates manages templates for step execution prompts
type StepTemplates struct {
	*TemplateManager
}

// NewStepTemplates creates a new step template manager
func NewStepTemplates() *StepTemplates {
	st := &StepTemplates{
		TemplateManager: NewTemplateManager(),
	}
	
	if err := st.registerTemplates(); err != nil {
		panic(fmt.Sprintf("Failed to register step templates: %v", err))
	}
	
	return st
}

// StepData represents data for step execution prompts
type StepData struct {
	StepName        string
	StepDescription string
	Purpose         string
	Deliverables    []string
	CompletionCriteria []string
	ReportPane      string
	ReportMessage   string
	Dependencies    []string
	Context         string
	Priority        string
	Deadline        string
	Resources       []string
}

func (st *StepTemplates) registerTemplates() error {
	// Basic step execution template
	stepTemplate := `サブタスク: {{.StepName}}
目的: {{.Purpose}}
成果物: 
{{range .Deliverables}}- {{.}}
{{end}}
完了条件: 
{{range .CompletionCriteria}}- {{.}}
{{end}}
{{if .Dependencies}}依存関係:
{{range .Dependencies}}- {{.}}
{{end}}{{end}}
{{if .Context}}追加コンテキスト:
{{.Context}}
{{end}}
{{if .Priority}}優先度: {{.Priority}}{{end}}
{{if .Deadline}}期限: {{.Deadline}}{{end}}
{{if .Resources}}リソース:
{{range .Resources}}- {{.}}
{{end}}{{end}}
報告方法: tmux send-keys -t {{.ReportPane}} '{{.ReportMessage}}' Enter; sleep 1; tmux send-keys -t {{.ReportPane}} '' Enter`

	if err := st.RegisterTemplate("step_execution", stepTemplate); err != nil {
		return err
	}

	// Code implementation step template
	codeTemplate := `サブタスク: {{.StepName}}
目的: {{.Purpose}}

## 実装要件
{{range .Deliverables}}- {{.}}
{{end}}

## 実装基準
{{range .CompletionCriteria}}- {{.}}
{{end}}

{{if .Dependencies}}## 依存関係
{{range .Dependencies}}- {{.}}
{{end}}{{end}}

{{if .Context}}## 技術的コンテキスト
{{.Context}}
{{end}}

## 実装ガイドライン
- 既存のコード規約に従う
- 適切なエラーハンドリングを実装
- 必要に応じてテストを作成
- コードの可読性を重視

{{if .Resources}}## 参考リソース
{{range .Resources}}- {{.}}
{{end}}{{end}}

報告方法: tmux send-keys -t {{.ReportPane}} '実装完了: {{.StepName}} - {{.ReportMessage}}' Enter; sleep 1; tmux send-keys -t {{.ReportPane}} '' Enter`

	if err := st.RegisterTemplate("code_implementation", codeTemplate); err != nil {
		return err
	}

	// Testing step template
	testTemplate := `サブタスク: {{.StepName}}
目的: {{.Purpose}}

## テスト対象
{{range .Deliverables}}- {{.}}
{{end}}

## テスト完了基準
{{range .CompletionCriteria}}- {{.}}
{{end}}

{{if .Dependencies}}## 前提条件
{{range .Dependencies}}- {{.}}
{{end}}{{end}}

## テスト方針
- 単体テストの実装
- 統合テストの実行
- エラーケースの検証
- パフォーマンステストの実施

{{if .Context}}## テストコンテキスト
{{.Context}}
{{end}}

報告方法: tmux send-keys -t {{.ReportPane}} 'テスト完了: {{.StepName}} - {{.ReportMessage}}' Enter; sleep 1; tmux send-keys -t {{.ReportPane}} '' Enter`

	if err := st.RegisterTemplate("testing", testTemplate); err != nil {
		return err
	}

	// Documentation step template
	docTemplate := `サブタスク: {{.StepName}}
目的: {{.Purpose}}

## ドキュメント成果物
{{range .Deliverables}}- {{.}}
{{end}}

## ドキュメント要件
{{range .CompletionCriteria}}- {{.}}
{{end}}

{{if .Dependencies}}## 対象機能・コンポーネント
{{range .Dependencies}}- {{.}}
{{end}}{{end}}

## ドキュメント方針
- 利用者の視点で記述
- 実例とサンプルコードを含める
- メンテナンス性を考慮
- 適切な構造化

{{if .Context}}## ドキュメントコンテキスト
{{.Context}}
{{end}}

報告方法: tmux send-keys -t {{.ReportPane}} 'ドキュメント完了: {{.StepName}} - {{.ReportMessage}}' Enter; sleep 1; tmux send-keys -t {{.ReportPane}} '' Enter`

	if err := st.RegisterTemplate("documentation", docTemplate); err != nil {
		return err
	}

	// Research step template
	researchTemplate := `サブタスク: {{.StepName}}
目的: {{.Purpose}}

## 調査範囲
{{range .Deliverables}}- {{.}}
{{end}}

## 調査完了基準
{{range .CompletionCriteria}}- {{.}}
{{end}}

{{if .Dependencies}}## 調査観点
{{range .Dependencies}}- {{.}}
{{end}}{{end}}

## 調査方針
- 既存コードベースの分析
- 関連技術の調査
- ベストプラクティスの特定
- 実装可能性の評価

{{if .Context}}## 調査コンテキスト
{{.Context}}
{{end}}

{{if .Resources}}## 調査リソース
{{range .Resources}}- {{.}}
{{end}}{{end}}

報告方法: tmux send-keys -t {{.ReportPane}} '調査完了: {{.StepName}} - {{.ReportMessage}}' Enter; sleep 1; tmux send-keys -t {{.ReportPane}} '' Enter`

	if err := st.RegisterTemplate("research", researchTemplate); err != nil {
		return err
	}

	// Review step template
	reviewTemplate := `サブタスク: {{.StepName}}
目的: {{.Purpose}}

## レビュー対象
{{range .Deliverables}}- {{.}}
{{end}}

## レビュー基準
{{range .CompletionCriteria}}- {{.}}
{{end}}

{{if .Dependencies}}## レビュー観点
{{range .Dependencies}}- {{.}}
{{end}}{{end}}

## レビュー方針
- コード品質の確認
- 仕様適合性の検証
- セキュリティ観点の確認
- パフォーマンス影響の評価

{{if .Context}}## レビューコンテキスト
{{.Context}}
{{end}}

報告方法: tmux send-keys -t {{.ReportPane}} 'レビュー完了: {{.StepName}} - {{.ReportMessage}}' Enter; sleep 1; tmux send-keys -t {{.ReportPane}} '' Enter`

	if err := st.RegisterTemplate("review", reviewTemplate); err != nil {
		return err
	}

	return nil
}

// BuildStepPrompt builds a step execution prompt
func (st *StepTemplates) BuildStepPrompt(templateType string, data StepData) (string, error) {
	return st.ExecuteTemplate(templateType, data)
}

// BuildCodeImplementationStep builds a code implementation step prompt
func (st *StepTemplates) BuildCodeImplementationStep(stepName, purpose string, deliverables, criteria []string, reportPane, reportMessage string) (string, error) {
	data := StepData{
		StepName:           stepName,
		Purpose:            purpose,
		Deliverables:       deliverables,
		CompletionCriteria: criteria,
		ReportPane:         reportPane,
		ReportMessage:      reportMessage,
	}
	return st.BuildStepPrompt("code_implementation", data)
}

// BuildTestingStep builds a testing step prompt
func (st *StepTemplates) BuildTestingStep(stepName, purpose string, deliverables, criteria []string, reportPane, reportMessage string) (string, error) {
	data := StepData{
		StepName:           stepName,
		Purpose:            purpose,
		Deliverables:       deliverables,
		CompletionCriteria: criteria,
		ReportPane:         reportPane,
		ReportMessage:      reportMessage,
	}
	return st.BuildStepPrompt("testing", data)
}

// BuildDocumentationStep builds a documentation step prompt
func (st *StepTemplates) BuildDocumentationStep(stepName, purpose string, deliverables, criteria []string, reportPane, reportMessage string) (string, error) {
	data := StepData{
		StepName:           stepName,
		Purpose:            purpose,
		Deliverables:       deliverables,
		CompletionCriteria: criteria,
		ReportPane:         reportPane,
		ReportMessage:      reportMessage,
	}
	return st.BuildStepPrompt("documentation", data)
}

// BuildResearchStep builds a research step prompt
func (st *StepTemplates) BuildResearchStep(stepName, purpose string, deliverables, criteria []string, reportPane, reportMessage string) (string, error) {
	data := StepData{
		StepName:           stepName,
		Purpose:            purpose,
		Deliverables:       deliverables,
		CompletionCriteria: criteria,
		ReportPane:         reportPane,
		ReportMessage:      reportMessage,
	}
	return st.BuildStepPrompt("research", data)
}

// BuildReviewStep builds a review step prompt
func (st *StepTemplates) BuildReviewStep(stepName, purpose string, deliverables, criteria []string, reportPane, reportMessage string) (string, error) {
	data := StepData{
		StepName:           stepName,
		Purpose:            purpose,
		Deliverables:       deliverables,
		CompletionCriteria: criteria,
		ReportPane:         reportPane,
		ReportMessage:      reportMessage,
	}
	return st.BuildStepPrompt("review", data)
}

// BuildCustomStep builds a custom step prompt with full control
func (st *StepTemplates) BuildCustomStep(templateType string, data StepData) (string, error) {
	return st.BuildStepPrompt(templateType, data)
}

// GetAvailableStepTemplates returns list of available step template names
func (st *StepTemplates) GetAvailableStepTemplates() []string {
	return []string{
		"step_execution",
		"code_implementation", 
		"testing",
		"documentation",
		"research",
		"review",
	}
}

// ValidateStepData validates step data for required fields
func (st *StepTemplates) ValidateStepData(data StepData) error {
	if data.StepName == "" {
		return fmt.Errorf("step name is required")
	}
	
	if data.Purpose == "" {
		return fmt.Errorf("purpose is required")
	}
	
	if len(data.Deliverables) == 0 {
		return fmt.Errorf("at least one deliverable is required")
	}
	
	if len(data.CompletionCriteria) == 0 {
		return fmt.Errorf("at least one completion criterion is required")
	}
	
	if data.ReportPane == "" {
		return fmt.Errorf("report pane is required")
	}
	
	return nil
}

// FormatDeliverables formats deliverables list for display
func (st *StepTemplates) FormatDeliverables(deliverables []string) string {
	if len(deliverables) == 0 {
		return "なし"
	}
	
	formatted := make([]string, len(deliverables))
	for i, deliverable := range deliverables {
		formatted[i] = fmt.Sprintf("- %s", deliverable)
	}
	
	return strings.Join(formatted, "\n")
}

// FormatCriteria formats completion criteria for display
func (st *StepTemplates) FormatCriteria(criteria []string) string {
	if len(criteria) == 0 {
		return "なし"
	}
	
	formatted := make([]string, len(criteria))
	for i, criterion := range criteria {
		formatted[i] = fmt.Sprintf("- %s", criterion)
	}
	
	return strings.Join(formatted, "\n")
}