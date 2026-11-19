# Local cc-connect features — build & restore notes

This branch carries three local features on top of upstream `main`:

1. `feat(core): fold busy-session messages into the running turn` — per-project
   `steer_while_busy`: a message sent while a turn is running is folded into
   that turn instead of waiting for the next one (attachments + history
   supported; Claude Code, Codex app-server, Pi RPC).
2. `feat(feishu): let the agent recall its own last message` — per-project
   `allow_agent_recall` + `cc-connect recall --last`.
3. `fix(core): honor the configured language in the injected agent prompt` —
   the top-level `language` is forwarded to the agent, and the prompt preamble
   is localized.

## Build

Requires Go 1.25 (the version in `go.mod`).

```bash
git clone git@github.com:Jonny-hzy/cc-connect.git
cd cc-connect
git checkout feat/agent-recall

# Web admin assets are not present in a plain clone; skip them.
go build -tags no_web -o cc-connect ./cmd/cc-connect
```

`-tags no_web` disables the embedded web admin. To include it, build the web
assets first (`cd web && pnpm install && pnpm build`) and run `make build`.

## Run

```bash
# The bridge must be launched with the custom binary, and the agent processes
# inherit PATH from it: this is what makes `cc-connect recall --last` resolve
# to the new CLI instead of the npm v1.5.0 wrapper (which has no recall).
export PATH="/path/to/cc-connect:$HOME/.local/bin:$PATH"
./cc-connect --force --config ~/.cc-connect/config.toml
```

## Config snippets (placeholders only — never commit real app secrets)

```toml
language = "zh"

[display]
mode = "compact"

[[projects]]
name = "demo"
steer_while_busy = true
allow_agent_recall = true

[projects.agent]
type = "claudecode"

[projects.agent.options]
work_dir = "/path/to/project"
mode = "bypassPermissions"          # or "default" / "acceptEdits"
context_window_tokens = 1048576     # only for 1M models; corrects the 1,000,000 heuristic
append_system_prompt = """
你是通过飞书与用户协作的工程师。
1. 开工时用一句话确认要做什么，不要复述需求。
2. 只在里程碑发进展，进展保持简短，不要贴日志或逐条工具播报。
3. 发现上一条回复有事实错误时，先运行 `/path/to/cc-connect recall --last` 撤回，再发更正。
   不要用自己的飞书应用凭据直接调 API——撤回必须由发出该消息的桥接应用执行。
"""

[projects.auto_compress]
enabled = true
max_tokens = 916576      # window - output reserve (32k) - 100k headroom
min_gap_mins = 10

[[projects.platforms]]
type = "feishu"

[projects.platforms.options]
app_id = "cli_xxx"
app_secret = "xxx"
allow_from = ["ou_xxx"]
admin_from = ["ou_xxx"]
enable_feishu_card = true
reaction_emoji = "OnIt"  # processing reaction; "none" disables
ack_emoji = "none"       # persistent receipt; "none" disables
done_emoji = "none"      # completion reaction; "none" disables
```

## Notes

- Recall requires the bridge process to hold the Feishu client of the app that
  sent the message. Calling the Feishu API with a different app returns
  `230026 no permission to recall this message`.
- `append_system_prompt` uses the absolute path to the binary so recall works
  regardless of the agent's PATH; the PATH export above is still recommended so
  `cc-connect send` / `cron` / `timer` resolve to the same build.
- Only the newest bot message in the chat is tracked for recall; multi-chunk
  long replies recall only the last chunk.
- A failed or rejected steer falls back to the FIFO queue with the usual
  "will process after the current task finishes" receipt.
