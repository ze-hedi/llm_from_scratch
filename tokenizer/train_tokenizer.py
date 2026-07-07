from tokenizers import Tokenizer, pre_tokenizers, models, trainers, decoders
from tokenizers.pre_tokenizers import Split
from tokenizers.normalizers import NFC


class TokenizerTrainer:
    def __init__(
        self,
        training_corpus: str,
        special_tokens: list[str],
        pattern: str,
        output: str,
        vocab_size: int = 32768,
        max_token_length: int = 64,
    ):
        self.training_corpus = training_corpus
        self.special_tokens = special_tokens
        self.pattern = pattern
        self.output = output
        self.vocab_size = vocab_size
        self.max_token_length = max_token_length

    def train(self) -> Tokenizer:
        tokenizer = Tokenizer(models.BPE())
        tokenizer.normalizer = NFC()
        tokenizer.pre_tokenizer = pre_tokenizers.Sequence([
            Split(pattern=self.pattern, behavior="isolated", invert=False),
            pre_tokenizers.ByteLevel(add_prefix_space=False, use_regex=False),
        ])
        tokenizer.decoder = decoders.ByteLevel()

        trainer = trainers.BpeTrainer(
            vocab_size=self.vocab_size,
            special_tokens=self.special_tokens,
            show_progress=True,
            max_token_length=self.max_token_length,
        )

        print("Training tokenizer...")
        tokenizer.train([self.training_corpus], trainer=trainer)
        tokenizer.save(self.output)
        print(f"Training complete — {self.output} written. Vocab size: {tokenizer.get_vocab_size()}")
        return tokenizer


if __name__ == "__main__":
    FRENCH_PATTERN = r"""[^\r\n\p{L}\p{N}]?[\p{L}\p{M}]+(?:['\u2019][\p{L}\p{M}]+)*|\p{N}{1,3}| ?[^\s\p{L}\p{N}]+[\r\n]*|\s*[\r\n]+|\s+(?!\S)|\s+"""

    trainer = TokenizerTrainer(
        training_corpus="tokenizer_training_corpus_small.txt",
        special_tokens=["<|pad|>", "<|eos|>", "<|user|>", "<|assistant|>"],
        pattern=FRENCH_PATTERN,
        output="fr_bpe_32k.json",
    )
    trainer.train()
