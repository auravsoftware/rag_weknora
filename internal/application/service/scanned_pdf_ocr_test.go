package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/Tencent/WeKnora/internal/infrastructure/chunker"
	"github.com/Tencent/WeKnora/internal/models/vlm"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type scannedPDFTestModelService struct {
	interfaces.ModelService
	models map[string]vlm.VLM
}

func (s *scannedPDFTestModelService) GetVLMModel(_ context.Context, modelID string) (vlm.VLM, error) {
	if m, ok := s.models[modelID]; ok {
		return m, nil
	}
	return nil, errors.New("vlm model not found: " + modelID)
}

type scannedPDFStubVLM struct {
	id         string
	mu         sync.Mutex
	calls      int
	prompts    []string
	thinking   []*bool
	byImgBytes map[string]string
	errByImg   map[string]error
}

func (m *scannedPDFStubVLM) Predict(ctx context.Context, imgBytes [][]byte, prompt string) (string, error) {
	return m.PredictWithOptions(ctx, imgBytes, prompt, nil)
}

func (m *scannedPDFStubVLM) PredictWithOptions(
	_ context.Context,
	imgBytes [][]byte,
	prompt string,
	opts *vlm.PredictOptions,
) (string, error) {
	m.mu.Lock()
	m.calls++
	m.prompts = append(m.prompts, prompt)
	if opts != nil {
		m.thinking = append(m.thinking, opts.Thinking)
	} else {
		m.thinking = append(m.thinking, nil)
	}
	m.mu.Unlock()

	key := ""
	if len(imgBytes) > 0 {
		key = string(imgBytes[0])
	}
	if err, ok := m.errByImg[key]; ok {
		return "", err
	}
	return m.byImgBytes[key], nil
}

func (m *scannedPDFStubVLM) GetModelName() string { return m.id }
func (m *scannedPDFStubVLM) GetModelID() string   { return m.id }

func TestShouldUseScannedPDFOCRFirst(t *testing.T) {
	t.Run("pure scanned pdf qualifies", func(t *testing.T) {
		res := &types.ReadResult{
			MarkdownContent: "![page 1](images/page_1.png)\n\n![page 2](images/page_2.png)",
			ImageRefs: []types.ImageRef{
				{OriginalRef: "images/page_1.png", ImageData: []byte("img1")},
				{OriginalRef: "images/page_2.png", ImageData: []byte("img2")},
			},
			Metadata: map[string]string{"image_source_type": "scanned_pdf"},
		}
		assert.True(t, shouldUseScannedPDFOCRFirst(res))
	})

	t.Run("non-scanned pdf does not qualify", func(t *testing.T) {
		res := &types.ReadResult{
			MarkdownContent: "![fig 1](images/fig_1.png)",
			ImageRefs:       []types.ImageRef{{OriginalRef: "images/fig_1.png", ImageData: []byte("img1")}},
			Metadata:        map[string]string{"image_source_type": "pdf_text_layer"},
		}
		assert.False(t, shouldUseScannedPDFOCRFirst(res))
	})

	t.Run("mixed pdf with real text layer does not qualify", func(t *testing.T) {
		res := &types.ReadResult{
			MarkdownContent: strings.Repeat("Điều khoản hợp đồng tín dụng. ", 20) + "\n![page 2](images/page_2.png)",
			ImageRefs:       []types.ImageRef{{OriginalRef: "images/page_2.png", ImageData: []byte("img2")}},
			Metadata:        map[string]string{"image_source_type": "scanned_pdf"},
		}
		assert.False(t, shouldUseScannedPDFOCRFirst(res))
	})

	t.Run("mixed image_source_types does not qualify", func(t *testing.T) {
		res := &types.ReadResult{
			MarkdownContent: "![fig](images/fig.jpg)\n\n![page 2](images/page_2.jpg)",
			ImageRefs: []types.ImageRef{
				{OriginalRef: "images/fig.jpg", ImageData: []byte("fig")},
				{OriginalRef: "images/page_2.jpg", ImageData: []byte("p2")},
			},
			Metadata: map[string]string{
				"image_source_type":  "scanned_pdf",
				"image_source_types": `{"images/fig.jpg":"pdf_text_layer","images/page_2.jpg":"scanned_pdf"}`,
			},
		}
		assert.False(t, shouldUseScannedPDFOCRFirst(res))
	})

	t.Run("scanned pdf without inline bytes does not qualify", func(t *testing.T) {
		res := &types.ReadResult{
			MarkdownContent: "![page 1](images/page_1.png)",
			ImageRefs:       []types.ImageRef{{OriginalRef: "images/page_1.png"}},
			Metadata:        map[string]string{"image_source_type": "scanned_pdf"},
		}
		assert.False(t, shouldUseScannedPDFOCRFirst(res))
	})

	t.Run("explicit pdf_classification=scanned always qualifies", func(t *testing.T) {
		res := &types.ReadResult{
			MarkdownContent: strings.Repeat("cho IlQnh viQn tla khoa Phu'ong Ddng ", 20) + "\n![page 1](images/page_1.jpg)",
			ImageRefs:       []types.ImageRef{{OriginalRef: "images/page_1.jpg", ImageData: []byte("img1")}},
			Metadata: map[string]string{
				"image_source_type":  "scanned_pdf",
				"pdf_classification": "scanned",
			},
		}
		assert.True(t, shouldUseScannedPDFOCRFirst(res))
	})

	t.Run("explicit pdf_classification=vector does not qualify", func(t *testing.T) {
		res := &types.ReadResult{
			MarkdownContent: "![fig 1](images/fig_1.jpg)",
			ImageRefs:       []types.ImageRef{{OriginalRef: "images/fig_1.jpg", ImageData: []byte("img1")}},
			Metadata: map[string]string{
				"image_source_type":  "scanned_pdf",
				"pdf_classification": "vector",
			},
		}
		assert.False(t, shouldUseScannedPDFOCRFirst(res))
	})

	t.Run("majority scanned pages without pdf_classification qualifies", func(t *testing.T) {
		res := &types.ReadResult{
			MarkdownContent: strings.Repeat("garbled scanner text ", 20) + "\n![page 1](images/page_1.jpg)",
			ImageRefs:       []types.ImageRef{{OriginalRef: "images/page_1.jpg", ImageData: []byte("img1")}},
			Metadata: map[string]string{
				"image_source_type":  "scanned_pdf",
				"page_count":         "97",
				"scanned_page_count": "95",
			},
		}
		assert.True(t, shouldUseScannedPDFOCRFirst(res))
	})
}

