@echo off
chcp 65001 >nul
echo ================================================================================
echo    BUNG DOCKER HỆ THỐNG RAG TIẾNG VIỆT 100%% OFFLINE (1x 24GB VRAM GPU)
echo ================================================================================

REM 1. Kiểm tra xem Docker Desktop đã được bật và đang chạy chưa
docker info >nul 2>&1
if errorlevel 1 (
    echo.
    echo ================================================================================
    echo    [X] LỖI: DOCKER DESKTOP CHƯA ĐƯỢC MỞ HOẶC CHƯA KHỞI ĐỘNG XONG!
    echo ================================================================================
    echo    Nguyên nhân: Docker Engine chưa phản hồi qua cổng pipe của Windows.
    echo.
    echo    Cách xử lý:
    echo    1. Mở ứng dụng "Docker Desktop" từ Start Menu hoặc màn hình Desktop.
    echo    2. Nhìn xuống góc dưới màn hình (hoặc cửa sổ Docker), đợi icon Docker chuyển
    echo       sang màu xanh lá (trạng thái "Engine running").
    echo    3. Sau khi Docker Desktop đã chạy xong, hãy chạy lại file start_offline.bat này.
    echo ================================================================================
    echo.
    pause
    exit /b 1
)

REM 2. Khởi tạo tệp .env nếu chưa có
if not exist .env (
    echo [+] Chưa tìm thấy .env, đang tự động tạo từ .env.offline-24gb.example...
    copy .env.offline-24gb.example .env >nul
) else (
    echo [✓] Tệp .env đã sẵn sàng.
)

REM 3. Đảm bảo các thư mục cache và dữ liệu tồn tại
if not exist "models_cache\huggingface" mkdir "models_cache\huggingface"
if not exist "data\files" mkdir "data\files"
echo [✓] Đã tạo thư mục lưu trữ: models_cache\huggingface và data\files.

REM 4. Khởi chạy Docker Compose (Tự động kéo image có sẵn, không ngốn RAM biên dịch)
echo [+] Đang bung Docker Compose (kéo và chạy các container dịch vụ)...
docker compose -f docker-compose.offline-24gb.yml up -d
if errorlevel 1 (
    echo.
    echo ================================================================================
    echo    [X] LỖI: QUÁ TRÌNH KHỞI CHẠY CONTAINER THẤT BẠI!
    echo ================================================================================
    echo    Vui lòng kiểm tra lại log báo lỗi ở trên.
    echo ================================================================================
    pause
    exit /b 1
)

echo.
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
