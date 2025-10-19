FROM oven/bun:latest

RUN cat > /root/.bunfig.toml <<EOF
telemetry = false

[install]
registry = "https://registry.npmmirror.com"
EOF

# 软链接 & 包装脚本：用 bun 顶替 node/npm/npx
RUN set -eux; \
  # 1) node -> bun（兼容 /usr/bin/env node）
  ln -sf /usr/local/bin/bun /usr/local/bin/node; \
  \
  # 2) npx -> bun x
  printf '#!/usr/bin/env sh\nexec bun x "$@"\n' > /usr/local/bin/npx; \
  chmod +x /usr/local/bin/npx; \
  \
  # 3) npm 智能包装：常见子命令直通 bun
  printf '%s\n' \
'#!/usr/bin/env sh' \
'# npm shim -> bun' \
'cmd="$1"; shift || true' \
'case "$cmd" in' \
'  run)            exec bun run "$@";;' \
'  ci)             exec bun install --frozen-lockfile "$@";;' \
'  install|"")     exec bun install "$@";;' \
'  exec)           exec bun x "$@";;' \
'  --version|-v|version) exec bun --version;;' \
'  *)              exec bun install "$cmd" "$@";;' \
'esac' \
  > /usr/local/bin/npm; \
  chmod +x /usr/local/bin/npm;
