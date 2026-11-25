// Package chunker - heuristic_splitter.go implements Tier 2: boundary-driven
// chunking for documents that lack proper Markdown headings but contain
// recognizable structural cues (page breaks, numbered sections, multilingual
// chapter markers, visual separators, all-caps section titles, page footers).
//
// The algorithm finds all candidate boundary positions, then performs greedy
// bin-packing — accumulating blocks between boundaries into chunks until the
// next block would exceed cfg.ChunkSize. Blocks larger than ChunkSize are
// recursively delegated to the legacy splitter for inner segmentation.
package chunker

import (
	"sort"
	"strings"
	"unicode/utf8"
)

func init() {
	splitByHeuristics = splitByHeuristicsImpl
}

// boundary marks a candidate split point in the document.
type boundary struct {
	runeStart int // rune offset where the next chunk should start
	priority  int
	level     int    // 1 = chapter/party/all-caps heading, 2 = short numbered sub-heading, 0 = other
	heading   string // trimmed heading text when level > 0
}

// heuristicHeadingEvent records a heading line's rune offset and hierarchy level
// for attaching ContextHeader breadcrumbs to downstream chunks.
type heuristicHeadingEvent struct {
	runeStart int
	level     int
	heading   string
}

// splitByHeuristicsImpl is the Tier-2 implementation. Falls through to the
// legacy splitter when no heuristic boundaries are found.
//
// profile is currently unused (this tier scans for boundaries directly) but
// is accepted to keep the splitByHeadings / splitByHeuristics signatures
// uniform — see strategy.runTier.
func splitByHeuristicsImpl(text string, cfg SplitterConfig, _ *DocProfile) []Chunk {
	if text == "" {
		return nil
	}
	runes := []rune(text)
	totalRunes := len(runes)
	if totalRunes <= cfg.ChunkSize {
		return SplitText(text, cfg)
	}

	bounds := findHeuristicBoundaries(text, cfg.Languages)
	// Drop any boundary that falls strictly inside a protected region (table,
	// fenced code block, LaTeX block, etc.) — splitting there would cut
	// through atomic content. Boundaries on a span edge are kept since they
	// align with the protected region start/end.
	if prot := protectedSpansRune(text, protectedSpans(text)); len(prot) > 0 {
		bounds = dropBoundsInsideSpans(bounds, prot)
	}
	if len(bounds) == 0 {
		return SplitText(text, cfg)
	}

	// Capture heading events before collapsing heading-only blocks so every
	// sub-chunk still resolves its exact chapter + section breadcrumb.
	headingEvents := extractHeadingEvents(bounds)

	// Append a sentinel at end-of-document so the bin-packer can flush.
	bounds = append(bounds, boundary{runeStart: totalRunes})
	// Always start with a boundary at offset 0 if not already there.
	if bounds[0].runeStart != 0 {
		bounds = append([]boundary{{runeStart: 0}}, bounds...)
	}

	// Collapse heading-only blocks (e.g. "ĐIỀU 4: GIAO NHẬN HÀNG HÓA\n\n"
	// immediately followed by "4.1 ...") so a heading never becomes an orphan
	// tail on the previous chunk or a standalone 1-line chunk.
	bounds = mergeHeadingOnlyBoundaries(bounds, runes)

	// Greedy bin-packing.
	var out []Chunk
	seq := 0
	chunkStart := bounds[0].runeStart
	curEnd := chunkStart
	minChunkSize := cfg.ChunkSize / 4
	if minChunkSize < 50 {
		minChunkSize = 50
	}

	for i := 1; i < len(bounds); i++ {
		nextEnd := bounds[i].runeStart
		blockLen := nextEnd - curEnd
		blockStartsMajorHeading := bounds[i-1].level == 1

		// Flush before starting a new major chapter/party heading when the
		// current chunk already has meaningful content, so major sections
		// don't bleed into the tail of the previous section.
		if blockStartsMajorHeading && curEnd > chunkStart && curEnd-chunkStart >= minChunkSize {
			out = appendChunk(out, runes, chunkStart, curEnd, &seq)
			chunkStart = curEnd
		}

		if blockLen > cfg.ChunkSize {
			// The block between the previous and this boundary is itself too
			// large to fit in any chunk. Flush current accumulation, then
			// recursively chunk the oversize block via the legacy splitter.
			if curEnd-chunkStart > 0 {
				out = appendChunk(out, runes, chunkStart, curEnd, &seq)
				chunkStart = curEnd
			}
			out = appendOversizeBlock(out, runes, curEnd, nextEnd, cfg, &seq)
			curEnd = nextEnd
			chunkStart = nextEnd
			continue
		}

		// Would adding this block exceed the budget?
		accumulated := nextEnd - chunkStart
		if accumulated > cfg.ChunkSize && curEnd-chunkStart >= minChunkSize {
			// Flush accumulated content as a chunk, restart at curEnd.
			out = appendChunk(out, runes, chunkStart, curEnd, &seq)
			// At a major chapter/party boundary, start cleanly at the heading
			// without pulling trailing overlap from the previous chapter.
			if blockStartsMajorHeading {
				chunkStart = curEnd
			} else {
				// Snap overlap start to the nearest semantic boundary, line,
				// sentence, or word break instead of slicing mid-word.
				chunkStart = applyOverlapAligned(runes, curEnd, cfg.ChunkOverlap, bounds)
			}
		}
		curEnd = nextEnd
	}

	// Flush remaining content.
	if curEnd > chunkStart {
		out = appendChunk(out, runes, chunkStart, curEnd, &seq)
	}

	// Attach hierarchical ContextHeader breadcrumbs from active chapter/section headings.
	attachHeuristicContextHeaders(out, headingEvents)

	// Merge tiny non-heading tail residues (e.g. a single trailing 50-char sentence
	// split off an oversize clause) back into the preceding chunk when safe.
	out = coalesceTinyHeuristicChunks(out, cfg.ChunkSize)

	return out
}

