import torch
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
        self.model.train() 
        for epoch in range(self.epochs) : 
            print(f"epoch : {epoch} ")
            for i in range(self.training_steps) :
                self.optimizer.zero_grad() 
                logits = self.model(self.training_batches[i])
                loss = torch.nn.functional.cross_entropy(
                    logits.flatten(0,1) , 
                    self.target_batches[i].flatten()
                )
                loss.backward()
                torch.nn.utils.clip_grad_norm_(self.model.parameters(),max_norm=1.0)
                self.optimizer.step()
                self.scheduler.step()

# def dff_for_target(N, V=32767, d=768, L=20, dkv=192, tied=True):
#     fixed = (2 - tied) * V * d + L * (2*d*d + 2*d*dkv + 2*d) + d
#     return (N - fixed) / (3 * L * d)

if __name__ == "__main__" : 

    model_config = {
        "d_model"        : 768,    
        "num_heads"      : 16,
        "num_kv_heads"   : 4,     
        "d_ff"           : 1792,  
        "vocab_size"     : 32767, 
        "n_layers"       : 30,
        "context_window" : 2048,  
    }

    llama_model = LlamaModel(model_config)
    total = sum(p.numel() for p in llama_model.parameters())
    trainable = sum(p.numel() for p in llama_model.parameters() if p.requires_grad)
    print(f"total number of parameters : {total:,}") 
    print(f"total number of trainable parameters :  {trainable:,}")
    data_loader = DataLoaderHF(tokenizer_file = "./fr_bpe_32k_422.json")
    npy_files = [
        "./training/training_tokens/manu_french_poetry_tokens.npy",
        # "./training/training_tokens/Volko76_french-classic-books_tokens.npy",
        # "./training/training_tokens/nirantk_french-books_tokens.npy",
        # "./training/training_tokens/PleIAs_French-PD-Newspapers.npy",
        # "./training/training_tokens/wikimedia_wikipedia_tokens.npy",
    ]
    training_batches, target_batches = data_loader.build_batches(npy_files=npy_files, batch_size=5)
    print(f"{len(training_batches)} batches of {len(training_batches[0])} sequences")


    
    


        


