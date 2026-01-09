📘《LLM 前置名词词典（程序员版）》

目标：

不讲历史八卦

不推公式

用程序员能直觉理解的方式解释

看完后：再看 Transformer / 论文不会“词都不认识”

你可以直接把这一份放进你的 Git 仓库，作为 docs/llm-glossary.md。

📘 LLM 前置名词词典（程序员版）
一、AI / ML 基础概念
Machine Learning（机器学习）

用数据自动拟合函数，而不是写规则

input → model → output

Model（模型）

一个带参数的函数

y = f(x; θ)

Parameters（参数）

模型内部可调的数字
LLM = 参数特别多（10⁹~10¹²）

Training（训练）

用数据反复调整参数，使输出更“像对的”

Loss（损失函数）

衡量模型有多“离谱”的函数

越小越好

LLM 常用：Cross Entropy

Backpropagation（反向传播）

自动算“哪个参数该为错误负责”

程序员视角：

自动求导 + 梯度分发

二、神经网络基础
Neural Network（神经网络）

多层函数的组合

Linear → Activation → Linear → ...

Activation Function（激活函数）

给模型“非线性能力”

常见：

ReLU

GELU（Transformer 常用）

Overfitting（过拟合）

模型记住了训练数据，但泛化能力差

三、经典模型（历史背景）
CNN（卷积神经网络）

专门处理“局部结构”的网络

常用于图像

不直接用于 LLM

重要在于：证明了深度网络可行

RNN（循环神经网络）

用循环结构处理序列

能“记住过去”

缺点：

难并行

长序列效果差

📌 Transformer 诞生的背景

四、语言模型（非常关键）
Language Model（语言模型）

预测下一个 token 的概率模型

P(tokenₙ | token₁…tokenₙ₋₁)

Token（词元）

模型实际处理的最小文本单位

不是“词”

可能是：

子词

单个字符

Tokenizer（分词器）

把文本 → token 序列

常见：

BPE

WordPiece

Vocabulary（词表）

所有 token 的集合

Embedding（嵌入）

把 token ID 映射成向量

token → vector

Autoregressive（自回归）

每次预测一个 token，并把它当输入

五、Transformer / LLM 核心概念（提前预热）
Attention（注意力）

决定“当前 token 应该关注谁”

不是人类注意力，是加权求和

Self-Attention（自注意力）

token 之间互相关注

Query / Key / Value（QKV）

用来算 attention 权重的三个向量

程序员类比：

Query：我在找什么

Key：我有什么

Value：真正要拿的内容

Transformer

基于 Attention 的序列模型

特点：

无循环

可并行

适合长文本

Decoder-only

只用 Transformer 的 Decoder 部分

GPT / LLaMA 都是这种

六、LLM 专属名词
LLM（Large Language Model）

参数规模极大的语言模型

本质没变，只是更大

Pre-training（预训练）

用海量文本学“语言结构”

Fine-tuning（微调）

用特定任务数据调整模型

Prompt

输入给模型的上下文文本

Inference（推理）

使用模型生成结果（不更新参数）

Hallucination（幻觉）

模型生成看似合理但错误的内容

七、你现在“可以暂时忽略”的词（放心）

BLEU / ROUGE

PPO 细节

Beam Search

混合专家（MoE）

Flash Attention

👉 现在忽略是正确的

🧠 如何使用这份词典（非常重要）
✅ 正确方式

看论文时：

碰到不熟的词 → 回来查

不要死记

❌ 错误方式

背诵

一次全懂