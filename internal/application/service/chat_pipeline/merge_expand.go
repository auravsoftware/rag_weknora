package chatpipeline

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/Tencent/WeKnora/internal/searchutil"
	"github.com/Tencent/WeKnora/internal/sourceloc"
	"github.com/Tencent/WeKnora/internal/types"
)

// isLikelyTruncated checks whether a chunk's content appears to be cut off
// mid-sentence, at a list boundary, or cross-page transition.
func isLikelyTruncated(content string) bool {
	trimmed := strings.TrimSpace(content)
	if len(trimmed) == 0 {
		return false
	}
	lastRune, _ := utf8.DecodeLastRuneInString(trimmed)
	if lastRune == ',' || lastRune == ';' || lastRune == ':' || lastRune == '-' || lastRune == '(' {
		return true
	}
	lines := strings.Split(trimmed, "\n")
	lastLine := strings.TrimSpace(lines[len(lines)-1])
	if len(lastLine) <= 4 && len(lastLine) > 0 {
		if lastLine[len(lastLine)-1] == '.' || lastLine[len(lastLine)-1] == ')' {
			return true
		}
	}
	if lastRune != '.' && lastRune != '?' && lastRune != '!' && lastRune != '"' && lastRune != '”' && lastRune != '}' && lastRune != ']' {
		return true
	}
	return false
}

// numberedItemKey identifies a numbered or lettered list item marker at the
// start of a line (e.g. "2.7." -> prefix="2.", num=7; "b)" -> prefix="alpha)", num='b').
type numberedItemKey struct {
	prefix string
	num    int
}

// stripLeadingListDecorators strips Markdown list bullets (-, *, +, •) and
// bold/italic wrappers (**, __, *, _) from the start of a trimmed line so OCR
// outputs like "**2.7.** Sĩ quan..." or "- 2.7. Sĩ quan..." are still parsed.
func stripLeadingListDecorators(t string) string {
	for {
		prev := t
		// Strip leading Markdown bullet + space.
		for _, bullet := range []string{"- ", "* ", "+ ", "• "} {
			if strings.HasPrefix(t, bullet) {
				t = strings.TrimSpace(t[len(bullet):])
			}
		}
		// Strip leading bold/italic markers.
		for _, wrap := range []string{"**", "__", "*", "_"} {
			if strings.HasPrefix(t, wrap) {
				t = strings.TrimSpace(t[len(wrap):])
			}
		}
		if t == prev {
			break
		}
	}
	return t
}

// isStandalonePageFooter reports whether a trimmed line is likely a PDF page
// number or horizontal rule (e.g. "1", "- 2 -", "Trang 1", "---").
func isStandalonePageFooter(t string) bool {
	if t == "" {
		return true
	}
	if strings.Trim(t, "-_=* ") == "" {
		return true
	}
	cleaned := strings.TrimSpace(strings.Trim(t, "-–— "))
	if _, ok := parsePositiveASCIIInt(cleaned); ok {
		return true
	}
	lower := strings.ToLower(cleaned)
	for _, prefix := range []string{"trang ", "page "} {
		if strings.HasPrefix(lower, prefix) {
			rest := strings.TrimSpace(lower[len(prefix):])
			if slash := strings.IndexByte(rest, '/'); slash > 0 {
				rest = strings.TrimSpace(rest[:slash])
			}
			if _, ok := parsePositiveASCIIInt(rest); ok {
				return true
			}
		}
	}
	return false
}

