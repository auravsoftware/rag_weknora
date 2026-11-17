# Hệ Thống Tra Cứu & Hỏi Đáp Tài Liệu RAG Tiếng Việt (100% Offline - 1x GPU 24GB VRAM)

Hệ thống **RAG (Retrieval-Augmented Generation)** tra cứu, hỏi đáp tài liệu nội bộ chuyên sâu có **trích dẫn nguyên văn tài liệu gốc**, quản lý kho tài liệu theo **cây thư mục đa cấp** kèm **5 trường siêu dữ liệu hành chính/pháp lý**, xử lý đa định dạng (**PDF Vector, Word `.docx`/`.doc`, PDF Scan ảnh**) và vận hành **100% Offline** trên máy chủ trang bị **01 GPU NVIDIA RTX PRO 4000 Blackwell (24GB VRAM)**.

---

## 1. Các tính năng cốt lõi sau khi "Đại phẫu" & Việt hóa

1. **Giao diện 100% Tiếng Việt (`vi-VN` mặc định), Tinh gọn & Chuyên biệt**:
   - Đã loại bỏ các phân hệ dư thừa (`miniprogram/`, `website-docs/`, Agents đa tác tử, Artifacts, Automations, Toolbox, Organizations, Cloud/IM Integrations, Web Search, Sandbox).
   - Tập trung hoàn toàn vào 2 luồng nghiệp vụ chính: **Hỏi đáp RAG có trích dẫn nguồn** và **Quản lý Kho tri thức Tài liệu**.

2. **Quản lý Tài liệu theo Cây thư mục & 5 Trường Metadata Pháp lý/Hành chính**:
   - **Cây thư mục (`KbFolderTree.vue`)**: Luôn hiển thị ở thanh bên trái của Kho tri thức, hỗ trợ tạo thư mục gốc, tạo thư mục con nhiều cấp và nhập tài liệu theo cấu trúc thư mục từ máy tính.
   - **5 Trường Metadata chuẩn hóa** (Hiển thị trực tiếp tại Hộp thoại Tải lên, Bảng danh sách tài liệu, Ngăn chi tiết tài liệu và **bơm tự động vào từng đoạn Chunk cho LLM suy luận**):
     1. **Tiêu đề văn bản** (`title` / `custom_metadata.document_title`)
     2. **Nơi ban hành** (`custom_metadata.issuing_authority`)
     3. **Ngày ban hành** (`custom_metadata.issue_date` - `YYYY-MM-DD`)
     4. **Ngày hiệu lực** (`custom_metadata.effective_date` - `YYYY-MM-DD`)
     5. **Ngày hết hiệu lực** (`custom_metadata.expiration_date` - `YYYY-MM-DD` hoặc để trống)
   - **Tự động kiểm tra tình trạng hiệu lực**: Hệ thống tự động gắn nhãn **`[Đang có hiệu lực]`**, **`[Chưa có hiệu lực]`**, hoặc **`[Hết hiệu lực]`** và cảnh báo rõ ràng trong câu trả lời của AI khi văn bản đã hết hiệu lực.

3. **Xử lý Đa định dạng Tài liệu Đầu vào (PDF Vector, Word, PDF Scan ảnh)**:
   - **PDF Vector có sẵn lớp chữ (`pdf_parser.py`)**: Bóc tách bố cục đa cột, tiêu đề, bảng biểu kèm chuẩn hóa Unicode **`NFC`** cho tiếng Việt (khắc phục lỗi tổ hợp dấu `NFD` khiến tìm kiếm BM25 bị trượt).
   - **Word `.docx` / `.doc` (`docx_parser.py`)**: Chuyển đổi `.doc` sang `.docx` qua LibreOffice headless, bóc tách đoạn văn & bảng biểu sang Markdown chuẩn hóa `NFC`.
   - **PDF Scan ảnh & Ảnh chụp văn bản**: Tự động phát hiện trang thiếu lớp văn bản, render ảnh độ phân giải cao và gửi qua mô hình thị giác **`Qwen/Qwen2.5-VL-7B-Instruct-AWQ`** để OCR toàn trang thành văn bản cấu trúc & bảng Markdown.
   - **Bộ cắt đoạn theo cấu trúc văn bản Việt Nam (`internal/infrastructure/chunker/patterns.go`)**: Nhận diện chính xác các cấp đề mục **`Phần` -> `Chương` -> `Mục` -> `Tiểu mục` -> `Điều` -> `Khoản`** của văn bản quy phạm pháp luật / quy chế nội bộ tiếng Việt.

