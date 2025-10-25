import re
import json
import sys
from datetime import datetime
from pathlib import Path
from statistics import mean, stdev

def parse_log_contents(text: str) -> list[float]:
    """1つのログから cycleごとの経過時間[ms] を抽出"""
    start_times = {}
    durations_ms = []
    ts_fmt = "%Y/%m/%d %H:%M:%S.%f"

    start_re = re.compile(
        r'^\[(?P<ts>[\d/ :\.]+)\]\s+Start upload test data \(cycle:(?P<cyc>\d+)\)'
    )
    finish_re = re.compile(
        r'^\[(?P<ts>[\d/ :\.]+)\]\s+Finish upload test data \(cycle:(?P<cyc>\d+)\)'
    )

    for line in text.splitlines():
        m_start = start_re.match(line)
        if m_start:
            cyc = int(m_start.group("cyc"))
            ts = datetime.strptime(m_start.group("ts"), ts_fmt)
            start_times[cyc] = ts
            continue

        m_finish = finish_re.match(line)
        if m_finish:
            cyc = int(m_finish.group("cyc"))
            ts_finish = datetime.strptime(m_finish.group("ts"), ts_fmt)
            if cyc in start_times:
                delta = ts_finish - start_times[cyc]
                durations_ms.append(delta.total_seconds() * 1000.0)

    return durations_ms


def parse_filename(fname: str) -> tuple[str, str]:
    """gentags-<FILESIZE>-<BLOCKSIZE>.log から FILESIZE, BLOCKSIZE を抽出"""
    m = re.match(r'^gentags-(?P<fsize>[^-]+)-(?P<bsize>[^.]+)\.log$', Path(fname).name)
    if not m:
        raise ValueError(f"unexpected filename format: {fname}")
    return m.group("fsize"), int(m.group("bsize"))

def calc_growth_stats(graph_data: dict) -> dict:
    growth_stats = {}
    for fsize, blocks in graph_data.items():
        # ブロック数を数値ソート
        sorted_blocks = sorted(blocks.keys(), key=lambda k: int(k))

        # 隣接する平均値の差分を計算
        deltas = [
            blocks[sorted_blocks[i + 1]]["avg_ms"] - blocks[sorted_blocks[i]]["avg_ms"]
            for i in range(len(sorted_blocks) - 1)
        ]

        if deltas:
            avg_delta = mean(deltas)
            std_delta = stdev(deltas) if len(deltas) > 1 else 0.0
            growth_stats[fsize] = {
                "avg_increase_ms": avg_delta,
                "std_increase_ms": std_delta
            }
    return growth_stats


def main(log_paths: list[Path]) -> dict:
    """複数ログを解析して FILESIZE→BLOCKSIZE→{avg_ms, std_ms} の辞書を作成"""
    result: dict[str, dict[str, dict[str, float]]] = {}

    for p in log_paths:
        fsize, bsize = parse_filename(p.name)
        text = p.read_text(encoding="utf-8")
        durations = parse_log_contents(text)
        if not durations:
            continue

        avg_ms = mean(durations)
        std_ms = stdev(durations) if len(durations) > 1 else 0.0

        result.setdefault(fsize, {})[bsize] = {
            "avg_ms": avg_ms,
            "std_ms": std_ms
        }

    return result

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python make_graph_data.py <input_logdir> <output_json>")
        sys.exit(1)

    log_dir = Path(sys.argv[1])
    out_dir = Path(sys.argv[2])

    if not log_dir.exists():
        print(f"Error: directory not found -> {log_dir}")
        sys.exit(1)

    logs = sorted(log_dir.glob("gentags-*-*.log"))
    if not logs:
        print(f"No log files found in {log_dir}")
        sys.exit(0)

    graph_data = main(logs)
    with open(out_dir / "graph-data.json", "w", encoding="utf-8") as f:
        json.dump(graph_data, f, indent=2, sort_keys=True, ensure_ascii=False)

    growth_stats = calc_growth_stats(graph_data)
    with open(out_dir / "growth-stats.json", "w", encoding="utf-8") as f:
        json.dump(growth_stats, f, indent=2, sort_keys=True, ensure_ascii=False)
