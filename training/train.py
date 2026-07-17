import torch
import math
import os
from pathlib import Path
from architecture.llama.llama_model import LlamaModel
from training.data_loader import DataLoaderHF
from typing import Dict, List, Tuple

class TrainingLoop :  
    def __init__(self,model:LlamaModel,model_config:Dict,
                 data_loader:DataLoaderHF,
                 data_set:List[str],
                 device = torch.device("cuda" if torch.cuda.is_available() else "cpu") , 
                 dtype=torch.bfloat16,
                 batch_size:int=5, lr:float=3e-4,
                 weight_decay:float=0.1,
                 adamw_betas:Tuple=(0.9,0.95),
                 warmup_steps:int=1000, 
                 warmup_step_percentage:float=0.02, 
                 decay_step_percentage:float=0.2, 
                 warm_up_start_factor:float=0.01, 
                 epochs:int=1) :

        self.model_config = model_config
        self.model = model(self.model_config)
        self.data_loader = data_loader
        self.device = device
        self.dtype = dtype
        self.optimizer = torch.optim.AdamW(
            self.model.parameters(),
            lr=lr ,
            weight_decay=weight_decay,
            betas=adamw_betas
        )
        self.epochs = epochs

        print("getting training tokens and target tokens")
        self.training_batches, self.target_batches = self.data_loader.build_batches(data_set, batch_size)
        
        self.training_batches = torch.from_numpy(self.training_batches)
        
        self.target_batches = torch.from_numpy(self.target_batches)
        
        self.training_steps =  self.target_batches.shape[0] 
        print(f"number of training iterations : {self.training_steps * self.epochs}")

        self.warmup_steps = int(warmup_step_percentage * self.training_steps)
        print(f"number of warmup iterations : {self.warmup_steps} ")

        self.decay_steps = int(decay_step_percentage * self.training_steps) 
        print(f"number of decay steps : {self.decay_steps}")

        self.stable_steps = self.training_steps - self.warmup_steps - self.decay_steps



        ##the learning rate start at 0.01 LR and start increasing until getting the peak which
        warm_up_phase = torch.optim.lr_scheduler.LinearLR(
            self.optimizer,
            start_factor=warm_up_start_factor,
            end_factor=1.0 , 
            total_iters=self.warmup_steps
        )

        stable_phase = torch.optim.lr_scheduler.ConstantLR(
            self.optimizer, 
            factor=1.0, 
            total_iters=self.stable_steps
        )

        cosine_decay_phase = torch.optim.lr_scheduler.CosineAnnealingLR(
            self.optimizer, 
            T_max=self.decay_steps, 
            eta_min=0.0
        )

        self.scheduler = torch.optim.lr_scheduler.SequentialLR(
            self.optimizer, 
            schedulers=[warm_up_phase,stable_phase,cosine_decay_phase] ,
            milestones=[self.warmup_steps, self.stable_steps+self.warmup_steps]
        )

    def train(self) :
        self.model.to(device=self.device, dtype=self.dtype)
        self.model.train() 
        for epoch in range(self.epochs) :
            print(f"epoch : {epoch} ")
            accumulated_loss = torch.zeros(1, device=self.device)
            for i in range(self.training_steps) :
                print(f"iteration num : {i}")
                self.optimizer.zero_grad()
                batch = self.training_batches[i].to(self.device, non_blocking=True)
                target = self.target_batches[i].to(self.device, non_blocking=True)
                logits = self.model(batch)
                loss = torch.nn.functional.cross_entropy(
                    logits.flatten(0,1) ,
                    target.flatten()
                )
                loss.backward()
                torch.nn.utils.clip_grad_norm_(self.model.parameters(),max_norm=1.0)
                self.optimizer.step()
                self.scheduler.step()
                accumulated_loss += loss.detach()
                if (i + 1) % 100 == 0 :
                    print(f"step {i + 1}/{self.training_steps} — accumulated loss: {accumulated_loss.item():.4f}")
                    accumulated_loss.zero_()
        torch.save(self.model.state_dict(), "model_weights.pt")
        print("model weights saved to model_weights.pt")

def estimate_d_model(n_layers, target_params=135_000_000, vocab_size=32767,
                     num_heads=16, num_kv_heads=4, ffn_dim_multiplier=1.0, tied=False):
    from architecture.FFN.llama_feed_forward import ffn_dim as compute_ffn_dim

    head_dim_8 = num_heads * 8

    # closed-form initial guess (approximates d_ff as exactly 8/3 * d * m)
    m = ffn_dim_multiplier if ffn_dim_multiplier is not None else 1.0
    c = 2 + 2 * (num_kv_heads / num_heads) + 8 * m
    b = (1 if tied else 2) * vocab_size + 2 * n_layers + 1
    d_approx = (math.sqrt(b * b + 4 * c * n_layers * target_params) - b) / (2 * c * n_layers)

    def actual_params(d):
        d_ff = compute_ffn_dim(d, ffn_dim_multiplier=ffn_dim_multiplier)
        attn = 2 * d * d + 2 * d * (num_kv_heads * (d // num_heads))
        ffn = 3 * d * d_ff
        norms = 2 * d
        per_layer = attn + ffn + norms
        emb = (1 if tied else 2) * vocab_size * d
        return n_layers * per_layer + emb + d  # +d for final RMSNorm

    # search nearby multiples of head_dim_8 around the initial guess
    best_d = round(d_approx / head_dim_8) * head_dim_8
    best_err = abs(actual_params(best_d) - target_params)
    for offset in range(-3, 4):
        candidate = round(d_approx / head_dim_8) * head_dim_8 + offset * head_dim_8
        if candidate <= 0:
            continue
        err = abs(actual_params(candidate) - target_params)
        if err < best_err:
            best_d = candidate
            best_err = err
    return best_d

if __name__ == "__main__" : 

    model_config = {
        "d_model"        : 768,    
        "num_heads"      : 16,
        "num_kv_heads"   : 4,     
        "vocab_size"     : 32768, 
        "n_layers"       : 22,
        "context_window" : 2048,  
    }

    d_model = estimate_d_model(model_config["n_layers"])
    print(f"estimated d_model {d_model}")
    model_config["d_model"] = d_model
    llama_model = LlamaModel(model_config)
    total = sum(p.numel() for p in llama_model.parameters())
    trainable = sum(p.numel() for p in llama_model.parameters() if p.requires_grad)
    print(f"total number of parameters : {total:,}") 
    print(f"total number of trainable parameters :  {trainable:,}")
    data_loader = DataLoaderHF(tokenizer_file = "./fr_bpe_32k_422.json",context_size=model_config["context_window"])
    num_tokens_per_batch = 70_000
    batch_size = num_tokens_per_batch // model_config["context_window"]
    npy_files = [
        "./training/training_tokens/manu_french_poetry_tokens.npy",
        "./training/training_tokens/Volko76_french-classic-books_tokens.npy",
        "./training/training_tokens/nirantk_french-books_tokens.npy",
        "./training/training_tokens/PleIAs_French-PD-Newspapers.npy",
        "./training/training_tokens/wikimedia_wikipedia_tokens.npy",
    ]
    training_loop = TrainingLoop(
        model=LlamaModel,
        model_config=model_config,
        data_loader=data_loader,
        data_set=npy_files,
        batch_size=batch_size,
    )
    training_loop.train()

