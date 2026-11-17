@echo off
chcp 65001 >nul
echo ================================================================================
echo    CÔNG CỤ ĐỒNG BỘ CƯỠNG CHẾ BẢN MỚI NHẤT TỪ GITHUB (RESET & PULL)
echo ================================================================================
echo.
echo [+] Kiểm tra remote GitHub...
git remote -v
echo.
echo [+] Đang fetch toàn bộ dữ liệu mới nhất từ origin main...
git fetch origin main
echo [+] Đang ghi đè sạch sẽ (git reset --hard origin/main) để cập nhật 100%%...
git reset --hard origin/main
echo.
echo ================================================================================
echo [✓] PHIÊN BẢN HIỆN TẠI TRÊN MÁY CỦA BẠN LÀ:
git log -1 --oneline
echo ================================================================================
echo Đã cập nhật xong toàn bộ code mới nhất, sửa lỗi vLLM OCR và giao diện RAG AI!
echo Bây giờ bạn hãy chạy lại file: start_offline.bat
echo ================================================================================
pause
