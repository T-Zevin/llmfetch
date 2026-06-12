#!/usr/bin/env python3
"""Collect and normalize mainstream local-LLM registry entries.

The registry should be curated, but not hand-written. This script uses the
Hugging Face public API as a broad discovery source, then filters and
normalizes models by mainstream family/provider heuristics.
"""

from __future__ import annotations

import argparse
import json
import math
import re
import shutil
import sys
import time
import urllib.parse
import urllib.request
from pathlib import Path
from typing import Any


ROOT = Path(__file__).resolve().parents[1]
OUT_REGISTRY = ROOT / "registry" / "models.json"
EMBED_REGISTRY = ROOT / "internal" / "registry" / "models.json"

HF_API = "https://huggingface.co/api/models"

PIPELINES = [
    "text-generation",
    "image-text-to-text",
    "text2text-generation",
    "feature-extraction",
    "sentence-similarity",
    "zero-shot-classification",
    "automatic-speech-recognition",
]

FAMILY_RULES = [
    ("qwen", "Alibaba", "Qwen"),
    ("deepseek", "DeepSeek", "DeepSeek"),
    ("llama", "Meta", "Llama"),
    ("gemma", "Google", "Gemma"),
    ("mistral", "Mistral", "Mistral"),
    ("mixtral", "Mistral", "Mixtral"),
    ("phi", "Microsoft", "Phi"),
    ("glm", "Zhipu", "GLM"),
    ("chatglm", "Zhipu", "GLM"),
    ("yi-", "01.AI", "Yi"),
    ("01-ai/yi", "01.AI", "Yi"),
    ("internlm", "InternLM", "InternLM"),
    ("minicpm", "OpenBMB", "MiniCPM"),
    ("baichuan", "Baichuan", "Baichuan"),
    ("falcon", "TII", "Falcon"),
    ("granite", "IBM", "Granite"),
    ("nemotron", "NVIDIA", "Nemotron"),
    ("starcoder", "BigCode", "StarCoder"),
    ("codellama", "Meta", "CodeLlama"),
    ("codegemma", "Google", "CodeGemma"),
    ("codestral", "Mistral", "Codestral"),
    ("openchat", "OpenChat", "OpenChat"),
    ("nous", "NousResearch", "Nous"),
    ("hermes", "NousResearch", "Hermes"),
    ("wizard", "WizardLM", "WizardLM"),
    ("vicuna", "LMSYS", "Vicuna"),
    ("zephyr", "HuggingFaceH4", "Zephyr"),
    ("smollm", "HuggingFaceTB", "SmolLM"),
    ("bge", "BAAI", "BGE"),
    ("e5", "Microsoft", "E5"),
    ("gte", "Alibaba", "GTE"),
    ("jina", "Jina AI", "Jina"),
    ("nomic", "Nomic", "Nomic"),
    ("colbert", "Stanford", "ColBERT"),
    ("reranker", "BAAI", "Reranker"),
    ("whisper", "OpenAI", "Whisper"),
    ("stablelm", "Stability AI", "StableLM"),
    ("command-r", "Cohere", "Command R"),
    ("aya", "Cohere", "Aya"),
    ("olmo", "AI2", "OLMo"),
    ("dbrx", "Databricks", "DBRX"),
    ("xverse", "XVERSE", "XVERSE"),
    ("solar", "Upstage", "SOLAR"),
    ("exaone", "LG AI Research", "EXAONE"),
]

SKIP_PATTERNS = [
    "adapter",
    "lora",
    "qlora",
    "merge",
    "merged",
    "test",
    "demo",
    "draft",
    "private",
    "uncensored",
    "claude",
    "sonnet",
    "opus",
    "gpt-",
    "chatgpt",
    "grok",
]


def fetch_json(url: str, timeout: int = 45) -> Any:
    request = urllib.request.Request(url, headers={"User-Agent": "llmfetch-registry-builder/0.1"})
    with urllib.request.urlopen(request, timeout=timeout) as response:
        return json.loads(response.read().decode("utf-8"))


def hf_models(params: dict[str, Any]) -> list[dict[str, Any]]:
    query = urllib.parse.urlencode(params, doseq=True)
    return fetch_json(f"{HF_API}?{query}")


