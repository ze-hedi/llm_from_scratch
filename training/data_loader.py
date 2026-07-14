from tokenizers import Tokenizer 
from typing import List, Dict, Tuple
from datasets import load_dataset
from pathlib import Path 
import numpy as np 

class DataLoader : 
    def __init__(self,tokenizer_file,training_corpus_file,context_size=1024) :
        self.tokenizer = Tokenizer.from_file(tokenizer_file) 
        self.token_ids = None
        self.training_tokens = []
        self.target_tokens = [] 

        print("reading training corpus .... ")
        with open(training_corpus_file,"r") as file :
            training_corpus = file.read() 
            self.token_ids = self.tokenizer.encode(training_corpus) 
        num_tokens = len(self.token_ids.ids) 
        print(f"number of tokens in corpus : {num_tokens}")
        
        print("building training strides .....")
        for i in range(0,num_tokens - context_size, context_size) : 
            self.training_tokens.append(self.token_ids.ids[i:i+context_size]) 
            self.target_tokens.append(self.token_ids.ids[i+1:i+context_size+1])

    def get_data(self) : 
        return self.training_tokens, self.target_tokens 



## this class is used to read directly HF datasets
class DataLoaderHF : 
    def __init__(self,tokenizer_file:str,hf_data_sets:List[Tuple],
                 context_size:int=1024,Tokenize:bool=False, 
                ) : 

        self.tokenizer = Tokenizer.from_file(tokenizer_file) 
        self.training_tokens = [] 
        self.target_tokens = []
        self.datasets = []
        self.context_size = context_size
        self.hf_data_sets = hf_data_sets

        Path("./training/training_tokens/").mkdir(parents=True, exist_ok=True)

        if Tokenize : 
            for dataset_tuple in hf_data_sets :
                print(f"loading dataset {dataset_tuple[0]}")
                self.datasets.append((dataset_tuple[0],load_dataset(dataset_tuple[0],**dataset_tuple[1] ),dataset_tuple[2]))
                print(self.datasets[-1][1])
                print("ended loading ...")



    def tokenize_datasets(self,parquet_num = None) : 
        total_tokens = 0
        for i in range(len(self.datasets)) : 
            print(f"tokenizing  : {self.datasets[i][0]}")
            tokens_per_data_set = []            
            for j in range(self.datasets[i][1]['train'].num_rows) :
                print(f"tokenizing file num {j}")
                encoding = None
                encoding = self.tokenizer.encode(self.datasets[i][1]['train'][j][self.datasets[i][2]]) 
                tokens_per_data_set.extend(encoding.ids)
                tokens_per_data_set.extend([1])
                total_tokens += len(encoding.ids)
            name = self.datasets[i][0].replace("/","_")
            has_parquet = self.hf_data_sets[i][1].get("data_files",None)
            if has_parquet is not None : 
                name = "./training/training_tokens/" f"{name}_{parquet_num}"
            print(f"total tokens : {len(tokens_per_data_set)}")
            np.save(name,np.array(tokens_per_data_set))

    def build_data_loader(self,npy_files:List[str]) :
        tokenized_corpus = []
        for npy_file in npy_files : 
            print(f"loading {npy_file}")
            tokenized_dataset = np.load(npy_file, allow_pickle=True)
            print(f"num tokens : {tokenized_dataset.size}")
            tokenized_corpus.extend(tokenized_dataset.tolist()) 
            print(f"ended loading {npy_file}")
        
        print(f"total tokens number : {len(tokenized_corpus)}")
        for i in range(0,len(tokenized_corpus)-self.context_size, self.context_size) : 
            self.training_tokens.append(tokenized_corpus[i:i+self.context_size]) 
            self.target_tokens.append(tokenized_corpus[i+1:i+self.context_size+1]) 
        return self.training_tokens, self.target_tokens 

        

        




 
if __name__ == "__main__" :     
    tokenizer_file = "./fr_bpe_32k_422.json" 
    
    for i in [29, 7, 71, 63, 58, 36] : 
        hf_data_sets_tuples = [("PleIAs/French-PD-Newspapers",{"data_files":f"gallica_presse_{i}.parquet"},"complete_text")]
    
        data_loader_hf = DataLoaderHF(tokenizer_file,hf_data_sets_tuples,Tokenize=True) 
        data_loader_hf.tokenize_datasets(i)
