@echo off
chcp 65001 >nul
echo ================================================================================
echo    DỌN DẸP AN TOÀN DOCKER CHO RIÊNG DỰ ÁN WEKNORA (KHÔNG ẢNH HƯỞNG DỰ ÁN KHÁC)
echo ================================================================================
echo.
echo [!] Lưu ý: Script này CHỈ xóa:
echo     1. Các image dangling (^<none^>) sinh ra khi rebuild app/docreader
echo     2. Ảnh tạm trong volume /tmp/docreader của container WeKnora-docreader
echo     3. BuildKit cache cũ hơn 24 giờ
echo     KHÔNG đụng tới container, volume, hay image đang dùng của các dự án khác!
echo.

docker info >nul 2>&1
if errorlevel 1 (
    echo [X] Docker Desktop chưa chạy. Hãy mở Docker Desktop trước.
    pause
    exit /b 1
)

echo [1/4] Dung lượng Docker TRƯỚC khi dọn:
docker system df
echo.

echo [2/4] Xóa các image ^<none^> (dangling layers từ những lần rebuild cũ)...
docker image prune -f
echo.

echo [3/4] Dọn ảnh tạm render PDF trong WeKnora-docreader (/tmp/docreader)...
docker inspect WeKnora-docreader >nul 2>&1
if not errorlevel 1 (
    docker exec WeKnora-docreader sh -c "rm -rf /tmp/docreader/*" 2>nul
    echo [✓] Đã dọn /tmp/docreader.
) else (
    echo [-] Container WeKnora-docreader chưa chạy, bỏ qua bước này.
)
echo.

echo [4/4] Dọn BuildKit cache cũ hơn 24h...
docker builder prune --filter "until=24h" -f
echo.

echo ================================================================================
echo    Dung lượng Docker SAU khi dọn:
docker system df
echo.
echo    [MẸO THU HỒI Ổ C TRÊN WINDOWS WSL2 - AN TOÀN CHO MỌI DỰ ÁN]:
echo    1. Quit Docker Desktop (chuột phải icon Docker ở taskbar -^> Quit Docker Desktop)
echo    2. Mở PowerShell (Run as Administrator) và chạy:
echo         wsl --shutdown
echo         Optimize-VHD -Path "$env:LOCALAPPDATA\Docker\wsl\disk\docker_data.vhdx" -Mode Full
echo       (Nếu Windows Home không có Optimize-VHD, dùng diskpart compact vdisk)
echo    3. Mở lại Docker Desktop.
echo ================================================================================
pause
