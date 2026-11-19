package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/Tencent/WeKnora/internal/logger"
	"github.com/Tencent/WeKnora/internal/models/vlm"
	"github.com/Tencent/WeKnora/internal/types"
)

const (
	// defaultScannedPDFOCRConcurrency bounds concurrent VLM OCR requests during
	// scanned-PDF ingestion so multi-page documents finish quickly without
	// overwhelming the VLM backend. Override via WEKNORA_SCANNED_PDF_OCR_CONCURRENCY.
	defaultScannedPDFOCRConcurrency = 4
)

func scannedPDFOCRConcurrency() int {
	return envPositiveInt("WEKNORA_SCANNED_PDF_OCR_CONCURRENCY", defaultScannedPDFOCRConcurrency)
}

type scannedPDFPage struct {
	pageNum     int
	ref         types.ImageRef
	pos         int
	originalIdx int
}

// shouldUseScannedPDFOCRFirst reports whether convertResult is a scanned PDF
// whose page images should be OCR'd up front and merged into a single Markdown
// document before chunking (OCR-First & Whole-Doc Chunking).
//
// Decision rules (vector-verified vs default-scanned):
//  1. If DocReader explicitly verified the PDF as non-vector
//     (metadata["pdf_classification"] == "scanned"), always run OCR-First
//     whenever inline page image bytes are present.
//  2. If DocReader explicitly verified the PDF as a genuine vector PDF
//     (metadata["pdf_classification"] == "vector"), keep its native text layer.
//  3. For older parsers without pdf_classification: if image_source_type is
//     "scanned_pdf" and either the majority of pages are scanned
//     (scanned_page_count/page_count >= 50%) or the extracted non-image text is
//     below 200 runes with no native pdf_text_layer images, run OCR-First.
func shouldUseScannedPDFOCRFirst(result *types.ReadResult) bool {
	if result == nil || result.Metadata == nil || len(result.ImageRefs) == 0 {
		return false
	}

	hasInlineImage := false
	for _, ref := range result.ImageRefs {
		if len(ref.ImageData) > 0 {
			hasInlineImage = true
			break
		}
	}
	if !hasInlineImage {
		return false
	}

	switch strings.ToLower(strings.TrimSpace(result.Metadata["pdf_classification"])) {
	case "scanned":
		return true
	case "vector":
		return false
	}

	if result.Metadata["image_source_type"] != "scanned_pdf" {
		return false
	}

	// Backward compatibility for parsers that emit scanned_page_count /
	// page_count without pdf_classification: when scanned pages are not a
	// small minority, treat the document as scanned rather than falling back
	// to per-image chunking.
	if scannedPages, err1 := strconv.Atoi(strings.TrimSpace(result.Metadata["scanned_page_count"])); err1 == nil && scannedPages > 0 {
		if totalPages, err2 := strconv.Atoi(strings.TrimSpace(result.Metadata["page_count"])); err2 == nil && totalPages > 0 {
			if scannedPages*100 >= totalPages*50 {
				return true
			}
		}
	}

	if approxTextContentRunes(result.MarkdownContent) >= temporaryDocumentLowTextRunes {
		return false
	}

	var imageSourceTypes map[string]string
	if raw := result.Metadata["image_source_types"]; raw != "" {
		if err := json.Unmarshal([]byte(raw), &imageSourceTypes); err != nil {
			imageSourceTypes = nil
		}
	}

	for _, ref := range result.ImageRefs {
		if len(ref.ImageData) == 0 {
			continue
		}
		if len(imageSourceTypes) > 0 {
			if src := imageSourceTypes[ref.OriginalRef]; src != "" && src != "scanned_pdf" {
				return false
			}
		}
	}
	return true
}

// orderedScannedPDFPageImages extracts page images with inline bytes in their
// document reading order (matching their appearance in MarkdownContent, falling
// back to their original slice order).
func orderedScannedPDFPageImages(result *types.ReadResult) []scannedPDFPage {
	if result == nil || len(result.ImageRefs) == 0 {
		return nil
	}
	pages := make([]scannedPDFPage, 0, len(result.ImageRefs))
	for idx, ref := range result.ImageRefs {
		if len(ref.ImageData) == 0 {
			continue
		}
		pos := -1
		if ref.OriginalRef != "" {
			pos = strings.Index(result.MarkdownContent, ref.OriginalRef)
		}
		pages = append(pages, scannedPDFPage{
			ref:         ref,
			pos:         pos,
			originalIdx: idx,
		})
	}
	sort.SliceStable(pages, func(i, j int) bool {
		pi, pj := pages[i].pos, pages[j].pos
		if pi >= 0 && pj >= 0 && pi != pj {
			return pi < pj
		}
		if (pi >= 0) != (pj >= 0) {
			return pi >= 0
		}
		return pages[i].originalIdx < pages[j].originalIdx
	})
	for i := range pages {
		pages[i].pageNum = i + 1
	}
	return pages
}

