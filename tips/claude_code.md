- [global agent guide](https://www.dzombak.com/blog/2025/08/getting-good-results-from-claude-code/)

    # Development Guidelines
    
    ## Philosophy
    
    ### Core Beliefs
    
    - **Incremental progress over big bangs** - Small changes that compile and pass tests
    - **Learning from existing code** - Study and plan before implementing
    - **Pragmatic over dogmatic** - Adapt to project reality
    - **Clear intent over clever code** - Be boring and obvious
    
    ### Simplicity Means
    
    - Single responsibility per function/class
    - Avoid premature abstractions
    - No clever tricks - choose the boring solution
    - If you need to explain it, it's too complex
    
    ## Process
    
    ### 1. Planning & Staging
    
    Break complex work into 3-5 stages. Document in `IMPLEMENTATION_PLAN.md`:
    
    ```markdown
    ## Stage N: [Name]
    **Goal**: [Specific deliverable]
    **Success Criteria**: [Testable outcomes]
    **Tests**: [Specific test cases]
    **Status**: [Not Started|In Progress|Complete]
    ```
    - Update status as you progress
    - Remove file when all stages are done
    
    ### 2. Implementation Flow
    
    1. **Understand** - Study existing patterns in codebase
    2. **Test** - Write test first (red)
    3. **Implement** - Minimal code to pass (green)
    4. **Refactor** - Clean up with tests passing
    5. **Commit** - With clear message linking to plan
    
    ### 3. When Stuck (After 3 Attempts)
    
    **CRITICAL**: Maximum 3 attempts per issue, then STOP.
    
    1. **Document what failed**:
        - What you tried
        - Specific error messages
        - Why you think it failed
    
    2. **Research alternatives**:
        - Find 2-3 similar implementations
        - Note different approaches used
    
    3. **Question fundamentals**:
        - Is this the right abstraction level?
        - Can this be split into smaller problems?
        - Is there a simpler approach entirely?
    
    4. **Try different angle**:
        - Different library/framework feature?
        - Different architectural pattern?
        - Remove abstraction instead of adding?
    
    ## Technical Standards
    
    ### Architecture Principles
    
    - **Composition over inheritance** - Use dependency injection
    - **Interfaces over singletons** - Enable testing and flexibility
    - **Explicit over implicit** - Clear data flow and dependencies
    - **Test-driven when possible** - Never disable tests, fix them
    
    ### Code Quality
    
    - **Every commit must**:
        - Compile successfully
        - Pass all existing tests
        - Include tests for new functionality
        - Follow project formatting/linting
    
    - **Before committing**:
        - Run formatters/linters
        - Self-review changes
        - Ensure commit message explains "why"
    
    ### Error Handling
    
    - Fail fast with descriptive messages
    - Include context for debugging
    - Handle errors at appropriate level
    - Never silently swallow exceptions
    
    ## Decision Framework
    
    When multiple valid approaches exist, choose based on:
    
    1. **Testability** - Can I easily test this?
    2. **Readability** - Will someone understand this in 6 months?
    3. **Consistency** - Does this match project patterns?
    4. **Simplicity** - Is this the simplest solution that works?
    5. **Reversibility** - How hard to change later?
    
    ## Project Integration
    
    ### Learning the Codebase
    
    - Find 3 similar features/components
    - Identify common patterns and conventions
    - Use same libraries/utilities when possible
    - Follow existing test patterns
    
    ### Tooling
    
    - Use project's existing build system
    - Use project's test framework
    - Use project's formatter/linter settings
    - Don't introduce new tools without strong justification
    
    ## Quality Gates
    
    ### Definition of Done
    
    - [ ] Tests written and passing
    - [ ] Code follows project conventions
    - [ ] No linter/formatter warnings
    - [ ] Commit messages are clear
    - [ ] Implementation matches plan
    - [ ] No TODOs without issue numbers
    
    ### Test Guidelines
    
    - Test behavior, not implementation
    - One assertion per test when possible
    - Clear test names describing scenario
    - Use existing test utilities/helpers
    - Tests should be deterministic
    
    ## Important Reminders
    
    **NEVER**:
    - Use `--no-verify` to bypass commit hooks
    - Disable tests instead of fixing them
    - Commit code that doesn't compile
    - Make assumptions - verify with existing code
    
    **ALWAYS**:
    - Commit working code incrementally
    - Update plan documentation as you go
    - Learn from existing implementations
    - Stop after 3 failed attempts and reassess


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
- Codex
  - codex --dangerously-bypass-approvals-and-sandbox 全自动跑
  - codex resume用来选择历史记录sessions
  - codex目前没有plan模式，可以/approvals选择read-only进行讨论，或者直接在提示词里要求codex进行plan，写到一个文档中
  - /review 命令很实用，在写完代码后让 codex review
- Codex Prompt
  - Make a pixel art game where I can walk around and talk to other villagers, and catch wild bugs
  - Give me a work management platform that helps teams organize, track, and manage their projects and tasks. Give me the platform with a kanban board, not the landing page.
  - Given this image as inspiration. Build a simple html page joke-site.html here that includes all the assets/javascript and content to implement a showcase version of this webapp. Delightful animations and a responsive design would be great but don't make things too busy
- Claude Code
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
  - 

















