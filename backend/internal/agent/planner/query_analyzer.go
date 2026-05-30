package planner

import (
	"strings"
)

// QueryAnalyzer determines if a query needs Agent mode (multi-step reasoning)
type QueryAnalyzer struct{}

func NewQueryAnalyzer() *QueryAnalyzer {
	return &QueryAnalyzer{}
}

// Analyze checks if the query requires multi-step TCM reasoning.
// Returns true for complex queries involving diagnosis, treatment, and lifestyle advice.
func (a *QueryAnalyzer) Analyze(query string) (bool, string) {
	keywords := []string{
		"分析", "诊断", "辨证", "调理", "方案",
		"症状", "病因", "治疗", "推荐", "建议",
		"怎么", "如何", "为什么", "什么原因",
		"帮我", "给我", "查看", "判断",
		"综合", "整体", "全面",
	}
	count := 0
	for _, kw := range keywords {
		if strings.Contains(query, kw) {
			count++
		}
	}
	needsAgent := count >= 2 || strings.Contains(query, "帮我分析") || strings.Contains(query, "调理方案")
	reason := "简单问答，使用RAG模式"
	if needsAgent {
		reason = "需要多步推理，使用Agent模式"
	}
	return needsAgent, reason
}
