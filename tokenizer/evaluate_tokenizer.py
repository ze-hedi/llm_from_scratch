"""Evaluate the French BPE tokenizer on compression, roundtrip, and linguistic quality."""

import sentencepiece as spm
from collections import Counter

MODEL_PATH = "fr_bpe_32k.model"

# French test corpus — mix of registers and edge cases
FRENCH_CORPUS = [
    # Formal / legal
    "L'article 1134 du Code civil dispose que les conventions légalement formées tiennent lieu de loi à ceux qui les ont faites.",
    "Le tribunal de grande instance de Paris a rendu son jugement le 15 mars 2024.",
    "Conformément aux dispositions de l'article L. 121-1 du Code de la consommation, tout contrat conclu à distance peut faire l'objet d'une rétractation.",
    # Literary
    "Longtemps, je me suis couché de bonne heure. Parfois, à peine ma bougie éteinte, mes yeux se fermaient si vite que je n'avais pas le temps de me dire : « Je m'endors. »",
    "Il est des parfums frais comme des chairs d'enfants, doux comme les hautbois, verts comme les prairies.",
    # Colloquial
    "T'as vu le match hier soir ? C'était vraiment n'importe quoi, l'arbitre a sifflé un penalty complètement bidon.",
    "J'suis pas sûr qu'on puisse y aller demain, ça dépend de la météo.",
    # Technical
    "L'algorithme de rétropropagation calcule le gradient de la fonction de coût par rapport aux poids du réseau de neurones.",
    "Le protocole TCP/IP assure la transmission fiable des paquets sur le réseau Internet.",
    # Numbers and special chars
    "Le PIB de la France en 2023 s'élevait à 2 803,04 milliards d'euros, soit une croissance de 0,9 %.",
    "Rendez-vous à 14h30 au 42, rue de la République — 3e étage, porte B.",
    # Accents and diacritics
    "L'élève a réussi l'épreuve grâce à sa maîtrise des règles de grammaire française.",
    "Les forêts méditerranéennes abritent une biodiversité considérable, menacée par les incendies récurrents.",
    # Apostrophes and contractions (French-specific)
    "Aujourd'hui, l'Assemblée nationale s'est réunie pour débattre de l'avenir de l'économie.",
    "Quelqu'un d'autre aurait-il pu prévoir qu'il n'y aurait pas d'issue ?",
    # Rare / morphologically complex words
    "L'anticonstitutionnellement célèbre mot est souvent cité comme le plus long de la langue française.",
    "La désindustrialisation progressive des régions septentrionales a entraîné un exode rural sans précédent.",
    # Mixed content
    "Version 3.2.1 — changelog : correction du bug #4521, amélioration des performances (+23 %), refactorisation du module d'authentification.",
    # Whitespace edge cases
    "  Texte avec   des espaces    multiples  et\ttabulations\t.",
    "",  # empty string
]


def evaluate(sp: spm.SentencePieceProcessor, corpus: list[str]):
    vocab_size = sp.get_piece_size()
    print(f"Vocab size: {vocab_size:,}")
    print(f"Special tokens — BOS: {sp.bos_id()}, EOS: {sp.eos_id()}, UNK: {sp.unk_id()}")
    print()

    total_chars = 0
    total_tokens = 0
    total_words = 0
    total_bytes = 0
    roundtrip_failures = []
    unk_count = 0
    char_fallback_count = 0
    token_freq = Counter()
    all_tokens_str = []

    for text in corpus:
        if not text:
            continue

        ids = sp.encode(text, out_type=int)
        pieces = sp.encode(text, out_type=str)
        decoded = sp.decode(ids)

        total_chars += len(text)
        total_bytes += len(text.encode("utf-8"))
        total_tokens += len(ids)
        total_words += len(text.split())
        token_freq.update(ids)
        all_tokens_str.extend(pieces)

        unk_count += ids.count(sp.unk_id())
        char_fallback_count += sum(1 for p in pieces if len(p.replace("▁", "")) <= 1 and p != "▁")

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
    print(f"  UNK tokens:                 {unk_count}/{total_tokens} ({unk_count/total_tokens:.4%})")
    print(f"  Single-char fallbacks:      {char_fallback_count}/{total_tokens} ({char_fallback_count/total_tokens:.2%})")
    print(f"  Unique tokens used:         {len(token_freq):,}/{vocab_size:,} ({len(token_freq)/vocab_size:.1%})")
    print()

    # --- Token length distribution ---
    print("=" * 60)
    print("TOKEN LENGTH DISTRIBUTION (by characters)")
    print("=" * 60)
    lengths = [len(p.replace("▁", "")) for p in all_tokens_str if p != "▁"]
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
        piece = sp.id_to_piece(tid)
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
        pieces = sp.encode(word, out_type=str)
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
        pieces = sp.encode(phrase, out_type=str)
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
        pieces = sp.encode(num, out_type=str)
        print(f"  {num:20s} → {pieces}")
    print()


if __name__ == "__main__":
    sp = spm.SentencePieceProcessor(model_file=MODEL_PATH)
    evaluate(sp, FRENCH_CORPUS)