// parseLeadingNumberedItem parses a line-leading item marker such as "2.7. ...",
// "**2.7.** ...", "- 2.7. ...", "7. ...", "7) ...", or "b) ...".
func parseLeadingNumberedItem(line string) (numberedItemKey, bool) {
	t := stripLeadingListDecorators(strings.TrimSpace(line))
	if len(t) < 3 {
		return numberedItemKey{}, false
	}
	// Multi-level numeric: "2.7. ..." or "2.7 ..." or "7.1.2. ..."
	if t[0] >= '0' && t[0] <= '9' {
		i := 0
		for i < len(t) && ((t[i] >= '0' && t[i] <= '9') || t[i] == '.') {
			i++
		}
		// Skip optional closing bold/italic marker right after "2.7." (e.g. "**2.7.** text").
		sepIdx := i
		for sepIdx < len(t) && (t[sepIdx] == '*' || t[sepIdx] == '_') {
			sepIdx++
		}
		if i > 0 && sepIdx < len(t) && (t[sepIdx] == ' ' || t[sepIdx] == '\t' || t[sepIdx] == ')') {
			token := strings.TrimRight(t[:i], ".")
			if token == "" {
				return numberedItemKey{}, false
			}
			if dot := strings.LastIndexByte(token, '.'); dot >= 0 {
				n, ok := parsePositiveASCIIInt(token[dot+1:])
				if ok {
					return numberedItemKey{prefix: token[:dot+1], num: n}, true
				}
			} else if t[sepIdx] == ')' || (i > 0 && t[i-1] == '.') {
				n, ok := parsePositiveASCIIInt(token)
				if ok {
					sep := "."
					if t[sepIdx] == ')' {
						sep = ")"
					}
					return numberedItemKey{prefix: "num" + sep, num: n}, true
				}
			}
		}
	}
	// Single-letter item: "a) ..." or "a. ..."
	if ((t[0] >= 'a' && t[0] <= 'z') || (t[0] >= 'A' && t[0] <= 'Z')) &&
		(t[1] == ')' || t[1] == '.') {
		sepIdx := 2
		for sepIdx < len(t) && (t[sepIdx] == '*' || t[sepIdx] == '_') {
			sepIdx++
		}
		if sepIdx < len(t) && (t[sepIdx] == ' ' || t[sepIdx] == '\t') {
			ch := t[0]
			if ch >= 'A' && ch <= 'Z' {
				ch += 'a' - 'A'
			}
			return numberedItemKey{prefix: "alpha" + string(t[1]), num: int(ch)}, true
		}
	}
	return numberedItemKey{}, false
}

func parsePositiveASCIIInt(s string) (int, bool) {
	if s == "" || len(s) > 4 {
		return 0, false
	}
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, false
		}
		n = n*10 + int(s[i]-'0')
	}
	return n, n > 0
}

// lastNumberedListItem returns the last numbered/lettered list item key in content
// if the final non-empty paragraph of content belongs to a numbered list.
// Handles multi-line wrapped items (common in OCR/PDF text) and trailing page numbers.
func lastNumberedListItem(content string) (numberedItemKey, bool) {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	seenText := false
	checked := 0
	for i := len(lines) - 1; i >= 0; i-- {
		t := strings.TrimSpace(lines[i])
		if t == "" {
			if seenText {
				// Reached the start of the final paragraph without finding a list marker.
				break
			}
			continue
		}
		if !seenText && isStandalonePageFooter(t) {
			continue
		}
		seenText = true
		checked++
		if key, ok := parseLeadingNumberedItem(t); ok {
			return key, true
		}
		if checked >= 12 {
			break
		}
	}
	return numberedItemKey{}, false
}

// endsInNumberedList reports whether content ends on a numbered/lettered list item.
func endsInNumberedList(content string) bool {
	_, ok := lastNumberedListItem(content)
	return ok
}

// continuesNumberedList reports whether nextContent continues the numbered list
// that currentContent ended on (e.g. current ends with "2.7." and next contains
// "2.8." near its start, even if next begins with a ContextHeader and overlapped
// lines of "2.7.").
func continuesNumberedList(currentContent, nextContent string) bool {
	lastKey, ok := lastNumberedListItem(currentContent)
	if !ok {
		return false
	}
	checked := 0
	runesScanned := 0
	for _, line := range strings.Split(nextContent, "\n") {
		t := strings.TrimSpace(line)
		if t == "" || isStandalonePageFooter(t) {
			continue
		}
		// Skip chunker context header lines (e.g. "[Ngữ cảnh: ...]" or "[Context: ...]").
		if strings.HasPrefix(t, "[") && strings.HasSuffix(t, "]") {
			continue
		}
		checked++
		runesScanned += utf8.RuneCountInString(t)
		if key, kOK := parseLeadingNumberedItem(t); kOK {
			if key.prefix == lastKey.prefix && key.num == lastKey.num+1 {
				return true
			}
			// Allow skipping overlapped repeats of lastKey (or earlier items in the same list).
			if key.prefix == lastKey.prefix && key.num <= lastKey.num {
				continue
			}
			// Allow nested lettered sub-items (e.g. a), b)) inside a numeric item.
			if strings.HasPrefix(key.prefix, "alpha") && !strings.HasPrefix(lastKey.prefix, "alpha") {
				continue
			}
			// Hit a different numeric list/section before seeing lastKey.num+1 -> list ended.
			break
		}
		if checked >= 12 || runesScanned >= 800 {
			break
		}
	}
	return false
}

