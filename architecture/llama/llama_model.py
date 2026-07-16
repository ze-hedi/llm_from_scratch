import torch 
import torch.nn as nn 
from architecture.llama.llama_transformer_block import LlamaTransformerBlock
from architecture.normalization.RMSNorm import RMSNorm

## configuration inputs
# "d_model" 
# "num_heads" 
# "num_kv_groups" 
# "context_window"
# vocab_size 
# n layers

class LlamaModel(nn.Module) : 
    def __init__(self,cfg) : 
        super().__init__()
        self.tok_emb = nn.Embedding(cfg["vocab_size"],cfg["d_model"])

        print(f"num number of layer of the model :{cfg['n_layers']}")
        self.llama_transformer_blocks = nn.Sequential(
            *[LlamaTransformerBlock(cfg) for _ in range(cfg["n_layers"])]
        )


        self.final_norm = RMSNorm(cfg["d_model"])
        self.out_head = nn.Linear(cfg["d_model"],cfg["vocab_size"])
    def forward(self,input_tokens) : 
        x = self.tok_emb(input_tokens)
        x = self.llama_transformer_blocks(x)
        x = self.final_norm(x) 
        logits = self.out_head(x)
        return logits



    