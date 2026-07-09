import torch
import torch.nn as nn

from architecture.attention.MHA.MHA import (
    MultiHeadAttention, MHA_wrapper, MHAFlashAttention,
    MHAEfficientAttention, MHAMathAttention, CausalAttention
)
from architecture.FFN.feed_forward import FeedForward, FeedForward_torch
from architecture.normalization.layer_norm import LayerNorm


MHA_REGISTRY = {
    "multihead": MultiHeadAttention,
    "wrapper": MHA_wrapper,
    "flash": MHAFlashAttention,
    "efficient": MHAEfficientAttention,
    "math": MHAMathAttention,
    "causal": CausalAttention,
}

FFN_REGISTRY = {
    "feedforward": FeedForward,
    "feedforward_torch": FeedForward_torch,
}


class GPTTransformerBlock(nn.Module) :
    def __init__(self, cfg) :
        super().__init__()

        mha_name = cfg.get("mha_variant", "multihead")
        ffn_name = cfg.get("ffn_variant", "feedforward")
        compiled = cfg.get("compiled", False)

        mha_cls = MHA_REGISTRY[mha_name]
        ffn_cls = FFN_REGISTRY[ffn_name]

        # CausalAttention and MHA_wrapper use slightly different param names
        if mha_name == "causal":
            self.att = mha_cls(
                d_in=cfg["emb_dim"],
                d_out=cfg["emb_dim"],
                context_length=cfg["context_length"],
                dropout=cfg["drop_rate"],
                qkv_bias=cfg["qkv_bias"],
                compile=compiled,
            )
        elif mha_name == "wrapper":
            self.att = mha_cls(
                d_in=cfg["emb_dim"],
                d_out=cfg["emb_dim"],
                context_length=cfg["context_length"],
                dropout=cfg["drop_rate"],
                n_heads=cfg["n_heads"],
                qkv_bias=cfg["qkv_bias"],
                compile=compiled,
            )
        else:
            self.att = mha_cls(
                d_in=cfg["emb_dim"],
                d_out=cfg["emb_dim"],
                context_length=cfg["context_length"],
                dropout=cfg["drop_rate"],
                num_heads=cfg["n_heads"],
                qkv_bias=cfg["qkv_bias"],
                compile=compiled,
            )

        self.ff = ffn_cls(cfg, 4 * cfg["emb_dim"], is_compiled=compiled)
        self.norm1 = LayerNorm(cfg["emb_dim"])
        self.norm2 = LayerNorm(cfg["emb_dim"])
        self.drop_shortcut = nn.Dropout(cfg["drop_rate"])

    def forward(self,x) :
        shortcut = x
        x = self.norm1(x)
        x = self.att(x)
        x = self.drop_shortcut(x)
        x = x + shortcut

        shortcut = x
        x = self.norm2(x)
        x = self.ff(x)
        x = self.drop_shortcut(x)
        x = x + shortcut

        return x


if __name__ == "__main__" :
    cfg = {
        "emb_dim" : 768,
        "context_length" : 1024 ,
        "n_heads" : 12,
        "drop_rate" : 0,
        "qkv_bias" : False,
        "mha_variant" : "flash",
        "ffn_variant" : "feedforward",
        "compiled" : False,
    }
    transformer_block = GPTTransformerBlock(cfg)
    x = torch.randn(3,12,cfg["emb_dim"])
    print(f"Using MHA: {cfg['mha_variant']}, FFN: {cfg['ffn_variant']}, compiled: {cfg['compiled']}")
    print(f"Output shape: {transformer_block(x).shape}")