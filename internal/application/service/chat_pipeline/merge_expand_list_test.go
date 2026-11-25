package chatpipeline

import (
	"context"
	"strings"
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
)

func TestContinuesNumberedList(t *testing.T) {
	t.Run("2.7 followed by overlapped 2.7 and 2.8", func(t *testing.T) {
		cur := "2.6. Công nhân, người lao động.\n\n2.7. Sĩ quan, quân nhân chuyên nghiệp đang công tác."
		next := "2.7. Sĩ quan, quân nhân chuyên nghiệp đang công tác.\n\n2.8. Cán bộ, công chức, viên chức."
		if !continuesNumberedList(cur, next) {
			t.Fatalf("expected continuesNumberedList to return true for 2.7 -> 2.8")
		}
	})

	t.Run("2.7 directly followed by 2.8", func(t *testing.T) {
		cur := "2.7. Sĩ quan, quân nhân chuyên nghiệp đang công tác."
		next := "2.8. Cán bộ, công chức, viên chức."
		if !continuesNumberedList(cur, next) {
			t.Fatalf("expected continuesNumberedList to return true for direct 2.7 -> 2.8")
		}
	})

	t.Run("multi-line wrapped 2.7 with page number followed by ContextHeader and overlap before 2.8", func(t *testing.T) {
		cur := "2.7. Sĩ quan, quân nhân chuyên nghiệp, hạ sĩ quan thuộc lực lượng vũ\n" +
			"trang nhân dân, công nhân công an, công chức, công nhân và viên chức quốc\n" +
			"phòng đang phục vụ tại ngũ; người làm công tác cơ yếu, người làm công tác\n" +
			"khác trong tổ chức cơ yếu hưởng lương từ ngân sách nhà nước đang công tác.\n1"
		next := "[Ngữ cảnh: 2. Đối tượng vay vốn]\n" +
			"trang nhân dân, công nhân công an, công chức, công nhân và viên chức quốc\n" +
			"phòng đang phục vụ tại ngũ; người làm công tác cơ yếu, người làm công tác\n" +
			"khác trong tổ chức cơ yếu hưởng lương từ ngân sách nhà nước đang công tác.\n\n" +
			"2.8. Cán bộ, công chức, viên chức theo quy định của pháp luật về cán bộ,\ncông chức, viên chức."
		if !continuesNumberedList(cur, next) {
			t.Fatalf("expected continuesNumberedList to return true for multi-line wrapped 2.7 -> 2.8")
		}
	})

	t.Run("markdown bold and bullet wrapped 2.7 -> 2.8", func(t *testing.T) {
		cur := "- **2.7.** Sĩ quan, quân nhân chuyên nghiệp đang công tác."
		next := "**2.8.** Cán bộ, công chức, viên chức."
		if !continuesNumberedList(cur, next) {
			t.Fatalf("expected continuesNumberedList to return true for markdown-decorated 2.7 -> 2.8")
		}
	})

	t.Run("2.8 followed by new section 3.1 does not continue", func(t *testing.T) {
		cur := "2.8. Cán bộ, công chức, viên chức."
		next := "Nguyên tắc vay vốn\n3.1. Khách hàng phải đúng đối tượng."
		if continuesNumberedList(cur, next) {
			t.Fatalf("expected continuesNumberedList to return false for 2.8 -> 3.1")
		}
	})
}

func TestExpandShortContextWithNeighbors_PullsNextNumberedListItemEvenWhenLongAndPeriodEnded(t *testing.T) {
	// Chunk 1 is > 1200 chars and ends with a period on item 2.7.
	// Without list-continuation expansion, merge_expand would stop and omit 2.8.
	longBody := "Đối tượng vay vốn\n" +
		strings.Repeat("2.1. Nội dung chi tiết về đối tượng vay vốn nhà ở xã hội theo quy định pháp luật hiện hành. ", 14) +
		"\n2.7. Sĩ quan, quân nhân chuyên nghiệp, hạ sĩ quan thuộc lực lượng vũ trang nhân dân đang công tác."
	chunk2Content := "2.8. Cán bộ, công chức, viên chức theo quy định của pháp luật về cán bộ, công chức, viên chức."

	repo := &expandChunkRepo{
		chunks: map[string]*types.Chunk{
			"c1": {
				ID:          "c1",
				KnowledgeID: "k1",
				ChunkType:   types.ChunkTypeText,
				Content:     longBody,
				NextChunkID: "c2",
			},
			"c2": {
				ID:          "c2",
				KnowledgeID: "k1",
				ChunkType:   types.ChunkTypeText,
				Content:     chunk2Content,
				PreChunkID:  "c1",
			},
		},
	}

	plugin := &PluginMerge{chunkRepo: repo}
	results := []*types.SearchResult{
		{
			ID:          "c1",
			KnowledgeID: "k1",
			ChunkType:   string(types.ChunkTypeText),
			Content:     longBody,
		},
	}

	expanded := plugin.expandShortContextWithNeighbors(context.Background(), results)
	if len(expanded) != 1 {
		t.Fatalf("expected 1 result, got %d", len(expanded))
	}
	if !strings.Contains(expanded[0].Content, "2.8. Cán bộ, công chức, viên chức") {
		t.Fatalf("expected expanded content to include 2.8 from next chunk, got tail: %q",
			expanded[0].Content[max(0, len(expanded[0].Content)-120):])
	}
}

