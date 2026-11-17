@echo off
chcp 65001 >nul
echo ================================================================================
echo    ĐANG DỪNG HỆ THỐNG VÀ GIẢI PHÓNG TOÀN BỘ CỔNG (7666 - 7672 & 6666 - 6672)...
echo ================================================================================

REM 1. Dừng và gỡ bỏ toàn bộ container qua Docker Compose
docker compose -f docker-compose.offline-24gb.yml down --remove-orphans >nul 2>&1

REM 2. Cưỡng chế dừng sạch mọi container nếu còn sót ngầm
docker stop WeKnora-frontend WeKnora-app WeKnora-postgres WeKnora-redis WeKnora-vllm-qwen25-vl WeKnora-infinity-models WeKnora-docreader >nul 2>&1
docker rm -f WeKnora-frontend WeKnora-app WeKnora-postgres WeKnora-redis WeKnora-vllm-qwen25-vl WeKnora-infinity-models WeKnora-docreader >nul 2>&1

REM 3. Quét và giải phóng cưỡng chế các tiến trình nếu còn chiếm dụng cổng
for %%P in (7666 7667 7668 7669 7670 7671 7672 6666 6667 6668 6669 6670 6671 6672) do (
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr /r /c:":%%P "') do (
        if not "%%a"=="0" if not "%%a"=="" (
            taskkill /F /PID %%a >nul 2>&1
        )
    )
)

echo ================================================================================
echo    [✓] ĐÃ DỪNG TOÀN BỘ DỊCH VỤ VÀ ĐÓNG HẾT CÁC CỔNG THÀNH CÔNG:
echo        • Cổng 7666 (Web Frontend) : ĐÃ ĐÓNG / GIẢI PHÓNG
echo        • Cổng 7667 (Backend App)  : ĐÃ ĐÓNG / GIẢI PHÓNG
echo        • Cổng 7668 (PostgreSQL)   : ĐÃ ĐÓNG / GIẢI PHÓNG
echo        • Cổng 7669 (Redis)        : ĐÃ ĐÓNG / GIẢI PHÓNG
echo        • Cổng 7670 (vLLM)         : ĐÃ ĐÓNG / GIẢI PHÓNG
echo        • Cổng 7671 (Infinity)     : ĐÃ ĐÓNG / GIẢI PHÓNG
echo        • Cổng 7672 (DocReader)    : ĐÃ ĐÓNG / GIẢI PHÓNG
echo ================================================================================
pause