4. **Kiến trúc Model 100% Offline tối ưu cho 1x GPU 24GB VRAM (RTX PRO 4000 Blackwell)**:
   - **Dùng chung 1 mô hình duy nhất cho cả OCR & LLM Hỏi đáp**: **`Qwen/Qwen2.5-VL-7B-Instruct-AWQ`** chạy trên **vLLM** (`gpu-memory-utilization=0.62`, chiếm ~15.0 GB VRAM), vừa đảm nhận OCR trang PDF Scan (`VLLM`) vừa trả lời hỏi đáp RAG có trích dẫn nguồn (`KnowledgeQA`).
   - **Embedding & Reranker chạy chung trên Infinity Server**: **`BAAI/bge-m3`** (Embedding 1024 chiều) + **`BAAI/bge-reranker-v2-m3`** (Cross-Encoder Reranker) chỉ chiếm ~2.5 GB VRAM.
   - **Hybrid Search (Vector + BM25)**: **ParadeDB PostgreSQL 17** (`pgvector` HNSW + `pg_search` BM25) chạy trên CPU/RAM (0 GB VRAM).

---

## 2. Bảng phân bổ tài nguyên phần cứng (1x RTX PRO 4000 Blackwell - 24GB VRAM)

| Thành phần | Model / Công nghệ | Dịch vụ chạy | VRAM tiêu thụ | Vai trò |
| :--- | :--- | :--- | :---: | :--- |
| **LLM Hỏi đáp RAG + VLM OCR** | `Qwen/Qwen2.5-VL-7B-Instruct-AWQ` | `vllm-qwen25-vl` (`:7670`) | **~15.0 GB** | Trả lời hỏi đáp RAG có trích dẫn `[ref:N]` **kiêm** OCR PDF Scan & bảng biểu |
| **Embedding Đa ngôn ngữ** | `BAAI/bge-m3` (FP16, 1024-dim) | `infinity-models` (`:7671`) | **~1.2 GB** | Mã hóa vector ngữ nghĩa tiếng Việt |
| **Reranker Chấm điểm lại** | `BAAI/bge-reranker-v2-m3` (FP16) | `infinity-models` (`:7671`) | **~1.3 GB** | Xếp hạng lại Top-K đoạn văn bản trước khi đưa vào LLM |
| **Vector + BM25 Hybrid DB** | `paradedb/paradedb:v0.25.11-pg17` | `postgres` (`:7668` host / `5432` container) | **0 GB (CPU)** | Tìm kiếm kết hợp Dense Vector (HNSW) + Từ khóa BM25 |
| **Redis Cache / Task Queue** | `redis:7.0-alpine` | `redis` (`:7669` host / `6379` container) | **0 GB (CPU)** | Hàng đợi tác vụ Asynq & xử lý luồng |
| **DocReader Parser** | Python DocReader + LibreOffice | `docreader` (`:7672` host / `50051` gRPC) | **0 GB (CPU)** | Bóc tách PDF Vector, Word .docx/.doc, render trang PDF Scan |
| **Backend & Frontend** | Go gin-gonic + Vue 3 UI Tiếng Việt | `app` (`:7667`) + `frontend` (`:7666`) | **0 GB (CPU)** | Web UI (`:7666`) & API Core (`:7667`) |
| **Tổng cộng VRAM** | | | **~17.5 GB / 24.0 GB** | **Dư ~6.5 GB VRAM an toàn cho hệ thống & đột biến tải** |

> [!NOTE]
> Toàn bộ dải cổng đã được chuyển đổi sang **`7666 - 7672`** để đảm bảo tuyệt đối an toàn cho trình duyệt (Chrome/Edge chặn dải 6665-6669 vì là cổng IRC) và không xung đột với các dịch vụ mặc định đang có trên máy chủ.

---

## 3. Hướng dẫn Triển khai Nhanh (Offline 24GB VRAM)

