#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
MP4動画ファイルから音声を抽出してWhisperを使用した文字起こしを行うプログラム

使用方法:
    python whisper_transcribe.py input.mp4 [output.txt]

引数:
    input.mp4: 文字起こしを行うMP4ファイル
    output.txt: 出力テキストファイル（省略可能、デフォルト: input_transcript.txt）
"""

import os
import sys
import argparse
import tempfile
import subprocess
from pathlib import Path
import whisper
import librosa
import soundfile as sf
import logging

# ログ設定
logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)


class MP4Transcriber:
    """MP4ファイルの文字起こしを行うクラス"""

    def __init__(self, model_name="base"):
        """
        初期化

        Args:
            model_name (str): Whisperモデル名 (tiny, base, small, medium, large)
        """
        self.model_name = model_name
        self.model = None

    def load_model(self):
        """Whisperモデルを読み込み"""
        logger.info(f"Whisperモデル '{self.model_name}' を読み込み中...")
        try:
            self.model = whisper.load_model(self.model_name)
            logger.info("モデルの読み込みが完了しました")
        except Exception as e:
            logger.error(f"モデルの読み込みに失敗しました: {e}")
            raise

    def extract_audio(self, video_path, audio_path):
        """
        MP4ファイルから音声を抽出

        Args:
            video_path (str): 入力MP4ファイルパス
            audio_path (str): 出力音声ファイルパス
        """
        logger.info(f"音声を抽出中: {video_path} -> {audio_path}")
        try:
            # Librosaを使用して音声を抽出・変換
            y, sr = librosa.load(video_path, sr=16000)

            # WAV形式で保存
            sf.write(audio_path, y, sr)

            logger.info("音声の抽出が完了しました")
        except Exception as e:
            logger.error(f"音声の抽出に失敗しました: {e}")
            logger.info(
                "Librosaでの処理に失敗しました。FFmpegが必要な可能性があります。"
            )
            raise

    def transcribe_audio(self, audio_path):
        """
        音声ファイルを文字起こし

        Args:
            audio_path (str): 音声ファイルパス

        Returns:
            dict: Whisperの転写結果
        """
        logger.info(f"文字起こし中: {audio_path}")
        try:
            if self.model is None:
                self.load_model()

            result = self.model.transcribe(audio_path, language="ja")
            logger.info("文字起こしが完了しました")
            return result
        except Exception as e:
            logger.error(f"文字起こしに失敗しました: {e}")
            raise

    def save_transcript(self, result, output_path, include_timestamps=True):
        """
        転写結果をファイルに保存

        Args:
            result (dict): Whisperの転写結果
            output_path (str): 出力ファイルパス
            include_timestamps (bool): タイムスタンプを含めるかどうか
        """
        logger.info(f"転写結果を保存中: {output_path}")
        try:
            with open(output_path, "w", encoding="utf-8") as f:
                # 全体のテキスト
                f.write("=== 完全な転写テキスト ===\n")
                f.write(result["text"])
                f.write("\n\n")

                if include_timestamps:
                    # セグメント毎の詳細（タイムスタンプ付き）
                    f.write("=== セグメント毎の詳細 ===\n")
                    for segment in result["segments"]:
                        start_time = format_timestamp(segment["start"])
                        end_time = format_timestamp(segment["end"])
                        f.write(f"[{start_time} - {end_time}] {segment['text']}\n")

            logger.info("転写結果の保存が完了しました")
        except Exception as e:
            logger.error(f"転写結果の保存に失敗しました: {e}")
            raise

    def transcribe_mp4(self, video_path, output_path=None, include_timestamps=True):
        """
        MP4ファイルを文字起こし（メイン処理）

        Args:
            video_path (str): 入力MP4ファイルパス
            output_path (str): 出力テキストファイルパス
            include_timestamps (bool): タイムスタンプを含めるかどうか

        Returns:
            dict: Whisperの転写結果
        """
        # 入力ファイルの存在確認
        if not os.path.exists(video_path):
            raise FileNotFoundError(f"ファイルが見つかりません: {video_path}")

        # 出力ファイル名の設定
        if output_path is None:
            video_name = Path(video_path).stem
            output_path = f"{video_name}_transcript.txt"

        try:
            # Whisperで直接MP4ファイルを処理
            logger.info(f"Whisperで直接ファイルを処理中: {video_path}")
            if self.model is None:
                self.load_model()

            # Whisperに直接MP4を渡す
            result = self.model.transcribe(video_path, language="ja")

            # 結果を保存
            self.save_transcript(result, output_path, include_timestamps)

            return result

        except Exception as e:
            logger.error(f"直接処理に失敗しました: {e}")
            logger.error("MP4ファイルの直接処理ができませんでした。")
            logger.info("解決策:")
            logger.info("1. FFmpegをインストールしてPATHに追加")
            logger.info(
                "2. または、VLCやAudacityなどで音声ファイル(.wav, .mp3)を手動で抽出"
            )
            logger.info("3. 抽出した音声ファイルでこのスクリプトを実行")

            # 詳細なエラー情報を提供
            file_ext = Path(video_path).suffix.lower()
            if file_ext in [".wav", ".mp3", ".flac", ".m4a"]:
                logger.info(f"音声ファイル（{file_ext}）として再試行...")
                result = self.model.transcribe(video_path, language="ja")
                self.save_transcript(result, output_path, include_timestamps)
                return result
            else:
                raise Exception(
                    f"MP4ファイルの処理には外部ツールが必要です。音声ファイル(.wav, .mp3など)に変換してから再実行してください。"
                )


def format_timestamp(seconds):
    """
    秒数をHH:MM:SS形式に変換

    Args:
        seconds (float): 秒数

    Returns:
        str: HH:MM:SS形式の時間文字列
    """
    hours = int(seconds // 3600)
    minutes = int((seconds % 3600) // 60)
    secs = int(seconds % 60)
    return f"{hours:02d}:{minutes:02d}:{secs:02d}"


def main():
    """メイン関数"""
    parser = argparse.ArgumentParser(
        description="MP4ファイルから音声を抽出してWhisperで文字起こしを行います"
    )
    parser.add_argument("input_file", help="入力MP4ファイル")
    parser.add_argument(
        "output_file", nargs="?", help="出力テキストファイル（省略可能）"
    )
    parser.add_argument(
        "--model",
        default="base",
        choices=["tiny", "base", "small", "medium", "large"],
        help="Whisperモデル（デフォルト: base）",
    )
    parser.add_argument(
        "--no-timestamps", action="store_true", help="タイムスタンプを出力しない"
    )

    args = parser.parse_args()

    try:
        # 転写処理を実行
        transcriber = MP4Transcriber(model_name=args.model)
        result = transcriber.transcribe_mp4(
            video_path=args.input_file,
            output_path=args.output_file,
            include_timestamps=not args.no_timestamps,
        )

        # 結果の表示
        output_file = args.output_file or f"{Path(args.input_file).stem}_transcript.txt"
        print(f"✅ 文字起こしが完了しました！")
        print(f"📁 出力ファイル: {output_file}")
        print(f"📝 検出された言語: {result.get('language', 'unknown')}")
        print(f"⏱️  処理時間: 約{len(result['segments'])}セグメント")

        # 短いプレビューを表示
        preview_text = (
            result["text"][:200] + "..."
            if len(result["text"]) > 200
            else result["text"]
        )
        print(f"📄 プレビュー: {preview_text}")

    except Exception as e:
        logger.error(f"処理中にエラーが発生しました: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
