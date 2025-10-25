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

# ===== FILESIZEごとのスタイル設定 =====
STYLE_MAP = {
    "1G": {"color": "blue",   "linestyle": "-",  "marker": "o", "linewidth": 2.0, "markersize": 6},
    "2G": {"color": "green",  "linestyle": "--", "marker": "s", "linewidth": 2.0, "markersize": 6},
    "3G": {"color": "red",    "linestyle": "-.", "marker": "D", "linewidth": 2.0, "markersize": 6}
}

# ===== 関数定義 =====
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

def plot_graph(grouped, output_path):
    plt.figure(figsize=FIG_SIZE)
    for filesize, values in grouped.items():
        if filesize not in STYLE_MAP:
            print(f"⚠️ スキップ: 未定義のファイルサイズ {filesize}")
            continue
        style = STYLE_MAP[filesize]
        values.sort(key=lambda x: x[0])
        x, y = zip(*values)
        plt.plot(
            x, y, label=f"{filesize}",
            color=style["color"],
            linestyle=style["linestyle"],
            marker=style["marker"],
            linewidth=style["linewidth"],
            markersize=style["markersize"]
        )

    plt.xlabel("Block Number", fontsize=FONT_SIZE)
    plt.ylabel("Mean (ms)", fontsize=FONT_SIZE)
    plt.title("Auditing Performance by Block Number", fontsize=FONT_SIZE + 2)
    plt.legend(title="File Size")
    if GRID_STYLE:
        plt.grid(True, linestyle="--", alpha=0.6)
    plt.tight_layout()
    plt.savefig(output_path, format='svg')
    print(f"✅ SVGファイルを出力しました: {output_path}")

# ===== メイン処理 =====
def main():
    parser = argparse.ArgumentParser(
        description="Plot auditing performance (Mean vs BlockNum) by FileSize."
    )
    parser.add_argument("input_json", help="入力JSONファイル (例: auditing_results.json)")
    parser.add_argument("output_svg", help="出力SVGファイル (例: auditing_graph.svg)")
    args = parser.parse_args()

    grouped = load_data(args.input_json)
    if not grouped:
        print("❌ データが見つかりません。JSON形式やファイル内容を確認してください。")
        return

    plot_graph(grouped, args.output_svg)

if __name__ == "__main__":
    main()
