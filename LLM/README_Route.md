「程序员友好 · 论文驱动 · 可执行」的 LLM 论文阅读清单，不是“列论文名”，而是明确告诉你：每一篇该看哪几页、哪些公式可以跳、哪些图必须看懂。

你只要按顺序走完，基本就完成了从 0 → 能读懂 LLM 论文 的跃迁。

LLM 论文阅读顺序清单（精读指引版）

说明

✅ = 必读

⚠️ = 可跳/略读

🎯 = 核心理解目标

🧱 第一阶段：Transformer 本体（地基）
1️⃣ Attention Is All You Need（2017）✅

阅读顺序

第 1 页（Introduction）✅

第 3 页（Model Architecture）✅

第 4 页（Attention 公式 + 图）✅

第 5 页（Position Encoding）⚠️（看思想即可）

第 6 页（Training）⚠️

重点看什么

图 1（整体结构）

Scaled Dot-Product Attention

Multi-Head Attention 为什么要多头

可以跳过

BLEU 分数

超参数细节

🎯 读完你应该能回答

Transformer 为什么不需要 RNN？

2️⃣ The Illustrated Transformer（博客）✅

全文顺读

把论文中的符号“翻译成人话”

🎯 目标

能在脑中跑一遍 forward 流程

🧬 第二阶段：Transformer → 语言模型
3️⃣ GPT-1: Improving Language Understanding…（2018）✅

阅读顺序

第 1 页（Motivation）✅

第 2 页（Language Model Pre-training）✅

Figure 1（Pretrain → Finetune）✅

实验部分 ⚠️

重点

Decoder-only

自回归建模

为什么 finetune 少量数据就有效

🎯 目标

明白“语言模型”本身就是一个通用特征提取器

4️⃣ GPT-2: Language Models are Unsupervised…（2019）✅

阅读顺序

Abstract

第 1 页（Zero-shot）

Figure 2（Prompt 示例）✅

其余实验 ⚠️

重点

Prompt = 输入的一部分

不再显式 finetune

🎯 目标

理解 Prompt Learning 的起点

🚀 第三阶段：规模 = 能力来源
5️⃣ Scaling Laws for Neural Language Models（2020）✅

阅读顺序

Abstract

Figure 1、Figure 2 ✅

Section 3（Scaling Law）⚠️（不推公式）

只看图，不推导

Loss vs 参数

Loss vs 数据

Loss vs 计算量

🎯 目标

理解「不是模型聪明，是规模足够大」

6️⃣ GPT-3: Language Models are Few-Shot Learners（2020）✅

阅读顺序

Abstract

Figure 1（Few-shot）✅

Section 2.3（Training）⚠️

各种任务结果 ⚠️

🎯 目标

理解 Few-shot 是 Prompt 技巧，而不是结构变化

🧭 第四阶段：指令对齐（从“会说话”到“听话”）
7️⃣ InstructGPT（2022）✅（极重要）

阅读顺序

Abstract

Figure 2（RLHF 流程图）✅

Section 3（Methods）✅

PPO 细节 ⚠️

重点

SFT → RM → PPO 三步

人类偏好 ≠ 真实标签

🎯 目标

明白 ChatGPT 为什么“好用但不一定更聪明”

8️⃣ Training Language Models to Follow Instructions（2022）✅

阅读顺序

Abstract

Table 1（数据来源）

方法部分（SFT）✅

🎯 目标

理解指令微调如何改变输出分布

🧠 第五阶段：推理能力的形成
9️⃣ Chain-of-Thought Prompting（2022）✅

阅读顺序

Abstract

Figure 1（有/无 CoT 对比）✅

实验结果 ⚠️

🎯 目标

推理不是“模型多聪明”，而是输出路径更长

🔟 Self-Consistency（2022）⚠️（选读）

看思想即可

1️⃣1️⃣ ReAct（2023）⚠️

看 Prompt 设计，不看实验

⚙️ 第六阶段：工程与效率（程序员优势）
1️⃣2️⃣ LoRA（2021）✅

Figure 1

Method 部分

🎯 目标

为什么低秩适合微调

1️⃣3️⃣ FlashAttention（2022）⚠️

看动机和 IO 优化思想

🧩 第七阶段：开源模型理解
1️⃣4️⃣ LLaMA（2023）✅

Model Architecture

Training 数据规模

🎯 目标

知道开源模型和 GPT 的差距在哪里

🧠 阅读方法建议（非常重要）
✅ 正确方式

先看图

再读文字

最后扫公式

❌ 错误方式

从公式开始

试图全懂

🗺️ 一句话总结你现在的路线

你不是在学“模型”，而是在学：语言如何被建模为概率程序