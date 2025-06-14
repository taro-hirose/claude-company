package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

// PaneType はペインの種類を表す
type PaneType int

const (
	PaneTypeConsole PaneType = iota // ユーザー操作用ターミナル
	PaneTypeManager                 // タスク分解・管理用Claude
	PaneTypeWorker                  // 実装作業用Claude
	PaneTypeUnknown
)

// String はPaneTypeの文字列表現を返す
func (pt PaneType) String() string {
	switch pt {
	case PaneTypeConsole:
		return "Console"
	case PaneTypeManager:
		return "Manager"
	case PaneTypeWorker:
		return "Worker"
	default:
		return "Unknown"
	}
}

// PaneFilter は統一ペインフィルター
type PaneFilter struct {
	// レガシーサポート用のマップ（段階的廃止予定）
	legacyParentPanes map[string]bool
}

// NewPaneFilter は新しいPaneFilterインスタンスを作成
func NewPaneFilter() *PaneFilter {
	return &PaneFilter{
		legacyParentPanes: make(map[string]bool),
	}
}

// NewPaneFilterWithLegacySupport はレガシーサポート付きのPaneFilterを作成
func NewPaneFilterWithLegacySupport(legacyParentPanes map[string]bool) *PaneFilter {
	if legacyParentPanes == nil {
		legacyParentPanes = make(map[string]bool)
	}
	return &PaneFilter{
		legacyParentPanes: legacyParentPanes,
	}
}