### Cách 1: Khởi chạy nhanh bằng 1 lệnh duy nhất (Khuyên dùng)
```bash
# Trên Windows CMD:
start_offline.bat

# Trên Linux:
./start_offline.sh
```
Script sẽ tự động tạo `.env` từ mẫu, đảm bảo thư mục dữ liệu, kiểm tra GPU và khởi chạy toàn bộ cụm Docker Compose.

### Cách 2: Triển khai thủ công bằng Docker Compose
1. **Khởi tạo tệp `.env`**:
   ```bash
   cp .env.offline-24gb.example .env
   ```
2. **(Tùy chọn cho môi trường ngắt mạng hoàn toàn - Air-gapped)** Tải trước model từ Hugging Face về `./models_cache/huggingface`:
   ```bash
   pip install -U "huggingface_hub[cli]"
   mkdir -p ./models_cache/huggingface

   # Các đường link model chính thức (đã kiểm tra hoạt động 100%):
   HF_HOME=./models_cache/huggingface huggingface-cli download Qwen/Qwen2.5-VL-7B-Instruct-AWQ
   HF_HOME=./models_cache/huggingface huggingface-cli download BAAI/bge-m3
   HF_HOME=./models_cache/huggingface huggingface-cli download BAAI/bge-reranker-v2-m3
   ```
   Sau khi tải, bật `HF_HUB_OFFLINE=1` trong `.env` để chạy hoàn toàn không cần kết nối mạng.
3. **Bung Docker Compose**:
   ```bash
   docker compose -f docker-compose.offline-24gb.yml up -d
   ```

- **Giao diện Web Tiếng Việt**: Truy cập `http://localhost:7666` (hoặc `http://<IP-Máy-Chủ>:7666`).
- **Tự động cấu hình Model**: Hệ thống tự động nạp tệp [`config/builtin_models.yaml`](file:///home/antigravity/auravsoftware/rag_weknora/config/builtin_models.yaml) khi khởi động, đăng ký sẵn 4 tác vụ (`KnowledgeQA`, `VLLM`, `Embedding`, `Rerank`) trỏ về `vllm-qwen25-vl:7670` và `infinity-models:7671` mà không cần thao tác gì thêm.

---

## 4. Cấu trúc Thư mục Chính

- [`config/builtin_models.yaml`](file:///home/antigravity/auravsoftware/rag_weknora/config/builtin_models.yaml): Khai báo 4 model mặc định dùng chung `Qwen2.5-VL-7B-Instruct-AWQ` (LLM + OCR) và `BAAI/bge-m3` + `bge-reranker-v2-m3`.
- [`config/prompt_templates/`](file:///home/antigravity/auravsoftware/rag_weknora/config/prompt_templates): Bộ Prompt hệ thống tiếng Việt ép buộc trích dẫn nguồn `[ref:N]` và đối chiếu ngày hiệu lực văn bản.
- [`docker-compose.offline-24gb.yml`](file:///home/antigravity/auravsoftware/rag_weknora/docker-compose.offline-24gb.yml): Cấu hình Docker Compose tích hợp trọn gói vLLM + Infinity + ParadeDB + DocReader + Backend + Frontend cho 1 GPU 24GB VRAM.
- [`internal/types/knowledge.go`](file:///home/antigravity/auravsoftware/rag_weknora/internal/types/knowledge.go): Định dạng 5 trường metadata hành chính/pháp lý (`CustomMetadataText()`) và tự động tính toán hiệu lực văn bản.
- [`internal/infrastructure/chunker/`](file:///home/antigravity/auravsoftware/rag_weknora/internal/infrastructure/chunker): Bộ chia đoạn thích ứng nhận diện cấu trúc văn bản pháp quy tiếng Việt (`Phần`, `Chương`, `Mục`, `Tiểu mục`, `Điều`, `Khoản`).
- [`docreader/parser/`](file:///home/antigravity/auravsoftware/rag_weknora/docreader/parser): Bộ giải mã PDF Vector, Word `.docx`/`.doc` chuẩn hóa Unicode `NFC` và render trang PDF Scan cho VLM OCR.
- [`frontend/src/i18n/locales/vi-VN.ts`](file:///home/antigravity/auravsoftware/rag_weknora/frontend/src/i18n/locales/vi-VN.ts): Bộ ngôn ngữ giao diện Tiếng Việt toàn diện (7.919 khóa).
