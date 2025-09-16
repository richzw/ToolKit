- [global agent guide](https://www.dzombak.com/blog/2025/08/getting-good-results-from-claude-code/)
```
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

```

- [ChatGPT 和 Claude 都有记忆功能，但两者实现原理截然不同](https://www.shloked.com/writing/claude-memory)
  - ChatGPT 的记忆模式是自动化、魔法般的个性记忆，不需要用户提醒，自动的悄悄记录用户的使用细节。
  - Claude的记忆模式是基于检索的，每次新开对话，都没有任何任何历史记忆，只有当你明确告诉 Claude 需要用到某条记忆内容，它才会从真实的历史记录中精准提取信息。
    - 一是基于关键词搜索历史对话，二是以时间线为索引检索近期聊天。
    - 两个检索工具：conversation_search（按关键词/主题跨全部历史查找、可多主题分别检索并汇总）与 recent_chats（按最近N次或特定时间窗口检索、可排序分页）。代价是会有可见延迟，但换来更强的可控性与可审计性
  - ChatGPT 走大众消费品路线：记忆默认开启、自动加载画像/摘要，追求“即刻个性化”和零等待，有利于规模化与留存；
  - Claude 面向开发者与专业工作流：用户理解并愿意显式触发工具（如搜索、长推理、记忆），更重视隐私与可预测性，记忆只是“按需调用”的工具之一




