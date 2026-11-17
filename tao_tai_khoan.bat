@echo off
chcp 65001 >nul
echo ================================================================================
echo    KHỞI TẠO TÀI KHOẢN NGƯỜI DÙNG RAG AI (3 KHÔNG GIAN LÀM VIỆC RIÊNG BIỆT)
echo ================================================================================
echo.
echo [+] Đang kết nối tới Backend Core (cổng 7667)...

echo [+] Đang tạo tài khoản phuongdong...
curl -s -X POST http://localhost:7667/api/v1/auth/register ^
  -H "Content-Type: application/json" ^
  -d "{\"username\":\"phuongdong\",\"email\":\"phuongdong@vbsp.local\",\"password\":\"itcn28ht\"}" >nul 2>&1

echo [+] Đang tạo tài khoản vbspnghean...
curl -s -X POST http://localhost:7667/api/v1/auth/register ^
  -H "Content-Type: application/json" ^
  -d "{\"username\":\"vbspnghean\",\"email\":\"vbspnghean@vbsp.local\",\"password\":\"itcn28ht\"}" >nul 2>&1

echo.
echo ================================================================================
echo    DANH SÁCH TÀI KHOẢN ĐĂNG NHẬP:
echo ================================================================================
echo   1. Quản trị viên : User: admin       ^| Pass: AdminPassword123!
echo   2. Người dùng 1  : User: phuongdong  ^| Pass: itcn28ht
echo   3. Người dùng 2  : User: vbspnghean  ^| Pass: itcn28ht
echo ================================================================================
echo   * Lưu ý: Mỗi tài khoản sở hữu một Không gian làm việc (Workspace) hoàn toàn
echo     độc lập, không nhìn thấy tài liệu và lịch sử hỏi đáp của nhau.
echo ================================================================================
echo.
pause