func TestOrderedScannedPDFPageImages(t *testing.T) {
	res := &types.ReadResult{
		MarkdownContent: "![p1](images/p1.png)\n\n![p2](images/p2.png)\n\n![p3](images/p3.png)",
		ImageRefs: []types.ImageRef{
			{OriginalRef: "images/p3.png", ImageData: []byte("p3")},
			{OriginalRef: "images/p1.png", ImageData: []byte("p1")},
			{OriginalRef: "images/empty.png"},
			{OriginalRef: "images/p2.png", ImageData: []byte("p2")},
		},
	}

	pages := orderedScannedPDFPageImages(res)
	require.Len(t, pages, 3)
	assert.Equal(t, 1, pages[0].pageNum)
	assert.Equal(t, "images/p1.png", pages[0].ref.OriginalRef)
	assert.Equal(t, 2, pages[1].pageNum)
	assert.Equal(t, "images/p2.png", pages[1].ref.OriginalRef)
	assert.Equal(t, 3, pages[2].pageNum)
	assert.Equal(t, "images/p3.png", pages[2].ref.OriginalRef)
}

func TestAssembleScannedPDFMarkdownAndCrossPageChunking(t *testing.T) {
	page1 := "CỘNG HÒA XÃ HỘI CHỦ NGHĨA VIỆT NAM\nĐiều 1. Phạm vi điều chỉnh và đối tượng áp dụng."
	page2 := "Điều 2. Các điều kiện vay vốn bao gồm:\n1. Có năng lực pháp luật dân sự đầy đủ theo quy định."
	page3 := "2. Có hộ khẩu thường trú tại địa bàn.\n3. Thuộc diện hộ nghèo hoặc cận nghèo.\n" +
		"4. Có phương án sử dụng vốn khả thi.\n5. Được tổ tiết kiệm và vay vốn bình xét.\n" +
		"6. Có xác nhận của UBND cấp xã.\n7. Không có nợ quá hạn tại thời điểm vay."

	merged, blocks, okPages := assembleScannedPDFMarkdown([]string{page1, page2, page3}, "hash-pdf-1")
	require.Equal(t, 3, okPages)
	require.Len(t, blocks, 3)
	assert.Contains(t, merged, "<!-- Page 1 -->\n"+page1)
	assert.Contains(t, merged, "<!-- Page 2 -->\n"+page2)
	assert.Contains(t, merged, "<!-- Page 3 -->\n"+page3)

	normalized := chunker.NormalizeLineEndings(merged)
	srcIdx := buildSourceIndex(merged, blocks, normalized, nil)
	require.NotNil(t, srcIdx)

	cfg := chunker.NormalizeSplitterConfig(chunker.SplitterConfig{
		ChunkSize:    2000,
		ChunkOverlap: 200,
		Separators:   []string{"\n\n", "\n"},
	})
	parts := chunker.Split(normalized, cfg)
	// Total length is well below 2000 chars, so the entire clause spanning
	// Page 2 (item 1) and Page 3 (items 2-7) stays intact in a single chunk.
	require.Len(t, parts, 1)
	assert.Contains(t, parts[0].Content, "1. Có năng lực pháp luật dân sự")
	assert.Contains(t, parts[0].Content, "7. Không có nợ quá hạn")

	locs := srcIdx.Locators(parts[0].Start, parts[0].End)
	require.Len(t, locs, 3)
	assert.Equal(t, 1, locs[0].Page)
	assert.Equal(t, 2, locs[1].Page)
	assert.Equal(t, 3, locs[2].Page)
	assert.Equal(t, types.SourceLocatorPDF, locs[0].Type)
	assert.Equal(t, "hash-pdf-1", locs[0].SourceHash)
}