def discover(limit_per_query: int) -> list[dict[str, Any]]:
    seen: dict[str, dict[str, Any]] = {}

    queries: list[dict[str, Any]] = []
    for pipeline in PIPELINES:
        queries.append(
            {
                "pipeline_tag": pipeline,
                "sort": "downloads",
                "direction": "-1",
                "limit": limit_per_query,
                "full": "true",
            }
        )
    for keyword, _, _ in FAMILY_RULES:
        queries.append(
            {
                "search": keyword,
                "sort": "downloads",
                "direction": "-1",
                "limit": min(limit_per_query, 200),
                "full": "true",
            }
        )

    for i, params in enumerate(queries, 1):
        try:
            models = hf_models(params)
        except Exception as exc:
            print(f"[warn] query failed {params}: {exc}", file=sys.stderr)
            continue
        for item in models:
            model_id = item.get("id") or item.get("modelId")
            if model_id:
                seen[model_id] = item
        if i % 8 == 0:
            time.sleep(0.2)

    return list(seen.values())


def family_for(model_id: str) -> tuple[str, str] | None:
    lowered = model_id.lower()
    for keyword, provider, family in FAMILY_RULES:
        if keyword in lowered:
            return provider, family
    return None


def should_skip(model_id: str) -> bool:
    lowered = model_id.lower()
    if any(pattern in lowered for pattern in SKIP_PATTERNS):
        return True
    if lowered.count("/") > 1:
        return True
    return False


def params_b(model_id: str, tags: list[str]) -> float:
    text = " ".join([model_id, *tags]).lower()
    patterns = [
        r"(\d+(?:\.\d+)?)\s*[bB]\b",
        r"(\d+(?:\.\d+)?)b(?:-|_|$)",
        r"-(\d+(?:\.\d+)?)b",
        r"_(\d+(?:\.\d+)?)b",
    ]
    for pattern in patterns:
        match = re.search(pattern, text)
        if match:
            return float(match.group(1))
    if "large" in text:
        return 7.0
    if "base" in text or "small" in text:
        return 1.5
    return 7.0


def context_window(model_id: str, tags: list[str]) -> str:
    text = " ".join([model_id, *tags]).lower()
    matches = re.findall(r"(\d+)\s*k", text)
    if matches:
        value = max(int(m) for m in matches)
        if value >= 4:
            return f"{value}K"
    if any(key in text for key in ["qwen3", "llama-3.3", "command-r", "deepseek-r1"]):
        return "128K"
    if any(key in text for key in ["coder", "long", "glm-4"]):
        return "128K"
    return "32K"


def task_type(model_id: str, pipeline: str, tags: list[str]) -> tuple[str, str]:
    id_lower = model_id.lower()
    lowered = " ".join([model_id, pipeline, *tags]).lower()
    if "rerank" in id_lower or "reranker" in id_lower:
        return "Reranker", "Rerank"
    if "embed" in lowered or pipeline in {"feature-extraction", "sentence-similarity"}:
        return "Embedding", "Embedding"
    if "whisper" in id_lower or pipeline == "automatic-speech-recognition":
        return "Audio", "ASR"
    vision_patterns = [
        r"(^|[-_/])vl($|[-_/0-9])",
        r"vision",
        r"visual",
        r"multimodal",
        r"image-text",
        r"ocr",
        r"pixtral",
    ]
    if any(re.search(pattern, id_lower) for pattern in vision_patterns):
        return "Vision", "Vision"
    if any(key in id_lower for key in ["coder", "code", "starcoder", "codestral"]):
        return "Coding", "Coding"
    if any(key in id_lower for key in ["r1", "reason", "math"]):
        return "Reasoning", "Reasoning"
    if any(key in id_lower for key in ["command-r", "rag", "colbert"]):
        return "RAG", "RAG"
    if any(key in id_lower for key in ["chat", "instruct", "it"]):
        return "General", "Chat"
    return "General", "General"


def license_name(item: dict[str, Any]) -> str:
    tags = item.get("tags") or []
    for tag in tags:
        if isinstance(tag, str) and tag.startswith("license:"):
            return tag.split(":", 1)[1]
    card = item.get("cardData") or {}
    lic = card.get("license") if isinstance(card, dict) else None
    if isinstance(lic, str) and lic:
        return lic
    return "unknown"


def runtime_for(model_id: str, task: str, provider: str) -> str:
    lowered = model_id.lower()
    if task in {"Embedding", "Reranker"}:
        return "Transformers"
    if task == "Audio":
        return "Whisper"
    if "mlx" in lowered or provider in {"Alibaba", "DeepSeek", "Google", "Mistral", "Meta", "Microsoft"}:
        return "MLX Native"
    return "Ollama"


def memory_estimate(params: float, task: str) -> int:
    if task in {"Embedding", "Reranker"}:
        return max(1, math.ceil(params * 1.2))
    if task == "Audio":
        return max(2, math.ceil(params * 1.5))
    return max(3, math.ceil(params * 0.72 + 2))


