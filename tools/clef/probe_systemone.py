#!/usr/bin/env python3
"""CLEF-002 live capability probe (non-production evidence tool).

Sends a fixed, bounded set of requests to Ollama's /v1/systemone (clef-flash)
and to the oMLX OpenAI-compatible server, and writes one JSON capture per probe
under docs/reports/clef-provider/captures/live/. Stdlib only.

    python3 tools/clef/probe_systemone.py [--ollama URL] [--omlx URL]
"""

import argparse
import datetime
import json
import math
import pathlib
import time
import urllib.error
import urllib.request

MAX_BODY = 8192
OUT = pathlib.Path(__file__).resolve().parents[2] / "docs/reports/clef-provider/captures/live"
STATE = "A change touches an untested code path and would deploy to production without review."

CHOICE = {"model": "clef-flash", "state": STATE, "questions": {"risk": {
    "type": "choice",
    "instructions": "Assess the operational risk of the described change.",
    "criteria": {"LOW": None, "MEDIUM": None, "HIGH": None},
    "choices": ["LOW", "MEDIUM", "HIGH"]}}}


def q(**question):
    return {"model": "clef-flash", "state": STATE, "questions": {"q": question}}


# (name, method, path, body). Bodies are dicts, raw strings (sent verbatim), or None (GET).
OLLAMA_PROBES = [
    ("ollama-version", "GET", "/api/version", None),
    ("ollama-show-clef-flash", "POST", "/api/show", {"model": "clef-flash"}),
    ("choice-run1", "POST", "/v1/systemone", CHOICE),
    ("choice-run2", "POST", "/v1/systemone", CHOICE),
    ("choice-run3", "POST", "/v1/systemone", CHOICE),
    ("score", "POST", "/v1/systemone", q(type="score", instructions="How safe is it to deploy this change?",
                                          criteria=["not safe", "somewhat safe", "very safe"])),
    ("noul", "POST", "/v1/systemone", q(type="noul", instructions="Should this change require human approval?")),
    ("noul-no-instructions", "POST", "/v1/systemone", q(type="noul")),
    ("multi-question", "POST", "/v1/systemone", {"model": "clef-flash", "state": STATE, "questions": {
        **CHOICE["questions"],
        "approval": {"type": "noul", "instructions": "Should this change require human approval?"}}}),
    # Request shapes recorded by the earlier transcribed captures; expected to be rejected.
    ("reject-transcribed-choice-shape", "POST", "/v1/systemone", {"model": "clef-flash", "state": STATE,
        "criteria": {"risk": {"type": "choice", "choices": {"LOW": "low", "MEDIUM": "medium", "HIGH": "high"}}}}),
    ("reject-score-legend", "POST", "/v1/systemone", q(type="score", instructions="x", legend={"0": "0", "10": "10"})),
    # Null / indeterminate and unsupported-operation behavior.
    ("reject-single-candidate", "POST", "/v1/systemone", q(type="choice", instructions="x", criteria={"ONLY": None})),
    ("reject-empty-candidates", "POST", "/v1/systemone", q(type="choice", instructions="x", criteria={})),
    ("reject-empty-state", "POST", "/v1/systemone", {**q(type="noul", instructions="x"), "state": ""}),
    ("reject-unknown-type", "POST", "/v1/systemone", q(type="rank", instructions="x")),
    ("reject-unknown-model", "POST", "/v1/systemone", {**q(type="noul", instructions="x"), "model": "no-such-model"}),
    ("reject-non-decision-model", "POST", "/v1/systemone", {**q(type="noul", instructions="x"), "model": "qwen3.5:4b"}),
    ("reject-malformed-json", "POST", "/v1/systemone", '{"model":"clef-flash",'),
    # Non-decision Ollama surfaces on the same model.
    ("chat", "POST", "/api/chat", {"model": "clef-flash", "stream": False,
                                   "messages": [{"role": "user", "content": "Reply with OK."}]}),
    ("embed", "POST", "/api/embed", {"model": "clef-flash", "input": "untested change"}),
]

OMLX_PROBES = [
    ("omlx-models", "GET", "/v1/models", None),
    ("omlx-models-status", "GET", "/v1/models/status", None),
    ("omlx-systemone-route", "POST", "/v1/systemone", CHOICE | {"model": "clef-4bit"}),
]