func TestExpandShortContextWithNeighbors_ParentResolvedListContinuation(t *testing.T) {
	// Simulates parent-child mode after resolveParentChunks:
	// SearchResult.ID is child "c1" (which only has 2.1..2.4), while SearchResult.Content
	// has been rewritten to parent "p1" (1. Đối tượng áp dụng ... 2.7 wrapped across lines).
	// Next parent "p2" starts with 2.8.
	child1Content := "2. Đối tượng vay vốn\n2.1. Người có công với cách mạng.\n2.2. Hộ gia đình nghèo."
	child2Content := "2.5. Người thu nhập thấp.\n2.7. Sĩ quan, quân nhân chuyên nghiệp."
	parent1Content := "1. Đối tượng áp dụng\n" +
		strings.Repeat("Nội dung hướng dẫn cho vay nhà ở xã hội theo quy định pháp luật hiện hành. ", 16) +
		"\n2. Đối tượng vay vốn\n" +
		"2.1. Người có công với cách mạng.\n" +
		"2.7. Sĩ quan, quân nhân chuyên nghiệp, hạ sĩ quan thuộc lực lượng vũ\n" +
		"trang nhân dân, công nhân công an, công chức, công nhân và viên chức quốc\n" +
		"phòng đang phục vụ tại ngũ; người làm công tác cơ yếu đang công tác."
	parent2Content := "2.8. Cán bộ, công chức, viên chức theo quy định của pháp luật về cán bộ, công chức, viên chức.\n" +
		"Các đối tượng được vay vốn nêu trên sau đây gọi chung là khách hàng."

	repo := &expandChunkRepo{
		chunks: map[string]*types.Chunk{
			"p1": {
				ID:          "p1",
				KnowledgeID: "k1",
				ChunkType:   types.ChunkTypeParentText,
				Content:     parent1Content,
				NextChunkID: "p2",
			},
			"p2": {
				ID:          "p2",
				KnowledgeID: "k1",
				ChunkType:   types.ChunkTypeParentText,
				Content:     parent2Content,
				PreChunkID:  "p1",
			},
			"c1": {
				ID:            "c1",
				KnowledgeID:   "k1",
				ChunkType:     types.ChunkTypeText,
				ParentChunkID: "p1",
				Content:       child1Content,
				NextChunkID:   "c2",
			},
			"c2": {
				ID:            "c2",
				KnowledgeID:   "k1",
				ChunkType:     types.ChunkTypeText,
				ParentChunkID: "p1",
				Content:       child2Content,
				PreChunkID:    "c1",
			},
		},
	}

	plugin := &PluginMerge{chunkRepo: repo}
	results := []*types.SearchResult{
		{
			ID:               "c1",
			KnowledgeID:      "k1",
			ChunkType:        string(types.ChunkTypeText),
			ParentChunkID:    "p1",
			Content:          parent1Content,
			ContentRewritten: true,
		},
	}

	expanded := plugin.expandShortContextWithNeighbors(context.Background(), results)
	if len(expanded) != 1 {
		t.Fatalf("expected 1 result, got %d", len(expanded))
	}
	if !strings.Contains(expanded[0].Content, "1. Đối tượng áp dụng") {
		t.Fatalf("expected parent1 prefix to be preserved, got: %q", expanded[0].Content[:min(120, len(expanded[0].Content))])
	}
	if !strings.Contains(expanded[0].Content, "2.8. Cán bộ, công chức, viên chức") {
		t.Fatalf("expected expanded parent content to include 2.8 from parent2, got tail: %q",
			expanded[0].Content[max(0, len(expanded[0].Content)-120):])
	}
	if !strings.Contains(expanded[0].Content, "Các đối tượng được vay vốn nêu trên") {
		t.Fatalf("expected trailing summary paragraph from parent2 to be included")
	}
}
