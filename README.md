# 🚀 Go-based Byte Pair Encoding (BPE)

A Go implementation of Byte Pair Encoding (BPE), inspired by [Andrej Karpathy's](https://www.youtube.com/watch?v=zduSFxRajkE) tutorial.  Support the `tiktoken` file format from [OpenAI](https://github.com/openai/tiktoken). You can fetch pretrained encodings directly from [OpenAI's github](https://github.com/openai/tiktoken/blob/main/tiktoken_ext/openai_public.py) 📦.

# ✨ Features
- 🔤 Tokenizes arbitrary byte sequences (not just text!)
- 🧩 Special token support with whitelisting
- 🧪 Regex-based input splitting
- ⚠️ Not a drop-in replacement for OpenAI’s tokenizer

# 🗂️ Project Structure

```
├── bpeprocessor.go          # Interface definition
├── go.mod                   # Module config
├── README.md                # You're reading it!
├── regextiktokenproc.go     # Regex-enhanced BPE processor
├── regextiktokenproc_test.go
├── tiktokenproc.go          # Core BPE for OpenAI's .tiktoken format
├── tiktokenproc_test.go
└── testdata/
    └── cl100k_base.tiktoken # Sample encoding data
```

# 💡 Key Takeaways
- 🧪 Fuzz testing in Go is powerful — used to test `decode(encode(x)) == x` across edge cases
- 🌐 UTF-8 is full of surprises — beware of multi-byte characters
- 🛠️ Byte slice manipulation in Go can be... tricky & annoying😅
-   🔍 Go’s regex capabilities are fundamentally different from Python’s 🐍 — beware of surprises!

# 📄 License

The file testdata/cl100k_base.tiktoken is under the MIT License.© 2022 OpenAI, Shantanu Jain.

This project itself is also licensed under the MIT License — feel free to use, fork, or contribute!