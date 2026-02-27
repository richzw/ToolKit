- AI Native 工作方式
  - 人负责：观点、方向、判断、决策。 
  - AI 负责：整理、结构化、补充细节、执行。 
  - 工具负责：连接和沉淀（Obsidian、Claude Code）
- [AI 比你更"熟练"时，如何划定控制权的边界](https://mp.weixin.qq.com/s/ZaVuD8zslb_knCvcR2WCIA)
  - 工欲善其事，必先利其器。然利器在手，当知进退，明取舍。善用者如庖丁解牛，游刃有余；不善用者如邯郸学步，反失其本
  - R-C-V 三维评估模型：给 AI 划红线
    - Risk（风险不对称性）、Context（语境封闭性） 和 Verification（验证成本）。
    - 风险不对称性：搞砸了会翻车吗？
      - 核心拷问： 如果 AI 搞砸了，后果是线性的还是指数级的？
      - “可逆性”是自动化的前提。 哪怕 AI 的准确率高达 99.99%，只要剩下的 0.01% 会导致”炸毁系统”，那么控制权必须在人手中
    - 语境封闭性：信息够用吗？
      - 核心拷问： 解决这个问题所需的信息，是否都在 Prompt 的窗口里？
      - AI 是局部最优的，人是全局最优的。 当任务需要”跳出当前窗口”去思考时，人绝不能离场。这不是”摆烂”，这是”破防”前的最后一道防线。
    - 验证成本 vs. 生成成本：检查比写还累吗？
      - 核心拷问： 检查 AI 的答案，是否比我自己做一遍更难？
      - 如果验证 AI 产出的成本（包含隐性风险成本）高于从头手写的成本，那么所谓的”自动化”就是伪命题。
  - 决策矩阵：你的行动指南
    - Copilot Zone（低风险 + 封闭语境）：AI 的”舒适区”
      典型场景： 正则表达式、SQL 查询优化、Boilerplate 代码、JSON 格式化、日志清洗脚本。
    - Gatekeeper Zone（高风险 + 封闭语境）：AI 的”内测区”
      典型场景： 核心支付逻辑、并发锁机制、自动驾驶的目标识别算法、医疗影像初筛。
    -  Inspiration Zone（低风险 + 开放语境）：AI 的”公测区”
       典型场景： 产品头脑风暴、UI 设计草图、文案初稿、测试用例的边缘场景发散。
  - AI 给出的是概率最高的答案，而我们需要的是逻辑上正确的答案
- [A Software Library with No Code](https://www.dbreunig.com/2026/01/08/a-software-library-with-no-code.html)
- [global agent guide](https://www.dzombak.com/blog/2025/08/getting-good-results-from-claude-code/)
 
- [ChatGPT 和 Claude 都有记忆功能，但两者实现原理截然不同](https://www.shloked.com/writing/claude-memory)
  - ChatGPT 的记忆模式是自动化、魔法般的个性记忆，不需要用户提醒，自动的悄悄记录用户的使用细节。
  - Claude的记忆模式是基于检索的，每次新开对话，都没有任何任何历史记忆，只有当你明确告诉 Claude 需要用到某条记忆内容，它才会从真实的历史记录中精准提取信息。
    - 一是基于关键词搜索历史对话，二是以时间线为索引检索近期聊天。
    - 两个检索工具：conversation_search（按关键词/主题跨全部历史查找、可多主题分别检索并汇总）与 recent_chats（按最近N次或特定时间窗口检索、可排序分页）。代价是会有可见延迟，但换来更强的可控性与可审计性
  - ChatGPT 走大众消费品路线：记忆默认开启、自动加载画像/摘要，追求“即刻个性化”和零等待，有利于规模化与留存；
  - Claude 面向开发者与专业工作流：用户理解并愿意显式触发工具（如搜索、长推理、记忆），更重视隐私与可预测性，记忆只是“按需调用”的工具之一
- [Claude Code放弃代码索引，使用grep技术](https://mp.weixin.qq.com/s/Fa15GoM3_2CUnjdHQ3I7Nw)
  - Anthropic 在《Claude Code: Agent in Your Terminal》中披露：评估向量索引、传统索引后，实时的 “agentic search”(glob+grep) 在吞吐、延迟与资源占用上全面优于其他方案，于是放弃长期持久索引
  - VS
    - Cursor：代码切块→向量嵌入→远程 Turbopuffer；优点语义召回，缺点需要上传、异步更新。
    -  JetBrains IDE：PSI 树 + stub 索引；优势精确重构、类型检查，但初次/增量索引耗时。
    -  Claude Code：本地 glob/grep；零配置、即时可用、结果确定、完全离线。
- Tips
  - Small context
    Keep conversations small+focused. After 60k tokens, start a new conversation.
  - CLAUDE.md files 
    - Use CLAUDE.md to tell Claude how you want it to interact with you
    - Use CLAUDE.md to tell Claude what kind of code you want it to produce
    - Use per-directory CLAUDE.md files to describe sub-components.
    - Keep per-directory CLAUDE.md files under 100 lines
    - Reminder to review your CLAUDE.md and keep it up to date
    - As you write CLAUDE.md, stay positive! Tell it what to do, not what not to do.
    - As you write CLAUDE.md, give it a decision-tree of what to do and when
  - Sub-agents 
    - Use sub-agents to delegate work
    - Keep your context small by using sub-agents
    - Use sub-agents for code-review
    - Use sub-agents just by asking! "Please use sub-agents to ..."
  - Planning 
    - Use Shift+Tab for planning mode before Claude starts editing code
    - Keep notes and plans in a .md file, and tell Claude about it
    - When you start a new conversation, tell Claude about the .md file where you're keeping plans+notes
    - Ask Claude to write its plans in a .md file
    - Use markdown files as a memory of a conversation (don't rely on auto-compacting)
    - When Claude does research, have it write down in a .md file
    - Keep a TODO list in a .md file, and have Claude check items off as it does them
  - Prompting 
    - Challenge yourself to not touch your editor, to have Claude do all editing!
    - Ask Claude to review your prompts for effectiveness
    - A prompting tip: have Claude ask you 2 important clarifying questions before it starts
    - Use sub-agents or /new when you want a fresh take, not biased by the conversation so far
  - [Claude Code Tips](https://agenticcoding.substack.com/p/32-claude-code-tips-from-basics-to)
    - https://github.com/ykdojo/claude-code-tips
- Codex
  - codex --dangerously-bypass-approvals-and-sandbox 全自动跑
  - codex resume用来选择历史记录sessions
  - codex目前没有plan模式，可以/approvals选择read-only进行讨论，或者直接在提示词里要求codex进行plan，写到一个文档中
  - [Unlocking the Codex harness](https://openai.com/index/unlocking-the-codex-harness/)
  - /review 命令很实用，在写完代码后让 codex review
  - Codex Prompt
    - Make a pixel art game where I can walk around and talk to other villagers, and catch wild bugs
    - Give me a work management platform that helps teams organize, track, and manage their projects and tasks. Give me the platform with a kanban board, not the landing page.
    - Given this image as inspiration. Build a simple html page joke-site.html here that includes all the assets/javascript and content to implement a showcase version of this webapp. Delightful animations and a responsive design would be great but don't make things too busy
  - Codex 当作“新入职的高级工程师” https://openai.com/index/shipping-sora-for-android-with-codex/
    - Codex 需要人类补齐的部分（团队的“管理动作”）
      - 不能凭空推断未告知的团队偏好：如架构风格、产品策略、真实用户行为、内部工程规范。
      - 无法“运行/体验”App：比如滚动手感、交互流程是否顺畅等，仍要人类在真机上验收。
      - 每个会话都要重新 onboarding（灌上下文）：需要清晰目标、约束、“我们团队怎么做事”。
      - 深层架构判断仍会偏向“先跑起来”：可能把逻辑塞进 UI、或引入不必要的 ViewModel 等，需要人类把关长期可维护性。
    - Codex 擅长的部分（团队把“产能”交出去的部分）
      - 快速读懂大代码库、多语言迁移理解。
      - 愿意写大量单元测试，覆盖面广，有助于防回归。
      - 善于根据 CI 失败日志迭代修复；并行开多会话推进多个模块（“可并行、可丢弃的执行”）。
      - 在设计讨论中提供新视角：例如为视频播放器做内存优化时，Codex 会去比较多个 SDK/方案，帮助降低最终内存占用
  - [Unrolling the Codex agent loop](https://openai.com/index/unrolling-the-codex-agent-loop/)
    - Codex 的本质是一个“编排器（harness）”：
      - 它不断把“用户输入 + 环境信息 + 历史对话 + 工具输出”组装成 input 发送给模型；模型要么给出最终答复，要么发出 tool call；Codex 执行工具后把结果再喂回模型，直到模型输出最终的 assistant message 结束本轮
    - Codex 通过 Responses API 的流式（SSE）事件接收模型输出（包括文本增量、输出项新增、完成事件等），并将与后续推理相关的输出项转为下一次请求的 input 项
    - 为了性能与隐私/合规，Codex 倾向保持请求无状态（不依赖 previous_response_id），并依靠 prompt caching 与 **compaction（对话压缩）**解决“请求体变大”和“上下文窗口耗尽”的问题
- Claude Code
  - [Ultimate Claude Starter Pack](https://x.com/aiedge_/article/2019153204869738897)
  - 三种“权限模式”（强烈建议先用 Plan Mode）
    - Normal：每次改文件/跑命令都会询问是否允许。 
    - Auto-accept：本会话内自动接受变更（原型阶段可用，谨慎）。 
    - Plan Mode：只读分析、先产出计划再执行，适合大型改造/安全走查。 切换：Shift+Tab；或直接用 claude --permission-mode plan 启动
  - 自定义命令
    - 把常用长提示做成一个 /命令，一键复用。
    - 选择作用域与目录
      - 项目级：.claude/commands/（随仓库共享）
      - 个人级：~/.claude/commands/（全项目可用） Claude 文档
    - 用 Markdown 创建命令
      - 文件名即命令名，fix.md → /fix。
  - 提供清晰的需求文档
    - 花时间写清楚你要它完成的功能点；
    - 明确涉及哪些接口、交互方式、边界条件；
    - 如果能画图（流程图、数据流）就更好了
  - 任务拆解细一点
    - 第一步：创建 API 接口
    - 第二步：添加字段验证
    - 第三步：编写测试用例
    - 第四步：写文档或 PR 描述
  - 防止过度思考
    - Claude Code 倾向 over-engineering 过度设计，所以Claude Code 内建了四档“思考深度。Claude 提供了从低到高四种指令：
      - think：快速简单的任务，适合查询或小型修改。 
      - think hard：中等複杂度，适合多步骤操作或中型重构。
      - think harder：跨模组或非同步架构的调整，适合深入思考。 
      - ultrathink：高複杂度场景，全域架构或演算法最佳化。
    - Anthropic 官方建议：
      - 先用中低档（think hard）测试，再根据实际需求调高档位。
      - 每次调整“档位”，清楚说明 Claude 要在哪些维度深入思考，
  - Tips
    - 如果Claude写的代码总是无法通过，可以在 CLAUDE.md 加了“请务必测试”；
    - /init 生出的 CLAUDE.md 太多废话，浪费 token，可以简要讲一下，我现在的不到 10 行。
    - 需要第二次纠正 AI 的就放进 CLAUDE.md；
- Claude Code 连续工作 8 小时的问题
  - 本质上就是一个 Manager 监控 Worker 干活。
    - Worker 要有 TODO List，并且 Agents/Claude Code MD 要有引导，这样每次固定提示词（continue）能继续任务
    - Worker 要开子进程避免上下文爆掉
    - Manager 去管理 Worker 干活要开子 Agent，避免 Manager 的上下文爆掉
  - Claude Code 有个特别的工具叫 Task tool，本质就是一个子 Agent，它可以有独立的上下文，所以哪怕它用了很多token，但也不会占用多少主Agent的上下文空间
  - claude code 支持 hook，理论上来说可以借助 hook 来自动化
    - claude code完成一个任务后，会写到一个完成文件，然后脚本里有监控流程，出现这个文件n秒后自动close claude，然后由脚本进行下一次task
- Claude Code 发布 v2.0 了，升级了 UI 界面，推出了全新的VS Code扩展插件。
  - 还有一个实用的新功能：检查点（checkpoints）。通过它，你可以快速撤销Claude刚刚做出的修改，只需轻松按下Esc+Esc快捷键，或者输入指令/rewind即可实现。
  - Sonnet 4.5模型，发现它有个非常明显的进步，那就是在压缩对话上下文（compacting conversations）方面，比其他用过的模型都要强不少。
  - Anthropic甚至专门建议用户可以让Sonnet 4.5以维护上下文文件的形式来记录状态，而不仅仅是简单的“上下文总结”（context summarization）。
  - Claude Code 读取 PDF、图片，都是把文件 base64 编码成字符串，一起传到 API，服务端解析，就这么简单粗暴
    - Base64 不是以字符的形式直接传入 prompt，而是在服务器端再还原成图片或 PDF，PDF 再转成图片
    - https://www.datastudios.org/post/how-claude-reads-pdf-files-in-2025-workflow-capabilities-and-limitations
  - Claude Code 默认是开启自动压缩上下文的，但是可以禁用： 输入“/config”，在菜单中将 Auto-compact 设置为 false
- Claude Code 插件系统
  - 有一个公开的 GitHub Repo，按照它的规范提供一个 .claude-plugin/marketplace.json 文件就好，官方也提供了官方插件市场，只要在 CC 中输入
     `> /plugin marketplace add anthropics/claude-code`
  - 插件可以简单地打包和分享以下自定义内容：
    - 斜杠命令（slash commands）：为常用的操作创建自定义快捷方式；
    - 专属智能体（subagents）：安装专为特定任务打造的智能体，协助你完成专业的开发工作；
    - MCP服务器：通过模型上下文协议（Model Context Protocol，简称MCP）连接外部工具和数据源；
    - 钩子函数（hooks）：在Claude Code的工作流关键节点，自定义它的行为。
  - 插件的典型应用场景
    - 强制团队规范：技术负责人可以通过插件设定统一的代码审查、测试流程等工作流标准；
    - 支持用户开发：开源项目维护者可以提供一些斜杠命令，帮助其他开发者更便捷地使用他们的代码库；
    - 分享实用工作流：开发者可以把自己精心设计的调试环境、部署流水线、测试框架等生产力工具打包分享；
    - 连接各种工具：团队可以用插件通过MCP快速、安全地连接内部数据和工具；
    - 打包个性化组合：框架作者或技术负责人能将针对特定场景的多种自定义设置进行打包，提供给团队成员使用
- [“榨干” Claude Code 和 OpenAI Codex 们的性能](https://simonwillison.net/2025/Sep/30/designing-agentic-loops/)
  - 在 AGENTS. md 中记录命令：“shot-scraper http://www.example. com/ -w 800 -o example.jpg”，让智能体轻松捕获网页截图
  - “Agentic loop” 概念
    - 作者把 LLM 代理定义为“在循环中调用工具以达成目标的系统”。
    - 这类代理本质上是一种“暴力搜索”——只要把问题拆成一个清晰目标＋可迭代的工具集，代理就能不停尝试直到找到可行解
  - YOLO mode：让代理全自动执行命令
    - Claude Code、Codex CLI 等默认每次执行 shell 命令前都要求人工确认，效率低。开启 “YOLO mode” 可自动批准一切命令，但极其危险
    - 未加监管时主要有三类风险：破坏性 shell 命令、数据外泄（源码/环境变量）、把本机当跳板发起攻击。
    - 作者给出三种缓解策略：
      - 在受限沙箱（Docker 或苹果新 container 工具）里运行代理；
      - 直接“用别人的电脑”——例如 GitHub Codespaces；
      - 承担风险、靠人工盯梢
  - 与其用复杂的 MCP（Model Context Protocol），作者更倾向直接暴露 shell 命令，因为 LLM 对它们最熟悉
- [Jina MCP](https://mp.weixin.qq.com/s/pZJr7-rfalOhZ1XRIOjGzw)
  - CC
    - `claude mcp add --transport sse jina https://mcp.jina.ai/sse --header "Authorization : Bearer ${JINA_API_KEY}"`
  - 在 OpenAI Codex 中配置： 编辑 ~/.codex/config.toml 文件，添加以下配置
- [Anthropic Skills vs. OpenAI AgentKit]
  - Skills 是为 Claude 定制的技能包，用户通过对话定义，Claude 会在需要时自动调用，无需手动编辑。
  - AgentKit 期望通过开发者构建和管理多步骤工作流，人工编排逻辑，成为企业 AI “自动化”的操作系统。
  - [Claude Skills](https://www.anthropic.com/engineering/equipping-agents-for-the-real-world-with-agent-skills)
    - 一个Skill就是一个文件夹，包含指令、脚本与资源。具体来说，每个Skill包含三样东西：
      - 指令(Instructions)告诉Claude该做什么、脚本(Scripts)执行具体任务、资源(Resources)提供模板和辅助内容。因为自然语言也是代码，指令和脚本其实是分不清的，都属于程序
    - Claude只会在Skill与当前任务相关时才会调用，并且采用渐进式披露：先加载元数据(约100词)，再加载http://SKILL.md主体(也比较小)，最后才是具体的资源文件。
    - Skills 的核心概念
      - 每个 Skill 是一个文件夹，至少包含一个 SKILL.md。文件首部必须是 YAML front-matter，含 name 与 description 两个字段。启动时，代理只把所有已安装技能的这两段元数据注入 system prompt，用于后续匹配任务场景。
      - 若代理判定某 Skill 相关，它会再读取完整 SKILL.md；若仍需更多细节，则按引用逐步打开同目录下的其他 Markdown、脚本或资源文件，实现「逐层披露（progressive disclosure）」的上下文加载策略，理论上可容纳无限量资料而不挤占上下文窗口
    - Skill 开发与评估最佳实践
      - 先做评估：用代表性任务找出代理能力缺口，再增量写 Skill。
      - 结构化扩展：当 SKILL.md 过长就拆分文件；互斥上下文放不同路径减少 token；把脚本既当工具也当文档。
      - 代理视角调试：观察 Claude 何时触发技能、是否走偏，并反复迭代 name/description。
      - 与 Claude 协同：让它把成功步骤、常见错误写回 Skill 以自我改进
- 在 Claude Code 中配置 GLM 4.6 的方法
  ```
  {
  "env": {
  "ANTHROPIC_AUTH_TOKEN": "your_zai_api_key",
  "ANTHROPIC_BASE_URL": "https://api.z.ai/api/anthropic",
  "API_TIMEOUT_MS": "3000000",
  "ANTHROPIC_DEFAULT_HAIKU_MODEL": "glm-4.5-air",
  "ANTHROPIC_DEFAULT_SONNET_MODEL": "glm-4.6",
  "ANTHROPIC_DEFAULT_OPUS_MODEL": "glm-4.6"
  }
  }
  ```
- [Claude Agent Skills: A First Principles Deep Dive](https://leehanchung.github.io/blogs/2025/10/26/claude-skills-deep-dive/)
  - Skills 不是可执行代码：不跑 Python/JS、不起 HTTP server；本质是“注入式指令”
  - 技能（skills）= Prompt 模板 + 对话上下文注入 + 执行环境修改。它们本质上是一段 Markdown（SKILL.md）而非可执行代码，通过“Skill”元工具在运行时注入到 Claude 的上下文中
  - 什么是Claude Skills？ 简单来说，它是一种Agent能力扩展机制，通过将指令、脚本和资源组织成文件夹（即一个Skill目录），让Agent能够动态发现和加载这些专业知识，从而将通用Agent转化为特定任务的专家。
  - 核心思想就是——信息分层设计哲学 - 细节分层 (LOD) 与 按需加载
    - LOD (Level of Detail)，即“细节层次”，是3D游戏渲染中的一项核心技术。它的基本思想是：一个物体离你越远，展示的细节就应该越少
    - LOD负责管理信息的精度，按需加载负责管理信息的时机。
  - Claude Agent SKILL 是一个结构化的、可重用的包，存储在你项目的 .claude/skills/ 文件夹中。它结合了以下内容：
    - 定义 Agent 角色和分步流程的精确指令
    - 参考文件（风格指南、示例、品牌语气）
    - 用于可靠、确定性操作的可执行脚本（Python, Node.js, Bash）
  - 一些流行的 SKILLs 示例：前端审查; PDF 提取; Web 抓取
  - [Docker Sandboxes 运行具有完全自主权的 Claude](https://mp.weixin.qq.com/s/H2Xhh2SJpKPp03tllfX-IQ)
- [Claude Code 核心](https://mp.weixin.qq.com/s/7g5DugzATAIX1by4yAYtTg)
  - Agent（主战力） + MCP（能力扩展） + Slash（效率） + Hook（可控 / 自动化）
- [How I Use Every Claude Code Feature](https://blog.sshh.io/p/how-i-use-every-claude-code-feature)
  - CLAUDE.md
    - Start with Guardrails, Not a Manual. CLAUDE.md 应该很短，只在 Claude 容易出错的地方加说明
    - 如果你有详细文档，可能想在 CLAUDE.md 里用 @ 引用。但这会把整个文件塞进上下文，非常臃肿
      - 正确做法是推销这个文件，告诉它为什么和何时该读。例如：遇到复杂用法或 FooBarError 错误时，参考 path/to/docs.md 的高级故障排除。
    - 不要写纯否定约束
    - Use CLAUDE.md as a Forcing Function 
    - Treat your CLAUDE.md as a high-level, curated set of guardrails and pointers. 
    - Use it to guide where you need to invest in more AI (and human) friendly tools, rather than trying to make it a comprehensive manual.
  - Compact, Context, & Clear
    - /compact (Avoid): I avoid this as much as possible
    - /clear + /catchup (Simple Restart): My default reboot. I /clear the state, then run a custom /catchup command to make Claude read all changed files in my git branch.
    - “Document & Clear” (Complex Restart): For large tasks. I have Claude dump its plan and progress into a .md, /clear the state, then start a new session by telling it to read the .md and continue.
  - Custom Slash Commands
    - Use slash commands as simple, personal shortcuts, not as a replacement for building a more intuitive CLAUDE.md and better-tooled agent.
  - Custom subagents are a brittle solution. Give your main agent the context (in CLAUDE.md) and let it use its own Task/Explore(...) feature to manage delegation.
  - Use claude --resume and claude --continue to restart sessions and uncover buried historical context
- [CLAUDE.md Masterclass: From Start to Pro-Level User with Hooks & Subagents](https://newsletter.claudecodemasterclass.com/p/claudemd-masterclass-from-start-to)
  - CLAUDE.md 的定位：不是 README，而是“配置/规则层
    - 当成“默认操作系统设置”，放跨任务稳定的内容（命令、约束、边界、流程），而不是一次性任务细节。官方 Best Practices 也强调“保持短、且只放广泛适用的规则”
  - Claude Code 会把 CLAUDE.md 包在一段 <system-reminder> 里，提醒模型“这些上下文可能无关；除非高度相关否则忽略”，所以写得越杂越容易被忽略
  - CLAUDE.md 的层级与加载规则（全局/项目/目录按需
    - 在多个位置放 CLAUDE.md；根部给全局规则，子目录放局部规则；Claude 访问子目录文件时才按需加载对应的嵌套 CLAUDE.md
  - “好用的 CLAUDE.md”的结构：回答 WHAT / WHY / HOW
    - 用清晰分区（项目背景、技术栈、关键目录、常用命令、标准、工作流、注意事项），让 Claude 快速上手且不误改
  - /init 生成 starter 配置，并明确提到用 # 来积累你反复重复的指令
  - 别让 Claude 反复“人工检查格式”；应把 lint/test 等放到 hook 或 pre-commit 自动执行，CLAUDE.md 只描述规则存在即可
  - 把高频通用规则放 CLAUDE.md；低频/任务型内容挪到独立文档，需要时再让 Claude 去读，避免主文件膨胀
- Claude Code 驱动任务执行 + Codex 深度代码分析与生成 https://github.com/Pluviobyte/Claude-Codex
  - 多agent协作 https://github.com/jeanchristophe13v/codex-mcp-async
    - 一个Claude Code作为orchestrator 同时调用gemini-cli看文档，调用2个codex做规划，最后再调用几个cc去执行
- 真实IP正在“裸奔”！难怪AI总提示“异常流量”
  - 看看是否存在WebRTC泄露和DNS泄露 https://ipcheck.ing/#/
  - 代理工具（Clash）一般有个TUN 模式
- 如何在 Codex CLI 里面用 SKILLs
  - 在你的项目目录下创建一个 “.claude/skills”目录，如果你不想提交到 git 就把 .claude 加到 .gitignore
    - 注：也可以是任意其他目录，放在“.claude/skills”目录下有个好处就是 claude code 默认能使用，不需要额外配置。
  - 把你要用到 skill 复制到“.claude/skills”目录下（可以去 http://github.com/anthropics/skills 这里找现成的）
  - 如果你需要用到哪个 skill，只需要手动 @ 一下相应的 skill 文件即可，比如：> 请使用 @.claude/skills/artifacts-builder/SKILL.md ，创建一个 whiteboard 项目
  - 在 CC 里优雅管理 Skills 的正确姿势是：一律“插件化 + marketplace化”，不要散落的文件。
    - Anthropic 官方 anthropics/skills 仓库已经给了非常明确的路线：通过 /plugin 把整个仓库当成一个 Plugin Marketplace 来挂载，再按需安装 Skill 套件。
    - /plugin marketplace add anthropics/skills 命令含义：
      - 告诉 Claude Code：anthropics/skills 仓库里有 .claude-plugin 配置，可以作为一个插件源。
      - 之后 /plugin 打开的 UI 里，你会看到一个叫 anthropic-agent-skills 的插件“市场”。
    - 1. 对于官方 Skills
      ```
      # 先从官方插件市场安装 Skills 插件
      /plugin marketplace add anthropics/skills
      
      # 从这个市场里按需安装插件化的 Skill 套件
      /plugin install example-skills@anthropic-agent-skills
      
      # 若有确定的文档处理需求，可以直接安装：
      /plugin install document-skills@anthropic-agent-skills
      ```
    - 2. 对于自定义 Skills
       - 在你自己的 GitHub org 建一个 org-claude-skills 仓库：
       - 初始化 .claude-plugin，定义 org-document-skills/org-dev-workflow 等插件。 把你最常用的两三类流程包装成 Skills（可以直接借鉴 skill-creator 模板）
    - 如何使用？
       - 安装完之后，Claude Code 会自动把插件里 skills/ 目录下的各个 Skill 注册进“可用 Skills”列表。
       - 你只需要“自然语言调用”即可，比如： “使用 PDF skill 从这个文档中提取表格：path/to/some-file.pdf” 不需要你手动 /skill xxx，也不需要写什么配置
  - Codex CLI 里的 Skills：本地 ~/.codex/skills + --enable skills
  - https://simonwillison.net/2025/Dec/12/openai-skills/
  - [awesome-claude-skills](https://github.com/ComposioHQ/awesome-claude-skills)
- [Antigravity Grounded! Security Vulnerabilities in Google's Latest IDE](https://embracethered.com/blog/posts/2025/security-keeps-google-antigravity-grounded/)
  - 1. **谨慎启用 MCP 服务器与工具**
    - 默认禁用高风险工具（尤其是具有写、执行、外联能力的）。
    - 根据实际业务需求最小化工具权限范围。
  - 2. **尽可能增加 Human in the Loop**
    - 在 Antigravity 中关闭或减少自动执行（Auto-Execute）：
      - 关闭终端命令的自动执行；
      - 对敏感命令、外联操作、文件读写等启用手动审批。
    - 使用“终端命令白名单”功能，只允许 AI 执行预先审核过的一小部分命令。
  - 3. **针对隐藏 Unicode 指令进行检测**
    - 在 CI/CD 中增加对 Unicode Tag Characters 等不可见字符的扫描，自动阻断或告警。
    - 不要仅依赖人工代码审查来应对提示注入，**视觉上看不到的东西需要自动化工具**来发现。
- [从「写代码」到「验代码 ](https://yousali.com/posts/20251124-how-to-coding-with-ai/)
- [Writing a good CLAUDE.md](https://www.humanlayer.dev/blog/writing-a-good-claude-md)
  - CLAUDE.md is for onboarding Claude into your codebase. It should define your project's WHY, WHAT, and HOW.
  - Less (instructions) is more. While you shouldn't omit necessary instructions, you should include as few instructions as reasonably possible in the file.
  - Keep the contents of your CLAUDE.md concise and universally applicable.
  - Use Progressive Disclosure - don't tell Claude all the information you could possibly want it to know. Rather, tell it how to find important information so that it can find and use it, but only when it needs to to avoid bloating your context window or instruction count.
  - Claude is not a linter. Use linters and code formatters, and use other features like Hooks and Slash Commands as necessary.
  - CLAUDE.md is the highest leverage point of the harness, so avoid auto-generating it. You should carefully craft its contents for best results.
- AI interview cheat
  - Whisper偷听面试官，Tesseract偷拍屏幕题，Claude两秒写完代码加口语解释，骨传导耳机低声报答案，或者干脆用Cluely的透明浮窗，连共享屏幕都看不到。
  - Anthropic Interviewer https://www.anthropic.com/research/anthropic-interviewer
    - https://huggingface.co/datasets/Anthropic/AnthropicInterviewer
- 后端代码，可以尝试用伪代码去提示词，试试TDD，先写测试代码，再去实现
  - 先和 code agent 在 plan 模式下把业务用 plantuml/mermaid 用 uml 或者 ddd 的语言沟通明白。
- [CC](https://blog.cosine.ren/post/my-claude-code-record-2)
  - 新功能从 0 到 1
    - 使用 Claude Code 的 Plan Mode，让模型只输出“变更计划（哪些文件、改动点、预期 diff）”，先不写代码。
    - ClaudeCode 的 Plan Mode 会生成计划，并询问一些你可能没讲清楚的地方
    - 补充并 Review 计划完毕后让他按计划生成代码，落到本地跑编译与最小样例。
    - 再次要求模型自检：列出潜在失败场景、边界条件和建议测试用例。
  - BugFix
    - 喂给他报错日志/最小复现工程。
    - 让模型列出“定位假说清单、验证步骤、最小改动方案”。
    - 实现、自检
  - 重构/迁移
    - 同样 Plan Mode 描述重构需求，让其生成文档计划等
    - 让模型先写 codemod，只在小部分上试跑。
    - 观察 diff，定义切分点和随时可回滚的边界。
    - 分批推进，并进行回归测试。
  - 模型与工具的选型与切换
    - 重要的架构设计/大重构：用强模型（质量优先）。
    - 批量生成测试/样例：用便宜模型（成本优先）。
    - 读 log / 写小脚本/摘要：用更快模型（速度优先）。
    - 小模型可以用 Plan Mode 先生产 “变更计划 + 验收用例” 的 PRD，spec 驱动开发，而大模型负责实现。
- [Martin Fowler警告：大模型正将编程拖入“非确定性深渊”](https://mp.weixin.qq.com/s/7PSQ3CsPuu-BPd3ck36-sw)
  - AI 的核心变化：不是抽象层级，而是“非确定性”
    - 传统软件大多是确定性的：同样输入 → 同样输出。
    - 大模型推理结果是非确定性的：同一输入，多次调用可能返回不同输出（取决于采样策略、温度、系统实现等
  - 把大模型视作：“高产但不可靠的初级开发者”：
    - 任务拆得足够小（小函数、小类、小变更）；
    - 模型的每次输出都当作 PR（pull request）；
    - 由人类做严格的代码审查和测试。
- [Vibe Engineering](https://forum.openai.com/public/videos/event-replay-vibe-engineering-with-openais-codex-2025-12-03)
  - AI 不只是“帮你写代码”，而是参与到工程的全链路里，帮你更快做出能上线、能维护、能扩展的生产级软件。前提是：每一行要发到生产环境的代码，仍然必须由人来负责。
  - 方法论：
    - 让 AI 写可读的产物，而不只是可运行的代码。
    - 并行化。Codex 提供的 “Best of N” 思路很像把一个问题同时交给四个候选工程师：让它们走不同路线，产出不同方案，然后人来选更符合目标、也更符合品味的那个。
- [Making Coding Agents Safe: Using LlamaIndex to Secure Filesystem Access](https://www.llamaindex.ai/blog/making-coding-agents-safe-using-llamaindex)
  - 编码代理（coding agents）如何安全地访问文件系统”以及“如何让代理更好理解非结构化文档（PDF/Office/Google Docs 等
    - 方案：用 AgentFS 把真实文件系统虚拟化隔离，再用 LlamaParse 把非结构化文件转成高质量文本，最后用 LlamaIndex Workflows 把整个流程“上安全带（harness）”并支持可恢复的人类介入
  - 用“虚拟文件系统”替代真实文件系统访问;  AgentFS 的文件操作封装成“受控工具”，并强制代理只能用这些工具
  - https://github.com/run-llama/agentfs-claude
- CC
  - AgentsTeams
    - settings.json
      ```
        "env": {
         "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1"
        }
      ```
      只需用自然语言告诉Claude创建智能体团队并描述任务结构即可
  - [Config](https://x.com/bcherny/status/2021701636075458648)
    - https://code.claude.com/docs/en/settings
    - Pre-approve common permissions /permissions
    - Enable Sandbox /sandbox
    - /keybindings
    -  Use output styles: explanatory or learning
  - [Deep Dive: How Claude Code's /insights Command Works](https://www.zolkos.com/2026/02/04/deep-dive-how-claude-codes-insights-command-works.html)
  - [The Shorthand Guide to Everything Claude Code](https://x.com/affaanmustafa/article/2012378465664745795)
    - https://github.com/affaan-m/everything-claude-code/tree/main
  - [Tips](https://x.com/bcherny/status/2017742759218794768)
    - Do more in parallel
      - Spin up 3–5 git worktrees at once, each running its own Claude session in parallel.
      - name their worktrees and set up shell aliases (za, zb, zc) so they can hop between them in one keystroke
      - Git worktree 是什么？ 让你在同一个仓库里同时打开多个分支的工作目录，不用来回切换。命令大概是 git worktree add ../feature-a feature-a
    - Start every complex task in plan mode. Pour your energy into the plan so Claude can 1-shot the implementation.
      - 让一个 Claude 写计划，另开一个 Claude 以“高级工程师”的身份审核这个计划。让 AI 审 AI
      - 事情一旦跑偏，立刻回到 Plan Mode 重新规划。不要硬推，不要让 Claude 在错误的方向上越走越远。
    - Invest in your http://CLAUDE.md. 
      - After every correction, end with: "Update your http://CLAUDE.md so you don't make that mistake again." Claude is eerily good at writing rules for itself.
      -  prompt 可以是：“Update your CLAUDE.md so you don't make that mistake again.”
      - 更系统：他为每个项目/任务维护一个 notes 目录，每次 PR 后更新。然后在 CLAUDE.md 里指向这些 notes，相当于给 Claude 建了一个持续更新的知识库。
      - CLAUDE MD 不建议放太多内容，只会适得其反，只放最重要的 AI 没训练过的内容，更多的内容作为文件链接按需读取。
      - Claude Code 官方项目中 CLAUDE md 文件也就大约 2.5k tokens：
        - 常用 Bash 指令：让 AI 知道如何像开发者一样操作命令行。
        - 代码风格规范 (Code Style Conventions)：确保 AI 写的代码符合团队编码标准。
        - UI 与内容设计准则：指导 AI 如何设计界面和编写文案。
        - 核心技术实现流程：教 AI 如何处理状态管理 (State Management)、日志记录 (Logging)、错误处理 (Error Handling)、功能门控 (Gating，即控制特定功能的开启与关闭) 以及调试 (Debugging)。
        - 代码合并请求 (Pull Request) 模板：规范提交代码时的文档格式。
    - Create your own skills and commit them to git. Reuse across every project. Tips from the team:
      - If you do something more than once a day, turn it into a skill or command
      - Build a /techdebt slash command and run it at the end of every session to find and kill duplicated code
      - Set up a slash command that syncs 7 days of Slack, GDrive, Asana, and GitHub into one context dump
      - Build analytics-engineer-style agents that write dbt models, review code, and test changes in dev
      - [Extend Claude with skills](https://code.claude.com/docs/en/skills#extend-claude-with-skills)
    - Claude fixes most bugs by itself. Here's how we do it:
      - Enable the Slack MCP, then paste a Slack bug thread into Claude and just say "fix." Zero context switching required.
      - /techdebt：在每次 session 结束时运行，让 Claude 检查并清理重复代码
    -  Level up your prompting
      - a. Challenge Claude. Say "Grill me on these changes and don't make a PR until I pass your test." Make Claude be your reviewer.  
        - Or, say "Prove to me this works" and have Claude diff behavior between main and your feature branch
      - b. After a mediocre fix, say: "Knowing everything you know now, scrap this and implement the elegant solution"
      - c. Write detailed specs and reduce ambiguity before handing work off. The more specific you are, the better the output
    - Use subagents
      - Append "use subagents" to any request where you want Claude to throw more compute at the problem
        -  在任何请求后面加上“use subagents”。Claude 会自动把任务拆分给多个 Subagents 并行处理，相当于让它“开更多的线程”来解决问题。
      - Offload individual tasks to subagents to keep your main agent's context window clean and focused
      - Route permission requests to Opus 4.5 via a hook — let it scan for attacks and auto-approve the safe ones
    - Use Claude for data & analytics
      - Ask Claude Code to use the "bq" CLI to pull and analyze metrics on the fly. 
      - We have a BigQuery skill checked into the codebase, and everyone on the team uses it for anlytics queries directly in Claude Code
    - Learning with Claude - few tips from the team to use Claude Code for learning:
      - Enable the "Explanatory" or "Learning" output style in /config to have Claude explain the *why* behind its changes
      - Have Claude generate a visual HTML presentation explaining unfamiliar code. It makes surprisingly good slides!
      - Ask Claude to draw ASCII diagrams of new protocols and codebases to help you understand them
      - Build a spaced-repetition learning skill: you explain your understanding, Claude asks follow-ups to fill gaps, stores the result
    - 记录解决方案
      - AI很擅长写文档。项目下放个docs目录，在某个子目录专门保存这些方案
      - 在当前会话（可以 claude --resume 找到），让它帮你记录到一个markdown文档
  - [How to Use Codex from Claude Code](https://gist.github.com/antirez/2e07727fb37e7301247e568b6634beff)
  - ClaudeCode开发http://laper.ai 的最核心技巧：
    - 根目录主md强调任何功能、架构、写法更新必须在工作结束后更新相关目录的子文档。
    - 每个，我是说每个，每个文件夹中都有一个极简的架构说明（3行以内），下面写下每个文件的名字、地位、功能。文件开头声明：一旦我所属的文件夹有所变化，请更新我。
    - 每个文件的开头，写下三行极简注释，文件input（依赖外部的什么）、文件ouput（对外提供什么）、文件pos（在系统局部的地位是什么）。并写下，一旦我被更新，务必更新我的开头注释，以及所属的文件夹的md。
  - 在文件头部写摘要这个策略很赞，code agent 读文件的策略
    - 小文件一次性载入;
    - 大文件先读头部，然后分块顺序载入，或者基于 grep/抽象语法树，查找载入；
      - 无论哪种都会第一时间把头部载入，这样就能第一时间获取重要信息； claude skill的 md 也是这种策略，《金字塔原理》
  - [本地 Claude Code 会话 → HTML 转录页](https://simonwillison.net/2025/Dec/25/claude-code-transcripts/) 
  - Claude Code 中开启 --chrome 后模型对着屏幕一顿疯狂截图
  - [Boris Cherny 公开了他的 CC 使用方法](https://x.com/bcherny/status/2007179832300581177)
    - 复利思维和验证
      - 复利思维体现在 CLAUDE. md 不是一次性写完的文档，而是团队在日常工作中持续积累的知识库。每次代码审查、每次发现问题，都在让这个文件变得更好。
    - 多实例并行：同时运行 15-20 个 Claude
      - 在终端和网页之间来回切换，用 & 符号把本地会话转到网页，或者用 --teleport 在两边传送。这种并行工作方式让他能同时推进多个任务。
      - 在 http://claude.ai/code 网页版上跑 5 到 10 个任务。终端和网页可以互相“交接”：用&符号把本地会话转到网页，或者用--teleport 在两边来回切换。
    - 模型选择：全程 Opus 4.5 with thinking
    - 团队知识库：共享的 CLAUDE .md 文件
    - 码审查集成：@.claude 标签触发改进
    - Plan 模式：先规划再执行 大部分会话都从 Plan 模式开始（按两次 shift+tab 进入）
    - 把每天重复做很多次的"内部循环"工作流都做成了 slash commands。这些命令保存在 .claude/commands/ 目录下，提交到 git
    - 常用几个 subagents：code-simplifier 在 Claude 完成工作后简化代码，verify-app 包含了端到端测试 Claude Code 的详细指令。
    - 用 PostToolUse hook 自动格式化 Claude 生成的代码
    - 不用 --dangerously-skip-permissions。他用 /permissions 预先允许那些在他环境里确定安全的常见 bash 命令，避免不必要的权限提示。 这些配置大部分都保存在 .claude/settings.json 
    - 长时间任务：后台代理和 Stop Hook
      - 让 Claude 在完成时用后台代理验证工作
      - 用 agent Stop hook 更确定性地做验证
      - 用 ralph-wiggum 插件
      - 在沙箱环境里用--permission-mode=dontAsk 或--dangerously-skip-permissions，让 Claude 不被权限确认打断，自己跑到底
  - [Ralph Wiggum 插件：让 Claude Code “通宵干活”](https://github.com/anthropics/claude-plugins-official/tree/main/plugins/ralph-wiggum)
    - /ralph-loop "你的任务描述" --completion-promise "DONE" --max-iterations 50
  - 跳过启动时候要求登录的限制没办法使用 Claude Code
    - 在 ~/.claude.json 这个配置文件里面加上 "hasCompletedOnboarding": true
  - [The complete claude code tutorial](https://x.com/eyad_khrais/article/2010076957938188661)
    - Think First
      - Thinking first, then typing, produces dramatically better results than typing first and hoping Claude figures it out.
      - click shift + tab twice, and you’re in plan mode
    - CLAUDE.md
      - Keep it short. Claude can only reliably follow around 150 to 200 instructions at a time, and Claude Code's system prompt already uses about 50 of those
      - Tell it why, not just what.
        - When you give it the reason behind an instruction, Claude implements it better than if you just tell it what to do.
        - "Use TypeScript strict mode because we've had production bugs from implicit any types" is better.
    - Context degrades at 30%, not 100%. Use external memory, scope conversations, and don't be afraid to clear and restart with the copy-paste reset trick.
    - 模型分工
      - Opus 4.5: 适合复杂推理、架构决策和制定计划。虽然慢且贵，但“脑子”最清楚。
      - Sonnet 4.5: 适合具体执行、写模板代码和重构。速度快，性价比高。
      - 建议方案： Opus 做计划，Sonnet 做执行。
  - [The claude code tutorial level 2](https://x.com/eyad_khrais/article/2010810802023141688)
    - Skills: Teaching Claude Your Specific Workflows
      - Create a folder with a SKILL.md file: ~/.claude/skills/your-skill-name/SKILL.md
      - Or for project-specific skills that you want to share with your team: .claude/skills/your-skill-name/SKILL.md
    - Subagents: Parallel Processing With Isolated Context
      - context degradation happens around 45% of your context window.
      - Claude Code includes three built-in subagents
        - Explore: A fast, read-only agent for searching and analyzing codebases. Claude delegates here when it needs to understand your code without making changes.
        - Plan: A research agent used during plan mode to gather context before presenting a plan.
        - General-purpose: A capable agent for complex, multi-step tasks requiring both exploration and action
      - add a markdown file to ~/.claude/agents/ (user-level, available in all projects) or .claude/agents/ (project-level, shared with your team).
    - MCP connectors eliminate context switching
      - Command: claude mcp add --transport http <name> <url>
    - Skills encode patterns, subagents handle subtasks, MCP connects services.
  - [How I Use Claude Code](https://x.com/ashpreetbedi/article/2011220028453241218)
  - [Advent of Claude: 31 Days of Claude Code](https://adocomplete.com/advent-of-claude-2025/)
  - Tool Search now in Claude Code
    - https://platform.claude.com/docs/en/agents-and-tools/tool-use/tool-search-tool
    - Claude Code detects when your MCP tool descriptions would use more than 10% of context
    - When triggered, tools are loaded via search instead of preloaded
  - Claude Code 避免上下文用满的经验
    - 从原理上来说，你可以把 Claude Code 的上下文窗口想成一块“内存条”：快、顺手、但容量有限。
      - 聊天当内存，文件当硬盘，Git 当时光机
    - 关掉自动压缩，我喜欢自己控制上下文。自动压缩有时候会把你最在意的细节当噪音裁掉
      - 中间结果存文件。你从我的 Skills 设计中也可以看得出来，我很喜欢保留中间文件，好处是新会话可以用，写作有 outline md、draft md
    - 如果很长时间一个会话内任务没完成，我不压缩会话，让 Claude 总结一下：目标、进度、卡点、下一步、关键约束。我看一眼，手动改几个关键地方，新开会话继续。
    - 用 Git。无论写代码还是写作，一个会话结束马上 commit。提示词也可以存进 prompts 目录，下次直接复用
    - 卡住了多半是思路不对。这时候别在原会话硬扛，借助 Git 回滚到上一个靠谱快照，找到原始提示词，以及一些中间产生的关键文件，从头来
  - [50 Claude Code Tips ](https://x.com/aiedge_/article/2014740607248564332)
    - Claude Skills Repo - A library of 80,000+ Claude Skills https:// skillsmp. com/
    - Claude Skills Library - A cool website with plug-and-play Skills and more https:// mcpservers. org/claude-skills
  - Claude Code 官方项目中 CLAUDE md 文件也就大约 2.5k tokens：
    - 常用 Bash 指令：让 AI 知道如何像开发者一样操作命令行。
    - 代码风格规范 (Code Style Conventions)：确保 AI 写的代码符合团队编码标准。
    - UI 与内容设计准则：指导 AI 如何设计界面和编写文案。
    - 核心技术实现流程：教 AI 如何处理状态管理 (State Management)、日志记录 (Logging)、错误处理 (Error Handling)、功能门控 (Gating，即控制特定功能的开启与关闭) 以及调试 (Debugging)。
    - 代码合并请求 (Pull Request) 模板：规范提交代码时的文档格式。
- Skill
  - 跟Claude聊天沟通把一个事情做完， 然后说一句“请把上面的推特写作方法写成Skill
    - Use the Skill Creator to build me a Skill for [X]
    - "Use the skill-creator skill to help me build a skill for [your use case]"
  - [Skills｜从概念到实操的完整指南](https://mp.weixin.qq.com/s/Bl4ODUxvwO8pYu9nXVmjuQ)
    - Skills 原理：沙盒 + 渐进式三层加载
      - Level 1 元数据：name/description（YAML）常驻加载，用于“能不能被选中”的索引
      - Level 2 说明文档：触发时用 bash 去读取 SKILL.md 正文进入上下文
      - Level 3 资源与代码：更深资源/脚本按需读取或执行；脚本代码本身不进入上下文，从而节省 token
    - Skill 可以包含三层内容：
      - 第一层：元数据。 就是 name 和 description，告诉 Agent 这个 Skill 是干嘛的、什么时候该用。这部分在启动时就加载，但只占几十个 token。
      - 第二层：指令。 SKILL.md 的主体内容，工作流程、最佳实践、注意事项。只有 Agent 判断需要用这个 Skill 时，才会读取这部分。
      - 第三层：资源和代码。 附带的脚本、模板、参考文档。Agent 按需读取，用的时候才加载。
    - Skills 调用逻辑：意图匹配 → 读取手册 → 按需执行 → 结果反馈
  - [MCP vs Skills](https://x.com/dani_avila7/article/2014409635370041517)
    - https://claude.com/blog/extending-claude-capabilities-with-skills-mcp-servers
  - A Claude agent SKILL is a structured, reusable package stored in your project's ".claude/skills/" folder. It combines the following.
    - Precise instructions defining the agent's role and step-by-step process
    - Reference files (style guides, examples, brand voice)
    - Executable scripts (Python, Node.js, Bash) for reliable, deterministic actions
  - [Avoid dependency hell for Claude SKILLs](https://x.com/juntao/status/2008945207946170471)
  - [Agent Skill](https://mp.weixin.qq.com/s/p-I5lcd43d_6zu3rFIyW0Q)
    - Agent Skills 更像一个操作手册，主要存在本地的文件里面，不需要调用外部接口，主要是用来告诉 AI 有哪些领域知识，然后教 AI 如何正确、高效地使用这些手，按照什么步骤去完成特定任务。
    - Agent Skills 解决了 MCP 无法解决的三个核心问题
      - 节省 token; 解决“会用工具但不懂业务”的问题（业务流程固化）
    - Agent Skills 最核心的创新是渐进式披露（Progressive Disclosure）机制。AI 在使用 Agent Skills 的时候并没有将整个知识库加载到人工智能有限的上下文窗口中，而是以智能的、高效的层级方式加载信息
      - 第一层：元数据（Metadata）：首先只看到每个可用Agent Skills的名称和描述，也就是 Frontmatter buff
      - 第二层：技能主体（Instructions）：一旦确定了相关技能，AI 就会读取主 SKILL.md 文件。该文件包含执行任务的分步指令和核心逻辑
      - 第三层：附加资源（Scripts & References）：如果说明中提到了其他文件（例如用于数据验证的 Python 脚本或报告模板），AI 会根据需要访问这些特定资源
  - [Agent Skills 实战](https://mp.weixin.qq.com/s/lcORy_qmfIv4CHl7tYL2-Q)
    - 从零构建 pdf-translator Skill
      - 目标：PDF → 提取文本 → 翻译 → 输出 Markdown
      - 工程准备：目录结构与 Python 环境
      - 编写 SKILL.md：Frontmatter（元数据）+ Instruction（步骤）
      - 编写脚本工具：extract_text.py / generate_md.py
      - 注册与验证：在 Claude 环境中加载 Skill 并执行
    - Skills 是如何“运行”的
      - 架构：Skill Meta-tool + Individual Skills；本质是 prompt/context 注入（prompt expansion）
      - 决策：通过展示元数据让模型做推理选择（非硬编码路由的设想）
      - 渐进式披露：先只暴露元数据，选中后再加载全文，降低上下文压力
      - 并发与状态：Tools 倾向无状态可并发；Skills 修改上下文因此更“有状态”、并发需谨慎
  - [Agent Skill on CC](https://x.com/wshuyi/article/2009451186039214388)
  - [CC存储体系](https://mp.weixin.qq.com/s/9OJnRhtWzz7MtwP9XRJxKA)
    - 传统的本地存储方案往往存在数据隔离性差、崩溃易丢数据、配置管理混乱、操作不可撤销等问题。
    - 多项目隔离问题：路径编码的项目目录 + Session文件独立存储 → 不同项目数据物理隔离，无交叉干扰；
    - 数据丢失问题：JSONL流式追加写入 + 每条消息实时持久化 → 崩溃时仅可能丢失最后一行未写入数据，损失最小化；
    - 对话追溯问题：uuid+parentUuid消息链 + 完整消息类型（thinking/tool_use/summary） → 可回溯每一轮交互的上下文、工具调用逻辑；
    - 操作不可撤销问题：file-history-snapshot前置备份 + 哈希存储原始内容 + 快捷键撤销 → 支持一键回滚代码修改，无风险；
    - 配置灵活度问题：三级配置体系（全局→本地→项目） + 权限优先级（deny>ask>allow） → 兼顾统一管理与局部定制，同时保障安全；
    - 功能扩展问题：plugins + skills三级架构 → 支持插件、Skill的灵活扩展，适配不同开发场景。
  - [Skill Agent 自动给文章配图](https://x.com/dotey/article/2011907793520116215)
    - 如果你已经有了 Claude Code 这样的 Agent，直接告诉 Agent： ``请帮我安装 github.com/JimLiu/baoyu-skills 中的 Skills``
    - 如果你只需要配图技能，就告诉它： ``请帮我安装宝玉的这个文章配图技能：github.com/JimLiu/baoyu-skills/blob/main/skills/baoyu-article-illustrator/SKILL.md``
  - [agent skill来编排workflow](https://mp.weixin.qq.com/s/jxgfOWfnjfgi5lu57yto7w)
    - Agent Skill 的核心理念：用模块化把复杂任务拆解成可编排 workflow
    - Skill 的几种“玩法”
      - skill 调用 skill（复用/分层）
        ```
        After all tasks complete and verified:
        - Announce: "I'm using the finishing-a-development-branch skill to complete this work."
        - **REQUIRED SUB-SKILL:** Use finishing-a-development-branch skill
        - Follow that skill to verify tests, present options, execute choice
        ```
      - skill 调用工具脚本（自动化执行）
      - 生成可验证的中间输出（文件化、可追溯、可续跑）分析 → 创建文件 → 验证 → 执行 → 验证
        ```
        When this skill is invoked:

        1. Create the `./input` directory if it doesn't exist
        2. Get the user's input message (passed as arguments or prompt for it)
        3. Generate a timestamp-based filename (format: `YYYY-MM-DD_HH-MM-SS.txt`)
        4. Save the input to `./input/<timestamp>.txt`
        5. Confirm the file has been saved with the full path
        ```
      - skill 调用 subagent（独立上下文、并行、专业化分工） Subagent 之间只传文件路径，不传内容，这条规则很重要
      - 用 reference 外挂文档（减少上下文、提升聚焦）
    - 常见 workflow pattern
      - 清单模式（Checklist）
      - 循环验证模式（validator loop）
      - 条件工作流模式（if/else 分支）
      - examples pattern（用输入输出示例约束行为）
      - 模板 pattern（强约束输出结构）
    - https://github.com/luozhiyun993/skill-workflow
  - [AGENTS.md outperforms skills in our agent evals](https://vercel.com/blog/agents-md-outperforms-skills-in-our-agent-evals)
    - Agent Skills（skills）：一种把说明、工具、脚本、参考资料打包成“按需调用”的能力包（open standard）
    - AGENTS.md：放在仓库根目录的 Markdown 文件，为 agent 提供每轮对话都常驻的项目上下文（类似 Claude Code 的 CLAUDE.md）
    - Skills 最大的问题：触发不稳定、并且“提示词很脆”
    - Compress aggressively. You don't need full docs in context. An index pointing to retrievable files works just as well.
    - Test with evals. Build evals targeting APIs not in training data. That's where doc access matters most.
- [Continuous Claude](https://github.com/parcadei/Continuous-Claude): 
  - 解决 Claude Code 等 AI Coding Agent 在长会话中面临的一个痛点：上下文丢失与“遗忘”
  -  原生机制：为了节省空间，Claude Code 会进行“压缩”，把之前的对话总结成摘要。
    - 后果：这种压缩是有损的。经过几次压缩后，“摘要的摘要”会丢失大量细节（比如某个函数的具体参数、之前的调试结果），导致 AI 开始“胡言乱语”或重复错误
  - Continuous-Claude 提出了一套类似人类工程师的工作流，用清晰的文档代替模糊的记忆。
  -  Ledger（账本/工作日志）：
     · 在当前会话中，它会实时记录“目标是什么”、“完成了什么”、“下一步做什么”。
     · 当你感觉到上下文快满时，你可以直接输入 /clear 清空上下文。但不用担心，Claude Code 会自动加载这个“账本”，瞬间找回工作状态，而不是依赖那个模糊的摘要。
  - Handoff（交接文档）：
     · 当你结束一天的工作，或者一个 Agent 完成了它的任务时，系统会生成一份详细的 markdown 交接文档。
     · 下一次启动（或下一个 Agent 接手）时，直接读取这份文档。这就像早班同事给晚班同事留了一份详细的交接条，确保无缝衔接。
  - 设计了一套完整的 Agent 编排 和 工具执行 逻辑：
    - A. 技能 vs 智能体
      · 技能：在当前会话中快速执行的动作。比如“写个 Commit”、“查个文档”。
      · Agent：当任务太复杂（比如“设计整个后端架构”），在当前窗口做会严重消耗 Token 时，系统会新开一个干净的上下文来运行 Agent。典型流程：Plan Agent (做计划) -> Validate Agent (验证/查资料) -> Implement Agent (执行代码)。
    - B. MCP 代码执行 (Token 节省)
      - 这是非常聪明的一点。通常 AI 运行工具会将工具的定义和结果塞满上下文。
      - Continuous-Claude 的做法：它通过脚本在外部运行工具（如 Python 脚本），只把必要的结果返回给 AI。这极大地减少了 Token 的占用，防止上下文被工具调用的冗余信息“污染”。
    - C. 钩子系统 (Hooks System)
      - 它利用 Claude Code 的生命周期钩子（Hooks）实现了自动化：
        · SessionStart：自动加载上次的“账本”和“交接文档”。
        · PreCompact：在 Claude 试图压缩上下文之前，自动拦截并保存当前状态，防止信息丢失。
        · StatusLine：在终端底部显示一个彩色的状态栏，实时告诉你 Token 用了多少，是否需要清理上下文了。
- [Claude Code power user customization: How to configure hooks](https://claude.com/blog/how-to-configure-hooks)
  - Hook
    - Hook 本质上是你编写的一段自定义 Shell 命令：当 Claude Code 会话内发生某个“事件”（例如即将写文件、你提交提示词等）时，Claude Code 会自动执行该命令
    - Hook 在本地环境以你的用户权限运行，通过 stdin 接收事件信息（JSON），并通过退出码 + stdout把结果反馈给 Claude Code，从而在“不修改 Claude Code 本体”的情况下精确改变其行为
  - Hook 主要解决三类摩擦
    - 消除重复手工步骤：如每次写文件后自动跑 Prettier；反复执行 npm test 不再每次弹权限确认。
    - 自动执行项目规则/护栏：如阻止危险命令、写入前校验路径、防止覆盖敏感文件。
    - 动态注入上下文：会话启动时注入 git status、TODO；每次发问自动附加当前 sprint 上下文
  - Hook 配置放在 JSON settings 文件中，文章强调有 三层：
    - 项目级（可提交共享）：.claude/settings.json
    - 用户级（全局生效）：~/.claude/settings.json
    - 项目本地（个人配置，不想提交）：.claude/settings.local.json

  | Hook 事件 | 触发时机 | 常见用途（文章示例/要点） |
  |---|---|---|
  | **PreToolUse** | Claude 选定要用某个工具后、工具执行前 | 阻止危险 Bash、校验写入路径、自动批准安全操作、甚至修改工具参数  |
  | **PermissionRequest** | 将要弹出权限对话框之前 | 对 `npm test*` 之类高频命令自动批准；阻止访问敏感目录/文件  |
  | **PostToolUse** | 工具成功执行后立刻触发 | 写/改文件后自动格式化（Prettier/Black/gofmt）、跑 lint、记录审计日志  |
  | **PreCompact** | Claude 即将做“上下文压缩（compaction）”前 | 备份完整 transcript、提取关键决策/代码片段，避免压缩丢细节  |
  | **SessionStart** | 会话开始或恢复时 | 把 `git status`、TODO、环境信息输出到 stdout 作为上下文注入  |
  | **Stop** | Claude 完成响应并准备等待下一步输入时 | 做“是否真的完成”的检查；可强制继续（多步骤工作流/检查清单）  |
  | **SubagentStop** | 子代理（subagent）完成时 | 验证子代理输出质量，决定 accept/reject，并触发后续动作  |
  | **UserPromptSubmit** | 你提交 prompt 后、Claude 处理前 | 每次提问自动附加 sprint context、最近错误日志/测试结果；也可做 prompt 校验/拦截  |

- [AI code guild](https://github.com/automata/aicodeguide)
  - AI 时代的代码审核：写两遍，反而更快
    - 先用最低成本把路趟一遍。第一版跑完，需求确认了，技术难点解决了，再来做设计，这时候你知道该设计什么、不该设计什么。少走很多弯路。
  -  vibe coding 真正的价值：不是让你变成工程师，而是让你能自己解决自己的问题。
  - 适合 vibe coding 的场景： 个人自动化工具、一次性脚本、快速验证想法的原型、不涉及敏感数据的内部工具
  - 不适合的场景： 金融、医疗等需要高可靠性的系统；需要长期维护迭代的产品；多人协作的代码库；涉及用户隐私和安全的应用

- [Cursor 的进阶用法](https://x.com/xiaokedada/status/1833132309496885434?s=46)
  - https://cursor101.com/zh
  - 1. Set 5-10 clear project rules upfront so Cursor knows your structure and constraints. Try /generate rules for existing codebases.
  - 2. Be specific in prompts. Spell out tech stack, behavior, and constraints like a mini spec.
  - 3. Work file by file; generate, test, and review in small, focused chunks.
  - 4. Write tests first, lock them, and generate code until all tests pass.
  - 5. Always review AI output and hard‑fix anything that breaks, then tell Cursor to use them as examples.
  - 6. Use @ file, @ folders, @ git to scope Cursor’s attention to the right parts of your codebase.
  - 7. Keep design docs and checklists in .cursor/ so the agent has full context on what to do next.
  - 8. If code is wrong, just write it yourself. Cursor learns faster from edits than explanations.
  - 9. Use chat history to iterate on old prompts without starting over.
  - 10. Choose models intentionally. Gemini for precision, Claude for breadth.
  - 11. In new or unfamiliar stacks, paste in link to documentation. Make Cursor explain all errors and fixes line by line.
  - 12.Let big projects index overnight and limit context scope to keep performance snappy.
  - 指令Prompt
    ```
    你是一个优秀的技术架构师和优秀的程序员，在进行架构分析、功能模块分析，以及进行编码的时候，请遵循如下规则：
    1. 分析问题和技术架构、代码模块组合等的时候请遵循“第一性原理”
    2. 在编码的时候，请遵循 “DRY原则”、“KISS原则”、“SOLID原则”、“YAGNI原则”
    3. 如果单独的类、函数或代码文件超过500行，请进行识别分解和分离，在识别、分解、分离的过程中请遵循以上原则
    ```
  - [Cursor AI编程神器：14个实用技巧](https://mp.weixin.qq.com/s/fGHyMzF9M5unuH7YNL1ADg)
    - 通过MCP获取最新知识: Context7 - 提供丰富的上下文信息 ; DeepWiki - 深度维基知识库
    - 善用.cursor/rules: 级联Cursor规则是一个强大的新功能，你可以组合多个规则文件
    - 灵活使用忽略文件: .cursorignore - 完全不索引的文件; .cursorindexignore - 不索引但可以在聊天中用@引用的文件
    - 掌握@符号的强大功能: @Files & Folders - 缩小上下文范围，帮助AI专注于相关文件 ; @git - 查看特定Git提交中发生的变化; @terminal - 访问日志和错误信息
    - 在.cursor/mcp.json中配置你的MCP服务器
    - 内联编辑功能
    - Settings > General > Privacy Mode
    - Homebrew安装最新版本的Cursor `brew install --cask --force cursor`
  - [cursor的codebase indexing](https://mp.weixin.qq.com/s/fj-9rOPEq_eF05VLQizX1g)
    - 什么是Merkle Tree 哈希树
      - 高效验证 数据完整性保证 增量同步
    - turbopuffer的serverless架构, 缓存/冷热策略，为Cursor实现了成本和性能的完美平衡。
    - Merkle tree 负责本地变更检测和高效同步，turbopuffer 负责云端的向量存储与检索。
  - [How I use Claude Code](https://www.reddit.com/r/ClaudeAI/comments/1lkfz1h/how_i_use_claude_code/)
    - 1. 维护 CLAUDE[.]md 文件
      - 建议为不同子目录（如测试、前端、后端）分别维护 CLAUDE[.]md，记录指令和上下文，便于 Claude 理解项目背景。
    - 2. 善用内置命令
      - ▫ Plan mode（shift+tab）：提升任务完成度和可靠性。
      - ▫ Verbose mode（CTRL+R）：查看 Claude 当前的全部上下文。
      - ▫ Bash mode（!前缀）：运行命令并将输出作为上下文。
      - ▫ Escape 键：中断或回溯对话历史。
    - 3. 并行运行多个实例: 前后端可分别用不同实例开发，提高效率，但复杂项目建议只用一个实例以减少环境配置麻烦。
    - 4. 使用子代理（subagents: 让多个子代理从不同角度解决问题，主代理负责整合和比较结果。
    - 5. 利用视觉输入: 支持拖拽截图，Claude Code 能理解视觉信息，适合调试 UI 或复现设计。
    - 6. 优先选择 Claude 4 Opus: 高级订阅用户建议优先用 Opus，体验和能力更强。
    - 7. 自定义项目专属 slash 命令: 在 `.claude/commands` 目录下编写常用任务、项目初始化、迁移等命令，提升自动化和复用性。
    - 8. 使用 Extended Thinking: 输入 `think`、`think harder` 或 `ultrathink`，让 Claude 分配更多“思考预算”，适合复杂任务。
    - 9. 文档化一切: 让 Claude 记录思路、任务、设计等到中间文档，便于后续追溯和上下文补充。
    - 10. 频繁使用 Git 进行版本控制: Claude 可帮写 commit message，AI 辅助开发时更要重视版本管理。
    - 11. 优化工作流
      - ▫ 用 `--resume` 继续会话，保持上下文。
      - ▫ 用 MCP 服务器或自建工具管理上下文。
      - ▫ 用 GitHub CLI 获取上下文而非 fetch 工具。
      - ▫ 用 ccusage 监控用量。
    - 12. 追求快速反馈循环: 给模型提供验证机制，减少“奖励劫持”（AI 取巧而非真正解决问题）。
    - 13. 集成到 IDE: 体验更像“结对编程”，Claude 可直接与 IDE 工具交互。
    - 14. 消息排队: Claude 处理任务时可继续发送消息，排队等待处理。
    - 15. 注意会话压缩与上下文长度 : 合理压缩对话，避免丢失重要上下文，建议在自然停顿点进行。
    - 16. 自定义 PR 模板 : 不要用默认模板，针对项目定制更合适的 PR（pull request) 模板。
  - [claude-code-cookbook](https://github.com/wasabeef/claude-code-cookbook/blob/main/README_zh.md)
  - [Getting Good Results from Claude Code](https://www.dzombak.com/blog/2025/08/getting-good-results-from-claude-code/)
    - Claude Code 新的 Learning mode 就是一个例子，你可以在启动Claude 后，输入 “/output-styles” 命令，选择 Learning 模式 “3. Learning” ，那么 Claude 就会只实现整体框架，留一个小模块让你自己实现。
    - 可以选择“2. Explanatory”，Claude 会在工作过程中生成其决策过程的摘要，让你有机会更好地理解它在做什么
  - [Claude Code 最佳实践](https://cc.deeptoai.com/docs/zh/best-practices/claude-code-best-practices)
    - `claude --permission-mode bypassPermissions`
    - `claude --dangerously-skip-permissions
       codex --dangerously-bypass-approvals-and-sandbox`
    - 在 ~/.claude/settings.json  加入下面的配置，就可以看你 Claude Code 的实时消耗了
      ```
      {
        "statusLine": {
          "type": "command",
          "command": "bun x ccusage statusline", // Use "npx -y ccusage statusline" if you prefer npm
          "padding": 0  // Optional: set to 0 to let status line go to edge
        }
      }
      ```
    - 第一，Plan先行。写任何代码之前，先把整件事的来龙去脉想清楚。构建系统分几层，依赖关系是什么样的，该从哪一层开始动手，这些都要捋明白
    - 第二，定义约束。这是最关键的一步，所谓约束，就是明确“什么叫做完了”。代码写完根本不算完，满足所有预设的约束条件，才叫真正的完成。
    - Review测试，不review代码。我会自己写测试用例，也会让Claude生成，但我审查的核心是测试
  - [Vibe Coding 有“最后一公里”知识幻觉](https://mp.weixin.qq.com/s/loRz_3N_N_fz58yFt_BanQ)
    - Milvus Code Helper MCP 服务外，开发者还可以选择如 Context7、DeepWiki 等新兴工具来解决这类问题
  - [Claude Code 如何做任务进度跟进](https://gist.github.com/richzw/ebeb0f8b39af64f2dd3a765aa4662150)
    - 每一个新需求，让Claude Code帮你自动生成一个对应md文件， 该文件包含plan和progress
    - Claude Code自带一个"内存版的todo list"，就是在面临新需求的时候， 它会自动拆解， 但是这个仅仅是用于更好的让用户查看当前进度，以及LLM自己保持前后一致性， 缺点是， 当前任务结束后用户并不好review。
    - 如何review呢？ 就是让Claude Code建一个plan and progress的同步版本 md文件。
      - 每次都要提醒一次吗？ 不用， 将prompt写入CLAUDE[.]md文件即可。 我一般都是放在 `docs/plan` 文件夹
  - [Claude Code: Best practices for agentic coding](https://www.anthropic.com/engineering/claude-code-best-practices)
  - [AI 写代码的深度体验](https://mp.weixin.qq.com/s/6dLnTlb0RfnLjrExa7j_zQ)
  - [How Anthropic teams use Claude Code](https://www-cdn.anthropic.com/58284b19e702b49db9302d5b6f135ad8871e7658.pdf)
  - [Claude Code Manual](https://docs.anthropic.com/zh-CN/docs/claude-code/overview)
  - [A curated list of awesome commands, files, and workflows for Claude Code](https://github.com/hesreallyhim/awesome-claude-code)
  - [claude-code-costs](https://github.com/philipp-spiess/claude-code-costs)
    - 如果你使用 Claude Code 并且是 Claude Pro/Max 订阅想要知道如果是 API 得花了多少钱，订阅费花的值不值，可以试试这个命令：$ npx claude-code-costs
  - [Claude Code 的自定义指令](https://docs.anthropic.com/en/docs/claude-code/slash-commands)
    - Claude Code 现在可以添加自定义指令，也就是你输入 “/” 可以出来命令提示，这个 ultrathink-task 可以调用架构智能体
    - https://www.reddit.com/r/ClaudeAI/comments/1lpvj7z/ultrathink_task_command/
  - [Claude Code如何引爆全员生产力](https://mp.weixin.qq.com/s/TsDK6-aM0HU33CdSitging)
  - 为了防止 claude code 习惯性代码过度膨胀，我的做法是使用一个 code-simplifie 的 sub agent ，要求每一个功能/todo之后都需要使 code-simplifie 优化代码。
  - 用好 Coding Agent 的一个经验技巧，就是为 Agent 提供验证结果的方法，这样 Agent 就会自己去测试去修改，直到完成任务，不需要自己反复测试修改。
    - 在用 Claude Code 或者 Copilot/Curosr 的 Agent mode，会在提示词中加一句类似的话：
      Please write tests and verify the tests by running
      `npx jest <testfilepath> -c './jest.config.ts' --no-coverage`
  - Cursor vs Claude Code
    - 用 Cursor 作为主要 IDE，享受熟悉的界面和顺滑的 Tab 补全；
    - 遇到复杂问题/bug时，在 Cursor 的终端中启动 Claude Code；
    - 让 Claude Code 负责思考和规划，Cursor 负责执行和微调；
  - [How to build a Claude Code like agent](https://minusx.ai/blog/decoding-claude-code/)
    - Control Loop
      - Keep one main loop (with max one branch) and one message history
      - Use a smaller model for all sorts of things. All. The. Frickin. Time.
    - Prompts
      - Use claude.md pattern to collaborate on and remember user preferences
      - Use special XML Tags, Markdown, and lots of examples
    - Tools
      - LLM search >>> RAG based search
      - How to design good tools? (High vs Low level tools)
      - Let your agent manage its own todo list
    - Steerability
      - Tone and style
      - "PLEASE THIS IS IMPORTANT" is unfortunately still state of the art
      - Write the algorithm, with heuristics and examples
  - [How I Use Claude Code](https://boristane.com/blog/how-i-use-claude-code/)
    - One core principle: never let Claude write code until you’ve reviewed and approved a written plan. 
    - 调研 -> 计划 -> 标注 -> 实施 -> 反馈
      - Research
        - read this folder in depth, understand how it works deeply, what it does and all its specificities.
        - when that’s done, write a detailed report of your learnings and findings in research.md
      - Planning
        - I want to build a new feature <name and description> that extends the system to perform <business outcome>. 
        - write a detailed plan.md document outlining how to implement this. include code snippets
      - The Todo List
        - add a detailed todo list to the plan, with all the phases and individual tasks necessary to complete the plan - don’t implement yet
- [Prompt auto-caching with Claude](https://x.com/RLanceMartin/article/2024573404888911886)
  - `"cache_control":  {"type": "ephemeral"} `















