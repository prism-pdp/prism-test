#!/usr/bin/env python3
import json
import re
import argparse
import matplotlib.pyplot as plt
from collections import defaultdict

# ===== グローバル設定 =====
FIG_SIZE = (8, 6)
FONT_SIZE = 12
GRID_STYLE = True

# ===== データセット1用スタイル設定 =====
STYLE_MAP_1 = {
    "1G":  {"color": "#017b4a", "linestyle": "-", "linewidth": 2.0},
    "2G":  {"color": "#fcb500", "linestyle": "-", "linewidth": 2.0},
    "3G":  {"color": "#093d9e", "linestyle": "-", "linewidth": 2.0},
}

# ===== データセット2用スタイル設定 =====
STYLE_MAP_2 = {
    "1G":  {"color": "#017b4a", "linestyle": "--", "linewidth": 2.0},
    "2G":  {"color": "#fcb500", "linestyle": "--", "linewidth": 2.0},
    "3G":  {"color": "#093d9e", "linestyle": "--", "linewidth": 2.0},
}

# ===== データ読み込み関数 =====
def load_data(json_path):
    with open(json_path, 'r') as f:
        data = json.load(f)
    pattern = re.compile(r"auditing-(?P<filesize>[^-]+)-(?P<blocknum>\d+)\.log")
    grouped = defaultdict(list)
    for item in data.get("ProcTime", []):
        m = pattern.match(item["Name"])
        if not m:
            continue
        filesize = m.group("filesize")
        blocknum = int(m.group("blocknum"))
        mean = float(item["Mean"])
        grouped[filesize].append((blocknum, mean))
    return grouped

# ===== グラフ描画関数 =====
def plot_grouped_data(grouped, style_map, label_prefix=""):
    for filesize, values in grouped.items():
        if filesize not in style_map:
            print(f"⚠️ スキップ: {label_prefix} の未定義ファイルサイズ {filesize}")
            continue
        values.sort(key=lambda x: x[0])
        x, y = zip(*values)
        style = style_map[filesize]
        plt.plot(
            x, y,
            label=f"{label_prefix} {filesize}",
            **style
        )

# ===== メイン関数 =====
def main():
    parser = argparse.ArgumentParser(
        description="Compare auditing performance (Mean vs BlockNum) for two datasets."
    )
    parser.add_argument("input_json1", help="入力JSONファイル1 (例: auditing_results_A.json)")
    parser.add_argument("input_json2", help="入力JSONファイル2 (例: auditing_results_B.json)")
    parser.add_argument("output_svg", help="出力SVGファイル (例: auditing_compare.svg)")
    args = parser.parse_args()

    # データ読み込み
    grouped1 = load_data(args.input_json1)
    grouped2 = load_data(args.input_json2)

    plt.figure(figsize=FIG_SIZE)

    # 1つ目のデータセット
    plot_grouped_data(grouped1, STYLE_MAP_1, label_prefix="[A]")

    # 2つ目のデータセット
    plot_grouped_data(grouped2, STYLE_MAP_2, label_prefix="[B]")

    # 見た目設定
    plt.xlabel("Block Number", fontsize=FONT_SIZE)
    plt.ylabel("Mean (ms)", fontsize=FONT_SIZE)
    plt.legend(title="File Size / Dataset")
    if GRID_STYLE:
        plt.grid(True, linestyle="--", alpha=0.6)
    plt.tight_layout()
    plt.savefig(args.output_svg, format="svg")
    print(f"✅ SVGファイルを出力しました: {args.output_svg}")

if __name__ == "__main__":
    main()
