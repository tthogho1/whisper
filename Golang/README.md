# Gladia Transcribe Go

MP4 音声ファイルからテキスト文字起こしを行う Go 言語版ツールです。Gladia API を使用してクラウド転写を実行し、AWS S3 バッチ処理にも対応しています。

## 特徴

- **高速処理**: Go 言語による高性能な実装
- **モジュール設計**: 機能ごとに分離された保守性の高い設計
- **シンプル操作**: コマンドライン一発で転写完了
- **クラウド API**: Gladia API による高精度転写
- **セキュア**: .env 方式で API キー管理
- **クロスプラットフォーム**: Windows、Linux、macOS 対応
- **S3 バッチ処理**: AWS S3 バケット内の複数 MP4 ファイルを一括処理

## プロジェクト構造

```
Golang/
├── main.go              # エントリーポイントとCLI定義
├── client/
│   ├── gladia.go        # Gladia API クライアント
│   └── s3.go           # AWS S3 クライアント
├── models/
│   └── types.go        # 構造体定義
├── processor/
│   └── batch.go        # バッチ処理ロジック
├── go.mod              # Go モジュール定義
├── .env.example        # 環境変数テンプレート
└── README.md           # このファイル
```

## セットアップ

### 1. 依存関係のインストール

```bash
go mod download
```

### 2. API キー設定

`.env.example`を`.env`にコピーして、Gladia API キーと AWS 設定を設定：

```bash
cp .env.example .env
```

`.env`ファイルを編集：

```
# 必須: Gladia API設定
GLADIA_API_KEY=your_actual_api_key_here

# オプション: AWS S3バッチ処理用設定
AWS_ACCESS_KEY_ID=your_aws_access_key_id
AWS_SECRET_ACCESS_KEY=your_aws_secret_access_key
AWS_REGION=us-east-1
AWS_INPUT_BUCKET=your-input-bucket-name
AWS_OUTPUT_BUCKET=your-output-bucket-name
```

### 3. ビルド

```bash
go build -o gladia-transcribe main.go
```

## 使用方法

### 単一ファイル処理

```bash
./gladia-transcribe input.mp4
```

### S3 バッチ処理

AWS S3 バケット内のすべての MP4 ファイルを一括処理：

```bash
./gladia-transcribe batch
```

### オプション

```bash
./gladia-transcribe --help
```

出力例：

```
MP4音声ファイルをGladia APIを使用してテキストに転写します

Usage:
  gladia-transcribe [flags] <input_file|batch>

Flags:
      --api-key string         Gladia API キー
      --language string        言語設定 (ja|en|zh|ko|es|fr|de|auto) (default "ja")
      --no-detect-language     自動言語検出を無効化
      --no-subtitles           字幕ファイル生成を無効化
      --output string          出力ファイル名（拡張子なし）
  -h, --help                   help for gladia-transcribe
```

### 使用例

#### 単一ファイル処理

```bash
# MP4ファイルを転写
./gladia-transcribe sample.mp4

# 出力ファイルが自動生成されます
# - sample_transcription.txt (転写テキスト)
# - sample_transcription.json (詳細情報)
```

#### S3 バッチ処理

```bash
# S3バケット内のすべてのMP4ファイルを処理
./gladia-transcribe batch

# 処理結果:
# - 入力: s3://input-bucket/*.mp4
# - 出力: s3://output-bucket/*_transcription.json
```

## コード構成

### Models (models/types.go)

- データ構造体の定義
- API レスポンス形式
- S3 ファイル情報

### Clients (client/)

- **gladia.go**: Gladia API との通信を担当
- **s3.go**: AWS S3 との通信を担当

### Processor (processor/)

- **batch.go**: S3 バッチ処理のメインロジック

### Main (main.go)

- CLI インターフェースの定義
- 引数解析とルーティング

## 出力ファイル

### 単一ファイル処理

- **`.txt`ファイル**: プレーンテキスト形式の転写結果
- **`.json`ファイル**: タイムスタンプ付きの詳細転写情報

### S3 バッチ処理

- **S3 JSON ファイル**: `s3://output-bucket/filename_transcription.json` 形式で保存

## AWS S3 バッチ処理の流れ

1. **入力バケット内の MP4 ファイル一覧を取得**
2. **各 MP4 ファイルを一時的にローカルにダウンロード**
3. **Gladia API で文字起こし実行**
4. **結果を JSON 形式で出力バケットに保存**
5. **一時ファイルを削除**

## エラー対処

### よくあるエラー

1. **API キー未設定**

   ```
   Error: GLADIA_API_KEY not found in .env file
   ```

   → `.env`ファイルに正しい API キーを設定してください

2. **AWS 設定未完了**

   ```
   Error: AWS credentials not found in environment variables
   ```

   → `.env`ファイルに AWS 認証情報を設定してください

3. **ファイルが見つからない**

   ```
   Error: file not found: input.mp4
   ```

   → ファイルパスを確認してください

4. **ビルドエラー**
   ```
   go build: cannot find module
   ```
   → `go mod download` を実行してください

## 開発

### 新機能の追加

1. **新しいクライアント**: `client/` ディレクトリに追加
2. **新しいモデル**: `models/types.go` に構造体を追加
3. **新しい処理**: `processor/` ディレクトリに追加

### テスト実行

```bash
go test ./...
```

### ビルド（クロスプラットフォーム）

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o gladia-transcribe.exe main.go

# Linux
GOOS=linux GOARCH=amd64 go build -o gladia-transcribe main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o gladia-transcribe main.go
```

## 対応ファイル形式

- MP4 (推奨)
- MP3
- WAV
- その他の音声・動画ファイル

## システム要件

- Go 1.21 以上
- インターネット接続 (Gladia API 使用のため)
- AWS アカウント (S3 バッチ処理使用の場合)

## ライセンス

MIT License

## 作者

GitHub Copilot