// assembleScannedPDFMarkdown joins per-page OCR text into a single continuous
// Markdown document with `<!-- Page N -->` markers and builds exact rune-range
// PDF SourceBlocks for each page so downstream chunking preserves accurate
// single-page and cross-page citations.
func assembleScannedPDFMarkdown(ocrTexts []string, sourceHash string) (string, []types.SourceBlock, int) {
	var (
		sb      strings.Builder
		blocks  []types.SourceBlock
		runePos int
		okPages int
	)
	for i, raw := range ocrTexts {
		text := strings.TrimSpace(raw)
		if text == "" {
			continue
		}
		pageNum := i + 1
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
			runePos += 2
		}
		pageStart := runePos
		marker := fmt.Sprintf("<!-- Page %d -->\n", pageNum)
		sb.WriteString(marker)
		sb.WriteString(text)
		runePos += utf8.RuneCountInString(marker) + utf8.RuneCountInString(text)
		blocks = append(blocks, types.SourceBlock{
			Start: pageStart,
			End:   runePos,
			Locator: types.SourceLocator{
				Type:       types.SourceLocatorPDF,
				Page:       pageNum,
				Mapping:    "exact",
				SourceHash: sourceHash,
			},
		})
		okPages++
	}
	return sb.String(), blocks, okPages
}

// resolveVLMForKnowledge resolves the VLM model configured for a knowledge
// ingestion run, honoring per-document overrides and falling back to the KB's
// VLM model on shared-KB resolution failures.
func (s *knowledgeService) resolveVLMForKnowledge(
	ctx context.Context,
	kb *types.KnowledgeBase,
	eff types.EffectiveProcessConfig,
) (vlm.VLM, types.VLMConfig, error) {
	vlmCfg := eff.VLMConfig
	if !vlmCfg.IsEnabled() {
		return nil, types.VLMConfig{}, fmt.Errorf("VLM is not enabled")
	}
	if vlmCfg.ModelID != "" {
		if s.modelService == nil {
			return nil, vlmCfg, fmt.Errorf("model service is not configured")
		}
		model, err := s.modelService.GetVLMModel(ctx, vlmCfg.ModelID)
		if err == nil {
			return model, vlmCfg, nil
		}
		if kb != nil && vlmCfg.ModelID != kb.VLMConfig.ModelID &&
			kb.VLMConfig.ModelID != "" && kb.VLMConfig.IsEnabled() {
			logger.Warnf(ctx,
				"[ScannedPDFOCR] Failed to resolve document VLM model %s: %v; falling back to KB VLM model %s",
				vlmCfg.ModelID, err, kb.VLMConfig.ModelID)
			fallbackModel, fallbackErr := s.modelService.GetVLMModel(ctx, kb.VLMConfig.ModelID)
			if fallbackErr != nil {
				return nil, vlmCfg, errors.Join(
					fmt.Errorf("resolve document VLM model %s: %w", vlmCfg.ModelID, err),
					fmt.Errorf("resolve KB fallback VLM model %s: %w", kb.VLMConfig.ModelID, fallbackErr),
				)
			}
			return fallbackModel, kb.VLMConfig, nil
		}
		return nil, vlmCfg, err
	}
	model, err := vlm.NewVLMFromLegacyConfig(vlmCfg, nil)
	return model, vlmCfg, err
}

