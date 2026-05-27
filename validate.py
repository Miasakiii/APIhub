"""
APIHub 数据管道验证：连接 cc-switch + 解析 JSONL + 计算费用

验证 spike findings 是否能在 Go 代码中复现。
"""

import json
import os
import sqlite3
import time
from collections import Counter
from pathlib import Path

CC_SWITCH_DB = Path.home() / ".cc-switch" / "cc-switch.db"
PROJECTS_DIR = Path.home() / ".claude" / "projects"


def banner(s):
    print(f"\n{'=' * 70}\n{s}\n{'=' * 70}")


def main():
    start = time.time()

    # 1. 连接 cc-switch (read-only + busy_timeout)
    banner("1. 连接 cc-switch.db")
    uri = f"file:{CC_SWITCH_DB.as_posix()}?mode=ro&_busy_timeout=5000"
    conn = sqlite3.connect(uri, uri=True, timeout=5.0)
    conn.row_factory = sqlite3.Row
    print(f"OK: {CC_SWITCH_DB}")

    # 2. 验证 schema
    banner("2. Schema 验证")
    version = conn.execute("PRAGMA user_version").fetchone()[0]
    print(f"PRAGMA user_version = {version} (need >= 10)")
    assert version >= 10, f"user_version {version} < 10"

    required = ["providers", "proxy_request_logs", "model_pricing", "provider_endpoints"]
    for tbl in required:
        n = conn.execute(
            "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", (tbl,)
        ).fetchone()[0]
        print(f"  table '{tbl}': {'OK' if n else 'MISSING'}")
        assert n, f"table {tbl} missing"
    print("Schema OK")

    # 3. 提取 providers
    banner("3. Providers")
    rows = conn.execute("SELECT id, app_type, name, category, settings_config FROM providers").fetchall()
    print(f"Total: {len(rows)}")
    with_config = 0
    for r in rows:
        cfg = json.loads(r["settings_config"])
        env = cfg.get("env", {})
        if env:
            with_config += 1
        print(f"  [{r['app_type']:10s}] {r['name']:30s} ({r['category']})")
    print(f"\n有 env 配置的: {with_config}")

    # 4. 价格表
    banner("4. Model Prices")
    prices = conn.execute("""
        SELECT model_id, display_name,
               input_cost_per_million, output_cost_per_million,
               COALESCE(cache_read_cost_per_million, 0) as cache_read,
               COALESCE(cache_creation_cost_per_million, 0) as cache_create
        FROM model_pricing
    """).fetchall()
    price_index = {}
    for p in prices:
        price_index[p["model_id"]] = {
            "in": float(p["input_cost_per_million"]),
            "out": float(p["output_cost_per_million"]),
            "cr": float(p["cache_read"]),
            "cc": float(p["cache_create"]),
        }
    print(f"Indexed {len(price_index)} models")

    # 5. 代理日志
    banner("5. Proxy Request Logs")
    logs = conn.execute("""
        SELECT request_id, provider_id, app_type, model, status_code as status,
               input_tokens, output_tokens,
               COALESCE(cache_read_tokens, 0) as cache_read,
               COALESCE(cache_creation_tokens, 0) as cache_create,
               COALESCE(total_cost_usd, 0) as cost,
               COALESCE(latency_ms, 0) as latency,
               data_source,
               created_at
        FROM proxy_request_logs
        ORDER BY created_at ASC
    """).fetchall()

    total_in = total_out = total_cr = total_cc = 0
    total_cost = 0.0
    model_counter = Counter()
    for l in logs:
        total_in += l["input_tokens"]
        total_out += l["output_tokens"]
        total_cr += l["cache_read"]
        total_cc += l["cache_create"]
        total_cost += float(l["cost"])
        model_counter[l["model"]] += 1

    print(f"Rows: {len(logs):>6,}")
    print(f"Input  tokens: {total_in:>12,}")
    print(f"Output tokens: {total_out:>12,}")
    print(f"Cache read  : {total_cr:>12,}")
    print(f"Cache create: {total_cc:>12,}")
    print(f"Total cost  : ${total_cost:>10.4f}")
    print(f"Unique models: {len(model_counter)}")

    print("\nTop 10 models by request count:")
    for m, n in model_counter.most_common(10):
        print(f"  {m:50s} {n:>5,}")

    # 6. JSONL 解析
    banner("6. JSONL 解析")
    jsonl_files = list(PROJECTS_DIR.rglob("*.jsonl"))
    print(f"Found {len(jsonl_files)} JSONL files")

    total_jsonl_recs = 0
    total_jsonl_cost = 0.0
    jsonl_model_counter = Counter()

    for fp in sorted(jsonl_files, key=lambda f: f.stat().st_mtime, reverse=True)[:5]:
        size = fp.stat().st_size
        with open(fp, encoding='utf-8') as f:
            records = 0
            file_cost = 0.0
            for line in f:
                try:
                    obj = json.loads(line)
                except json.JSONDecodeError:
                    continue
                if obj.get("type") != "assistant":
                    continue
                msg = obj.get("message", {})
                usage = msg.get("usage", {})
                if not usage:
                    continue
                records += 1
                model = msg.get("model", "unknown")
                jsonl_model_counter[model] += 1
                in_t = usage.get("input_tokens", 0) or 0
                out_t = usage.get("output_tokens", 0) or 0
                cr_t = usage.get("cache_read_input_tokens", 0) or 0
                cc_t = usage.get("cache_creation_input_tokens", 0) or 0

                # 计算费用
                p = price_index.get(model, {})
                if p:
                    m = 1_000_000
                    cost = (in_t * p.get("in", 0) + out_t * p.get("out", 0) +
                            cr_t * p.get("cr", 0) + cc_t * p.get("cc", 0)) / m
                    file_cost += cost

            total_jsonl_recs += records
            total_jsonl_cost += file_cost
            print(f"  {fp.name:45s} {size:>8,}B  {records:>4,} recs  ${file_cost:>8.4f}")

    print(f"\nJSONL summary (top 5): {total_jsonl_recs:,} records, ${total_jsonl_cost:.4f}")
    print("\nJSONL top models:")
    for m, n in jsonl_model_counter.most_common(10):
        print(f"  {m:50s} {n:>5,}")

    # 7. 对比 cc-switch logs 和 JSONL
    banner("7. 数据对比")
    print(f"cc-switch proxy logs: {len(logs):>6,} rows, ${total_cost:>10.4f}")
    print(f"JSONL parsed records: {total_jsonl_recs:>6,} rows, ${total_jsonl_cost:>10.4f}")
    print(f"\n差异原因: cc-switch 代理了所有 CLI 请求 (Codex/Gemini/Claude)")
    print(f"        JSONL 只含 Claude Code 的 assistant 消息")

    # 8. 耗时
    banner("验证完成")
    print(f"Total time: {time.time() - start:.2f}s")

    conn.close()


if __name__ == "__main__":
    main()
