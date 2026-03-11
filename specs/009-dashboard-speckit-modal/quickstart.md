# Quickstart: Dashboard SpecKit Modal

**Branch**: `009-dashboard-speckit-modal` | **Date**: 2026-03-11

## Implementation Order

### Step 1: Layout Changes (standalone, no dependencies)

1. Edit `internal/tui/layout.go` — change Specs panel ratio from 40% to 55%
2. Edit `internal/tui/panels/specs.go` — add 1-char inner padding to `View()`
3. Update `internal/tui/layout_test.go` — adjust expected ratios
4. Update `internal/tui/panels/specs_test.go` — verify padding in rendered output
5. Run `go test ./internal/tui/...` — confirm passing

### Step 2: Modal Component (no subprocess dependency)

1. Create `internal/tui/modal.go` — SpecKitModal struct, Update(), View()
2. Create `internal/tui/modal_test.go` — test navigation, wrapping, esc/enter, rendering
3. Add message types to `internal/tui/msg.go` — SpecKitActionMsg
4. Wire into `internal/tui/app.go`:
   - Add `modal SpecKitModal` field
   - Add `S` to `GlobalKeyBindings` in `keymap.go`
   - Handle `S` in `handleKey()` — open modal with selected spec
   - Handle modal key interception when `modal.visible`
   - Render modal overlay in `View()`
5. Run `go test ./internal/tui/...` — confirm passing

### Step 3: SpecKit Runner (subprocess integration)

1. Create `internal/tui/speckit_runner.go` — subprocess launch, event bridging
2. Create `internal/tui/speckit_runner_test.go` — test with fake subprocess
3. Add remaining message types to `msg.go` — SpecKitOutputMsg, SpecKitDoneMsg
4. Wire into `app.go`:
   - Handle `SpecKitActionMsg` → create runner, start subprocess
   - Handle `SpecKitOutputMsg` → append to mainView.outputLog
   - Handle `SpecKitDoneMsg` → clean up runner, update state
5. Add `StateSpecKit` to `focus.go` LoopState
6. Run `go test ./internal/tui/...` — confirm passing

### Step 4: Interactive Clarify (builds on Step 3)

1. Add stdin pipe to runner for clarify action
2. Add question detection logic (pattern matching on output)
3. Add `SpecKitInputRequestMsg` / `SpecKitInputResponseMsg` to `msg.go`
4. Modify `internal/tui/panels/main_view.go` — add input prompt mode
5. Wire input flow in `app.go`:
   - Handle `SpecKitInputRequestMsg` → switch MainView to input mode
   - Handle `SpecKitInputResponseMsg` → forward to runner stdin
6. Test interactive flow with fake subprocess
7. Run `go test ./internal/tui/...` — confirm passing

## Verification

```sh
go build ./cmd/ralph/         # builds
go test ./...                 # all tests pass
go vet ./...                  # zero warnings
```

Manual test:
1. `ralph` — launch dashboard
2. Select a spec in Specs panel
3. Press `S` — modal appears with spec name
4. Navigate with `j`/`k`, press `enter` on "Plan"
5. Verify output streams to Output tab
6. Verify header shows `SPECKIT:PLAN`