def call(base, method, path, body):
    data = None if body is None else (body if isinstance(body, str) else json.dumps(body)).encode()
    req = urllib.request.Request(base + path, data=data, method=method,
                                 headers={"Content-Type": "application/json"})
    start = time.monotonic()
    try:
        with urllib.request.urlopen(req, timeout=180) as resp:
            status, raw = resp.status, resp.read()
    except urllib.error.HTTPError as err:
        status, raw = err.code, err.read()
    except (urllib.error.URLError, TimeoutError) as err:
        status, raw = None, str(err).encode()
    elapsed_ms = round((time.monotonic() - start) * 1000)
    text = raw.decode(errors="replace")
    try:
        body = json.loads(text)
    except json.JSONDecodeError:
        body = text
    if isinstance(body, dict):  # /api/show embeds license/template dumps; drop them before bounding
        for key in ("license", "modelfile", "template", "tensors"):
            body.pop(key, None)
    serialized = body if isinstance(body, str) else json.dumps(body)
    truncated = len(serialized) > MAX_BODY
    return {"http_status": status, "elapsed_ms": elapsed_ms, "_truncated": truncated,
            "response": serialized[:MAX_BODY] if truncated else body}


def checks(response):
    """Recompute invariants of each SystemOne answer from its own probabilities."""
    out = {}
    answers = response.get("answers") if isinstance(response, dict) else None
    for qid, ans in (answers or {}).items():
        p = ans.get("probabilities")
        if not p:
            continue
        n = len(p)
        entropy = -sum(v * math.log(v) for v in p.values() if v > 0)
        c = {"probability_sum": round(sum(p.values()), 9),
             "one_minus_normalized_entropy": round(1 - entropy / math.log(n), 9),
             "reported_confidence": ans.get("confidence")}
        c["confidence_matches_entropy_formula"] = abs(c["one_minus_normalized_entropy"] - (ans.get("confidence") or -1)) < 1e-6
        if ans.get("type") == "choice":
            c["choice_is_argmax"] = ans.get("choice") == max(p, key=p.get)
        if ans.get("type") == "score":
            expected = sum(int(k) * v for k, v in p.items())
            c["score_is_expected_index"] = abs(expected - ans.get("score", -1)) < 1e-6
        out[qid] = c
    return out


def mlx_metadata(model_dir):
    """Architecture facts for the local MLX checkpoint, for base-model comparison."""
    d = pathlib.Path(model_dir).expanduser()
    text = json.loads((d / "config.json").read_text()).get("text_config", {})
    readme = (d / "README.md").read_text().split("---")[1]  # YAML front matter only
    return {"_capture": "clef-4bit-local-metadata", "_observed_live": True, "model_dir": str(d),
            "timestamp": datetime.datetime.now(datetime.timezone.utc).isoformat(timespec="seconds"),
            "text_config": {k: text.get(k) for k in ("model_type", "hidden_size", "num_hidden_layers",
                                                      "num_attention_heads", "vocab_size", "max_position_embeddings")},
            "joint_head_config": json.loads((d / "joint_head_config.json").read_text()),
            "readme_front_matter": readme.strip().splitlines(),
            "weight_shards_bytes": {p.name: p.stat().st_size for p in sorted(d.glob("*.safetensors"))}}


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--ollama", default="http://localhost:11434")
    ap.add_argument("--omlx", default="http://localhost:8000")
    ap.add_argument("--mlx-model-dir", default="~/.omlx/omlx/models/mlx-community/clef-4bit")
    args = ap.parse_args()
    OUT.mkdir(parents=True, exist_ok=True)
    (OUT / "clef-4bit-local-metadata.json").write_text(json.dumps(mlx_metadata(args.mlx_model_dir), indent=2) + "\n")
    for base, probes in ((args.ollama, OLLAMA_PROBES), (args.omlx, OMLX_PROBES)):
        for name, method, path, body in probes:
            capture = {"_capture": name, "_observed_live": True, "_bounded": True, "_max_response_bytes": MAX_BODY,
                       "timestamp": datetime.datetime.now(datetime.timezone.utc).isoformat(timespec="seconds"),
                       "endpoint": base, "method": method, "path": path, "request": body}
            capture |= call(base, method, path, body)
            if inv := checks(capture["response"]):
                capture["invariants"] = inv
            (OUT / f"{name}.json").write_text(json.dumps(capture, indent=2) + "\n")
            print(f"{name:34} {capture['http_status']} {capture['elapsed_ms']}ms")


if __name__ == "__main__":
    main()
