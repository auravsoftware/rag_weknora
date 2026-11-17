@echo off
chcp 65001 >nul
echo ================================================================================
echo    BUNG DOCKER HỆ THỐNG RAG TIẾNG VIỆT 100%% OFFLINE [1x 24GB VRAM GPU]
echo ================================================================================

REM 1. Kiem tra Docker Desktop
docker info >nul 2>&1
if errorlevel 1 goto DOCKER_NOT_RUNNING
goto DOCKER_IS_READY

:DOCKER_NOT_RUNNING
echo.
echo ================================================================================
echo    [X] LỖI: DOCKER DESKTOP CHƯA ĐƯỢC MỞ HOẶC CHƯA KHỞI ĐỘNG XONG!
echo ================================================================================
echo    Nguyên nhân: Docker Engine chưa phản hồi qua cổng pipe của Windows.
echo.
echo    Cách xử lý:
echo    1. Mở ứng dụng Docker Desktop từ Start Menu hoặc màn hình Desktop.
echo    2. Đợi icon Docker ở thanh taskbar báo Engine running [màu xanh lá].
echo    3. Sau khi Docker Desktop đã chạy xong, hãy chạy lại file này.
echo ================================================================================
echo.
pause
exit /b 1

:DOCKER_IS_READY
echo [✓] Docker Engine đang hoạt động tốt.

REM 2. Khoi tao tep .env neu chua co
if exist .env goto ENV_EXISTS
echo [+] Chưa tìm thấy .env, đang tự động tạo từ .env.offline-24gb.example...
copy .env.offline-24gb.example .env >nul
:ENV_EXISTS
echo [✓] Tệp .env đã sẵn sàng.

REM 3. Dam bao cac thu muc ton tai
if not exist "models_cache\huggingface" mkdir "models_cache\huggingface"
if not exist "data\files" mkdir "data\files"
echo [✓] Đã tạo thư mục lưu trữ: models_cache\huggingface và data\files.

REM 4. Khoi chay Docker Compose
echo [+] Đang bung Docker Compose [kéo image và chạy các container]...
docker compose -f docker-compose.offline-24gb.yml up -d
if errorlevel 1 goto COMPOSE_FAILED
goto COMPOSE_OK

:COMPOSE_FAILED
echo.
echo ================================================================================
echo    [X] LỖI: QUÁ TRÌNH KHỞI CHẠY CONTAINER THẤT BẠI!
echo ================================================================================
echo    Vui lòng kiểm tra lại log báo lỗi ở trên.
echo ================================================================================
pause
exit /b 1

:COMPOSE_OK
echo.
echo ================================================================================
echo    KHỞI CHẠY HỆ THỐNG THÀNH CÔNG!
echo ================================================================================
echo   - Giao diện Web RAG Tiếng Việt : http://localhost:6666
echo   - Backend Core API             : http://localhost:6667
echo   - PostgreSQL [ParadeDB HNSW]   : Cổng 6668
echo   - Redis Cache                  : Cổng 6669
echo   - vLLM [Qwen2.5-VL-7B OCR+LLM] : http://localhost:6670/v1
echo   - Infinity [BGE-M3 + Reranker] : http://localhost:6671/v1
echo   - DocReader Parser             : Cổng 6672
echo ================================================================================
echo   Mở trình duyệt truy cập ngay: http://localhost:6666
echo ================================================================================
pause