// extractHeadingEvents collects all heading boundaries for ContextHeader lookup.
func extractHeadingEvents(bounds []boundary) []heuristicHeadingEvent {
	var events []heuristicHeadingEvent
	for _, b := range bounds {
		if b.level > 0 && b.heading != "" {
			events = append(events, heuristicHeadingEvent{
				runeStart: b.runeStart,
				level:     b.level,
				heading:   b.heading,
			})
		}
	}
	return events
}

// isHeadingOnlySlice reports whether runes[start:end] contains only short
// heading lines (no body paragraphs). Used to prevent orphan headings like
// "ĐIỀU 4: GIAO NHẬN HÀNG HÓA" or stacked "Mục II\nVAY VỐN...\nĐiều kiện..."
// from being split off from the clauses that follow them.
func isHeadingOnlySlice(runes []rune, start, end int) bool {
	if end <= start || start < 0 || end > len(runes) {
		return false
	}
	trimmed := strings.TrimSpace(string(runes[start:end]))
	if trimmed == "" || utf8.RuneCountInString(trimmed) > 250 {
		return false
	}
	for _, line := range strings.Split(trimmed, "\n") {
		lt := strings.TrimSpace(line)
		if lt == "" {
			continue
		}
		// Any line that looks like a full prose sentence or numbered clause body
		// disqualifies the slice from being heading-only.
		if utf8.RuneCountInString(lt) > 120 || strings.HasSuffix(lt, ".") {
			return false
		}
	}
	return true
}

// mergeHeadingOnlyBoundaries drops the split boundary immediately after a
// heading-only block so the heading stays attached to the first clause/paragraph
// that follows it. bounds must start at 0 and end with the sentinel at len(runes).
func mergeHeadingOnlyBoundaries(bounds []boundary, runes []rune) []boundary {
	if len(bounds) <= 2 {
		return bounds
	}
	merged := make([]boundary, 0, len(bounds))
	merged = append(merged, bounds[0])
	i := 1
	for i < len(bounds) {
		prev := merged[len(merged)-1]
		// If the span [prev.runeStart, bounds[i].runeStart) contains only heading
		// lines (including stacked headings like "Mục II" + ALL CAPS title + sub-title)
		// and bounds[i] is not the end-of-document sentinel, absorb bounds[i] so the
		// heading block stays attached to the content that follows.
		if prev.level > 0 && i < len(bounds)-1 &&
			isHeadingOnlySlice(runes, prev.runeStart, bounds[i].runeStart) {
			i++
			continue
		}
		merged = append(merged, bounds[i])
		i++
	}
	return merged
}

