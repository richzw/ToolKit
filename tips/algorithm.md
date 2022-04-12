
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
- [数组遍历题](https://mp.weixin.qq.com/s/57lf24oMyNIwBxV2kuW9cg)
  - 给定一个包含 n + 1 个整数的数组 nums，其数字都在 1 到 n 之间（包括 1 和 n），可知至少存在一个重复的整数。假设只有一个重复的整数，找出这个重复的数。
    - 不能更改原数组（假设数组是只读的）。
    - 只能使用额外的 O(1) 的空间。
    - 时间复杂度小于 O(n^2) 。
    - 数组中只有一个重复的数字，但它可能不止重复出现一次。
  - 解析
    - 不能改变数组 导致无法排序，也无法用 index 和元素建立关系；
    - 只能使用 O(1) 的空间 意味着使用哈希表去计数这条路也走不通；
    - 时间复杂度必须小于 O(n^2) 表示暴力求解也不行；
    - 重复的元素可重复多次 这一条加上后，本来可以通过累加求和然后做差 sum(array) - sum(1,2,...,n) 的方式也变得不可行。
  - Deep dive
    - 什么样的算法可以不使用额外的空间解决数组上面的搜索问题？ - 二分查找
    - 二分法的使用 并不一定 需要在排序好的数组上面进行，不要让常见的例题限制了你的思路，二分法还有一个比较高级的用法叫做 按值二分。
      - 如果选中的数 小于 我们要找的答案，那么整个数组中小于或等于该数的元素个数必然小于或等于该元素的值;
      - 如果选中的数 大于或等于 我们要找的答案，那么整个数组中小于或等于该数的元素个数必然 大于 该元素的值
    - 另外一种 O(n) 的解法借鉴快慢指针找交点的思想
  - Code
    ```c++
    //二分查找
    class Solution {
        public int findDuplicate(int[] nums) {
             int len = nums.length;
            int start = 1;
            int end = len - 1;
    
            while (start < end) {
                int mid = start + (end - start) / 2;
                int counter = 0;
                for (int num:nums) {
                    if (num <= mid) {
                        counter++;
                    }
                }
                if (counter > mid) {
                    end = mid;
                } else {
                    start = mid + 1;
                }
            }
            return start;
        }
    }
    ```
    ```c
    //快慢指针
    public int findDuplicate(int[] nums) {        
        int fast = nums[nums[0]];
        int slow = nums[0];
    
        while (fast != slow) {
            fast = nums[nums[fast]];
            slow = nums[slow];
        }
    
        slow = 0;
        while (fast != slow) {
            fast = nums[fast];
            slow = nums[slow];
        }
    
        return slow;
    }
    ```
- [数组双指针](https://mp.weixin.qq.com/s/Z-oYzx9O1pjiym6HtKqGIQ)
  - 双指针技巧在处理数组和链表相关问题时经常用到，主要分为两类：左右指针和快慢指针。
  - 数组问题中比较常见且难度不高的的快慢指针技巧，是让你原地修改数组
    - 删除有序数组中的重复项（简单）
    - 删除排序链表中的重复元素（简单）
    - 移除元素（简单）
      ```shell
      int removeElement(int[] nums, int val) {
          int fast = 0, slow = 0;
          while (fast < nums.length) {
              if (nums[fast] != val) {
                  nums[slow] = nums[fast];
                  slow++;
              }
              fast++;
          }
          return slow;
      }
      ```
    - 移动零（简单）- 题目让我们将所有 0 移到最后，其实就相当于移除nums中的所有 0，然后再把后面的元素都赋值为 0 即可。
    - [滑动窗口算法核心框架详解](https://mp.weixin.qq.com/s?__biz=MzAxODQxMDM0Mw==&mid=2247485141&idx=1&sn=0e4583ad935e76e9a3f6793792e60734&scene=21#wechat_redirect)
      ```shell
      /* 滑动窗口算法框架 */
      void slidingWindow(string s, string t) {
          unordered_map<char, int> need, window;
          for (char c : t) need[c]++;
      
          int left = 0, right = 0;
          int valid = 0; 
          while (right < s.size()) {
              char c = s[right];
              // 右移（增大）窗口
              right++;
              // 进行窗口内数据的一系列更新
      
              while (window needs shrink) {
                  char d = s[left];
                  // 左移（缩小）窗口
                  left++;
                  // 进行窗口内数据的一系列更新
              }
          }
      }
      ```
  - 左右指针的常用算法
    - 二分查找
    - 反转字符串（简单）
    - 两数之和 II - 输入有序数组（中等）
    - 最长回文子串（中等）
      ```shell
      for 0 <= i < len(s):
          找到以 s[i] 为中心的回文串
          找到以 s[i] 和 s[i+1] 为中心的回文串
          更新答案
      ```