def speed_estimate(params: float, task: str) -> int:
    if task in {"Embedding", "Reranker"}:
        return max(120, int(700 / max(1, params)))
    if params <= 3:
        return 120
    if params <= 8:
        return 90
    if params <= 14:
        return 72
    if params <= 34:
        return 55
    if params <= 72:
        return 28
    return 14


def score_for(downloads: int, likes: int, family_bonus: int, task: str) -> int:
    score = 55
    if downloads > 0:
        score += min(25, int(math.log10(downloads + 1) * 5))
    if likes > 0:
        score += min(10, int(math.log10(likes + 1) * 3))
    score += family_bonus
    if task in {"Coding", "Reasoning", "Vision"}:
        score += 2
    if task in {"Embedding", "Reranker", "Audio"}:
        score -= 8
    return max(60, min(99, score))


def fit_for(memory: int) -> str:
    if memory <= 24:
        return "Best"
    if memory <= 40:
        return "Good"
    return "Near"


def normalize(items: list[dict[str, Any]], target: int) -> list[dict[str, Any]]:
    candidates: list[dict[str, Any]] = []
    seen: set[str] = set()
    for item in items:
        model_id = item.get("id") or item.get("modelId") or ""
        if not model_id or model_id in seen or should_skip(model_id):
            continue
        fam = family_for(model_id)
        if fam is None:
            continue
        seen.add(model_id)
        provider, family = fam
        tags = [tag for tag in item.get("tags", []) if isinstance(tag, str)]
        pipeline = item.get("pipeline_tag") or ""
        model_type, best_for = task_type(model_id, pipeline, tags)
        p = params_b(model_id, tags)
        memory = memory_estimate(p, model_type)
        downloads = int(item.get("downloads") or 0)
        likes = int(item.get("likes") or 0)
        score = score_for(downloads, likes, 5, model_type)
        trend = min(99, max(-9, int(math.log10(downloads + 1)) - 2))
        runtime = runtime_for(model_id, model_type, provider)
        out_tps = speed_estimate(p, model_type)
        candidates.append(
            {
                "model_id": model_id,
                "rank": 0,
                "name": display_name(model_id),
                "provider": provider,
                "family": family,
                "best_for": best_for,
                "type": model_type,
                "score": score,
                "runtime": runtime,
                "out_tps": out_tps,
                "in_tps": out_tps * 5,
                "memory_gb": memory,
                "fit": fit_for(memory),
                "context": context_window(model_id, tags),
                "license": normalize_license(license_name(item)),
                "trend": trend,
                "downloads": downloads,
                "likes": likes,
            }
        )

    candidates.sort(key=lambda m: (m["score"], m["downloads"], m["likes"]), reverse=True)
    result = candidates[:target]
    for idx, item in enumerate(result, 1):
        item["rank"] = idx
        # Keep runtime fields lean for the Go struct. Extra fields are harmless
        # for json.Unmarshal, but public registry should stay readable.
        item.pop("downloads", None)
        item.pop("likes", None)
        item.pop("family", None)
        item.pop("model_id", None)
    return result


def display_name(model_id: str) -> str:
    name = model_id.split("/")[-1]
    return name.replace("_", "-")


def normalize_license(value: str) -> str:
    mapping = {
        "apache-2.0": "Apache-2",
        "mit": "MIT",
        "cc-by-nc-4.0": "CC-BY-NC",
        "llama3.1": "Llama",
        "llama3.2": "Llama",
        "llama3.3": "Llama",
        "gemma": "Gemma",
    }
    return mapping.get(value.lower(), value[:16] if value else "unknown")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--target", type=int, default=500)
    parser.add_argument("--limit-per-query", type=int, default=500)
    parser.add_argument("--output", type=Path, default=OUT_REGISTRY)
    args = parser.parse_args()

    print(f"[info] discovering models from Hugging Face API...", file=sys.stderr)
    raw = discover(args.limit_per_query)
    print(f"[info] discovered {len(raw)} unique candidates before filtering", file=sys.stderr)
    models = normalize(raw, args.target)
    if len(models) < args.target:
        print(f"[warn] only produced {len(models)} models; target was {args.target}", file=sys.stderr)
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(models, indent=2, ensure_ascii=False) + "\n")
    if args.output.resolve() == OUT_REGISTRY.resolve():
        shutil.copyfile(OUT_REGISTRY, EMBED_REGISTRY)
    print(f"[info] wrote {len(models)} models to {args.output}", file=sys.stderr)
    return 0 if len(models) >= min(args.target, 1) else 1


if __name__ == "__main__":
    raise SystemExit(main())
