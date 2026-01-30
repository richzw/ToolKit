- Tips
  - [Sam Altman 与开发者的一小时](https://x.com/dotey/article/2016153776713863599)
    - 给你一个去年这个时候一个人两周都做不到的任务，然后看你 10 分钟或 20 分钟做完
    - 变得高能动性（high agency）、善于产生想法、非常有韧性、非常能适应快速变化的世界
  - AI is the fastest way to forget how to code and how to think.
    - If AI is doing the thinking for you, your skills are getting worse... fast.
    - If AI is helping you think better, you’re still learning.
  - Use it as a reviewer, a rubber duck, a teacher. Not as an author:
    - Explain the decisions behind a piece of code
    - Explain the style and conventions
    - Ask whether your code is good enough, simple enough
    - Ask about edge cases you might have missed
    - Ask for trade-offs, not just answers
    - Ask for alternative implementations, then compare
    - Ask for tests, not production code: "What tests would give me confidence this works?"
    - Ask it to review, not generate
    - Ask it to explain errors you already hit
    - Ask it to challenge your assumptions
- [AI Era, 保留“认知摩擦”是你最后的护城河？](https://mp.weixin.qq.com/s/0-ygACSXvpG8XSSE4JR-Hw)
  - 背景判断：AI 时代的“知识通胀”
    - 获取答案/产出内容的成本极低，带来“顺滑”的工作流，但也引发思维能力的退化风险。
  - 核心问题：AI Brain（AI 大脑）= 变“钝”而非变“笨”
    - 过度依赖 AI 像“认知轮椅”，减少深度思考与处理混乱的训练。
  - 关键观点：智慧诞生于“认知摩擦”
    - 真正学习与创造发生在“第一遍思考（first-pass thinking）”的挣扎中；跳过挣扎等于跳过心理模型构建。
    - AI 消除了“摩擦”，但人类的智慧，恰恰诞生于“摩擦”之中
  - 解决方案：4 条“反直觉”的自救法则
    - 拥抱第一遍思考（先写/先画/先做初稿）
    - 人为保留认知摩擦（让问题“飞一会儿”，设置无 AI 窗口）
    - 做少但做深（用 AI 换时间，押注 1% 核心深度）
    - 回归物理世界（体验、交流、身体感与真实系统延迟等）
  - 结论：人与 AI 的边界决定你的未来
    - AI 可做“外骨骼”（放大你），也可成“拐杖”（替代你）。关键在于：你拒绝让 AI 做什么
- AI trade
  - AI-Trader🚀 - the fully autonomous trading benchmark where AI agents make real financial decisions without any human help.
    - https://github.com/HKUDS/AI-Trader
  - TradingAgent - Multi-Agents LLM Financial Trading Framework
    - https://github.com/richzw/TradingAgents
- Network access
  - 电脑端使用Gemini网页正常
    - iPhone手机使用Gemini app无法访问，百思不得其解
    - 研究发现把Google Analytics的统计域名加入代理手机就正常了
      - DOMAIN-SUFFIX,http://app-analytics-services.com,Proxy
- 安全指南
  - [Clawdbot 安全指南](https://x.com/katherineq1212/article/2016043617098363114)
    - 第一层：网络层防护
      - 本地部署：默认 localhost:8080，只在本地可访问
      - 远程访问：用 Tailscale 建立虚拟局域网，或者 SSH 隧道转发
      - 如果必须公网暴露：配置防火墙白名单（只允许你的 IP），加上强认证
    - 认证层防护
      - 配对模式（Pairing）：最推荐。你需要先在本地和机器人"握手"，双方交换密钥，之后只有配对成功的设备能控制它
      - Token 模式：次推荐。设置一个长随机字符串作为 token，每次请求都要带上。但 token 可能被泄露，需要定期更换
    - 工具权限层 - 最小权限原则 来配置
      - 沙箱模式：把工具限制在特定目录，比如 ~/clawdbot-sandbox
      - 工具分级：不同 agent 不同权限。比如"代码助手"只能访问 ~/code，"系统管理员"才能执行系统命令
      - 禁用高风险工具：如果不需要浏览器控制，直接禁用









