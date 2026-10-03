#!/usr/bin/env bash
# verify-docs.sh — 端到端校验 .kiro/steering/{product,tech,structure}.md
# 与 AC-1 ~ AC-15 的 grep 模式一致。
#
# 用法：bash scripts/verify-docs.sh
# 退出：0 = 全部 PASS；1 = 至少 1 条 FAIL

set -u

ROOT="$(cd "$(dirname "$0")/../../../.." && pwd)"
PRODUCT="$ROOT/.kiro/steering/product.md"
TECH="$ROOT/.kiro/steering/tech.md"
STRUCTURE="$ROOT/.kiro/steering/structure.md"
UPGRADE="$ROOT/internal/upgrade/upgrade.go"

pass=0
fail=0

ok() { echo "PASS: $1"; pass=$((pass + 1)); }
ng() { echo "FAIL: $1 — $2"; fail=$((fail + 1)); }

# AC-1 product.md 状态机：planning 阶段 + 可互转 + APPROVED → TASKS 回填字符串存在
# 注意：'planning 阶段' 与 '可互转' 之间隔了 '（requirements / design / tasks）'，不能作为子串模式匹配
if grep -qE 'planning 阶段' "$PRODUCT" && grep -q '可互转' "$PRODUCT" && grep -q 'APPROVED → TASKS' "$PRODUCT"; then
  ok "AC-1 product.md 状态机描述含 planning 阶段 + 可互转 + APPROVED 回填"
else
  ng "AC-1" "product.md 缺 'planning 阶段' / '可互转' / 'APPROVED → TASKS' 中任一字符串"
fi

# AC-2 product.md EARS 数量：6 类正则分支 + 出现 ubiquitous
if grep -qE '6 类正则分支' "$PRODUCT" && grep -q 'ubiquitous' "$PRODUCT"; then
  ok "AC-2 product.md EARS 数量为 5+1=6 类"
else
  ng "AC-2" "product.md EARS 数量描述缺 6 类正则分支或缺 ubiquitous"
fi

# AC-3 product.md advisory 数量：≥ 3 个不同 finding code（用 -oE 计数独立命中）
adv_count=$(grep -oE 'vague-language|duplicate-acceptance-criteria|uncovered-acceptance-criteria|tasks-without-requirements' "$PRODUCT" | wc -l | tr -d ' ')
if [[ $adv_count -ge 3 ]]; then
  ok "AC-3 product.md advisory 列出 $adv_count 类 finding（≥3）"
else
  ng "AC-3" "product.md advisory 仅 $adv_count 类，需 ≥ 3"
fi

# AC-4 product.md issue→spec：声明 gh CLI（允许反引号包裹）+ auth login
if grep -qE '`gh` CLI|gh CLI' "$PRODUCT" && grep -qE '`gh auth login`|gh auth login' "$PRODUCT"; then
  ok "AC-4 product.md issue→spec 声明 gh CLI 前置依赖"
else
  ng "AC-4" "product.md issue→spec 缺 gh CLI 或 gh auth login 声明"
fi

# AC-5 product.md browser→spec：声明 bsk + browser-skill
if grep -q 'bsk' "$PRODUCT" && grep -q 'browser-skill' "$PRODUCT"; then
  ok "AC-5 product.md browser→spec 声明 bsk 前置依赖"
else
  ng "AC-5" "product.md browser→spec 缺 bsk 或 browser-skill 声明"
fi

# AC-6 product.md watch：5 个 preset 全部列出（用 -oE 计数独立命中）
preset_count=$(grep -oE '`(default|lint|status|reactive|full)`' "$PRODUCT" | wc -l | tr -d ' ')
if [[ $preset_count -ge 5 ]]; then
  ok "AC-6 product.md watch 列出 $preset_count 个 preset"
else
  ng "AC-6" "product.md watch preset 仅 $preset_count 个，需 ≥ 5"
fi

# AC-7 product.md '4 widget'：≥ 3 个 widget 名字（用 -oE 计数独立命中）
widget_count=$(grep -oE 'stat-card|sparkline-cell|phase-badge|waves-progress' "$PRODUCT" | wc -l | tr -d ' ')
if [[ $widget_count -ge 3 ]]; then
  ok "AC-7 product.md '4 widget' 明确指代 $widget_count 个 widget"