func TestOCRScannedPDFDocument_EndToEnd(t *testing.T) {
	ctx := context.Background()
	stubVLM := &scannedPDFStubVLM{
		id: "vlm-ocr-1",
		byImgBytes: map[string]string{
			"page-1-bytes": "Nội dung trang 1: Hợp đồng mẫu.",
			"page-2-bytes": "No text content.", // blank scanned page
			"page-3-bytes": "Nội dung trang 3: Điều khoản thanh toán.",
		},
	}
	svc := &knowledgeService{
		modelService: &scannedPDFTestModelService{
			models: map[string]vlm.VLM{"vlm-ocr-1": stubVLM},
		},
	}

	kb := &types.KnowledgeBase{
		ID: "kb-1",
		VLMConfig: types.VLMConfig{
			Enabled: true,
			ModelID: "vlm-ocr-1",
		},
	}
	knowledge := &types.Knowledge{
		ID:       "know-1",
		FileHash: "sha256-test",
	}
	eff := ResolveProcessConfig(kb, nil)
	res := &types.ReadResult{
		MarkdownContent: "![p1](images/p1.png)\n\n![p2](images/p2.png)\n\n![p3](images/p3.png)",
		ImageRefs: []types.ImageRef{
			{OriginalRef: "images/p1.png", ImageData: []byte("page-1-bytes")},
			{OriginalRef: "images/p2.png", ImageData: []byte("page-2-bytes")},
			{OriginalRef: "images/p3.png", ImageData: []byte("page-3-bytes")},
		},
		Metadata: map[string]string{
			"image_source_type": "scanned_pdf",
		},
	}

	applied, err := svc.ocrScannedPDFDocument(ctx, kb, knowledge, eff, res)
	require.NoError(t, err)
	require.True(t, applied)

	assert.Equal(t, 3, stubVLM.calls)
	for _, p := range stubVLM.prompts {
		assert.Equal(t, vlmOCRScannedPDFPrompt, p)
	}
	for _, th := range stubVLM.thinking {
		require.NotNil(t, th)
		assert.False(t, *th, "OCR calls should disable thinking by default")
	}

	// Blank page 2 is omitted while page 3 preserves its real page number (3).
	assert.Equal(t,
		"<!-- Page 1 -->\nNội dung trang 1: Hợp đồng mẫu.\n\n<!-- Page 3 -->\nNội dung trang 3: Điều khoản thanh toán.",
		res.MarkdownContent,
	)
	assert.Nil(t, res.ImageRefs, "ImageRefs must be cleared to prevent duplicate per-page multimodal OCR")
	assert.Equal(t, "ocr_first", res.Metadata["scanned_pdf_ocr"])
	assert.Equal(t, "2", res.Metadata["scanned_pdf_ocr_pages"])
	assert.Equal(t, "3", res.Metadata["scanned_pdf_total_pages"])

	require.Len(t, res.SourceBlocks, 2)
	assert.Equal(t, 1, res.SourceBlocks[0].Locator.Page)
	assert.Equal(t, 3, res.SourceBlocks[1].Locator.Page)
}

func TestOCRScannedPDFDocument_FallbackWhenVLMDisabled(t *testing.T) {
	ctx := context.Background()
	svc := &knowledgeService{}
	kb := &types.KnowledgeBase{ID: "kb-no-vlm"}
	knowledge := &types.Knowledge{ID: "know-no-vlm"}
	eff := ResolveProcessConfig(kb, nil)
	res := &types.ReadResult{
		MarkdownContent: "![p1](images/p1.png)",
		ImageRefs: []types.ImageRef{
			{OriginalRef: "images/p1.png", ImageData: []byte("page-1-bytes")},
		},
		Metadata: map[string]string{"image_source_type": "scanned_pdf"},
	}

	applied, err := svc.ocrScannedPDFDocument(ctx, kb, knowledge, eff, res)
	require.NoError(t, err)
	assert.False(t, applied)
	assert.Len(t, res.ImageRefs, 1, "legacy fallback must preserve ImageRefs")
}