// GetPaneTitle はペインのタイトルを取得
func (pf *PaneFilter) GetPaneTitle(paneID string) (string, error) {
	cmd := exec.Command("tmux", "display-message", "-t", paneID, "-p", "#{pane_title}")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get pane title for %s: %v", paneID, err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetPaneType はペインIDからペインタイプを判定
func (pf *PaneFilter) GetPaneType(paneID string) PaneType {
	// まずレガシーマップをチェック（段階的廃止予定）
	if pf.legacyParentPanes[paneID] {
		return PaneTypeManager
	}

	// ペインタイトルから判定
	title, err := pf.GetPaneTitle(paneID)
	if err != nil {
		return PaneTypeUnknown
	}

	switch {
	case strings.Contains(title, "[CONSOLE]"):
		return PaneTypeConsole
	case strings.Contains(title, "[MANAGER]"):
		return PaneTypeManager
	case strings.Contains(title, "[WORKER]"):
		return PaneTypeWorker
	default:
		return PaneTypeUnknown
	}
}

// IsConsolePane はコンソールペインかどうかを判定
func (pf *PaneFilter) IsConsolePane(paneID string) bool {
	return pf.GetPaneType(paneID) == PaneTypeConsole
}

// IsManagerPane は管理ペインかどうかを判定
func (pf *PaneFilter) IsManagerPane(paneID string) bool {
	return pf.GetPaneType(paneID) == PaneTypeManager
}

// IsWorkerPane はワーカーペインかどうかを判定
func (pf *PaneFilter) IsWorkerPane(paneID string) bool {
	return pf.GetPaneType(paneID) == PaneTypeWorker
}

// IsParentPane は親ペイン（ConsoleまたはManager）かどうかを判定
func (pf *PaneFilter) IsParentPane(paneID string) bool {
	paneType := pf.GetPaneType(paneID)
	return paneType == PaneTypeConsole || paneType == PaneTypeManager
}

// IsChildPane は子ペイン（Worker）かどうかを判定
func (pf *PaneFilter) IsChildPane(paneID string) bool {
	return pf.GetPaneType(paneID) == PaneTypeWorker
}

// GetAllPanes は全てのペインIDを取得
func (pf *PaneFilter) GetAllPanes() ([]string, error) {
	cmd := exec.Command("tmux", "list-panes", "-F", "#{pane_id}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get panes: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var panes []string
	for _, line := range lines {
		if strings.HasPrefix(line, "%") {
			panes = append(panes, line)
		}
	}
	return panes, nil
}

// FilterPanesByType は指定されたタイプのペインのみを返す
func (pf *PaneFilter) FilterPanesByType(paneType PaneType) ([]string, error) {
	allPanes, err := pf.GetAllPanes()
	if err != nil {
		return nil, err
	}

	var filteredPanes []string
	for _, pane := range allPanes {
		if pf.GetPaneType(pane) == paneType {
			filteredPanes = append(filteredPanes, pane)
		}
	}
	return filteredPanes, nil
}

// GetManagerPanes は管理ペインのリストを取得
func (pf *PaneFilter) GetManagerPanes() ([]string, error) {
	return pf.FilterPanesByType(PaneTypeManager)
}

// GetWorkerPanes はワーカーペインのリストを取得
func (pf *PaneFilter) GetWorkerPanes() ([]string, error) {
	return pf.FilterPanesByType(PaneTypeWorker)
}

// GetConsolePanes はコンソールペインのリストを取得
func (pf *PaneFilter) GetConsolePanes() ([]string, error) {
	return pf.FilterPanesByType(PaneTypeConsole)
}

// TaskType はタスクの種類を表す
type TaskType int

const (
	TaskTypeImplementation TaskType = iota // 実装作業（Worker向け）
	TaskTypeManagement                     // 管理作業（Manager向け）
	TaskTypeReview                         // レビュー作業（Manager向け）
	TaskTypeUnknown
)

// ClassifyTask はタスク内容からタスクタイプを分類
func (pf *PaneFilter) ClassifyTask(taskDescription string) TaskType {
	taskLower := strings.ToLower(taskDescription)

	// 実装関連キーワード
	implementationKeywords := []string{
		"実装", "コード", "プログラム", "ファイル作成", "関数", "メソッド",
		"implement", "code", "create file", "function", "method",
		"write", "develop", "build", "test", "debug",
	}

	// 管理関連キーワード
	managementKeywords := []string{
		"分析", "計画", "設計", "分解", "割り当て", "監視",
		"analyze", "plan", "design", "breakdown", "assign", "monitor",
		"manage", "coordinate", "organize",
	}

	// レビュー関連キーワード
	reviewKeywords := []string{
		"レビュー", "確認", "検証", "品質", "チェック",
		"review", "verify", "check", "quality", "validate",
		"audit", "inspect",
	}

	// 実装タスクの判定
	for _, keyword := range implementationKeywords {
		if strings.Contains(taskLower, keyword) {
			return TaskTypeImplementation
		}
	}

	// レビュータスクの判定
	for _, keyword := range reviewKeywords {
		if strings.Contains(taskLower, keyword) {
			return TaskTypeReview
		}
	}

	// 管理タスクの判定
	for _, keyword := range managementKeywords {
		if strings.Contains(taskLower, keyword) {
			return TaskTypeManagement
		}
	}

	return TaskTypeUnknown
}

// ValidateTaskAssignment はタスクとペインの組み合わせが適切かを検証
func (pf *PaneFilter) ValidateTaskAssignment(taskDescription, paneID string) (bool, string, error) {
	taskType := pf.ClassifyTask(taskDescription)
	paneType := pf.GetPaneType(paneID)

	switch taskType {
	case TaskTypeImplementation:
		if paneType == PaneTypeWorker {
			return true, "Implementation task correctly assigned to worker pane", nil
		}
		return false, fmt.Sprintf("Implementation task should be assigned to worker pane, not %s pane", paneType.String()), nil

	case TaskTypeManagement:
		if paneType == PaneTypeManager {
			return true, "Management task correctly assigned to manager pane", nil
		}
		return false, fmt.Sprintf("Management task should be assigned to manager pane, not %s pane", paneType.String()), nil

	case TaskTypeReview:
		if paneType == PaneTypeManager {
			return true, "Review task correctly assigned to manager pane", nil
		}
		return false, fmt.Sprintf("Review task should be assigned to manager pane, not %s pane", paneType.String()), nil

	default:
		return true, "Unknown task type, allowing assignment", nil
	}
}

// GetBestPaneForTask はタスクに最適なペインを取得
func (pf *PaneFilter) GetBestPaneForTask(taskDescription string) (string, error) {
	taskType := pf.ClassifyTask(taskDescription)

	switch taskType {
	case TaskTypeImplementation:
		workerPanes, err := pf.GetWorkerPanes()
		if err != nil {
			return "", fmt.Errorf("failed to get worker panes: %v", err)
		}
		if len(workerPanes) == 0 {
			return "", fmt.Errorf("no worker panes available for implementation task")
		}
		return workerPanes[0], nil

	case TaskTypeManagement, TaskTypeReview:
		managerPanes, err := pf.GetManagerPanes()
		if err != nil {
			return "", fmt.Errorf("failed to get manager panes: %v", err)
		}
		if len(managerPanes) == 0 {
			return "", fmt.Errorf("no manager panes available for management/review task")
		}
		return managerPanes[0], nil

	default:
		// 不明なタスクタイプの場合、利用可能な任意のペインを返す
		allPanes, err := pf.GetAllPanes()
		if err != nil {
			return "", fmt.Errorf("failed to get panes: %v", err)
		}
		if len(allPanes) == 0 {
			return "", fmt.Errorf("no panes available")
		}
		return allPanes[0], nil
	}
}

// GetPaneStatistics はペイン統計情報を取得
func (pf *PaneFilter) GetPaneStatistics() (map[string]interface{}, error) {
	allPanes, err := pf.GetAllPanes()
	if err != nil {
		return nil, fmt.Errorf("failed to get all panes: %v", err)
	}

	managerPanes, _ := pf.GetManagerPanes()
	workerPanes, _ := pf.GetWorkerPanes()
	consolePanes, _ := pf.GetConsolePanes()

	stats := map[string]interface{}{
		"total_panes":   len(allPanes),
		"manager_panes": len(managerPanes),
		"worker_panes":  len(workerPanes),
		"console_panes": len(consolePanes),
		"all_panes":     allPanes,
		"manager_list":  managerPanes,
		"worker_list":   workerPanes,
		"console_list":  consolePanes,
	}

	return stats, nil
}