else
  ng "AC-7" "product.md 4 widget 仅 $widget_count 个，需 ≥ 3"
fi

# AC-8 product.md Windows re-exec 限制
if grep -qiE 'windows.*(手动|重启)|(手动|重启).*windows' "$PRODUCT"; then
  ok "AC-8 product.md 声明 Windows re-exec 限制"
else
  ng "AC-8" "product.md 缺 Windows re-exec 限制声明"
fi

# AC-9 product.md fsnotify fallback
if grep -qE 'fallback|降级' "$PRODUCT" && grep -qE '2s' "$PRODUCT"; then
  ok "AC-9 product.md 声明 fsnotify fallback"
else
  ng "AC-9" "product.md 缺 fsnotify fallback 或 2s 声明"
fi

# AC-10 tech.md：原"保留 ~/.prev 回滚"句消失 + 替换为升级失败行为段落 + 代码 os.Rename 保留
if ! grep -q '保留上一份二进制到.*\.prev' "$TECH" \
   && grep -qE '升级失败行为|不替换旧二进制' "$TECH" \
   && grep -q 'os\.Rename' "$UPGRADE"; then
  ok "AC-10 tech.md 旧 .prev 回滚句消失 + 替换为升级失败行为 + 代码现状保留"
else
  ng "AC-10" "tech.md 仍含旧回滚句 / 缺升级失败行为 / 代码异常（仍 os.Rename）"
fi

# AC-11 tech.md 覆盖率阈值：不强制声明
if grep -qE 'CI 当前仅打印|不强制阈值' "$TECH"; then
  ok "AC-11 tech.md 覆盖率阈值已声明不强制"
else
  ng "AC-11" "tech.md 缺覆盖率阈值不强制声明"
fi

# AC-12 tech.md 性能预算：待基准验证
if grep -qE '待基准验证|无.*基准|无.*常量' "$TECH"; then
  ok "AC-12 tech.md 性能预算已声明待基准验证"
else
  ng "AC-12" "tech.md 缺性能预算待基准声明"
fi

# AC-13 tech.md GoReleaser：5 个平台 + windows+arm64 标注
if grep -qE '5 个平台' "$TECH" && grep -q 'windows+arm64' "$TECH"; then
  ok "AC-13 tech.md GoReleaser 平台数改 5 个"
else
  ng "AC-13" "tech.md 缺 5 个平台或 windows+arm64 标注"
fi

# AC-14 tech.md fsnotify fallback
if grep -qE 'fallback|降级' "$TECH" && grep -qE '2s' "$TECH"; then
  ok "AC-14 tech.md 声明 fsnotify fallback"
else
  ng "AC-14" "tech.md 缺 fsnotify fallback 或 2s 声明"
fi

# AC-15 structure.md β 方案
ac15_ok=1
grep -q 'cobra 命令文件' "$STRUCTURE" || ac15_ok=0
grep -q '~22 个' "$STRUCTURE" || ac15_ok=0
for f in ears.go bugfix.go requirements.go tasks.go linter.go baseline.go quality.go; do
  grep -q "$f" "$STRUCTURE" || ac15_ok=0
done
grep -qE 'Vite|FSD' "$STRUCTURE" || ac15_ok=0
grep -qE 'src/.*dist/.*scripts/' "$STRUCTURE" || ac15_ok=0
# 6 个根目录文件：BSD grep \b 对 / 不友好，用不带 \b 的等价模式
root_count=$(grep -cE 'LICENSE|CONTRIBUTING\.md|Makefile|AGENTS\.md|CLAUDE\.md|examples/' "$STRUCTURE")
if [[ $root_count -lt 5 ]]; then ac15_ok=0; fi
if [[ $ac15_ok -eq 1 ]]; then
  ok "AC-15 structure.md β 方案齐全（root_files_lines=$root_count）"
else
  ng "AC-15" "structure.md β 方案要素不全（root_files_lines=$root_count）"
fi

echo "---"
echo "RESULT: PASS=$pass FAIL=$fail TOTAL=$((pass + fail))"
[[ $fail -eq 0 ]] && exit 0 || exit 1
