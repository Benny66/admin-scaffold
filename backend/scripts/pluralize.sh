#!/usr/bin/env bash
# 资源名复数化：模块名 → REST 路由路径与权限码前缀使用的复数形式。
#
# 这是复数规则的「单一真相」：gen-module.sh 与 internal/guard 的范例一致性护栏
# 共同调用本脚本，避免规则在两处各写一份而漂移（见 gen-module-closure design D1）。
#
# 用法：pluralize.sh <单数名词>
#   pluralize.sh asset     → assets
#   pluralize.sh box       → boxes
#   pluralize.sh category  → categories
#   pluralize.sh day       → days
#
# 规则：
#   以 s / x / z / ch / sh 结尾  → 加 es
#   以「辅音 + y」结尾            → y 改 ies
#   其余                          → 加 s
#
# 已知限制：不处理不规则名词（person/people、datum/data）。模块名由开发者自取，
# 遇到不规则词时在生成后手工微调即可，不为此引入英语词典。
set -euo pipefail

word="${1:-}"
if [[ -z "$word" ]]; then
  echo "用法：$0 <单数名词>" >&2
  exit 1
fi

case "$word" in
  *s|*x|*z|*ch|*sh) printf '%ses\n' "$word" ;;
  *[^aeiou]y)       printf '%sies\n' "${word%y}" ;;
  *)                printf '%ss\n' "$word" ;;
esac
