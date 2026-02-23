#!/bin/bash
# RalphKing Bootstrap Loop — feeds PROMPT_build.md to claude until Ralph can run himself
# Usage:
#   ./loop.sh             # unlimited iterations
#   ./loop.sh 10          # max 10 iterations

set -euo pipefail

export PATH=~/.npm-global/bin:$PATH

MAX_ITERATIONS=${1:-0}
ITERATION=0
CURRENT_BRANCH=$(git branch --show-current)
PROMPT_FILE="PROMPT_build.md"

if [ ! -f "$PROMPT_FILE" ]; then
    echo "Error: $PROMPT_FILE not found"
    exit 1
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "👑 RalphKing Bootstrap Loop"
echo "Branch: $CURRENT_BRANCH"
[ "$MAX_ITERATIONS" -gt 0 ] && echo "Max:    $MAX_ITERATIONS iterations"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

while true; do
    if [ "$MAX_ITERATIONS" -gt 0 ] && [ "$ITERATION" -ge "$MAX_ITERATIONS" ]; then
        echo "Reached max iterations: $MAX_ITERATIONS"
        break
    fi

    # Pull at start of each iteration — Ralph always works on latest code
    echo "⬇️  Pulling latest changes..."
    STASHED=0
    if ! git diff --quiet || ! git diff --cached --quiet; then
        echo "⚠️  Uncommitted changes — stashing before pull..."
        git stash push -m "loop-pre-pull-stash" && STASHED=1
    fi
    git pull --rebase origin "$CURRENT_BRANCH" || {
        echo "⚠️  Rebase conflict — falling back to merge..."
        git rebase --abort 2>/dev/null || true
        git pull --no-rebase origin "$CURRENT_BRANCH"
    }
    if [ "$STASHED" -eq 1 ]; then
        echo "📦 Restoring stashed changes..."
        git stash pop || echo "⚠️  Stash pop failed — check git stash list"
    fi

    # Run Claude
    cat "$PROMPT_FILE" | claude -p \
        --dangerously-skip-permissions \
        --output-format=stream-json \
        --verbose 2>&1 | python3 -c "
import sys, json
for line in sys.stdin:
    line = line.strip()
    if not line:
        continue
    try:
        obj = json.loads(line)
        t = obj.get('type', '')
        if t == 'assistant' and 'message' in obj:
            for block in obj['message'].get('content', []):
                if block.get('type') == 'text' and block.get('text','').strip():
                    print(block['text'].strip())
                elif block.get('type') == 'tool_use':
                    name = block.get('name','')
                    inp = block.get('input', {})
                    if name == 'write_file':
                        print(f'  ✏️  write: {inp.get(\"path\",\"\")}')
                    elif name == 'bash':
                        cmd = inp.get('command','')[:80]
                        print(f'  🔧 bash: {cmd}')
                    elif name in ('read_file','list_directory'):
                        print(f'  📖 {name}: {inp.get(\"path\",inp.get(\"directory\",\"\"))}')
                    elif name:
                        print(f'  → {name}')
        elif t == 'result':
            print(f'✅ Done — cost: \${obj.get(\"cost_usd\", 0):.4f}')
        elif t == 'system' and obj.get('subtype') == 'error':
            print(f'❌ Error: {obj.get(\"error\",\"\")}')
    except json.JSONDecodeError:
        print(line)
"

    # Push only if there's something new
    if git diff --quiet origin/"$CURRENT_BRANCH" HEAD 2>/dev/null; then
        echo "ℹ️  Nothing new to push."
    else
        echo "⬆️  Pushing to $CURRENT_BRANCH..."
        git push origin "$CURRENT_BRANCH" || git push -u origin "$CURRENT_BRANCH"
    fi

    ITERATION=$((ITERATION + 1))
    echo -e "\n\n======================== LOOP $ITERATION ========================\n"
done
