package chunker

import (
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

const sampleHD30 = `Cộng hòa xã hội chủ nghĩa Việt Nam
Độc lập - Tự do - Hạnh phúc
Hợp đồng mua bán
Số: 30/2026/YS-PĐ

Căn cứ Bộ luật Dân sự số: 91/2015/QH13 được Quốc hội khoá 13 nước CHXHCN Việt Nam thông qua ngày 24/11/2015.
Căn cứ Luật thương mại số: 36/2005/QH11 được Quốc hội khoá 11 nước CHXHCN Việt Nam thông qua ngày 14/06/2005.
Căn cứ vào nhu cầu sử dụng và khả năng cung cấp của hai bên
Hôm nay, ngày 24 tháng 02 năm 2026, chúng tôi gồm:

BÊN MUA (BÊN A): BỆNH VIỆN ĐA KHOA PHƯƠNG ĐÔNG – CHI NHÁNH
CÔNG TY TNHH TỔ HỢP Y TẾ PHƯƠNG ĐÔNG
Đại diện: Ông Nguyễn Công Minh Chức vụ: Tổng giám đốc Công ty kiêm phó giám đốc bệnh viện
Địa chỉ: Số 9 Phố Viên, phường Đông Ngạc, thành phố Hà Nội, Việt Nam
Điện thoại: 024.32753960
Tài khoản số: 1290153594 tại Ngân hàng thương mại cổ phần đầu tư và phát triển Việt Nam (BIDV) - Chi nhánh Hoàng Mai Hà Nội
Mã số thuế: 0101816147-001

BÊN BÁN (BÊN B): CÔNG TY TNHH DỊCH VỤ KỸ THUẬT Y SINH
Đại diện: Bà Lê Minh Ngọc Chức vụ: Giám đốc
Địa chỉ: Tầng 10, tòa nhà Licogi 13, Số 164 Khuất Duy Tiến, phường Thanh Xuân, thành phố Hà Nội, Việt Nam
Điện thoại: 024.37876052
Tài khoản: 02773055001 tại Ngân hàng TMCP Tiên Phong – chi nhánh Hà Nội
Mã NH: 01358001
Mã số thuế: 0102773055

Sau khi thỏa thuận, hai bên nhất trí cùng ký hợp đồng mua trang thiết bị với các điều kiện và điều khoản sau:

Điều 1 : Tên hàng - số lượng - đơn giá - trị giá
Bên B nhận bán cho bên A và bên A đồng ý mua các loại thiết bị, vật tư với số lượng như sau:

| Số TT | Tên hàng hóa | ĐV tính | Số lượng | Đơn giá (VNĐ) | Thành tiền (VNĐ) |
| --- | --- | --- | --- | --- | --- |
| 1 | Máy điện tim Model: ECG-3250 Hãng sản xuất: Shanghai Kohden Medical Electronic Instrument Corp. Xuất xứ: Trung Quốc Mới 100%, sản xuất từ năm 2025 trở đi Cấu hình: máy chính kèm phụ kiện sau: - Giấy ghi - Lau dầu nhiệt - Cáp điện tim - Dây nguồn - Điện cực trước ngục, 3mm - Điện cực chi - Ắc qui khô | bộ | 08 | 50.000.000 | 400.000.000 |
| 3 | Xe đạp (mua trong nước) | Cái | 08 | 3.000.000 | 24.000.000 |

Điều 3: Phương thức thanh toán

Đợt 1: Sau khi ký hợp đồng Bên A tạm ứng cho bên B 30% giá trị hợp đồng và nhận được đề nghị tạm ứng của bên B.
Đợt 2: Bên A thanh toán 70% giá trị còn lại của hợp đồng trong vòng 10 ngày kể từ ngày Bên B cung cấp cho bên A đầy đủ hồ sơ thanh toán gồm:

Hồ sơ thanh toán gồm:
- Hoá đơn tài chính hợp lệ
- Giấy đề nghị thanh toán
- Giấy chứng nhận nguồn gốc hàng hóa - CO: 02 bản sao công chứng
- Giấy chứng nhận chất lượng - CQ: 01 Bản gốc hoặc 02 bản sao công chứng
- Tờ khai hải quan: 02 Bản sao y công ty, đã xóa giá
- Vận đơn vận chuyển (Bill of landing / Air way bill): 02 Bản sao y công ty
- Danh mục đóng gói (Packing list): 02 Bản sao y công ty
- Tài liệu hướng dẫn sử dụng thiết bị bằng tiếng Anh hoặc tiếng Việt.
- Bảng tính giá trị thanh toán.
- Biên bản bàn giao nghiệm thu đưa vào sử dụng
- Hình thức thanh toán: bằng tiền mặt, séc hoặc chuyển khoản

Điều 4: Giao hàng

4.1. Điều kiện đảm bảo giao, nhận hàng hóa: Hàng hóa bên B bàn giao cho bên A phải đảm bảo đủ điều kiện theo điều 2 Hợp đồng này.

4.2. Địa điểm giao hàng và lắp đặt: Bệnh viện đa khoa Phương Đông, số 9 phố Viên, phường Cổ Nhuế 2, quận Bắc Từ Liêm, Hà Nội

4.3. Thời gian giao hàng: Trong vòng 15 ngày sau khi nhận được tiền tạm ứng.

4.4. Phương thức giao hàng: Bên B chịu trách nhiệm vận chuyển, lắp đặt, hướng dẫn sử dụng cho bên A.

Điều 5: Quyền và nghĩa vụ của bên A

5.1. Có quyền từ chối không nhận hoặc trả lại hàng hóa trong trường hợp hàng hóa giao không đúng chủng loại, chất lượng theo quy định của hợp đồng này.

5.2. Trong quá trình kiểm đếm tại địa điểm nhận hàng, khi phát hiện sản phẩm bị lỗi, sứt vỡ, hư hỏng thì hai bên cùng nhau tiến hành lập biên bản. Bên B sẽ chịu hoàn toàn trách nhiệm về những sản phẩm bị lỗi, sứt vỡ, hư hỏng và thay thế sản phẩm mới đáp ứng điều kiện theo như quy định của hợp đồng này

5.3. Bên A có trách nhiệm thanh toán đúng hạn cho bên B theo quy định tại Điều 3 của hợp đồng này. Nếu sản phẩm không đạt tiêu chuẩn theo quy định tại hợp đồng thì có quyền từ chối không thanh toán

5.4. Cử cán bộ cùng bên B giao nhận, kiểm tra chất lượng, chủng loại, quy cách, số lượng hàng hóa trước khi nhận

Điều 6: Quyền và nghĩa vụ của bên B

6.1. Được thanh toán theo đúng hình thức và thời hạn thanh toán đã quy định của hợp đồng này.

6.2. Đảm bảo hàng hóa giao cho bên A thuộc sở hữu hợp pháp của bên B, hiện không có tranh chấp và được phép lưu hành ở Việt Nam.

6.3. Bên B có trách nhiệm thông báo cho bên A về thời gian giao hàng cụ thể.

6.4. Bên B có trách nhiệm chuyển hàng đến địa điểm giao nhận hàng theo khoản 4.2 của Điều 4 trong hợp đồng này. Hai bên sẽ cùng nhau tiến hành kiểm đếm, giao nhận cả về số lượng và chất lượng, quy cách từng sản phẩm tại kho Bên A.

6.5. Cung cấp kịp thời, đúng, đủ số lượng, chất lượng hàng và giao hàng theo đúng tiến độ yêu cầu của bên A.

6.6. Cung cấp đầy đủ hồ sơ thanh toán hợp lệ theo quy định của hợp đồng này và các giấy tờ chứng minh chất lượng sản phẩm và chịu trách nhiệm về sự trung thực và tính hợp pháp của chứng từ.

6.7. Chịu trách nhiệm về hàng hóa hư hỏng, sứt, vỡ...trong quá trình vận chuyển và thay thế bằng sản phẩm tương tự mới 100% cho bên A.

6.8. Bố trí nhân viên kỹ thuật lắp đặt và hướng dẫn sử dụng, bảo quản thiết bị cho bên A.

6.9. Bảo mật thông tin trong quá trình thực hiện hợp đồng với bên A.`

const sampleHD70 = `ĐIỀU 4: GIAO NHẬN HÀNG HÓA

4.1 Thời gian giao hàng, lắp đặt và bàn giao cho Bên A:
a. Thời gian hoàn thành giao hàng và lắp đặt là 20-24 tuần. trong đó:
- Thời gian giao hàng: 16 - 18 tuần kể từ ngày bên B nhận được đầy đủ khoản thanh toán tạm ứng và Bảo lãnh thanh toán được quy định tại điều 7 và điều 9 của hợp đồng và xác nhận của bên A trên Bản vẽ sơ bộ phòng lắp đặt máy, tùy theo điều kiện nào đến sau cùng.
- Thời gian lắp đặt: Với điều kiện bên A hoàn thành các nghĩa vụ của mình đúng hạn, bên B sẽ lắp đặt hoặc hoàn tất lắp đặt Sản Phẩm trong vòng 04 - 06 tuần kể từ ngày các bên ký Biên Bản Giao hàng.
b. Địa điểm giao hàng và lắp đặt: Tại Bệnh viện đa khoa Phương Đông, có địa chỉ tại Số 9 Phố Viên, phường Cổ Nhuế 2, quận Bắc Từ Liêm, thành phố Hà Nội.
c. Trước khi giao hàng và lắp đặt tối đa 15 ngày, Bên B phải thông báo bằng văn bản cho Bên A để chuẩn bị các điều kiện cần thiết cho việc tiếp nhận hàng.
d. Thời gian giao hàng nếu trên được tính từ lúc Bên B nhận được số tiền tạm ứng của Bên A theo quy định tại mục 7.1.1 Điều 7 của Hợp đồng này.
e. Bên B thông báo bằng văn bản cho Bên A về Ngày xuất xưởng của thiết bị. Bên A có quyền yêu cầu thay đổi Ngày Xuất Xưởng trong vòng 01 tuần (bao gồm cả ngày nghỉ, ngày lễ tết) kể từ ngày nhận được thông báo Ngày Xuất Xưởng của Bên B. Trường hợp Bên A thay đổi Ngày Xuất Xưởng sau thời hạn trên, Bên B có quyền tính thêm chi phí bằng: 150.000.000 VND (Bằng chữ: Một trăm năm mươi triệu đồng) cho mỗi lần yêu cầu. Khoản phí này sẽ được tính cộng vào giá trị của đơn hàng.
f. Trường hợp Bên A trì hoãn nhận hàng sau khi đã xác nhận Ngày Xuất Xưởng với Bên B hoặc trì hoãn thời gian nhận hàng sau khi hàng về đến cảng do lỗi của Bên A trong bất kỳ trường hợp nào, Bên A phải chịu toàn bộ chi phí lưu kho, bến bãi, vận chuyển và các khoản chi phí phát sinh khác. Khoản phí này sẽ được thanh toán trước khi Hàng Hóa được giao.
g. Trường hợp khi Bên B đã gửi thông báo giao hàng cho Bên A, được Bên A xác nhận, mà Bên A hoặc đại diện của Bên A không có mặt như lịch hẹn do Bên B đưa ra hoặc không ký biên bản giao hàng mặc dù hàng hoá đã được chuyển tới địa điểm nói trên, biên bản giao hàng xem như đã được ký vào thời gian như lịch hẹn do Bên B đưa ra. Biên bản giao hàng được lập làm 04 bản, mỗi bên giữ 02 bản.
Biên bản xác nhận hàng hóa nguyên đai nguyên kiện được ký trong vòng 03 ngày làm việc kể từ thời điểm hàng đến Bệnh Viện. Trong trường hợp bên A không ký biên bản xác nhận mà không do lỗi của bên B thì Biên bản xác nhận hàng hóa nguyên đai nguyên kiện mặc nhiên được xác nhận. Đồng thời bên A phải có nghĩa vụ thanh toán theo quy định tại Hợp đồng này.
h. Trường hợp hàng hóa chậm bàn giao và lắp đặt do lỗi của Bên A thì Bên A vẫn phải thanh toán cho Bên B.

ĐIỀU 6: GIÁ TRỊ HỢP ĐỒNG

6.1 Tổng Giá trị hợp đồng (đã bao gồm thuế GTGT): 23.992.500.000 đồng
Bằng chữ: Hai mươi ba tỷ, chín trăm chín hai triệu, năm trăm nghìn đồng
(Chi tiết tại Bảng phụ lục 1: Bảng giá trị hợp đồng kèm theo).

6.2 Hình thức hợp đồng: Hợp đồng trọn gói.

ĐIỀU 7: TẠM ỨNG VÀ THANH TOÁN HỢP ĐỒNG

7.1.2. Thanh toán đợt 1:
a. Giá trị thanh toán: Bên A sẽ thanh toán cho Bên B số tiền 40% giá trị Hợp đồng (không kể giá trị tạm ứng) trong vòng 20 ngày kể từ ngày hàng hoá về đến Bệnh viện.
b. Hồ sơ đề nghị thanh toán đợt 1:
- Giấy chứng nhận chất lượng của Nhà sản xuất (CQ) - Bản scan + 04 bản sao y Công ty.
- Giấy chứng nhận xuất xứ hàng hóa (CO) - Bản scan + 04 bản sao y Công ty.
- Danh mục đóng gói hàng hoá (Packing List) - Bản scan + 04 bản sao y Công ty.
- Vận đơn tàu biển/ hàng không - Bản scan + 04 bản sao y Công ty.
- Tờ khai hải quan xóa giá – 05 bản có dấu của bên B (tờ khai điện tử);
- Biên bản xác nhận hàng hóa nguyên đai nguyên kiện theo thông báo giao hàng của bên B: được hai bên ký xác nhận (bản gốc).
- Giấy đề nghị thanh toán;

7.1.3 Thanh toán đợt 2:
a. Giá trị thanh toán: Bên A thanh toán cho Bên B 30% giá trị còn lại của hợp đồng, trong vòng 30 ngày kể từ ngày nghiệm thu đưa vào sử dụng.
b. Hồ sơ đề nghị thanh toán đợt 2:
- Giấy chứng nhận chất lượng của Nhà sản xuất (CQ) - Bản gốc + 04 bản sao Công chứng.
- Giấy chứng nhận xuất xứ hàng hóa (CO) - Bản gốc + 04 bản sao Công chứng.
- Danh mục đóng gói hàng hoá (Packing List) - 05 bản sao của nhà nhập khẩu.
- Vận đơn tàu biển/ hàng không/ tài liệu tương đương - 05 bản sao của nhà nhập khẩu.
- Tờ khai hải quan xóa giá – 05 bản sao của nhà nhập khẩu.
- Biên bản nghiệm thu hoàn thành bàn giao đưa vào sử dụng được ký bởi đại diện có thẩm quyền của Bên A và Bên B kèm theo xác nhận đã hoàn thành công tác đào tạo, hướng dẫn vận hành cho cán bộ chuyên môn của Bên A: Bản gốc
- 7. Thư bảo lãnh bảo hành: Bản gốc.
- 8. Công văn đề nghị thanh toán.
- 9. Tài liệu hướng dẫn vận hành, bảo trì, bảo dưỡng thiết bị (Tiếng Anh và tiếng Việt) - 02 bộ.
- 10. Hóa đơn tài chính hợp lệ tương ứng với 100 % giá trị hàng hóa được giao.

ĐIỀU 9: ĐẢM BẢO THỰC HIỆN HỢP ĐỒNG

9.1. Bên B cam kết cung cấp cho Bên A bảo đảm thực hiện Hợp đồng bằng hình thức chứng thư bảo lãnh của ngân hàng có giá trị bằng 5% giá trị hợp đồng trong thời gian chậm nhất 10 ngày kể từ ngày hai Bên ký kết hợp đồng.`

// assertNoMidWordStart checks that no chunk starts inside a word in doc.
func assertNoMidWordStart(t *testing.T, doc string, chunks []ChildChunk) {
	t.Helper()
	runes := []rune(doc)
	for i, c := range chunks {
		if c.Start > 0 && c.Start < len(runes) {
			prev := runes[c.Start-1]
			curr := runes[c.Start]
			if unicode.IsLetter(prev) && unicode.IsLetter(curr) {
				preview := strings.TrimSpace(c.Content)
				if len(preview) > 60 {
					preview = preview[:60]
				}
				t.Errorf("chunk[%d] starts mid-word at rune %d (prev=%q curr=%q): %q",
					i, c.Start, prev, curr, preview)
			}
		}
	}
}

func TestVietnameseContractHD30_ParentChildQuality(t *testing.T) {
	base := NormalizeSplitterConfig(SplitterConfig{Strategy: StrategyAuto})
	parentCfg, childCfg := DeriveParentChildConfigs(base, 4096, 1024)
	res := SplitParentChild(sampleHD30, parentCfg, childCfg)
	if len(res.Children) == 0 {
		t.Fatal("expected child chunks")
	}

	// 1. Never cut mid-word in overlap (previously "o nhận...", "n phẩm...").
	assertNoMidWordStart(t, sampleHD30, res.Children)

	// 2. Payment document bullet list ("Hồ sơ thanh toán gồm:") stays intact in one chunk.
	foundFullList := false
	for _, c := range res.Children {
		if strings.Contains(c.Content, "Hoá đơn tài chính hợp lệ") &&
			strings.Contains(c.Content, "Biên bản bàn giao nghiệm thu đưa vào sử dụng") {
			foundFullList = true
			break
		}
	}
	if !foundFullList {
		t.Errorf("expected 'Hồ sơ thanh toán gồm:' bullet list to remain intact in a single chunk")
	}

	// 3. No orphan chapter heading at the end of a chunk, and no heading-only chunk.
	for i, c := range res.Children {
		trimmed := strings.TrimSpace(c.Content)
		lines := strings.Split(trimmed, "\n")
		if len(lines) == 1 && VietnameseChapterPattern.MatchString(lines[0]) {
			t.Errorf("chunk[%d] is a heading-only chunk: %q", i, trimmed)
		}
		lastLine := strings.TrimSpace(lines[len(lines)-1])
		if len(lines) > 1 && VietnameseChapterPattern.MatchString(lastLine) {
			t.Errorf("chunk[%d] ends with an orphan chapter heading: %q", i, lastLine)
		}
	}

	// 4. Long table row ("Máy điện tim...") keeps its Markdown table header columns.
	foundTableRowWithHeader := false
	for _, c := range res.Children {
		if strings.Contains(c.Content, "Máy điện tim Model: ECG-3250") {
			if strings.Contains(c.Content, "Số TT") && strings.Contains(c.Content, "Tên hàng hóa") {
				foundTableRowWithHeader = true
			} else {
				t.Errorf("table row chunk lost its table header:\n%s", c.Content)
			}
		}
	}
	if !foundTableRowWithHeader {
		t.Errorf("did not find table row 'Máy điện tim Model: ECG-3250' with header")
	}

	// 5. Sub-chunks under Điều 5 / Điều 6 carry chapter context via Content or ContextHeader.
	for i, c := range res.Children {
		emb := c.EmbeddingContent()
		if strings.Contains(c.Content, "5.3. Bên A có trách nhiệm") && !strings.Contains(emb, "Điều 5") {
			t.Errorf("chunk[%d] with 5.3 lacks Điều 5 context in EmbeddingContent:\n%s", i, emb)
		}
		if strings.Contains(c.Content, "6.6. Cung cấp đầy đủ hồ sơ") && !strings.Contains(emb, "Điều 6") {
			t.Errorf("chunk[%d] with 6.6 lacks Điều 6 context in EmbeddingContent:\n%s", i, emb)
		}
	}
}

func TestVietnameseContractHD70_ParentChildQuality(t *testing.T) {
	base := NormalizeSplitterConfig(SplitterConfig{Strategy: StrategyAuto})
	parentCfg, childCfg := DeriveParentChildConfigs(base, 4096, 1024)
	res := SplitParentChild(sampleHD70, parentCfg, childCfg)
	if len(res.Children) == 0 {
		t.Fatal("expected child chunks")
	}

	// 1. Never cut mid-word (previously "ng tác đào tạo...").
	assertNoMidWordStart(t, sampleHD70, res.Children)

	// 2. No heading-only chunks (previously ĐIỀU 4, ĐIỀU 6, ĐIỀU 9 standalone).
	for i, c := range res.Children {
		trimmed := strings.TrimSpace(c.Content)
		lines := strings.Split(trimmed, "\n")
		if len(lines) == 1 && VietnameseChapterPattern.MatchString(lines[0]) {
			t.Errorf("chunk[%d] is a heading-only chunk: %q", i, trimmed)
		}
		if utf8.RuneCountInString(trimmed) < 60 {
			t.Errorf("chunk[%d] is unexpectedly tiny (%d runes): %q", i, utf8.RuneCountInString(trimmed), trimmed)
		}
	}

	// 3. No duplicate clause "h. Trường hợp hàng hóa chậm bàn giao" across parent seam.
	hCount := 0
	for _, c := range res.Children {
		if strings.Contains(c.Content, "h. Trường hợp hàng hóa chậm bàn giao") {
			hCount++
		}
	}
	if hCount != 1 {
		t.Errorf("expected clause h. to appear exactly once, appeared %d times", hCount)
	}

	// 4. Every chunk containing "Bản gốc + 04 bản sao Công chứng" (đợt 2) must carry
	// "7.1.3" or "đợt 2" context in EmbeddingContent so it never confuses đợt 1 vs đợt 2.
	for i, c := range res.Children {
		if strings.Contains(c.Content, "Bản gốc + 04 bản sao Công chứng") {
			emb := c.EmbeddingContent()
			if !strings.Contains(emb, "7.1.3") && !strings.Contains(emb, "đợt 2") {
				t.Errorf("chunk[%d] with đợt 2 docs lacks 7.1.3/đợt 2 context:\n%s", i, emb)
			}
		}
	}
}

func TestDetectLanguage_Vietnamese(t *testing.T) {
	got := DetectLanguage("Điều 3: Phương thức thanh toán. Bên A tạm ứng cho bên B 30% giá trị hợp đồng.")
	if got != LangVietnamese {
		t.Errorf("expected %s, got %s", LangVietnamese, got)
	}
}

func TestApplyOverlapAligned_NeverSplitsMidWordAtSmallSize(t *testing.T) {
	// Even with small legacy child size 384 (like existing KBs), overlap must never split mid-word.
	base := NormalizeSplitterConfig(SplitterConfig{Strategy: StrategyAuto})
	parentCfg, childCfg := DeriveParentChildConfigs(base, 4096, 384)
	res := SplitParentChild(sampleHD30, parentCfg, childCfg)
	assertNoMidWordStart(t, sampleHD30, res.Children)
}