// attachHeuristicContextHeaders populates Chunk.ContextHeader with the active
// chapter/section breadcrumb at each chunk's start offset when that heading is
// not already present at the top of the chunk's Content.
func attachHeuristicContextHeaders(chunks []Chunk, events []heuristicHeadingEvent) {
	if len(chunks) == 0 || len(events) == 0 {
		return
	}
	for i := range chunks {
		start := chunks[i].Start
		var lvl1, lvl2 string
		for _, ev := range events {
			if ev.runeStart > start {
				break
			}
			if ev.level == 1 {
				lvl1 = ev.heading
				lvl2 = ""
			} else if ev.level == 2 {
				lvl2 = ev.heading
			}
		}
		var parts []string
		trimmedContent := strings.TrimSpace(chunks[i].Content)
		if lvl1 != "" && !strings.HasPrefix(trimmedContent, lvl1) {
			parts = append(parts, lvl1)
		}
		if lvl2 != "" && !strings.HasPrefix(trimmedContent, lvl2) && !strings.Contains(trimmedContent, lvl2) {
			parts = append(parts, lvl2)
		}
		if len(parts) > 0 {
			chunks[i].ContextHeader = strings.Join(parts, "\n")
		}
	}
}

// coalesceTinyHeuristicChunks merges adjacent tiny residue chunks (< 80 runes)
// into the preceding chunk when they are contiguous (cur.End == next.Start),
// share the same ContextHeader, and fit within a 1.25x soft cap.
func coalesceTinyHeuristicChunks(in []Chunk, chunkSize int) []Chunk {
	if len(in) <= 1 || chunkSize <= 0 {
		return in
	}
	const tinyThreshold = 80
	softMax := chunkSize + chunkSize/4

	out := make([]Chunk, 0, len(in))
	cur := in[0]
	for i := 1; i < len(in); i++ {
		next := in[i]
		nextTrimLen := utf8.RuneCountInString(strings.TrimSpace(next.Content))
		curLen := utf8.RuneCountInString(cur.Content)
		nextLen := utf8.RuneCountInString(next.Content)
		if nextTrimLen > 0 && nextTrimLen < tinyThreshold &&
			cur.End == next.Start &&
			cur.ContextHeader == next.ContextHeader &&
			curLen+nextLen <= softMax {
			cur.Content += next.Content
			cur.End = next.End
			continue
		}
		out = append(out, cur)
		cur = next
	}
	out = append(out, cur)

	for i := range out {
		out[i].Seq = i
	}
	return out
}

// isShortSubHeadingLine reports whether a numbered section line looks like a
// short sub-heading (e.g. "7.1.2. Thanh toán đợt 1:", "4.1 Thời gian giao hàng:")
// rather than a full multi-sentence paragraph.
func isShortSubHeadingLine(trimmed string) bool {
	rLen := utf8.RuneCountInString(trimmed)
	if rLen == 0 || rLen > 120 {
		return false
	}
	if strings.HasSuffix(trimmed, ":") {
		return true
	}
	// Short title line without sentence-ending period in the middle.
	return rLen <= 80 && !strings.Contains(trimmed, ". ")
}

// startsNumberedSubClauseOne reports whether trimmed starts a new numbered
// sub-clause series like "1.1. ...", "2.1. ...", or "3.1 ...".
func startsNumberedSubClauseOne(trimmed string) bool {
	if len(trimmed) < 4 || trimmed[0] < '0' || trimmed[0] > '9' {
		return false
	}
	i := 0
	for i < len(trimmed) && trimmed[i] >= '0' && trimmed[i] <= '9' {
		i++
	}
	if i >= len(trimmed)-2 || trimmed[i] != '.' || trimmed[i+1] != '1' {
		return false
	}
	next := trimmed[i+2]
	return next == '.' || next == ' ' || next == '\t'
}

// isImplicitSectionHeading reports whether trimmed is a short unnumbered title
// line (e.g. "Đối tượng vay vốn", "Nguyên tắc vay vốn") immediately preceding
// a numbered sub-clause series ("2.1.", "3.1.").
func isImplicitSectionHeading(trimmed string, lines []string, idx int) bool {
	rLen := utf8.RuneCountInString(trimmed)
	if rLen < 3 || rLen > 100 {
		return false
	}
	// Must not start with a bullet, table pipe, or digit.
	first := trimmed[0]
	if first == '-' || first == '*' || first == '|' || (first >= '0' && first <= '9') || strings.HasPrefix(trimmed, "•") {
		return false
	}
	// Must not end like a prose sentence or list lead-in.
	lastR, _ := utf8.DecodeLastRuneInString(trimmed)
	switch lastR {
	case '.', ';', '?', '!', ':', ',', '。', '？', '！':
		return false
	}
	// Look ahead to the next non-empty line.
	for j := idx + 1; j < len(lines); j++ {
		nt := strings.TrimSpace(lines[j])
		if nt == "" {
			continue
		}
		return startsNumberedSubClauseOne(nt)
	}
	return false
}