// expandShortContextWithNeighbors expands the short context with neighbors.
//
// Lookups are not tenant scoped: hits from an org-shared KB belong to the
// sharing workspace, and a caller-tenant filter silently skipped their
// expansion. The base IDs come from retrieval results and neighbors are only
// accepted from the same knowledge, the same rule parent resolution uses
// (#3342).
func (p *PluginMerge) expandShortContextWithNeighbors(
	ctx context.Context,
	results []*types.SearchResult,
) []*types.SearchResult {
	const (
		minLen = 1200
		maxLen = 6000
	)

	if len(results) == 0 || p.chunkRepo == nil {
		return results
	}

	type targetInfo struct {
		result *types.SearchResult
	}

	targets := make([]targetInfo, 0)
	baseIDsSet := make(map[string]struct{})

	for _, r := range results {
		if r == nil || r.ID == "" || r.Content == "" {
			continue
		}
		if r.ChunkType != string(types.ChunkTypeText) && r.ChunkType != string(types.ChunkTypeImageOCR) {
			continue
		}
		needsExpand := runeLen(r.Content) < minLen ||
			isLikelyTruncated(r.Content) ||
			endsInNumberedList(r.Content) ||
			r.ChunkType == string(types.ChunkTypeImageOCR)
		if !needsExpand || runeLen(r.Content) >= maxLen {
			continue
		}
		targets = append(targets, targetInfo{result: r})
		baseIDsSet[r.ID] = struct{}{}
		pipelineInfo(ctx, "Merge", "need_expand", map[string]interface{}{
			"chunk_id":   r.ID,
			"content":    r.Content,
			"chunk_type": r.ChunkType,
			"len":        runeLen(r.Content),
		})
	}

	if len(targets) == 0 {
		return results
	}

	baseIDs := make([]string, 0, len(baseIDsSet))
	for id := range baseIDsSet {
		baseIDs = append(baseIDs, id)
	}

	chunkMap := make(map[string]*types.Chunk, len(baseIDs))
	chunks, err := p.chunkRepo.ListChunksByIDOnly(ctx, baseIDs)
	if err != nil {
		pipelineWarn(ctx, "Merge", "expand_list_base_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return results
	}
	for _, chunk := range chunks {
		chunkMap[chunk.ID] = chunk
	}

	neighborIDsSet := make(map[string]struct{})
	for _, chunk := range chunkMap {
		if chunk == nil {
			continue
		}
		if chunk.PreChunkID != "" {
			if _, exists := chunkMap[chunk.PreChunkID]; !exists {
				neighborIDsSet[chunk.PreChunkID] = struct{}{}
			}
		}
		if chunk.NextChunkID != "" {
			if _, exists := chunkMap[chunk.NextChunkID]; !exists {
				neighborIDsSet[chunk.NextChunkID] = struct{}{}
			}
		}
	}

	if len(neighborIDsSet) > 0 {
		neighborIDs := make([]string, 0, len(neighborIDsSet))
		for id := range neighborIDsSet {
			neighborIDs = append(neighborIDs, id)
		}
		neighbors, err := p.chunkRepo.ListChunksByIDOnly(ctx, neighborIDs)
		if err != nil {
			pipelineWarn(ctx, "Merge", "expand_list_neighbor_failed", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			for _, chunk := range neighbors {
				chunkMap[chunk.ID] = chunk
				pipelineInfo(ctx, "Merge", "expand_list_neighbor_success", map[string]interface{}{
					"neighbor_chunk_id":   chunk.ID,
					"neighbor_content":    chunk.Content,
					"neighbor_chunk_type": chunk.ChunkType,
					"neighbor_len":        runeLen(chunk.Content),
				})
			}
		}
	}

	for _, target := range targets {
		res := target.result
		p.fetchChunksIfMissing(ctx, chunkMap, res.ID)
		baseChunk := chunkMap[res.ID]
		if baseChunk == nil || baseChunk.Content == "" || (baseChunk.ChunkType != types.ChunkTypeText && baseChunk.ChunkType != types.ChunkTypeImageOCR) {
			continue
		}

		// When resolveParentChunks (step 4) or sequential merge (step 5) already
		// expanded res.Content from fresh DB chunks, preserve that rewritten body
		// instead of reverting to the smaller child baseChunk.Content, and walk
		// the parent's PreChunkID/NextChunkID chain rather than sibling children
		// already contained inside the parent.
		baseContent := baseChunk.Content
		prevCursor := baseChunk.PreChunkID
		nextCursor := baseChunk.NextChunkID
		if res.ContentRewritten && res.Content != "" {
			baseContent = res.Content
			if baseChunk.ParentChunkID != "" {
				p.fetchChunksIfMissing(ctx, chunkMap, baseChunk.ParentChunkID)
				if parentChunk := chunkMap[baseChunk.ParentChunkID]; parentChunk != nil &&
					parentChunk.ChunkType == types.ChunkTypeParentText {
					prevCursor = parentChunk.PreChunkID
					nextCursor = parentChunk.NextChunkID
				}
			}
		} else if (prevCursor == "" || nextCursor == "") && baseChunk.ParentChunkID != "" {
			p.fetchChunksIfMissing(ctx, chunkMap, baseChunk.ParentChunkID)
			if parentChunk := chunkMap[baseChunk.ParentChunkID]; parentChunk != nil {
				if prevCursor == "" {
					prevCursor = parentChunk.PreChunkID
				}
				if nextCursor == "" {
					nextCursor = parentChunk.NextChunkID
				}
			}
		}

		// A long, non-truncated chunk that only qualified via endsInNumberedList
		// should expand forward ONLY when the next chunk continues that numbered list.
		listOnlyExpand := runeLen(baseContent) >= minLen &&
			!isLikelyTruncated(baseContent) &&
			res.ChunkType != string(types.ChunkTypeImageOCR)

		prevContent := ""
		nextContent := ""
		prevIDs := []string{}
		nextIDs := []string{}

		if listOnlyExpand {
			prevCursor = ""
		}

		p.fetchChunksIfMissing(ctx, chunkMap, prevCursor, nextCursor)

		if prevCursor != "" {
			if prevChunk := chunkMap[prevCursor]; prevChunk != nil && prevChunk.KnowledgeID == baseChunk.KnowledgeID {
				prevContent = prevChunk.Content
				prevIDs = append(prevIDs, prevChunk.ID)
				prevCursor = prevChunk.PreChunkID
			} else {
				prevCursor = ""
			}
		}

		if nextCursor != "" {
			if nextChunk := chunkMap[nextCursor]; nextChunk != nil && nextChunk.KnowledgeID == baseChunk.KnowledgeID {
				if !listOnlyExpand || continuesNumberedList(baseContent, nextChunk.Content) {
					nextContent = nextChunk.Content
					nextIDs = append(nextIDs, nextChunk.ID)
					nextCursor = nextChunk.NextChunkID
				} else {
					nextCursor = ""
				}
			} else {
				nextCursor = ""
			}
		}

		targetMinLen := minLen
		if isLikelyTruncated(baseContent) || res.ChunkType == string(types.ChunkTypeImageOCR) {
			targetMinLen = maxLen
		}

		var merged string
		for {
			merged = mergeOrderedContent(prevContent, baseContent, nextContent, maxLen)
			if merged == "" {
				break
			}
			if prevCursor == "" && nextCursor == "" {
				break
			}

			// Check if the next chunk continues an active numbered list (e.g. 2.7. -> 2.8.)
			// even after targetMinLen has already been satisfied.
			tailForList := baseContent
			if nextContent != "" {
				tailForList = nextContent
			}
			listContinues := false
			if nextCursor != "" && runeLen(merged) < maxLen && endsInNumberedList(tailForList) {
				p.fetchChunksIfMissing(ctx, chunkMap, nextCursor)
				if nextChunk := chunkMap[nextCursor]; nextChunk != nil &&
					nextChunk.KnowledgeID == baseChunk.KnowledgeID &&
					continuesNumberedList(tailForList, nextChunk.Content) {
					listContinues = true
				}
			}

			if runeLen(merged) >= targetMinLen && !listContinues {
				break
			}

			expanded := false
			if !listContinues && prevCursor != "" {
				p.fetchChunksIfMissing(ctx, chunkMap, prevCursor)
				if prevChunk := chunkMap[prevCursor]; prevChunk != nil &&
					prevChunk.KnowledgeID == baseChunk.KnowledgeID {
					prevContent = searchutil.JoinChunkContent(prevChunk.Content, prevContent, "\n\n")
					prevIDs = append([]string{prevChunk.ID}, prevIDs...)
					prevCursor = prevChunk.PreChunkID
					expanded = true
				} else {
					prevCursor = ""
				}
			}

			merged = mergeOrderedContent(prevContent, baseContent, nextContent, maxLen)
			if runeLen(merged) >= targetMinLen && !listContinues {
				break
			}

			if nextCursor != "" {
				p.fetchChunksIfMissing(ctx, chunkMap, nextCursor)
				if nextChunk := chunkMap[nextCursor]; nextChunk != nil &&
					nextChunk.KnowledgeID == baseChunk.KnowledgeID {
					nextContent = searchutil.JoinChunkContent(nextContent, nextChunk.Content, "\n\n")
					nextIDs = append(nextIDs, nextChunk.ID)
					nextCursor = nextChunk.NextChunkID
					expanded = true
				} else {
					nextCursor = ""
				}
			}

			if !expanded {
				break
			}
		}

		if merged == "" || merged == baseContent {
			continue
		}

		beforeLen := runeLen(res.Content)
		res.Content = merged
		res.ContentRewritten = true

		for _, id := range prevIDs {
			if id != "" && !containsID(res.SubChunkID, id) {
				res.SubChunkID = append(res.SubChunkID, id)
			}
			if c := chunkMap[id]; c != nil {
				res.SourceLocators = sourceloc.MergeLocators(c.SourceLocators, res.SourceLocators)
			}
		}
		for _, id := range nextIDs {
			if id != "" && !containsID(res.SubChunkID, id) {
				res.SubChunkID = append(res.SubChunkID, id)
			}
			if c := chunkMap[id]; c != nil {
				res.SourceLocators = sourceloc.MergeLocators(res.SourceLocators, c.SourceLocators)
			}
		}

		pipelineInfo(ctx, "Merge", "expand_short_chunk", map[string]interface{}{
			"chunk_id":       res.ID,
			"prev_ids":       prevIDs,
			"next_ids":       nextIDs,
			"before_len":     beforeLen,
			"after_len":      runeLen(res.Content),
			"base_content":   baseChunk.Content,
			"after_content":  res.Content,
			"chunk_type":     res.ChunkType,
			"remaining_prev": prevCursor,
			"remaining_next": nextCursor,
		})
	}

	return results
}

// runeLen returns the length of a string in runes
func runeLen(s string) int {
	return len([]rune(s))
}

// mergeOrderedContent joins prev, base and next in document order within
// roughly maxLen runes. base is the retrieval hit and is never cut; the budget
// left after it goes to the end of prev and the start of next, the text
// adjacent to the hit, with any share one side cannot use passed to the other.
// Truncating the joined string instead let a long previous neighbor push the
// hit out of the context entirely.
func mergeOrderedContent(prev, base, next string, maxLen int) string {
	prevRunes, nextRunes := []rune(prev), []rune(next)
	budget := max(maxLen-runeLen(base), 0)
	nextKeep := min(len(nextRunes), budget)
	prevKeep := min(len(prevRunes), budget-nextKeep)

	content := base
	if prevKeep > 0 {
		content = searchutil.JoinChunkContent(string(prevRunes[len(prevRunes)-prevKeep:]), content, "\n\n")
	}
	if nextKeep > 0 {
		content = searchutil.JoinChunkContent(content, string(nextRunes[:nextKeep]), "\n\n")
	}
	return content
}

func containsID(ids []string, target string) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}

func (p *PluginMerge) fetchChunksIfMissing(
	ctx context.Context,
	chunkMap map[string]*types.Chunk,
	chunkIDs ...string,
) {
	missing := make([]string, 0, len(chunkIDs))
	for _, id := range chunkIDs {
		if id == "" {
			continue
		}
		if _, exists := chunkMap[id]; !exists {
			missing = append(missing, id)
		}
	}
	if len(missing) == 0 {
		return
	}

	chunks, err := p.chunkRepo.ListChunksByIDOnly(ctx, missing)
	if err != nil {
		pipelineWarn(ctx, "Merge", "expand_fetch_missing_failed", map[string]interface{}{
			"missing_cnt": len(missing),
			"error":       err.Error(),
		})
	}

	found := make(map[string]struct{}, len(chunks))
	for _, chunk := range chunks {
		chunkMap[chunk.ID] = chunk
		found[chunk.ID] = struct{}{}
	}

	for _, id := range missing {
		if _, ok := found[id]; !ok {
			chunkMap[id] = nil
		}
	}
}
