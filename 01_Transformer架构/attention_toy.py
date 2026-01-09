# attention_toy.py
# 目标：不使用 PyTorch/TensorFlow，只用 NumPy 模拟 Self-Attention 的核心逻辑
import numpy as np
import logging

# 配置日志输出，使用中文记录计算步骤
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')

def softmax(x):
    """计算 Softmax 归一化"""
    e_x = np.exp(x - np.max(x))
    return e_x / e_x.sum(axis=-1, keepdims=True)

def self_attention_demo():
    logging.info("开始 Self-Attention 模拟实验")
    
    # 假设我们有 3 个 token，每个 token 的向量维度是 4
    # 示例词汇: "I", "love", "AI"
    x = np.array([
        [1, 0, 1, 0], # "I" 的嵌入向量
        [0, 2, 0, 1], # "love" 的嵌入向量
        [1, 1, 1, 1]  # "AI" 的嵌入向量
    ], dtype=float)

    logging.info("步骤 1: 准备输入特征矩阵 X")
    print("1. 输入特征 (X):")
    print(x)

    # 在标准的 Self-Attention 中，Q (Query), K (Key), V (Value) 通常由 X 经过不同的线性变换得到
    # 为了简化演示核心逻辑，我们直接设定 Q = K = V = X
    q = k = v = x
    logging.info("步骤 2: 设定 Q, K, V (当前简化为 Q=K=V=X)")

    # 2. 计算注意力得分 (Scores) = Q * K^T
    # 这一步是在计算词与词之间的相似度或关联度
    scores = np.dot(q, k.T)
    logging.info("步骤 3: 计算注意力得分 (Q * K.T)")
    print("\n2. 原始注意力得分 (Q * K^T):")
    print(scores)

    # 3. Softmax 归一化，得到权重 (Weights)
    # 权重代表了每个 token 对序列中其他 token 的关注程度（总和为 1）
    weights = softmax(scores)
    logging.info("步骤 4: 执行 Softmax 归一化，获取注意力权重")
    print("\n3. 注意力权重 (Softmax 后的比例):")
    print(weights)
    print("注：每一行加起来和为 1.0")

    # 4. 加权求和 (Output) = Weights * V
    # 最终的输出是根据注意力权重对 Value 进行加权融合的结果
    output = np.dot(weights, v)
    logging.info("步骤 5: 根据权重对 V 进行加权求和，得到最终输出")
    print("\n4. 输出特征 (加权后的新向量):")
    print(output)

if __name__ == "__main__":
    self_attention_demo()
    print("\n[原理提示]：")
    print("- 注意力机制本质上是在做『动态加权求和』。")
    print("- 如果某个位置的权重很大，说明模型在处理当前词时，重点『关注』了那个位置的信息。")
    print("- 通过这种方式，模型可以捕捉到序列中长距离的依赖关系。")