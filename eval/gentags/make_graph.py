import json
import sys
import matplotlib.pyplot as plt
from matplotlib.ticker import FuncFormatter

# =========================================
# グローバル設定：線の見た目（あとで調整しやすい用）
# =========================================
LINE_STYLES = {
    "1G": {"label": "1G", "linestyle": "-", "color": "#017b4a"},
    "2G": {"label": "2G", "linestyle": "-", "color": "#fcb500"},
    "3G": {"label": "3G", "linestyle": "-", "color": "#093d9e"},
}

FONT_SIZE = 16

def plot_data(data: dict, output_path: str) -> None:
    """
    data: {
        "1G": {
            "100": {"avg_ms": 5956.16, "std_ms": 638.21},
            "200": {"avg_ms": 6508.91, "std_ms": 31.77},
            ...
        },
        "2G": { ... },
        ...
    }
    output_path: 保存先SVGパス
    """

    fig, ax = plt.subplots()

    for filesize, series in data.items():
        if filesize not in LINE_STYLES:
            print(f"⚠️ スキップ: 未定義ファイルサイズ {filesize}")
            continue
        style = LINE_STYLES[filesize]

        # ブロック数キーをint化してソート
        blocks_sorted = sorted(series.keys(), key=lambda k: int(k))

        x_vals = [int(b) for b in blocks_sorted]
        y_avg  = [series[b]["avg_ms"] for b in blocks_sorted]
        y_std  = [series[b]["std_ms"] for b in blocks_sorted]

        ax.errorbar(
            x_vals,
            y_avg,
            yerr=y_std,
            capsize=5,          # 誤差バーの先端の横棒
            **style
        )

    ax.set_xlabel("# of Blocks", fontsize = FONT_SIZE)
    ax.set_ylabel("Time [ms]",   fontsize = FONT_SIZE)

    ax.tick_params(axis='x', labelsize=FONT_SIZE*0.8)
    ax.tick_params(axis='y', labelsize=FONT_SIZE*0.8)

    ax.grid(True, which="both", linestyle="--", alpha=0.4)

    ax.legend(title="File Size", loc="lower center", bbox_to_anchor=(0.5, 1.02), ncol=5, fontsize=FONT_SIZE*0.6)

    ax.yaxis.set_major_formatter(FuncFormatter(lambda x, _: f"{x:,.0f}"))

    plt.tight_layout()
    fig.savefig(output_path, format="svg")


def main():
    # 使い方:
    #   python plot_upload_time.py input.json output.svg
    if len(sys.argv) != 3:
        print("Usage: python make_graph.py <input_json> <output_svg>")
        sys.exit(1)

    input_json = sys.argv[1]
    output_svg = sys.argv[2]

    # JSON読み込み
    with open(input_json, "r", encoding="utf-8") as f:
        data = json.load(f)

    plot_data(data, output_svg)


if __name__ == "__main__":
    main()
