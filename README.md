# MP4 音声文字起こしツール (Whisper)

このプロジェクトは、OpenAI の Whisper を使用して MP4 動画ファイルから音声を抽出し、自動で文字起こしを行う Python アプリケーションです。

## 🎯 機能

- **MP4 ファイルからの音声抽出**: MoviePy を使用して MP4 ファイルから音声データを抽出
- **高精度な文字起こし**: OpenAI Whisper を使用した高品質な音声認識
- **日本語対応**: 日本語音声の文字起こしに最適化
- **タイムスタンプ付き出力**: 各セグメントの開始・終了時刻を記録
- **複数のモデルサイズ対応**: tiny, base, small, medium, large から選択可能
- **コマンドライン操作**: 簡単なコマンドラインインターフェース

## 📋 必要条件

- Python 3.8 以上
- FFmpeg（MoviePy で音声処理に必要）

## 🚀 インストール

### 1. リポジトリのクローン

```bash
git clone <repository-url>
cd whisper
```

### 2. Python ライブラリのインストール

```bash
# 仮想環境の作成（推奨）
python -m venv .venv

# 仮想環境のアクティベート
# Windows:
.venv\Scripts\activate
# macOS/Linux:
source .venv/bin/activate

# 必要なパッケージのインストール
pip install -r requirements.txt
```

### 3. FFmpeg のインストール

MoviePy が音声処理を行うために FFmpeg が必要です：

**Windows:**

- [FFmpeg 公式サイト](https://ffmpeg.org/download.html)からダウンロード
- 環境変数 PATH に追加

**macOS:**

```bash
brew install ffmpeg
```

**Ubuntu/Debian:**

```bash
sudo apt update
sudo apt install ffmpeg
```

## 📖 使用方法

### 基本的な使用方法

```bash
# MP4ファイルを文字起こし
python whisper_transcribe.py input_video.mp4

# 出力ファイル名を指定
python whisper_transcribe.py input_video.mp4 output_transcript.txt

# より高精度なモデルを使用
python whisper_transcribe.py input_video.mp4 --model medium

# タイムスタンプなしで出力
python whisper_transcribe.py input_video.mp4 --no-timestamps
```

### コマンドラインオプション

| オプション        | 説明                                          | デフォルト                    |
| ----------------- | --------------------------------------------- | ----------------------------- |
| `input_file`      | 入力 MP4 ファイル                             | 必須                          |
| `output_file`     | 出力テキストファイル                          | `{input_name}_transcript.txt` |
| `--model`         | Whisper モデル (tiny/base/small/medium/large) | `base`                        |
| `--no-timestamps` | タイムスタンプを出力しない                    | False                         |

### Whisper モデルの選択

| モデル | サイズ  | 相対速度 | 精度 | 推奨用途         |
| ------ | ------- | -------- | ---- | ---------------- |
| tiny   | 39 MB   | 最高速   | 低   | 高速プロトタイプ |
| base   | 74 MB   | 高速     | 中   | 一般的な用途     |
| small  | 244 MB  | 中速     | 高   | バランス重視     |
| medium | 769 MB  | 低速     | 高   | 高品質重視       |
| large  | 1550 MB | 最低速   | 最高 | 最高品質         |

## 📄 出力形式

生成される文字起こしファイルには以下の情報が含まれます：

```
=== 完全な転写テキスト ===
こんにちは、これはサンプルの音声です。Whisperの文字起こし機能をテストしています。

=== セグメント毎の詳細 ===
[00:00:00 - 00:00:03] こんにちは、これはサンプルの音声です。
[00:00:03 - 00:00:08] Whisperの文字起こし機能をテストしています。
```

## 🧪 テスト用サンプルファイルの作成

実際の MP4 ファイルがない場合、テスト用のサンプルファイルを生成できます：

```bash
# テキスト付きサンプル動画を作成
python create_sample_mp4.py sample_video.mp4

# 音声のみのサンプル動画を作成
python create_sample_mp4.py sample_audio.mp4 --audio-only

# 30秒のサンプル動画を作成
python create_sample_mp4.py long_sample.mp4 --duration 30
```

## 🔧 プログラムの使用例

### Python スクリプトとして使用

```python
from whisper_transcribe import MP4Transcriber

# 転写処理の実行
transcriber = MP4Transcriber(model_name="base")
result = transcriber.transcribe_mp4("input_video.mp4", "output.txt")

# 結果の確認
print(f"言語: {result['language']}")
print(f"テキスト: {result['text']}")
```

### バッチ処理

複数のファイルを一度に処理する場合：

```bash
# PowerShellの場合
Get-ChildItem "*.mp4" | ForEach-Object { python whisper_transcribe.py $_.Name }

# Bashの場合
for file in *.mp4; do python whisper_transcribe.py "$file"; done
```

## ⚡ パフォーマンスの最適化

### GPU 加速（CUDA 対応 GPU 使用時）

```bash
# CUDA対応のPyTorchをインストール
pip install torch torchvision torchaudio --index-url https://download.pytorch.org/whl/cu118
```

### メモリ使用量の削減

- より小さなモデル（tiny, base）を使用
- 長い動画は分割して処理
- 不要なプロセスを終了してメモリを確保

## 🚨 トラブルシューティング

### よくある問題

**問題**: `FFmpeg not found`エラー
**解決策**: FFmpeg をインストールし、PATH に追加してください

**問題**: メモリ不足エラー
**解決策**: より小さなモデルを使用するか、動画を分割してください

**問題**: 音声が検出されない
**解決策**: 入力ファイルに音声トラックが含まれているか確認してください

**問題**: 文字起こしの精度が低い
**解決策**: より大きなモデル（medium, large）を試してください

### ログの確認

プログラムは詳細なログを出力します。問題が発生した場合は、ログメッセージを確認してください。

## 📚 依存関係

- `openai-whisper`: OpenAI の音声認識モデル
- `moviepy`: 動画・音声処理ライブラリ
- `torch`: PyTorch ディープラーニングフレームワーク
- `transformers`: Hugging Face Transformers ライブラリ

## 📝 ライセンス

このプロジェクトは MIT ライセンスの下で公開されています。

## 🤝 貢献

バグ報告や機能追加の提案は、GitHub の Issues までお願いします。

## 📞 サポート

問題が発生した場合は、以下の情報を含めて Issue を作成してください：

- 使用している OS
- Python のバージョン
- エラーメッセージの全文
- 実行したコマンド

---

**注意**: 初回実行時は Whisper モデルのダウンロードが必要です。インターネット接続を確認してください。
