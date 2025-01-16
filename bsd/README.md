# Bsd - Binary Similarity Detection

By leveraging the power of original Binshot and Ghidra IR, Bsd is focusing on cross-platform and optimization-free binary similarity detection.

# 模型微调流程
## 1. 样本转换
样本文件命名规则：`文件名-架构{arm,mips,x86,x64}-编译器{clang,gcc}-优化等级{O0,O1,O2,O3}`
示例：
```
openssl-1.0.1f-arm-gcc-O2
openssl-1.0.1f-mips-gcc-O2
libuv.so-x64-clang-O1
```
使用如下命令，将二进制文件样本转化为结构化指令，只会转换该文件夹中具有执行权限（x）的文件：
```
./gen_ghidra.sh 样本目录路径
```
该命名会对目录下的每个可执行文件生产同名的JSON文件。

## 2. 语料构建
```
python3 corpusgen_ghidra.py -d 样本目录路径 -pkl 样本目录路径 -o 语料目录路径 --binsim_gen
python3 corpusgen.py -f 语料目录路径/binsim.xxxx.corpus.txt -y binsimtask --split_data
```

此步骤会在语料目录路径下生成三个文件，其中`xxxx`是样本目录路径的最后一段（目录名称）。
```
binsim.xxxx.train.corpus.txt
binsim.xxxx.valid.corpus.txt
binsim.xxxx.test.corpus.txt
```

## 3. 模型微调
TODO: 改个名字，尝试加载sim_epx.model文件？

`基础模型文件`为本次微调的原模型中，名称以`bert_ep`开头的文件，如`原模型/model_sim/bert_ep0.model`，而非名称以`sim`开头的文件。输出模型目录不能和原模型相同。

微调后在系统中加载使用的是`输出模型目录/model_sim/sim_epx.model`，其中`x`是从0开始的数字，与模型训练进度有关，一般可以取数字最大的文件。

在训练过程中，按下Ctrl+C可以中断训练，进入测试流程。
```
python3 binshot.py \
    --bert_model_path 基础模型文件 \
    --vocab_path /media/xmoe/storage/buildroot-elf-5arch/buildroot-elf-5arch/corpus/pretrain.combined.corpus.voca \
    --output_path 输出模型目录 \
    --result_path result \
    --train_dataset 语料目录路径/binsim.xxxx.train.corpus.txt \
    --valid_dataset 语料目录路径/binsim.xxxx.valid.corpus.txt \
    --test_dataset 语料目录路径/binsim.xxxx.test.corpus.txt
```

### Special Thanks
* binshot