package orchestrator

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

type ContextSummarizer struct {
	MinWords    int
	MaxWords    int
	Templates   map[string]string
	StopWords   []string
	stopWordSet map[string]bool
}

type SummaryOptions struct {
	WordLimit     int
	IncludeSteps  bool
	IncludeStatus bool
	Template      string
	Focus         []string
}

func NewContextSummarizer() *ContextSummarizer {
	stopWords := []string{
		"の", "を", "に", "が", "は", "で", "と", "から", "まで", "より", "へ",
		"こと", "もの", "ため", "など", "として", "について", "による", "において",
		"a", "an", "the", "and", "or", "but", "in", "on", "at", "to", "for",
		"of", "with", "by", "from", "as", "is", "are", "was", "were", "be",
	}

	stopWordSet := make(map[string]bool)
	for _, word := range stopWords {
		stopWordSet[word] = true
	}

	return &ContextSummarizer{
		MinWords:    10,
		MaxWords:    200,
		StopWords:   stopWords,
		stopWordSet: stopWordSet,
		Templates: map[string]string{
			"default":    "タスク「{title}」: {description}。現在のステータス: {status}。{details}",
			"detailed":   "【{title}】{description}。進捗: {status}。詳細: {details}。次のステップ: {next_steps}",
			"brief":      "{title}: {status}。{key_points}",
			"technical":  "タスク{id}: {title}。実装内容: {description}。状態: {status}。技術的詳細: {details}",
			"management": "プロジェクト要約 - {title}: {description}。進捗状況: {status}。重要事項: {key_points}",
		},
	}
}

func (cs *ContextSummarizer) SummarizeTask(task *TaskSummary, options *SummaryOptions) (string, error) {
	if task == nil {
		return "", fmt.Errorf("タスクがnilです")
	}

	if options == nil {
		options = &SummaryOptions{
			WordLimit:     cs.MaxWords,
			IncludeSteps:  true,
			IncludeStatus: true,
			Template:      "default",
		}
	}

	template := cs.getTemplate(options.Template)
	summary := cs.buildSummary(task, template, options)

	if options.WordLimit > 0 {
		summary = cs.limitWords(summary, options.WordLimit)
	}

	wordCount := countWords(summary)
	if wordCount < cs.MinWords {
		return "", fmt.Errorf("生成された要約が短すぎます (現在: %d語, 下限: %d語)", wordCount, cs.MinWords)
	}
	if wordCount > cs.MaxWords {
		summary = cs.limitWords(summary, cs.MaxWords)
	}

	return summary, nil
}

func (cs *ContextSummarizer) SummarizeMultipleTasks(tasks []*TaskSummary, options *SummaryOptions) (string, error) {
	if len(tasks) == 0 {
		return "", fmt.Errorf("タスクが提供されていません")
	}

	if options == nil {
		options = &SummaryOptions{
			WordLimit: cs.MaxWords,
			Template:  "management",
		}
	}

	var summaries []string
	completed := 0
	inProgress := 0
	pending := 0

	for _, task := range tasks {
		switch task.Status {
		case StatusCompleted:
			completed++
		case StatusInProgress:
			inProgress++
		case StatusPending:
			pending++
		}

		taskSummary, err := cs.SummarizeTask(task, &SummaryOptions{
			WordLimit: 50,
			Template:  "brief",
		})
		if err == nil {
			summaries = append(summaries, taskSummary)
		}
	}

	overallSummary := fmt.Sprintf("プロジェクト概要: 全%dタスク（完了: %d、進行中: %d、待機: %d）。",
		len(tasks), completed, inProgress, pending)

	if len(summaries) > 0 {
		overallSummary += " 主要タスク: " + strings.Join(summaries, "; ")
	}

	if options.WordLimit > 0 {
		overallSummary = cs.limitWords(overallSummary, options.WordLimit)
	}

	return overallSummary, nil
}

func (cs *ContextSummarizer) ExtractKeywords(text string, maxKeywords int) []string {
	if text == "" {
		return []string{}
	}

	words := cs.tokenize(text)
	wordFreq := make(map[string]int)

	for _, word := range words {
		cleanWord := strings.ToLower(strings.TrimSpace(word))
		if len(cleanWord) > 2 && !cs.stopWordSet[cleanWord] {
			wordFreq[cleanWord]++
		}
	}

	type wordCount struct {
		word  string
		count int
	}

	var wordCounts []wordCount
	for word, count := range wordFreq {
		wordCounts = append(wordCounts, wordCount{word, count})
	}

	for i := 0; i < len(wordCounts)-1; i++ {
		for j := i + 1; j < len(wordCounts); j++ {
			if wordCounts[i].count < wordCounts[j].count {
				wordCounts[i], wordCounts[j] = wordCounts[j], wordCounts[i]
			}
		}
	}

	var keywords []string
	limit := maxKeywords
	if len(wordCounts) < limit {
		limit = len(wordCounts)
	}

	for i := 0; i < limit; i++ {
		keywords = append(keywords, wordCounts[i].word)
	}

	return keywords
}

