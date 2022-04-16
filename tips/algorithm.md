
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
- [nSum问题](https://mp.weixin.qq.com/s?__biz=MzAxODQxMDM0Mw==&mid=2247485789&idx=1&sn=efc1167b85011c019e05d2c3db1039e6&scene=21#wechat_redirect)
  - twoSum 问题
    - 暴力
    - hashmap
    - 先对 nums 排序，然后利用前文「双指针技巧汇总」写过的左右双指针技巧，从两端相向而行就行了
  - 魔改题目
    - nums 中可能有多对儿元素之和都等于 target，请你的算法返回所有和为 target 的元素对儿，其中不能出现重复。
      比如说输入为 nums = [1,3,1,2,2,3], target = 4，那么算法返回的结果就是：[[1,3],[2,2]]
      ```c++
      vector<vector<int>> twoSumTarget(vector<int>& nums, int target) {
          // nums 数组必须有序
          sort(nums.begin(), nums.end());
          int lo = 0, hi = nums.size() - 1;
          vector<vector<int>> res;
          while (lo < hi) {
              int sum = nums[lo] + nums[hi];
              int left = nums[lo], right = nums[hi];
              if (sum < target) {
                  while (lo < hi && nums[lo] == left) lo++;
              } else if (sum > target) {
                  while (lo < hi && nums[hi] == right) hi--;
              } else {
                  res.push_back({left, right});
                  while (lo < hi && nums[lo] == left) lo++;
                  while (lo < hi && nums[hi] == right) hi--;
              }
          }
          return res;
      }
      ```
  - 3Sum 4Sum nSum
    ```go
    /* 注意：调用这个函数之前一定要先给 nums 排序 */
    vector<vector<int>> nSumTarget(
        vector<int>& nums, int n, int start, int target) {
    
        int sz = nums.size();
        vector<vector<int>> res;
        // 至少是 2Sum，且数组大小不应该小于 n
        if (n < 2 || sz < n) return res;
        // 2Sum 是 base case
        if (n == 2) {
            // 双指针那一套操作
            int lo = start, hi = sz - 1;
            while (lo < hi) {
                int sum = nums[lo] + nums[hi];
                int left = nums[lo], right = nums[hi];
                if (sum < target) {
                    while (lo < hi && nums[lo] == left) lo++;
                } else if (sum > target) {
                    while (lo < hi && nums[hi] == right) hi--;
                } else {
                    res.push_back({left, right});
                    while (lo < hi && nums[lo] == left) lo++;
                    while (lo < hi && nums[hi] == right) hi--;
                }
            }
        } else {
            // n > 2 时，递归计算 (n-1)Sum 的结果
            for (int i = start; i < sz; i++) {
                vector<vector<int>> 
                    sub = nSumTarget(nums, n - 1, i + 1, target - nums[i]);
                for (vector<int>& arr : sub) {
                    // (n-1)Sum 加上 nums[i] 就是 nSum
                    arr.push_back(nums[i]);
                    res.push_back(arr);
                }
                while (i < sz - 1 && nums[i] == nums[i + 1]) i++;
            }
        }
        return res;
    }
    ```
- [单链表的六大解题套路](https://mp.weixin.qq.com/s?__biz=MzAxODQxMDM0Mw==&mid=2247492022&idx=1&sn=35f6cb8ab60794f8f52338fab3e5cda5&scene=21#wechat_redirect)
  - 合并两个有序链表
  - 合并k个有序链表 
    - 用到 优先级队列（二叉堆） 这种数据结构，把链表节点放入一个最小堆，就可以每次获得k个节点中的最小节点
    - 优先队列pq中的元素个数最多是k，所以一次poll或者add方法的时间复杂度是O(logk)；所有的链表节点都会被加入和弹出pq，所以算法整体的时间复杂度是O(Nlogk)，其中k是链表的条数，N是这些链表的节点总数。
  - 寻找单链表的倒数第k个节点
  - 寻找单链表的中点
  - 判断单链表是否包含环并找出环起点
    - 我们假设快慢指针相遇时，慢指针slow走了k步，那么快指针fast一定走了2k步：
    - fast一定比slow多走了k步，这多走的k步其实就是fast指针在环里转圈圈，所以k的值就是环长度的「整数倍」。
    - 假设相遇点距环的起点的距离为m，那么结合上图的 slow 指针，环的起点距头结点head的距离为k - m，也就是说如果从head前进k - m步就能到达环起点。
    - 巧的是，如果从相遇点继续前进k - m步，也恰好到达环起点。因为结合上图的 fast 指针，从相遇点开始走k步可以转回到相遇点，那走k - m步肯定就走到环起点了：
  - 判断两个单链表是否相交并找出交点
    - 直接的想法可能是用HashSet记录一个链表的所有节点，然后和另一条链表对比，但这就需要额外的空间。
    - 我们可以让p1遍历完链表A之后开始遍历链表B，让p2遍历完链表B之后开始遍历链表A，这样相当于「逻辑上」两条链表接在了一起。
- [二分查找的妙用]()
  - [最长递增子序列](https://mp.weixin.qq.com/s?__biz=MzAxODQxMDM0Mw==&mid=2247484498&idx=1&sn=df58ef249c457dd50ea632f7c2e6e761&source=41#wechat_redirect)
    - 最长递增子序列（Longest Increasing Subsequence，简写 LIS）
    - 比较容易想到的是动态规划解法(动态规划的核心设计思想是数学归纳法)，时间复杂度 O(N^2)
      - dp[i] 表示以 nums[i] 这个数结尾的最长递增子序列的长度
      - dp 数组应该全部初始化为 1，因为子序列最少也要包含自己，所以长度最小为 1
        ```go
        for i := 0; i < len(nums); i++ {
            for j := 0; i < i; j++ {
                if nums[j] < nums[i] {
                    dp[i] = max(dp[i], dp[j] + 1)
                }   
            }
        }
        ```
    - [二分查找解法](https://mp.weixin.qq.com/s?__biz=MzAxODQxMDM0Mw==&mid=2247484507&idx=1&sn=36b8808fb8fac0e1906493347d3c96e6&source=41#wechat_redirect)
      - patience game 的纸牌游戏有关，甚至有一种排序方法就叫做 patience sorting
        - 只能把点数小的牌压到点数比它大的牌上。如果当前牌点数较大没有可以放置的堆，则新建一个堆，把这张牌放进去。如果当前牌有多个堆可供选择，则选择最左边的堆放置。
        - 按照上述规则执行，可以算出最长递增子序列，牌的堆数就是我们想求的最长递增子序列的长度
      - 我们只要把处理扑克牌的过程编程写出来即可。每次处理一张扑克牌不是要找一个合适的牌堆顶来放吗，牌堆顶的牌不是有序吗，这就能用到二分查找了：用二分查找来搜索当前牌应放置的位置
      ![img.png](algorithm_LIS.png)
  - [判定子序列](https://mp.weixin.qq.com/s?__biz=MzAxODQxMDM0Mw==&mid=2247484479&idx=1&sn=31a3fc4aebab315e01ea510e482b186a&source=41#wechat_redirect)
    - 如何判定字符串s是否是字符串t的子序列
    - 利用双指针i, j分别指向s, t，一边前进一边匹配子序列
      ```go
      bool isSubsequence(string s, string t) {
          int i = 0, j = 0;
          while (i < s.size() && j < t.size()) {
              if (s[i] == t[j]) i++;
              j++;
          }
          return i == s.size();
      }
      ```
    - 如果给你一系列字符串s1,s2,...和字符串t，你需要判定每个串s是否是t的子序列（可以假定s相对较短，t很长）
    - 二分思路主要是对t进行预处理，用一个字典index将每个字符出现的索引位置按顺序存储下来
      - 现在借助index中记录的信息，可以二分搜索index[c]中比 j 大的那个索引，寻找左侧边界的二分搜索就可以做到
    - ![img.png](algorithm_bs.png)







