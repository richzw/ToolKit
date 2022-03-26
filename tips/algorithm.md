
- [第 N 个数](https://mp.weixin.qq.com/s/YaJ0nf7Y0juf6YTpVrV34g)
  - 给你一个整数 n ，请你在无限的整数序列 [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, ...] 中找出并返回第 n 位上的数字
  - 思路
    - 1、先找出 n 是落在长度为多少的数字上：`n = n - 9 * 1 * 1 - 9 * 10 * 2 - 9 * 100 * 3 - 9 * 1000 * 4 ...` 
    - 2、再找出 n 落在哪个数字上： `curNum = 10^(len-1) + ( n - 1 ) / len`
    - 3、再找出 n 落在这个数字的第几位
     ```c++
     class Solution {
         public int findNthDigit(int n) {
             if ( n < 10 ) return n;
             // len 表示数字的长度，即位数
             // 比如数字 123 长度为 3 ，即位数为 3 
             // 比如数字 5678 长度为 4 ，即位数为 4 
             int len = 1;
     
             // weight 表示数字所在的位数的数字的权重
             // 比如长度为 2 的数字的权重是 10
             // 比如长度为 3 的数字的权重是 100
             // 即 10^(len-1)
             int weight = 1;
     
             // 1、先找出 n 是落在长度为多少的数字上
             // 公式就是：n =  n - 9 * 1 * 1 - 9 * 10 * 2 - 9 * 100 * 3 - 9 * 1000 * 4 ... 
             // 直到在减的过程中发现 n 再去剪后面的数字为变成负数为止。
             // 由于 n 会很大，避免溢出，转一下类型
             while( n >  9 * len * (long)weight ){
                 // 公式就是：n =  n - 9 * 1 * 1 - 9 * 10 * 2 - 9 * 100 * 3 - 9 * 1000 * 4 ... 
                 n = n - 9 * len * weight ;
                  
                 // 数字的位数在不断增加
                 len += 1;
                 // 数字的权重也在不断增加
                 weight *= 10;
             }
             // 2、再找出 n 落在哪个数字上
             int curNum = weight + (n - 1 ) / len ;
             // 3、再找出 n 落在这个数字的第几位
             int count  =  (n - 1) % len;
             // 4、最后计算出这个数位来
             return (curNum / (int) Math.pow(10,len - count - 1 )) % 10;
         }
     }
     ```
- [环形链表](https://mp.weixin.qq.com/s/MLGauAOe2fpq1d18A69naQ)
  - 当 slow = 1 ， fast = 2 ，为什么 fast 和 slow 一定会相遇？
    - 1、假设，fast 在 slow 后方一个节点的位置，那么它们都跳一次之后，fast 和 slow 相遇了。
    - 2、假设，fast 在 slow 后方两个节点的位置，那么它们都跳一次之后，fast 和 slow 的距离缩短为 1，变成了上述假设 1 的问题，可以相遇。
    - 3、假设，fast 在 slow 后方 N 个节点的位置，那么它们都跳一次之后，fast 和 slow 的距离缩短为 N - 1，每条一次，都可以缩短一个单位，不断缩短，最终变成上述的假设 1。
    - 所以，fast 和 slow 一定会相遇。
  - 当 slow = 1 ， fast = 2 ，fast 和 slow 相遇时，slow 指针是否绕环超过一圈？
    - 当 slow 跑完一圈来到起始位置节点 2 时，由于 fast 速度是 slow 的两倍，那么 fast 必然跑了两圈也来到起始位置节点 2 ，两者相遇。
    - 而此时，fast 在 slow 的前方位置，意味着两者的距离是小于一个完整的环的节点数的，说明 fast 可以更快的追上 slow。
  - slow 和 fast 的移动步数有什么规则？
    - 只需要两者的步数不一致就行。
  - 能否设置为 slow 每次移动 1 步，fast 每次移动  3、4、5...步？
    - 假设环的周长为整数 L，slow 每次走 a 步，fast 每次走 b 步，a ≠ b。
    - 进入环后开始统计，相遇时两者都走了 t 次，如果把环拉成一根直线，相当于 fast 在距离 slow 为 t * (b – a ) 这样远的距离开始追击，每次追近 b – a 步，追了 t 次追上了。
    - 那么 t * (b – a ) / L 就是 fast 需要跑多少圈才能追上 slow 并刚好相遇的答案。
    - n 表示圈数，为整数才合理。
    - 因为如果 n 不是整数，代表 fast 和 slow 不在某个节点相遇，而是在节点与节点之间的位置相遇了。也就是说 n = t * (b – a ) / L 。
    - 很显然，L 为常数值，(b – a ) 为常数值，必然可以找到 t，使得 t * (b – a ) 是 L 的倍数。
  - 为什么设置 slow = 1 ， fast = 2 
    - 由于 n = t * (b – a ) / L 。
    - 则 n 越小越合适，也就是希望 t * (b – a ) 越小，希望 b – a 越小，因为 b 和 a 才是我们可以控制的。
    - 由此 b – a = 1 是一个最小的答案，即 fast 和 slow 相差一步最合理。
- [重排链表]
  - 给定一个单链表 L：L0→L1→…→Ln-1→Ln， 将其重新排列后变为：`L0→Ln→L1→Ln-1→L2→Ln-2→...
  - 找链表的中点
  - 反转链表
  - 合并链表
    ```c++
    public void reorderList(ListNode head) {
        if (head == null || head.next == null) {
            return;
        }
    
        // 步骤 1: 通过快慢指针找到链表中点
        // 通过调节快慢指针的起始位置，可以保证前半部分的长度大于等于后半部分
        ListNode slow = head, fast = head.next;
        while (fast != null && fast.next != null) {
            slow = slow.next;
            fast = fast.next.next;
        }
    
        // 步骤 2: 反转后半部分的链表
        // 在反转之前需要的一个操作是将前后半部分断开
        ListNode second = slow.next;
        slow.next = null;
        second = reverseList(second);
    
        // 步骤 3: 合并前半部分链表以及反转后的后半部分链表
        mergeList(head, second);
    }
    
    private ListNode reverseList(ListNode head) {
        ListNode prev = null, tmp = null, pointer = head;
        while (pointer != null) {
            tmp = pointer.next;
            pointer.next = prev;
            prev = pointer;
            pointer = tmp;
        }
    
        return prev;
    }
    
    private void mergeList(ListNode first, ListNode second) {
        ListNode dummy = new ListNode(0);
        ListNode pointer = dummy;
    
        while (first != null && second != null) {
            pointer.next = first;
            first = first.next;
            pointer.next.next = second;
            second = second.next;
            pointer = pointer.next.next;
        }
    
        // 因为我们之前找中点的时候保证了前半部分的长度不小于后半部分的长度
        // 因此交叉后，多出来的部分只可能是前半部分，判断前半部分即可
        if (first != null) {
            pointer.next = first;
        }
    }
    ```