func (cs *ContextSummarizer) CompressSummary(summary string, targetWords int) (string, error) {
	if summary == "" {
		return "", fmt.Errorf("要約が空です")
	}

	currentWords := countWords(summary)
	if currentWords <= targetWords {
		return summary, nil
	}

	sentences := cs.splitSentences(summary)
	if len(sentences) <= 1 {
		return cs.limitWords(summary, targetWords), nil
	}

	keywords := cs.ExtractKeywords(summary, 10)
	keywordSet := make(map[string]bool)
	for _, keyword := range keywords {
		keywordSet[keyword] = true
	}

	type sentenceScore struct {
		sentence string
		score    int
		words    int
	}

	var scored []sentenceScore
	for _, sentence := range sentences {
		score := 0
		words := countWords(sentence)
		sentenceWords := cs.tokenize(sentence)

		for _, word := range sentenceWords {
			if keywordSet[strings.ToLower(word)] {
				score++
			}
		}

		scored = append(scored, sentenceScore{sentence, score, words})
	}

	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[i].score < scored[j].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	var result []string
	wordCount := 0
	for _, item := range scored {
		if wordCount+item.words <= targetWords {
			result = append(result, item.sentence)
			wordCount += item.words
		}
	}

	if len(result) == 0 {
		return cs.limitWords(summary, targetWords), nil
	}

	return strings.Join(result, " "), nil
}

func (cs *ContextSummarizer) getTemplate(templateName string) string {
	if template, exists := cs.Templates[templateName]; exists {
		return template
	}
	return cs.Templates["default"]
}

func (cs *ContextSummarizer) buildSummary(task *TaskSummary, template string, options *SummaryOptions) string {
	summary := template

	replacements := map[string]string{
		"{id}":          task.ID,
		"{title}":       task.Title,
		"{description}": task.Description,
		"{status}":      string(task.Status),
		"{priority}":    string(task.Priority),
		"{assigned_to}": task.AssignedTo,
	}

	if task.StepContext != nil {
		replacements["{step_number}"] = fmt.Sprintf("%d", task.StepContext.StepNumber)
		replacements["{dependencies}"] = strings.Join(task.StepContext.Dependencies, ", ")
		replacements["{outputs}"] = strings.Join(task.StepContext.Outputs, ", ")
	}

	details := []string{}
	if options.IncludeSteps && task.StepContext != nil {
		details = append(details, fmt.Sprintf("ステップ%d", task.StepContext.StepNumber))
	}
	if len(task.Tags) > 0 {
		details = append(details, "タグ: "+strings.Join(task.Tags, ", "))
	}
	if task.AssignedTo != "" {
		details = append(details, "担当: "+task.AssignedTo)
	}

	replacements["{details}"] = strings.Join(details, "。")
	replacements["{key_points}"] = cs.extractKeyPoints(task)
	replacements["{next_steps}"] = cs.generateNextSteps(task)

	for placeholder, value := range replacements {
		summary = strings.ReplaceAll(summary, placeholder, value)
	}

	return summary
}

func (cs *ContextSummarizer) extractKeyPoints(task *TaskSummary) string {
	points := []string{}

	if task.Priority == PriorityHigh || task.Priority == PriorityCritical {
		points = append(points, "高優先度")
	}

	if task.Status == StatusInProgress {
		points = append(points, "進行中")
	} else if task.Status == StatusCompleted {
		points = append(points, "完了済み")
	}

	if len(task.Tags) > 0 {
		points = append(points, "関連: "+strings.Join(task.Tags, ", "))
	}

	if len(points) == 0 {
		return "通常進行"
	}

	return strings.Join(points, ", ")
}

func (cs *ContextSummarizer) generateNextSteps(task *TaskSummary) string {
	switch task.Status {
	case StatusPending:
		return "開始待ち"
	case StatusInProgress:
		return "継続実行"
	case StatusCompleted:
		return "完了"
	case StatusFailed:
		return "再実行検討"
	default:
		return "状況確認"
	}
}

func (cs *ContextSummarizer) limitWords(text string, limit int) string {
	words := cs.tokenize(text)
	if len(words) <= limit {
		return text
	}

	limitedWords := words[:limit]
	result := strings.Join(limitedWords, "")

	if utf8.ValidString(result) {
		return result + "..."
	}

	return text[:min(len(text), limit*3)] + "..."
}

func (cs *ContextSummarizer) tokenize(text string) []string {
	re := regexp.MustCompile(`[\p{Han}\p{Hiragana}\p{Katakana}]|[a-zA-Z]+`)
	return re.FindAllString(text, -1)
}

func (cs *ContextSummarizer) splitSentences(text string) []string {
	re := regexp.MustCompile(`[。！？.!?]+`)
	sentences := re.Split(text, -1)

	var result []string
	for _, sentence := range sentences {
		trimmed := strings.TrimSpace(sentence)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}