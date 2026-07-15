import torch
from architecture.llama.llama_model import LlamaModel
from training.data_loader import DataLoaderHF
from typing import Dict, List, Tuple

class TrainingLoop :  
    def __init__(self,model:LlamaModel,model_config:Dict,data_loader:DataLoaderHF,data_set:List[str],device,dtype=torch.bfloat16,
                 batch_size:int=5, lr:float=3e-4, weight_decay:float=0.1,adamw_betas:Tuple=(0.9,0.95), warmup_steps:int=1000, 
                 warmup_step_percentage=0.05 ) :

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

        print("getting training tokens and target tokens")
        self.training_batches, self.target_batches = self.data_loader.build_batches(data_set, batch_size)
        
        self.training_batches = torch.tensor(self.training_batches).to(device=self.device, 
                                                                        dtype=torch.int64)
        
        self.target_batches = torch.tensor(self.target_batches).to(device=self.device, 
                                                                   dtype=torch.int64)

        self.training_steps =  self.target_batches.shape(0) 
        print(f"number of training iterations : {self.training_steps}")

        self.warmup_steps = warmup_step_percentage * self.training_steps
        print(f"number of warmup iterations : {self.warmup_steps} ")

        


        ##the learning rate start at 0.01 LR and start increasing until getting the peak which
        warm_up_phase = torch.optim.LinearLR(
            self.optimizer,
            start_factor=0.01, 
            total_iters=0.05*self.num_iter
        )

        cosine_decay_phase = torch.optim.ConsineAnnealingLR(
            self.optimizer, 
            T_max=self.training_steps - self.warmup_steps, 
            eta_min=0.1*self.lr 
        )

        self.scheduler = torch.optim.lr_scheduler.SequentialLR(
            self.optimizer, 
            schedulers=[warm_up_phase,consine_decay_phase] ,
            milestones=[warm_up_phase]
        )

    def train() : 
        self.model.train() 
        for i in self.batch_size :
            self.optimizer.zero_grad() 
            logits = self.model(self.training_batches[i])
            loss = torch.nn.functional.cross_entropy(
                logits.flatten(0,1) , 
                self.target_batches[i].flatten()
            )
            loss.backward()
            self.optimizer.step()
            self.scheduler.step()


if __name__ == "__main__" : 

    model_config = {
        "d_model" : 576 , 
        "num_heads" : 9, 
        "num_kv_heads" : 3, 
        "vocab_size" : 32767 , 
        "n_layers" : 12, 
        "context_window" : 2048
    }

    llama_model = LlamaModel(model_config)
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


    
    


        


