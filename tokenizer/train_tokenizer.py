from pathlib import Path
from tokenizers import Tokenizer, pre_tokenizers, models, trainers, decoders
from tokenizers.pre_tokenizers import Split
from tokenizers.normalizers import NFC
from dataset_loader import all_docs

GPT4_PATTERN = r"""'(?i:[sdmt]|ll|ve|re)|[^\r\n\p{L}\p{N}]?+\p{L}+|\p{N}{1,3}| ?[^\s\p{L}\p{N}]++[\r\n]*|\s*[\r\n]|\s+(?!\S)|\s+"""
FRENCH_PATTERN = r"""[^\r\n\p{L}\p{N}]?[\p{L}\p{M}]+(?:['\u2019][\p{L}\p{M}]+)*|\p{N}{1,3}| ?[^\s\p{L}\p{N}]+[\r\n]*|\s*[\r\n]+|\s+(?!\S)|\s+"""

CORPUS_FILE = Path("tokenizer_corpus.txt")

# Step 1: write corpus to disk (skip if already exists)
if CORPUS_FILE.exists():
    print(f"Corpus file already exists at {CORPUS_FILE}, skipping write.")
else:
    print("Writing corpus to disk...")
    count = 0
    with open(CORPUS_FILE, "w") as f:
        for doc in all_docs():
            f.write(doc + "\n")
            count += 1
            if count % 1_000_000 == 0:
                print(f"  {count:,} paragraphs written...")
    print(f"Done — {count:,} paragraphs written to {CORPUS_FILE}")

# Step 2: train from file
tokenizer = Tokenizer(models.BPE())
tokenizer.normalizer = NFC()
tokenizer.pre_tokenizer = pre_tokenizers.Sequence([
    Split(pattern=FRENCH_PATTERN, behavior="isolated", invert=False),
    pre_tokenizers.ByteLevel(add_prefix_space=False, use_regex=False),
])
tokenizer.decoder = decoders.ByteLevel()

trainer = trainers.BpeTrainer(
    vocab_size=32768,
    special_tokens=["<|pad|>", "<|eos|>", "<|eof|>, <|user|>, <|assistant|>"],
    show_progress=True,
    max_token_length=64,
)

print("Training tokenizer...")
tokenizer.train([str(CORPUS_FILE)], trainer=trainer)
tokenizer.save("fr_bpe_32k.json")

print(f"Training complete — fr_bpe_32k.json written. Vocab size: {tokenizer.get_vocab_size()}")