// findHeuristicBoundaries scans text and returns boundary positions in
// ascending order. Lower-priority duplicates at the same offset are dropped.
func findHeuristicBoundaries(text string, langs []string) []boundary {
	var bounds []boundary

	// Form feeds — strongest single-character boundary.
	for _, idx := range allRuneIndices(text, "\f") {
		bounds = append(bounds, boundary{runeStart: idx, priority: PrioFormFeed})
	}

	// Per-line patterns walk the text once, line by line.
	lines := strings.Split(text, "\n")
	chapterPatterns := ChapterPatternsForLangs(langs)
	pos := 0
	inFence := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
		} else if !inFence {
			runeStart := pos
			added := false
			for _, pat := range chapterPatterns {
				if pat.MatchString(line) {
					bounds = append(bounds, boundary{
						runeStart: runeStart,
						priority:  PrioChapterMarker,
						level:     1,
						heading:   trimmed,
					})
					added = true
					break
				}
			}
			if !added && VietnamesePartyPattern.MatchString(line) {
				bounds = append(bounds, boundary{
					runeStart: runeStart,
					priority:  PrioChapterMarker,
					level:     1,
					heading:   trimmed,
				})
				added = true
			}
			if !added && NumberedSectionPattern.MatchString(line) {
				b := boundary{runeStart: runeStart, priority: PrioNumberedHead}
				if isShortSubHeadingLine(trimmed) {
					b.level = 2
					b.heading = trimmed
				}
				bounds = append(bounds, b)
				added = true
			}
			if !added && AllCapsHeadingPattern.MatchString(line) {
				bounds = append(bounds, boundary{
					runeStart: runeStart,
					priority:  PrioAllCapsHeading,
					level:     1,
					heading:   trimmed,
				})
				added = true
			}
			if !added && isImplicitSectionHeading(trimmed, lines, i) {
				bounds = append(bounds, boundary{
					runeStart: runeStart,
					priority:  PrioChapterMarker,
					level:     1,
					heading:   trimmed,
				})
				added = true
			}
			if !added && VisualSeparatorPattern.MatchString(line) {
				bounds = append(bounds, boundary{runeStart: runeStart, priority: PrioVisualSep})
				added = true
			}
			if !added && PageFooterPattern.MatchString(line) {
				bounds = append(bounds, boundary{runeStart: runeStart, priority: PrioPageFooter})
			}
		}
		pos += utf8.RuneCountInString(line)
		if i < len(lines)-1 {
			pos++ // \n
		}
	}

	// Excessive blank blocks (\n{3,}). Match at the *start* of the run so we
	// drop into the next paragraph cleanly.
	for _, idx := range ExcessiveBlanksPattern.FindAllStringIndex(text, -1) {
		runeStart := utf8.RuneCountInString(text[:idx[1]])
		bounds = append(bounds, boundary{runeStart: runeStart, priority: PrioBlankBlock})
	}

	if len(bounds) == 0 {
		return nil
	}

	// Sort by position; drop near-duplicate offsets keeping the highest priority.
	sort.Slice(bounds, func(i, j int) bool {
		if bounds[i].runeStart != bounds[j].runeStart {
			return bounds[i].runeStart < bounds[j].runeStart
		}
		return bounds[i].priority > bounds[j].priority
	})
	deduped := bounds[:0]
	prev := -1
	for _, b := range bounds {
		if b.runeStart != prev {
			deduped = append(deduped, b)
			prev = b.runeStart
		}
	}
	return deduped
}

// dropBoundsInsideSpans returns bounds with entries that fall strictly
// inside any of the (rune-offset) protected spans removed. Bounds at a
// span's start or end are kept — they align with the span edge and don't
// split protected content. spans must be sorted by start.
func dropBoundsInsideSpans(bounds []boundary, spans []span) []boundary {
	if len(spans) == 0 {
		return bounds
	}
	out := bounds[:0]
boundLoop:
	for _, b := range bounds {
		for _, s := range spans {
			if s.start >= b.runeStart {
				break // remaining spans start at or after b — can't contain b
			}
			if b.runeStart < s.end {
				continue boundLoop
			}
		}
		out = append(out, b)
	}
	return out
}

