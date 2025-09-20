@echo off
echo.
echo ===============================================
echo   Whisper MP4 文字起こしツール - セットアップ
echo ===============================================
echo.

echo FFmpegのインストール状況を確認中...
ffmpeg -version >nul 2>&1
if %errorlevel%==0 (
    echo ✅ FFmpegがインストールされています
    goto :run_program
) else (
    echo ❌ FFmpegが見つかりません
)

echo.
echo FFmpegのインストールが必要です。以下の手順に従ってください：
echo.
echo 1. https://ffmpeg.org/download.html にアクセス
echo 2. Windows用のFFmpegをダウンロード
echo 3. ダウンロードしたファイルを展開
echo 4. binフォルダのパスを環境変数PATHに追加
echo.
echo または、Chocolateyを使用してインストール：
echo    choco install ffmpeg
echo.
echo Wingetを使用してインストール：
echo    winget install ffmpeg
echo.

pause
exit /b 1

:run_program
echo.
echo 使用方法:
echo.
echo   基本使用:
echo     python whisper_transcribe.py your_video.mp4
echo.
echo   高精度モデル使用:
echo     python whisper_transcribe.py your_video.mp4 --model medium
echo.
echo   利用可能なモデル: tiny, base, small, medium, large
echo.
pause