// ocrScannedPDFDocument runs OCR-First & Whole-Doc assembly for scanned PDFs:
// it OCRs all page images with bounded concurrency using the configured VLM,
// stitches the extracted text into one continuous Markdown document with
// `<!-- Page N -->` markers and matching PDF SourceBlocks, and clears ImageRefs
// so the downstream pipeline chunks and embeds the unified text directly
// without enqueueing duplicate per-page multimodal OCR tasks.
//
// Returns (true, nil) when OCR-First replaced convertResult.MarkdownContent,
// (false, nil) to fall back safely to the legacy per-image multimodal path, or
// (false, err) when the context was canceled/expired mid-OCR.
func (s *knowledgeService) ocrScannedPDFDocument(
	ctx context.Context,
	kb *types.KnowledgeBase,
	knowledge *types.Knowledge,
	eff types.EffectiveProcessConfig,
	result *types.ReadResult,
) (bool, error) {
	if !shouldUseScannedPDFOCRFirst(result) {
		if result != nil && result.Metadata != nil && result.Metadata["image_source_type"] == "scanned_pdf" {
			logger.Infof(ctx,
				"[ScannedPDFOCR] Skipped OCR-First for knowledge %s: pdf_classification=%q image_refs=%d text_runes=%d",
				knowledge.ID,
				result.Metadata["pdf_classification"],
				len(result.ImageRefs),
				approxTextContentRunes(result.MarkdownContent),
			)
		}
		return false, nil
	}

	vlmModel, vlmCfg, err := s.resolveVLMForKnowledge(ctx, kb, eff)
	if err != nil {
		logger.Warnf(ctx,
			"[ScannedPDFOCR] Skipping OCR-First for knowledge %s (VLM unavailable: %v); falling back to legacy pipeline",
			knowledge.ID, err)
		return false, nil
	}

	pages := orderedScannedPDFPageImages(result)
	if len(pages) == 0 {
		return false, nil
	}

	started := time.Now()
	concurrency := scannedPDFOCRConcurrency()
	if concurrency > len(pages) {
		concurrency = len(pages)
	}
	logger.Infof(ctx,
		"[ScannedPDFOCR] Starting OCR-First for knowledge %s: pages=%d concurrency=%d vlm=%s",
		knowledge.ID, len(pages), concurrency, vlmCfg.ModelID)

	s.beginStage(ctx, knowledge.ID, types.StageMultimodal, types.JSONMap{
		"image_count":  len(pages),
		"mode":         "scanned_pdf_ocr_first",
		"vlm_model_id": vlmCfg.ModelID,
	})

	prompt := buildVLMOCRPrompt("scanned_pdf", vlmCfg)
	ocrTexts := make([]string, len(pages))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i, page := range pages {
		if ctx.Err() != nil {
			break
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, p scannedPDFPage) {
			defer wg.Done()
			defer func() { <-sem }()
			if ctx.Err() != nil {
				return
			}
			thinkingOff := false
			ocrCtx := types.WithLLMCallMetadata(ctx, "image_ocr", "")
			raw, ocrErr := vlmModel.PredictWithOptions(
				ocrCtx,
				[][]byte{p.ref.ImageData},
				prompt,
				&vlm.PredictOptions{Thinking: &thinkingOff},
			)
			if ocrErr != nil {
				logger.Warnf(ctx,
					"[ScannedPDFOCR] VLM OCR failed on page %d/%d for knowledge %s: %v",
					p.pageNum, len(pages), knowledge.ID, ocrErr)
				return
			}
			cleaned, valErr := validateOCRText(raw)
			if valErr != nil {
				logger.Warnf(ctx,
					"[ScannedPDFOCR] VLM OCR output rejected on page %d/%d for knowledge %s: %v",
					p.pageNum, len(pages), knowledge.ID, valErr)
				return
			}
			ocrTexts[idx] = cleaned
		}(i, page)
	}
	wg.Wait()

	if err := ctx.Err(); err != nil {
		logger.Warnf(ctx, "[ScannedPDFOCR] Interrupted for knowledge %s: %v", knowledge.ID, err)
		return false, err
	}

	mergedMarkdown, pageBlocks, okPages := assembleScannedPDFMarkdown(ocrTexts, knowledge.FileHash)
	if okPages == 0 || strings.TrimSpace(mergedMarkdown) == "" {
		logger.Warnf(ctx,
			"[ScannedPDFOCR] Produced no text across %d pages for knowledge %s; falling back to legacy pipeline",
			len(pages), knowledge.ID)
		return false, nil
	}

	result.MarkdownContent = mergedMarkdown
	result.SourceBlocks = pageBlocks
	// Clear raw page image refs now that all scanned pages have been converted
	// into unified Markdown text, preventing duplicate image storage and
	// redundant per-page multimodal OCR tasks.
	result.ImageRefs = nil
	if result.Metadata == nil {
		result.Metadata = make(map[string]string)
	}
	result.Metadata["scanned_pdf_ocr"] = "ocr_first"
	result.Metadata["scanned_pdf_ocr_pages"] = strconv.Itoa(okPages)
	result.Metadata["scanned_pdf_total_pages"] = strconv.Itoa(len(pages))

	s.endStage(ctx, knowledge.ID, types.StageMultimodal, types.JSONMap{
		"mode":         "scanned_pdf_ocr_first",
		"pages_total":  len(pages),
		"pages_ocr_ok": okPages,
		"text_length":  len(mergedMarkdown),
	})

	logger.Infof(ctx,
		"[ScannedPDFOCR] Completed OCR-First for knowledge %s in %s: pages_ok=%d/%d markdown_bytes=%d",
		knowledge.ID, time.Since(started), okPages, len(pages), len(mergedMarkdown))
	return true, nil
}
