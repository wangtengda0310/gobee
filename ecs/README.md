
# 问题
 - entity迭代的时候如何避免多次for循环

   在迭代entity时查找当前entity挂载component对应的system
 - 多个system关心同一个component时如何避免重复迭代

 - system关心的component有依赖关心时如何处理

    分组分回合,同一组的system在同一回合内执行
    有依赖的system在前一次components迭代完成后再次迭代
# 位运算
  1. 通过位运算来表示entity挂载了哪些component
2. 通过位运算来表示system需要哪些component
  system使用一个bitmask来表示需要哪些component
  与entity的bitmask进行比较,`与`位运算后结果还是system的bitmask,则entity满足system的要求