"""FNV-1a hash for A/B variant assignment (mirrors Go server)."""

from __future__ import annotations


def hash_index(value: str) -> int:
    h = 2166136261
    for ch in value.encode("utf-8"):
        h ^= ch
        h = (h * 16777619) & 0xFFFFFFFF
    return int(h)
