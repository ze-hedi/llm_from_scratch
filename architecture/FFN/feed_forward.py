import torch 
import torch.nn as nn 
from architecture.FFN.activation import GELU, GELU_torch

class FeedForward(nn.Module) : 
    
    def __init__(self,cfg,hidden_size,is_compiled=False) : 
        super().__init__() 
        self.layers = nn.Sequential(
            nn.Linear(cfg["emb_dim"],hidden_size) ,
            GELU() ,
            nn.Linear(hidden_size,cfg["emb_dim"])
        )
        if is_compiled :
            self.layers = torch.compile(self.layers) 
 
    def forward(self,x) : 
        return self.layers(x)

class FeedForward_torch(nn.Module) : 
    
    def __init__(self,cfg,hidden_size,is_compiled=False) : 
        super().__init__() 
        self.layers = nn.Sequential(
            nn.Linear(cfg["emb_dim"],hidden_size) ,
            GELU_torch() ,
            nn.Linear(hidden_size,cfg["emb_dim"])
        )
        if is_compiled :
            self.layers = torch.compile(self.layers) 
 
    def forward(self,x) : 
        return self.layers(x)




if __name__ == "__main__" : 
    cfg = {
        "emb_dim" : 768
    }
    hidden_size = 4 * 768 

    x = torch.randn(4,12,768) 

    feed_forward = FeedForward(cfg,hidden_size) 

    result = feed_forward(x) 