#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
テスト用のサンプルMP4ファイルを生成するスクリプト

使用方法:
    python create_sample_mp4.py [output_file.mp4]

このスクリプトは短い音声付きの動画ファイルを生成します。
実際のテストには、実際のMP4ファイルを用意することをお勧めします。
"""

import os
import argparse
import numpy as np
from moviepy.editor import VideoClip, AudioClip, CompositeVideoClip, TextClip
import logging

# ログ設定
logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)


def create_sample_video(output_path="sample_video.mp4", duration=10):
    """
    テスト用のサンプル動画を作成

    Args:
        output_path (str): 出力ファイルパス
        duration (int): 動画の長さ（秒）
    """
    logger.info(f"サンプル動画を作成中: {output_path} (時間: {duration}秒)")

    try:
        # テキストクリップを作成
        text_clips = []

        # 複数のテキストセグメントを作成
        segments = [
            {"text": "こんにちは、これはテスト用の動画です", "start": 0, "duration": 3},
            {
                "text": "Whisperの文字起こし機能をテストしています",
                "start": 3,
                "duration": 4,
            },
            {"text": "この動画は自動生成されました", "start": 7, "duration": 3},
        ]

        for segment in segments:
            text_clip = (
                TextClip(
                    segment["text"],
                    fontsize=50,
                    color="white",
                    font="Arial",
                    size=(800, 600),
                )
                .set_position("center")
                .set_start(segment["start"])
                .set_duration(segment["duration"])
            )

            text_clips.append(text_clip)

        # 背景色のクリップを作成
        def make_frame(t):
            # 時間に応じて色を変える
            color_intensity = int(100 + 50 * np.sin(t))
            return np.full((600, 800, 3), [color_intensity, 50, 100], dtype=np.uint8)

        background = VideoClip(make_frame, duration=duration)

        # テキストと背景を合成
        video = CompositeVideoClip([background] + text_clips)

        # 音声を生成（簡単なトーン）
        def generate_audio(t):
            # 基本周波数440Hz（ラの音）にノイズを加える
            frequency = 440
            tone = np.sin(2 * np.pi * frequency * t)

            # テキストの内容に応じて周波数を変える
            if t < 3:
                frequency = 440  # ラ
            elif t < 7:
                frequency = 523  # ド
            else:
                frequency = 659  # ミ

            tone = 0.1 * np.sin(2 * np.pi * frequency * t)  # 音量を下げる
            return np.array([tone, tone])  # ステレオ

        audio = AudioClip(generate_audio, duration=duration)

        # 音声を動画に追加
        final_video = video.set_audio(audio)

        # ファイルに書き出し
        final_video.write_videofile(
            output_path,
            fps=24,
            audio_codec="aac",
            codec="libx264",
            verbose=False,
            logger=None,
        )

        logger.info(f"サンプル動画の作成が完了しました: {output_path}")

    except Exception as e:
        logger.error(f"サンプル動画の作成に失敗しました: {e}")
        raise


def create_audio_sample(output_path="sample_audio.mp4", duration=10):
    """
    音声のみのサンプルファイルを作成（文字起こしテスト用）

    Args:
        output_path (str): 出力ファイルパス
        duration (int): 音声の長さ（秒）
    """
    logger.info(f"音声サンプルを作成中: {output_path} (時間: {duration}秒)")

    try:
        # 音声信号を生成（複数の周波数を組み合わせて擬似的な音声を作成）
        def generate_speech_like_audio(t):
            # 基本的な音声パターンを模擬
            if t < 3:
                # 「こんにちは」の部分
                freq1, freq2 = 300, 800
            elif t < 7:
                # 「Whisperのテスト」の部分
                freq1, freq2 = 250, 1200
            else:
                # 「自動生成」の部分
                freq1, freq2 = 350, 900

            # 複数の周波数を重ねて音声らしく
            tone1 = 0.05 * np.sin(2 * np.pi * freq1 * t)
            tone2 = 0.03 * np.sin(2 * np.pi * freq2 * t)
            noise = 0.01 * np.random.random()  # 軽いノイズ

            signal = tone1 + tone2 + noise
            return np.array([signal, signal])  # ステレオ

        # 音声クリップを作成
        audio = AudioClip(generate_speech_like_audio, duration=duration)

        # 黒い画面の動画を作成
        def make_black_frame(t):
            return np.zeros((480, 640, 3), dtype=np.uint8)

        video = VideoClip(make_black_frame, duration=duration)

        # 音声を動画に追加
        final_video = video.set_audio(audio)

        # ファイルに書き出し
        final_video.write_videofile(
            output_path,
            fps=24,
            audio_codec="aac",
            codec="libx264",
            verbose=False,
            logger=None,
        )

        logger.info(f"音声サンプルの作成が完了しました: {output_path}")

    except Exception as e:
        logger.error(f"音声サンプルの作成に失敗しました: {e}")
        raise


def main():
    """メイン関数"""
    parser = argparse.ArgumentParser(
        description="テスト用のサンプルMP4ファイルを生成します"
    )
    parser.add_argument(
        "output_file",
        nargs="?",
        default="sample_video.mp4",
        help="出力MP4ファイル（デフォルト: sample_video.mp4）",
    )
    parser.add_argument(
        "--duration", type=int, default=10, help="動画の長さ（秒）（デフォルト: 10）"
    )
    parser.add_argument(
        "--audio-only", action="store_true", help="音声のみのサンプルを作成"
    )

    args = parser.parse_args()

    try:
        if args.audio_only:
            create_audio_sample(args.output_file, args.duration)
        else:
            create_sample_video(args.output_file, args.duration)

        print(f"✅ サンプルファイルが作成されました: {args.output_file}")
        print(
            f"📁 ファイルサイズ: {os.path.getsize(args.output_file) / (1024*1024):.2f} MB"
        )
        print("\n🔧 使用方法:")
        print(f"python whisper_transcribe.py {args.output_file}")

    except Exception as e:
        logger.error(f"処理中にエラーが発生しました: {e}")
        print("❌ サンプルファイルの作成に失敗しました")
        print("\n💡 実際のMP4ファイルを使用してテストすることをお勧めします")


if __name__ == "__main__":
    main()
