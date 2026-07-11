def estimate_flops_per_layer(config: dict) -> dict:
    """
    Estimate the number of floating point operations per transformer layer.

    config keys:
        batch_size (B): batch size
        seq_len (T): sequence length
        d_model: model dimension
        num_heads (H): number of attention heads
        d_ff: feed-forward hidden dim (defaults to 4 * d_model if not provided)
        num_kv_heads: number of KV heads for GQA (defaults to num_heads = standard MHA)

    Each multiply-add counts as 2 FLOPs.
    Returns a dict with per-operation and total FLOPs for one layer.
    """
    B = config["batch_size"]
    T = config["seq_len"]
    D = config["d_model"]
    H = config["num_heads"]
    d_head = D // H
    d_ff = config.get("d_ff", 4 * D)
    num_kv_heads = config.get("num_kv_heads", H)

    # Q projection: B * T * D -> B * T * D
    qkv_q = 2 * B * T * D * D
    # K projection: B * T * D -> B * T * (num_kv_heads * d_head)
    qkv_k = 2 * B * T * D * (num_kv_heads * d_head)
    # V projection: same as K
    qkv_v = 2 * B * T * D * (num_kv_heads * d_head)

    # Attention scores: Q @ K^T -> (B, H, T, T)
    # Each head does (T, d_head) @ (d_head, T) = 2 * T * T * d_head
    attn_scores = 2 * B * H * T * T * d_head

    # Attention @ V -> (B, H, T, d_head)
    attn_v = 2 * B * H * T * T * d_head

    # Output projection: B * T * D -> B * T * D
    out_proj = 2 * B * T * D * D

    # FFN: two linear layers (D -> d_ff -> D)
    ffn_up = 2 * B * T * D * d_ff
    ffn_down = 2 * B * T * d_ff * D

    ops = {
        "qkv_q": qkv_q,
        "qkv_k": qkv_k,
        "qkv_v": qkv_v,
        "attn_scores": attn_scores,
        "attn_v": attn_v,
        "out_proj": out_proj,
        "ffn_up": ffn_up,
        "ffn_down": ffn_down,
    }
    ops["total"] = sum(ops.values())
    return ops


def estimate_total_flops(config: dict) -> dict:
    """
    Estimate total FLOPs for the full model (all layers).

    Additional config key:
        num_layers: number of transformer layers
        training: if True, multiply by 3 (forward + backward ≈ 3x forward)
    """
    num_layers = config["num_layers"]
    training = config.get("training", False)

    per_layer = estimate_flops_per_layer(config)
    total = {k: v * num_layers for k, v in per_layer.items()}

    if training:
        total = {k: v * 3 for k, v in total.items()}

    return total


def estimate_mfu(config: dict, elapsed_sec: float, gpu_peak_tflops: float) -> float:
    """
    Estimate Model FLOPs Utilization.

    Args:
        config: model config dict (same as estimate_total_flops, with training=True for training)
        elapsed_sec: wall-clock time in seconds for the measured pass
        gpu_peak_tflops: GPU peak throughput in TFLOP/s (e.g. 35.6 for 3060, 142 for 3090 in fp16)

    Returns:
        MFU as a fraction (0.0 to 1.0)
    """
    flops = estimate_total_flops(config)["total"]
    achieved_tflops = flops / elapsed_sec / 1e12
    return achieved_tflops / gpu_peak_tflops


if __name__ == "__main__":
    config = {
        "batch_size": 8,
        "seq_len": 2048,
        "d_model": 1024,
        "num_heads": 16,
        "num_layers": 24,
        "training": True,
    }

    per_layer = estimate_flops_per_layer(config)
    print("FLOPs per layer:")
    for k, v in per_layer.items():
        print(f"  {k}: {v:.2e}")

    total = estimate_total_flops(config)
    print(f"\nTotal FLOPs ({config['num_layers']} layers, training={config['training']}):")
    print(f"  {total['total']:.2e}")
