# ECS并发与数据竞争相关讨论资料汇总

本文件整理了ECS（Entity-Component-System）架构下关于system并发修改component时的数据竞争、并发安全、调度与工程实践的相关文章与讨论，便于查阅和深入理解。

---

## 1. Exploring the Theory and Practice of Concurrency in the Entity-Component-System Pattern
**概要**：系统性分析ECS模式下的并发与确定性，提出了Core ECS形式化模型，讨论了system并发调度、数据竞争的理论基础与主流ECS框架的实际实现差异。
- [论文PDF（英文）](https://users.soe.ucsc.edu/~lkuper/papers/core-ecs-draft.pdf)

## 2. Data Structures for Entity Systems: Multi-threading and Networking
**概要**：讨论ECS在多线程和网络环境下的数据结构设计，强调只读数据的多线程安全、数据副本与事件驱动、Copy-on-Write等工程实践，分析多线程下的内存一致性与数据竞争问题。
- [T-machine.org原文](https://t-machine.org/index.php/2015/05/02/data-structures-for-entity-systems-multi-threading-and-networking/)

## 3. Why by-convention read-only shared state is bad
**概要**：以游戏引擎多线程渲染为例，分析"约定只读"带来的并发隐患，提出RCU（Read Copy Update）、深拷贝等可持续工程方案，强调复杂系统中并发约束的显式化。
- [Richard Geldreich's Blog](https://richg42.blogspot.com/2015/12/why-by-convention-read-only-shared.html)

## 4. Please don't put ECS into your game engine（Rust论坛ECS争议与并发讨论）
**概要**：Rust社区关于ECS架构的争议，涉及ECS并发、状态一致性、事务性、可扩展性等问题，包含多位开发者对ECS并发与数据一致性的深入讨论。
- [Rust官方论坛讨论](https://users.rust-lang.org/t/please-dont-put-ecs-into-your-game-engine/49305)

## 5. Hacker News：ECS、OOP与并发的哲学与工程讨论
**概要**：Hacker News上关于ECS本质、数据与行为分离、并发与数据一致性的讨论，涉及ECS与OOP、数据导向设计、并发可扩展性等话题。
- [Hacker News讨论串](https://news.ycombinator.com/item?id=27663361)

---

> 建议结合理论模型与主流ECS框架的实际实现，深入理解ECS并发调度、数据竞争与工程可持续性问题。 