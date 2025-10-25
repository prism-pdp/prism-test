#!/usr/bin/env python3
import json
import re
import argparse
import matplotlib.pyplot as plt
from collections import defaultdict

# ===== グローバル設定 =====
FIG_SIZE = (8, 4)
FONT_SIZE = 15
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
            label=f"{label_prefix} ({filesize})",
            **style
        )

def sort_key(label):
    # 条件の種類で優先順位付け（Block Ratio が先）
    order = 1 if "Block Ratio" in label else 0

    # ファイルサイズ（例: 1G, 2G, 3G）を数値で取得
    m_size = re.search(r"\((\d+)G\)", label)
    size = int(m_size.group(1)) if m_size else 0

    # まず条件タイプでソートし、その中でサイズ順
    return (order, size)

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
    plot_grouped_data(grouped1, STYLE_MAP_1, label_prefix="# of Blocks = 100")

    # 2つ目のデータセット
    plot_grouped_data(grouped2, STYLE_MAP_2, label_prefix="Block Ratio = 0.1")

    # 見た目設定
    plt.xlabel("Test File Split Count (Number of Blocks)", fontsize=FONT_SIZE)
    plt.ylabel("Average Processing Time [ms]", fontsize=FONT_SIZE)
    plt.tick_params(axis='x', labelsize=FONT_SIZE*0.8)
    plt.tick_params(axis='y', labelsize=FONT_SIZE*0.8)
    if GRID_STYLE:
        plt.grid(True, linestyle="--", alpha=0.6)

    handles, labels = plt.gca().get_legend_handles_labels()
    print(handles)
    print(labels)
    sorted_pairs = sorted(zip(labels, handles), key=lambda x: sort_key(x[0]))
    print(sorted_pairs)
    labels, handles = zip(*sorted_pairs)
    plt.legend(handles, labels, ncol=2)

    plt.xlim(100, 1000)

    plt.tight_layout()
    plt.savefig(args.output_svg, format="svg")
    print(f"✅ SVGファイルを出力しました: {args.output_svg}")

if __name__ == "__main__":
    main()
