# Stage 1: Transformer 本体（地基）

### 🎯 核心理解目标
- [ ] **图 1 (整体结构)**：能分辨出什么是 Encoder，什么是 Decoder。
- [ ] **Scaled Dot-Product Attention**：理解为什么叫“注意力”，本质上就是算权重。
- [ ] **Multi-Head Attention**：理解为什么要多头（就像多个人从不同角度看同一段话）。
- [ ] **位置编码 (Position Encoding)**：明白为什么没有 RNN 的 Transformer 还能知道词的顺序。

### 📄 必读建议
- **论文**: *Attention Is All You Need*
- **重点**: 第 1-4 页是黄金内容，第 5 页看思想，后面实验部分可以快速跳过。

### 💻 实验建议
- 尝试运行 `attention_toy.py` (即将创建)，手动改改数值，看注意力权重如何变化。
