# tokenizer_lab.py
# 目标：直观理解文本是如何变成数字序列的
import logging

# 配置日志输出格式，使用中文记录过程
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')

def simple_tokenizer_demo():
    text = "LLM is powerful and amazing"
    logging.info("开始分词实验，原始文本: %s", text)
    
    # 1. 简单的空格分词 (Word-level)
    words = text.split()
    logging.info("步骤 1: 进行空格分词")
    print(f"原始文本: {text}")
    print(f"1. 空格分词结果: {words}")

    # 2. 模拟词表 (Vocabulary)
    # 在实际的大模型中，这个词表有数万个词，涵盖了各种子词或字符
    vocab = {
        "<PAD>": 0,
        "LLM": 1,
        "is": 2,
        "powerful": 3,
        "and": 4,
        "amazing": 5,
        "<UNK>": 6
    }
    logging.info("步骤 2: 加载模拟词表，当前词表大小: %d", len(vocab))
    
    # 3. 编码 (Encoding): Text -> IDs
    # 将每个词转换成词表中对应的索引数字
    ids = [vocab.get(word, vocab["<UNK>"]) for word in words]
    logging.info("步骤 3: 执行编码 (Text -> IDs)")
    print(f"2. 编码后的 ID 序列: {ids}")

    # 4. 解码 (Decoding): IDs -> Text
    # 程序员视角：这就是一个简单的 Map 逆向查表
    reverse_vocab = {v: k for k, v in vocab.items()}
    decoded_text = " ".join([reverse_vocab.get(i, "<UNK>") for i in ids])
    logging.info("步骤 4: 执行解码 (IDs -> Text)")
    print(f"3. 解码还原的文本: {decoded_text}")

if __name__ == "__main__":
    simple_tokenizer_demo()
    print("\n[原理提示]：")
    print("- 大模型不认识汉字或单词，它们只认识数字序列。")
    print("- 现代模型（如 GPT）使用的是 BPE（Byte Pair Encoding）分词，")
    print("- 它们会将词拆分成更小的“子词”（Subwords），以便在词表大小和语义完整性之间取得平衡。")