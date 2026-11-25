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