// allRuneIndices returns every rune offset where needle starts in text.
// Only used for single-rune needles like form-feed.
func allRuneIndices(text, needle string) []int {
	var out []int
	if needle == "" {
		return out
	}
	pos := 0
	for _, r := range text {
		if string(r) == needle {
			out = append(out, pos)
		}
		pos++
	}
	return out
}

// appendChunk slices runes[start:end] into a Chunk and appends it to out.
// Pure-whitespace slices are skipped (boundary clustering can occasionally
// produce them). The Content stored is the raw slice — Start/End rune
// offsets must match utf8.RuneCountInString(Content) for downstream
// reconstruction code; whitespace stripping for embedding happens in
// Chunk.EmbeddingContent.
func appendChunk(out []Chunk, runes []rune, start, end int, seq *int) []Chunk {
	if end <= start {
		return out
	}
	raw := string(runes[start:end])
	if strings.TrimSpace(raw) == "" {
		return out
	}
	c := Chunk{Content: raw, Seq: *seq, Start: start, End: end}
	*seq++
	return append(out, c)
}

// appendOversizeBlock recursively chunks a region that is itself larger than
// cfg.ChunkSize, using the legacy splitter so internal length budgets and
// protected patterns are still respected.
func appendOversizeBlock(out []Chunk, runes []rune, start, end int, cfg SplitterConfig, seq *int) []Chunk {
	if end <= start {
		return out
	}
	subText := string(runes[start:end])
	subs := SplitText(subText, cfg)
	for _, s := range subs {
		out = append(out, Chunk{
			Content:       s.Content,
			ContextHeader: s.ContextHeader,
			Seq:           *seq,
			Start:         start + s.Start,
			End:           start + s.End,
		})
		*seq++
	}
	return out
}

// applyOverlapAligned returns the rune offset where the next chunk should
// start. The target is `curEnd - overlap`, snapped to the nearest semantic
// boundary, newline, sentence end, or word boundary so chunks never begin
// mid-word (e.g. "o nhận..." or "ng tác...").
//
// curEnd itself is always a boundary (the bin-packer flushes at boundary
// positions), so we exclude it from the search — picking it would yield
// zero overlap, defeating the purpose of this function.
func applyOverlapAligned(runes []rune, curEnd, overlap int, bounds []boundary) int {
	if overlap <= 0 {
		return curEnd
	}
	target := curEnd - overlap
	if target < 0 {
		target = 0
	}
	// Allowed search window: [curEnd - 2*overlap, curEnd)
	windowStart := curEnd - 2*overlap
	if windowStart < 0 {
		windowStart = 0
	}

	// 1. Prefer a semantic boundary strictly inside the window.
	bestBound := -1
	for _, b := range bounds {
		if b.runeStart >= windowStart && b.runeStart < curEnd && b.runeStart > bestBound {
			bestBound = b.runeStart
		}
	}
	if bestBound >= 0 {
		return bestBound
	}

	// 2. Scan backwards from `target` to the previous newline within windowStart.
	for i := target; i > windowStart && i < len(runes); i-- {
		if runes[i] == '\n' && i+1 < curEnd {
			return i + 1
		}
	}

	// 3. Scan backwards from `target` to the previous sentence boundary within windowStart.
	for i := target; i > windowStart && i < len(runes); i-- {
		switch runes[i] {
		case '。', '？', '！':
			if i+1 < curEnd {
				return i + 1
			}
		case '.', '?', '!', ';':
			if i+1 < curEnd && (runes[i+1] == ' ' || runes[i+1] == '\n') {
				next := i + 2
				for next < curEnd && runes[next] == ' ' {
					next++
				}
				if next < curEnd {
					return next
				}
			}
		}
	}

	// 4. Word-safe fallback: never split mid-word. Scan backwards to the
	// previous space/tab within windowStart, or forward to the next space
	// before curEnd.
	for i := target; i > windowStart && i < len(runes); i-- {
		if runes[i] == ' ' || runes[i] == '\t' {
			if i+1 < curEnd {
				return i + 1
			}
		}
	}
	for i := target; i < curEnd && i < len(runes); i++ {
		if runes[i] == ' ' || runes[i] == '\t' || runes[i] == '\n' {
			if i+1 < curEnd {
				return i + 1
			}
		}
	}
	return target
}
