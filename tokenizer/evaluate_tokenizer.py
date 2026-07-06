"""Evaluate the French BPE tokenizer on compression, roundtrip, and linguistic quality."""

from tokenizers import Tokenizer
from collections import Counter
from datasets import load_dataset
import random
import sys


def load_eval_corpus(n_sentences=500, seed=42):
    """Load held-out sentences from nirantk/french-books for evaluation."""
    random.seed(seed)
    ds = load_dataset("nirantk/french-books")
    train = ds["train"]

    # Pick from the last 200 books (held-out from small training corpus)
    rows = list(range(len(train) - 200, len(train)))
    random.shuffle(rows)

    sentences = []
    for idx in rows:
        text = train[idx]["complete_text"]
        for paragraph in text.split("\n\n"):
            for sentence in paragraph.split(". "):
                s = sentence.strip()
                if 30 < len(s) < 500:
                    sentences.append(s)
                if len(sentences) >= n_sentences:
                    return sentences
    return sentences


def evaluate(tok: Tokenizer, corpus: list[str]):
    vocab_size = tok.get_vocab_size()
    print(f"Vocab size: {vocab_size:,}")
    print()

    total_chars = 0
    total_tokens = 0
    total_words = 0
    total_bytes = 0
    roundtrip_failures = []
    char_fallback_count = 0
    token_freq = Counter()
    all_tokens_str = []

    for text in corpus:
        if not text:
            continue

        encoding = tok.encode(text)
        ids = encoding.ids
        pieces = encoding.tokens
        decoded = tok.decode(ids)

        total_chars += len(text)
        total_bytes += len(text.encode("utf-8"))
        total_tokens += len(ids)
        total_words += len(text.split())
        token_freq.update(ids)
        all_tokens_str.extend(pieces)

        char_fallback_count += sum(1 for p in pieces if len(p) <= 1)

        if decoded != text:
            roundtrip_failures.append((text, decoded))

    # --- Compression ---
    print("=" * 60)
    print("COMPRESSION")
    print("=" * 60)
    fertility = total_tokens / total_words
    chars_per_token = total_chars / total_tokens
    bytes_per_token = total_bytes / total_tokens
    print(f"  Fertility (tokens/word):    {fertility:.2f}")
    print(f"  Chars per token:            {chars_per_token:.2f}")
    print(f"  Bytes per token:            {bytes_per_token:.2f}")
    print(f"  Total: {total_chars:,} chars → {total_tokens:,} tokens ({total_words:,} words)")
    print()

    # --- Roundtrip ---
    print("=" * 60)
    print("ROUNDTRIP FIDELITY")
    print("=" * 60)
    n_texts = sum(1 for t in corpus if t)
    if roundtrip_failures:
        print(f"  FAILURES: {len(roundtrip_failures)}/{n_texts}")
        for orig, dec in roundtrip_failures[:5]:
            print(f"    original: {orig[:80]!r}")
            print(f"    decoded:  {dec[:80]!r}")
            print()
    else:
        print(f"  All {n_texts} texts round-trip perfectly.")
    print()

    # --- Unknown / fallback ---
    print("=" * 60)
    print("COVERAGE")
    print("=" * 60)
    print(f"  Single-char fallbacks:      {char_fallback_count}/{total_tokens} ({char_fallback_count/total_tokens:.2%})")
    print(f"  Unique tokens used:         {len(token_freq):,}/{vocab_size:,} ({len(token_freq)/vocab_size:.1%})")
    print()

    # --- Token length distribution ---
    print("=" * 60)
    print("TOKEN LENGTH DISTRIBUTION (by characters)")
    print("=" * 60)
    lengths = [len(p) for p in all_tokens_str if p.strip()]
    length_counts = Counter(lengths)
    for length in sorted(length_counts):
        bar = "█" * min(length_counts[length], 60)
        print(f"  {length:2d} chars: {length_counts[length]:4d}  {bar}")
    print()

    # --- Most / least common tokens ---
    print("=" * 60)
    print("TOP 20 MOST FREQUENT TOKENS")
    print("=" * 60)
    for tid, count in token_freq.most_common(20):
        piece = tok.id_to_token(tid)
        print(f"  {piece!r:20s}  (id={tid:5d})  count={count}")
    print()

    # --- Morphological spot checks ---
    print("=" * 60)
    print("MORPHOLOGICAL ALIGNMENT (spot checks)")
    print("=" * 60)
    spot_checks = [
        "anticonstitutionnellement",
        "désindustrialisation",
        "incontestablement",
        "aujourd'hui",
        "l'Assemblée",
        "quelqu'un",
        "rétropropagation",
        "méditerranéennes",
        "authentification",
    ]
    for word in spot_checks:
        pieces = tok.encode(word).tokens
        print(f"  {word:30s} → {pieces}")
    print()

    # --- French-specific: apostrophe handling ---
    print("=" * 60)
    print("FRENCH APOSTROPHE HANDLING")
    print("=" * 60)
    apostrophe_tests = [
        "l'homme",
        "l'économie",
        "d'abord",
        "n'est-ce pas",
        "quelqu'un",
        "aujourd'hui",
        "c'est-à-dire",
    ]
    for phrase in apostrophe_tests:
        pieces = tok.encode(phrase).tokens
        print(f"  {phrase:25s} → {pieces}")
    print()

    # --- Number tokenization ---
    print("=" * 60)
    print("NUMBER TOKENIZATION")
    print("=" * 60)
    number_tests = [
        "42",
        "3.14159",
        "1 000 000",
        "2023",
        "14h30",
        "0,9 %",
        "2 803,04",
        "+23 %",
    ]
    for num in number_tests:
        pieces = tok.encode(num).tokens
        print(f"  {num:20s} → {pieces}")
    print()


if __name__ == "__main__":
    if len(sys.argv) != 2:
        print(f"Usage: python {sys.argv[0]} <tokenizer.json>")
        sys.exit(1)
    tok = Tokenizer.from_file(sys.argv[1])
    print("Loading evaluation corpus from nirantk/french-books...")
    corpus = load_eval_corpus()
    print(f"Loaded {len(corpus)} sentences\n")
    evaluate(tok, corpus)
