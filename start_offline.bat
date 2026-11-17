@echo off
chcp 65001 >nul
echo ================================================================================
echo    BUNG DOCKER HỆ THỐNG RAG TIẾNG VIỆT 100%% OFFLINE [1x 24GB VRAM GPU]
echo ================================================================================

REM 1. Tu dong kiem tra va cap nhat ma nguon neu may co Git
where git >nul 2>&1
if not errorlevel 1 (
    echo [+] Đang kiểm tra và đồng bộ cập nhật mới nhất từ GitHub...
    git fetch origin main
    git reset --hard origin/main
    echo [✓] Phiên bản Git hiện tại trên máy của bạn:
    git log -1 --oneline
)

REM 2. Kiem tra Docker Desktop
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

REM 3. Khoi tao hoac dong bo tep .env
if not exist .env (
    echo [+] Chưa tìm thấy .env, đang tự động tạo từ .env.offline-24gb.example...
    copy .env.offline-24gb.example .env >nul
)
findstr "6666" .env >nul 2>&1
if not errorlevel 1 (
    echo [+] Phát hiện dải cổng 6666 cũ [bị Chrome chặn ERR_UNSAFE_PORT], tự động nâng cấp lên 7666...
    powershell -Command "(Get-Content .env) -replace '666([6-9])', '766$1' -replace '667([0-2])', '767$1' | Set-Content .env" >nul 2>&1
)
REM Nang cap VLLM_MAX_MODEL_LEN len 24576 va VLLM_GPU_MEMORY_UTILIZATION len 0.60
findstr "VLLM_MAX_MODEL_LEN=8192" .env >nul 2>&1
if not errorlevel 1 (
    echo [+] Tự động cập nhật trần context VLLM_MAX_MODEL_LEN=24576 trong .env...
    powershell -Command "(Get-Content .env) -replace 'VLLM_MAX_MODEL_LEN=8192', 'VLLM_MAX_MODEL_LEN=24576' -replace 'VLLM_GPU_MEMORY_UTILIZATION=0.50', 'VLLM_GPU_MEMORY_UTILIZATION=0.60' | Set-Content .env" >nul 2>&1
)
findstr "VLLM_MAX_MODEL_LEN=16384" .env >nul 2>&1
if not errorlevel 1 (
    echo [+] Tự động cập nhật trần context VLLM_MAX_MODEL_LEN=24576 trong .env...
    powershell -Command "(Get-Content .env) -replace 'VLLM_MAX_MODEL_LEN=16384', 'VLLM_MAX_MODEL_LEN=24576' -replace 'VLLM_GPU_MEMORY_UTILIZATION=0.50', 'VLLM_GPU_MEMORY_UTILIZATION=0.60' | Set-Content .env" >nul 2>&1
)
echo [✓] Tệp .env đã sẵn sàng [dải cổng an toàn 7666-7672].

REM 4. Dam bao cac thu muc ton tai
if not exist "models_cache\huggingface" mkdir "models_cache\huggingface"
if not exist "data\files" mkdir "data\files"
echo [✓] Đã tạo thư mục lưu trữ: models_cache\huggingface và data\files.

REM 5. Khoi chay Docker Compose [kich hoat du 7/7 container]
echo [+] Đang bung Docker Compose [kich hoat du 7/7 container]...
docker compose -f docker-compose.offline-24gb.yml up -d frontend app docreader postgres redis vllm-qwen25-vl infinity-models
if errorlevel 1 goto COMPOSE_FAILED

REM 6. Cho doi mo hinh AI vLLM hoan tat nap vao VRAM GPU
echo [+] Đang kiểm tra và đợi mô hình AI vLLM nạp vào GPU (khoảng 30-90 giây)...
set /a VLLM_WAIT_COUNT=0
:WAIT_VLLM_LOOP
timeout /t 5 >nul 2>&1
set /a VLLM_WAIT_COUNT+=1
curl -s -m 2 http://localhost:7670/health >nul 2>&1
if not errorlevel 1 goto VLLM_READY
if %VLLM_WAIT_COUNT% geq 24 goto VLLM_TIMEOUT
echo     ... [%VLLM_WAIT_COUNT%/24] Đang nạp trọng số mô hình vào VRAM, vui lòng đợi...
goto WAIT_VLLM_LOOP

:VLLM_READY
echo [✓] Mô hình AI vLLM đã nạp hoàn tất và sẵn sàng phục vụ!
goto REGISTER_USERS

:VLLM_TIMEOUT
echo [!] Lưu ý: Mô hình vLLM đang tiếp tục nạp. Bạn có thể mở xem_log_ocr_llm.bat để theo dõi.

:REGISTER_USERS
REM 7. Tu dong tao 2 tai khoan nguoi dung neu chua ton tai
curl -s -X POST http://localhost:7667/api/v1/auth/register -H "Content-Type: application/json" -d "{\"username\":\"phuongdong\",\"email\":\"phuongdong@vbsp.local\",\"password\":\"itcn28ht\"}" >nul 2>&1
curl -s -X POST http://localhost:7667/api/v1/auth/register -H "Content-Type: application/json" -d "{\"username\":\"vbspnghean\",\"email\":\"vbspnghean@vbsp.local\",\"password\":\"itcn28ht\"}" >nul 2>&1
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
echo    KHỞI CHẠY HỆ THỐNG THÀNH CÔNG! [7/7 CONTAINER SẴN SÀNG]
echo ================================================================================
echo   - Giao diện Web RAG Tiếng Việt : http://localhost:7666
echo   - Backend Core API             : http://localhost:7667
echo   - PostgreSQL [ParadeDB HNSW]   : Cổng 7668
echo   - Redis Cache                  : Cổng 7669
echo   - vLLM [Qwen2.5-VL-7B OCR+LLM] : http://localhost:7670/v1
echo   - Infinity [BGE-M3 + Reranker] : http://localhost:7671/v1
echo   - DocReader Parser             : Cổng 7672
echo ================================================================================
echo    DANH SÁCH TÀI KHOẢN ĐĂNG NHẬP (3 KHÔNG GIAN LÀM VIỆC ĐỘC LẬP):
echo ================================================================================
echo   [1] Quản trị viên: User: admin       ^| Pass: AdminPassword123!
echo   [2] Người dùng 1 : User: phuongdong  ^| Pass: itcn28ht
echo   [3] Người dùng 2 : User: vbspnghean  ^| Pass: itcn28ht
echo ================================================================================
echo   Mở trình duyệt truy cập ngay: http://localhost:7666
echo ================================================================================
echo   * LƯU Ý KHI KHỞI CHẠY LẦN ĐẦU:
echo     Container vLLM và Infinity sẽ tải model (~8GB) về models_cache\huggingface.
echo     Để theo dõi tiến trình tải model vLLM (Qwen2.5-VL), mở CMD mới và gõ:
echo        docker logs -f WeKnora-vllm-qwen25-vl
echo     Để theo dõi tiến trình tải model Infinity (Embedding/Reranker), gõ:
echo        docker logs -f WeKnora-infinity-models
echo ================================================================================
pause
