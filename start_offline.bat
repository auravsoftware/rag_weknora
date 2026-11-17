@echo off
chcp 65001 >nul
echo ================================================================================
echo    BUNG DOCKER HỆ THỐNG RAG TIẾNG VIỆT 100%% OFFLINE (1x 24GB VRAM GPU)
echo ================================================================================

REM 1. Khởi tạo tệp .env nếu chưa có
if not exist .env (
    echo [+] Chưa tìm thấy .env, đang tự động tạo từ .env.offline-24gb.example...
    copy .env.offline-24gb.example .env >nul
) else (
    echo [✓] Tệp .env đã sẵn sàng.
)

REM 2. Đảm bảo các thư mục cache và dữ liệu tồn tại
if not exist "models_cache\huggingface" mkdir "models_cache\huggingface"
if not exist "data\files" mkdir "data\files"
echo [✓] Đã tạo thư mục lưu trữ: models_cache\huggingface và data\files.

REM 3. Khởi chạy Docker Compose (Tự động kéo image có sẵn, không ngốn RAM biên dịch)
echo [+] Đang bung Docker Compose...
docker compose -f docker-compose.offline-24gb.yml up -d

echo ================================================================================
echo    KHỞI CHẠY HỆ THỐNG THÀNH CÔNG!
echo ================================================================================
echo   • Giao diện Web RAG Tiếng Việt : http://localhost:6666
echo   • Backend Core API             : http://localhost:6667
echo   • PostgreSQL (ParadeDB HNSW)   : Cổng 6668
echo   • Redis Cache                  : Cổng 6669
echo   • vLLM (Qwen2.5-VL-7B OCR+LLM) : http://localhost:6670/v1
echo   • Infinity (BGE-M3 + Reranker) : http://localhost:6671/v1
echo   • DocReader Parser             : Cổng 6672
echo ================================================================================
echo   Mở trình duyệt truy cập ngay: http://localhost:6666
echo ================================================================================
pause
