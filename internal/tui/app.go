package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/loop"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/spec"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/store"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/tui/panels"
)

// Model is the root bubbletea model for the multi-panel Ralph TUI.
type Model struct {
	// Event source
	events      <-chan loop.LogEntry
	storeReader store.Reader

	// Sub-panels
	specsPanel      panels.SpecsPanel
	iterationsPanel panels.IterationsPanel
	mainView        panels.MainView
	secondary       panels.SecondaryPanel

	// Layout and focus
	layout Layout
	focus  FocusTarget
	theme  Theme
	width  int
	height int

	// Loop state
	loopState  LoopState
	iteration  int
	maxIter    int
	mode       string
	branch     string
	totalCost  float64
	lastCommit string

	// Time
	startedAt time.Time
	now       time.Time

	// Identity
	projectName string
	workDir     string

	// Graceful stop
	requestStop   func()
	stopRequested bool

	// Loop control (nil when launched from ralph build/plan/run)
	controller LoopController

	// Error/done
	err  error
	done bool
}

// New creates the multi-panel TUI Model.
// storeReader may be nil if no session log is available.
// specFiles is the initial list of specs for the sidebar; nil is allowed.
// requestStop, if non-nil, is called once when the user presses 's'.
func New(events <-chan loop.LogEntry, storeReader store.Reader, accentColor, projectName, workDir string, specFiles []spec.SpecFile, requestStop func(), controller LoopController) Model {
	now := time.Now()
	th := NewTheme(accentColor)
	layout := Calculate(80, 24)

	specsW, specsH := innerDims(layout.Specs)
	itersW, itersH := innerDims(layout.Iterations)
	mainW, mainH := innerDims(layout.Main)
	secW, secH := innerDims(layout.Secondary)

	return Model{
		events:          events,
		storeReader:     storeReader,
		specsPanel:      panels.NewSpecsPanel(specFiles, specsW, specsH),
		iterationsPanel: panels.NewIterationsPanel(itersW, itersH),
		mainView:        panels.NewMainView(mainW, mainH),
		secondary:       panels.NewSecondaryPanel(secW, secH),
		layout:          layout,
		focus:           FocusMain,
		theme:           th,
		width:           80,
		height:          24,
		loopState:       StateIdle,
		startedAt:       now,
		now:             now,
		projectName:     projectName,
		workDir:         workDir,
		requestStop:     requestStop,
		controller:      controller,
	}
}

// Err returns any error recorded from the loop.
func (m Model) Err() error { return m.err }

// Init returns the initial commands: event listener + clock ticker.
func (m Model) Init() tea.Cmd {
	return tea.Batch(waitForEvent(m.events), tickCmd())
}

// tickCmd schedules the next one-second clock tick.
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// waitForEvent blocks on the event channel and returns the next message.
func waitForEvent(ch <-chan loop.LogEntry) tea.Cmd {
	return func() tea.Msg {
		entry, ok := <-ch
		if !ok {
			return loopDoneMsg{}
		}
		return logEntryMsg(entry)
	}
}

// Update handles all incoming bubbletea messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)
	case tea.KeyMsg:
		return m.handleKey(msg)
	case logEntryMsg:
		return m.handleLogEntry(msg)
	case tickMsg:
		m.now = time.Time(msg)
		return m, tickCmd()
	case loopDoneMsg:
		// Channel closed — loop finished. Transition to idle but keep TUI open.
		// In dashboard mode the channel is never closed; for ralph build/plan/run
		// this fires once when the single loop finishes. User presses q to exit.
		if m.loopState.CanTransitionTo(StateIdle) {
			m.loopState = StateIdle
		}
		return m, nil
	case loopErrMsg:
		m.err = msg.err
		m.done = true
		return m, tea.Quit
	case panels.SpecSelectedMsg:
		return m.handleSpecSelected(msg)
	case panels.EditSpecRequestMsg:
		return m.handleEditSpecRequest(msg)
	case panels.CreateSpecRequestMsg:
		return m.handleCreateSpecRequest(msg)
	case specsRefreshedMsg:
		return m.handleSpecsRefreshed(msg)
	case panels.IterationSelectedMsg:
		return m.handleIterationSelected(msg)
	case iterationLogLoadedMsg:
		return m.handleIterationLogLoaded(msg)
	}
	return m.delegateToFocused(msg)
}

func (m Model) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	m.layout = Calculate(msg.Width, msg.Height)
	if !m.layout.TooSmall {
		specsW, specsH := innerDims(m.layout.Specs)
		itersW, itersH := innerDims(m.layout.Iterations)
		mainW, mainH := innerDims(m.layout.Main)
		secW, secH := innerDims(m.layout.Secondary)
		m.specsPanel = m.specsPanel.SetSize(specsW, specsH)
		m.iterationsPanel = m.iterationsPanel.SetSize(itersW, itersH)
		m.mainView = m.mainView.SetSize(mainW, mainH)
		m.secondary = m.secondary.SetSize(secW, secH)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "s":
		if m.requestStop != nil && !m.stopRequested {
			m.stopRequested = true
			m.requestStop()
		}
		return m, nil
	case "b":
		if m.controller != nil && !m.controller.IsRunning() {
			if m.loopState.CanTransitionTo(StateBuilding) {
				m.loopState = StateBuilding
			}
			m.controller.StartLoop("build")
		}
		return m, nil
	case "p":
		if m.controller != nil && !m.controller.IsRunning() {
			if m.loopState.CanTransitionTo(StatePlanning) {
				m.loopState = StatePlanning
			}
			m.controller.StartLoop("plan")
		}
		return m, nil
	case "R":
		if m.controller != nil && !m.controller.IsRunning() {
			if m.loopState.CanTransitionTo(StateBuilding) {
				m.loopState = StateBuilding
			}
			m.controller.StartLoop("smart")
		}
		return m, nil
	case "x":
		if m.controller != nil {
			m.controller.StopLoop()
		}
		return m, nil
	case "tab":
		m.focus = m.focus.Next()
		return m, nil
	case "shift+tab":
		m.focus = m.focus.Prev()
		return m, nil
	case "1":
		m.focus = FocusSpecs
		return m, nil
	case "2":
		m.focus = FocusIterations
		return m, nil
	case "3":
		m.focus = FocusMain
		return m, nil
	case "4":
		m.focus = FocusSecondary
		return m, nil
	}
	return m.delegateToFocused(msg)
}

func (m Model) delegateToFocused(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.focus {
	case FocusSpecs:
		m.specsPanel, cmd = m.specsPanel.Update(msg)
	case FocusIterations:
		m.iterationsPanel, cmd = m.iterationsPanel.Update(msg)
	case FocusMain:
		m.mainView, cmd = m.mainView.Update(msg)
	case FocusSecondary:
		m.secondary, cmd = m.secondary.Update(msg)
	}
	return m, cmd
}

func (m Model) handleLogEntry(msg logEntryMsg) (tea.Model, tea.Cmd) {
	entry := loop.LogEntry(msg)

	// Update loop metadata from entry
	if entry.Branch != "" {
		m.branch = entry.Branch
	}
	if entry.Mode != "" {
		m.mode = entry.Mode
	}
	if entry.MaxIter > 0 {
		m.maxIter = entry.MaxIter
	}
	if entry.Iteration > 0 {
		m.iteration = entry.Iteration
	}
	if entry.TotalCost > 0 {
		m.totalCost = entry.TotalCost
	}
	if entry.Commit != "" {
		m.lastCommit = entry.Commit
	}

	// Derive LoopState transitions from log kind
	switch entry.Kind {
	case loop.LogIterStart:
		next := StateBuilding
		if entry.Mode == "plan" {
			next = StatePlanning
		}
		if m.loopState.CanTransitionTo(next) {
			m.loopState = next
		}
		m.iterationsPanel = m.iterationsPanel.SetCurrent(entry.Iteration)

	case loop.LogIterComplete:
		summary := store.IterationSummary{
			Number:   entry.Iteration,
			Mode:     entry.Mode,
			CostUSD:  entry.CostUSD,
			Duration: entry.Duration,
			Subtype:  entry.Subtype,
			Commit:   entry.Commit,
		}
		m.iterationsPanel = m.iterationsPanel.AddIteration(summary).SetCurrent(0)

	case loop.LogDone, loop.LogStopped:
		if m.loopState.CanTransitionTo(StateIdle) {
			m.loopState = StateIdle
		}

	case loop.LogError:
		if m.loopState.CanTransitionTo(StateFailed) {
			m.loopState = StateFailed
		}

	case loop.LogRegent:
		if m.loopState.CanTransitionTo(StateRegentRestart) {
			m.loopState = StateRegentRestart
		}
	}

	// Render once at current width; route by kind
	rendered := m.theme.RenderLogLine(entry, m.layout.Main.Width)
	switch entry.Kind {
	case loop.LogRegent:
		m.secondary = m.secondary.AppendLine(rendered, panels.TabRegent)
	case loop.LogGitPull, loop.LogGitPush:
		m.secondary = m.secondary.AppendLine(rendered, panels.TabGit)
		m.mainView = m.mainView.AppendLine(rendered)
	default:
		m.mainView = m.mainView.AppendLine(rendered)
	}

	return m, waitForEvent(m.events)
}

func (m Model) handleSpecSelected(msg panels.SpecSelectedMsg) (tea.Model, tea.Cmd) {
	content := m.readSpecContent(msg.Spec)
	m.mainView = m.mainView.ShowSpec(content)
	return m, nil
}

func (m Model) handleEditSpecRequest(msg panels.EditSpecRequestMsg) (tea.Model, tea.Cmd) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return m, nil
	}
	path := msg.Path
	if !filepath.IsAbs(path) && m.workDir != "" {
		path = filepath.Join(m.workDir, path)
	}
	parts := strings.Fields(editor)
	args := append(parts[1:], path)
	cmd := exec.Command(parts[0], args...) //nolint:gosec
	workDir := m.workDir
	return m, tea.ExecProcess(cmd, func(_ error) tea.Msg {
		specs, _ := spec.List(workDir)
		return specsRefreshedMsg{Specs: specs}
	})
}

func (m Model) handleCreateSpecRequest(msg panels.CreateSpecRequestMsg) (tea.Model, tea.Cmd) {
	workDir := m.workDir
	name := msg.Name
	return m, func() tea.Msg {
		_, _ = spec.New(workDir, name)
		specs, _ := spec.List(workDir)
		return specsRefreshedMsg{Specs: specs}
	}
}

func (m Model) handleSpecsRefreshed(msg specsRefreshedMsg) (tea.Model, tea.Cmd) {
	specsW, specsH := innerDims(m.layout.Specs)
	m.specsPanel = panels.NewSpecsPanel(msg.Specs, specsW, specsH)
	return m, nil
}

func (m Model) readSpecContent(sf spec.SpecFile) string {
	path := sf.Path
	if !filepath.IsAbs(path) && m.workDir != "" {
		path = filepath.Join(m.workDir, path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("(cannot read %s: %v)", sf.Path, err)
	}
	return string(data)
}

func (m Model) handleIterationSelected(msg panels.IterationSelectedMsg) (tea.Model, tea.Cmd) {
	if m.storeReader == nil {
		return m, nil
	}
	n := msg.Number
	return m, func() tea.Msg {
		entries, err := m.storeReader.IterationLog(n)
		var summary store.IterationSummary
		if summaries, sErr := m.storeReader.Iterations(); sErr == nil {
			for _, s := range summaries {
				if s.Number == n {
					summary = s
					break
				}
			}
		}
		return iterationLogLoadedMsg{Number: n, Entries: entries, Summary: summary, Err: err}
	}
}

func (m Model) handleIterationLogLoaded(msg iterationLogLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		return m, nil
	}
	rendered := make([]string, len(msg.Entries))
	for i, e := range msg.Entries {
		rendered[i] = m.theme.RenderLogLine(e, m.layout.Main.Width)
	}
	m.mainView = m.mainView.ShowIterationLog(rendered)
	m.mainView = m.mainView.SetIterationSummary(renderIterationSummary(msg.Summary))
	return m, nil
}

// renderIterationSummary formats an IterationSummary as key-value lines for the Summary tab.
func renderIterationSummary(s store.IterationSummary) []string {
	lines := []string{
		fmt.Sprintf("%-12s %d", "Iteration:", s.Number),
		fmt.Sprintf("%-12s %s", "Mode:", s.Mode),
		fmt.Sprintf("%-12s $%.4f", "Cost:", s.CostUSD),
		fmt.Sprintf("%-12s %.1fs", "Duration:", s.Duration),
	}
	if s.Subtype != "" {
		lines = append(lines, fmt.Sprintf("%-12s %s", "Exit:", s.Subtype))
	}
	if s.Commit != "" {
		lines = append(lines, fmt.Sprintf("%-12s %s", "Commit:", s.Commit))
	}
	return lines
}

// View renders the full multi-panel TUI.
func (m Model) View() string {
	if m.layout.TooSmall {
		msg := fmt.Sprintf("Terminal too small (%dx%d).\nPlease resize to at least 80x24.", m.width, m.height)
		return lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Center).
			Render(msg)
	}

	header := panels.RenderHeader(panels.HeaderProps{
		ProjectName: m.projectName,
		WorkDir:     m.workDir,
		Branch:      m.branch,
		Mode:        m.mode,
		Iteration:   m.iteration,
		MaxIter:     m.maxIter,
		TotalCost:   m.totalCost,
		StateSymbol: m.loopState.Symbol(),
		StateLabel:  m.loopState.Label(),
		Elapsed:     m.now.Sub(m.startedAt),
		Clock:       m.now,
	}, m.layout.Header.Width, m.theme.AccentHeaderStyle())

	footer := panels.RenderFooter(panels.FooterProps{
		Focus:         m.focus.String(),
		LastCommit:    m.lastCommit,
		StopRequested: m.stopRequested,
		StateLabel:    m.loopState.Label(),
	}, m.layout.Footer.Width)

	// Left sidebar: specs (top) + iterations (bottom)
	specsW, specsH := innerDims(m.layout.Specs)
	itersW, itersH := innerDims(m.layout.Iterations)
	mainW, mainH := innerDims(m.layout.Main)
	secW, secH := innerDims(m.layout.Secondary)

	sidebar := lipgloss.JoinVertical(lipgloss.Left,
		m.theme.PanelBorderStyle(m.focus == FocusSpecs).
			Width(specsW).Height(specsH).
			Render(m.specsPanel.View()),
		m.theme.PanelBorderStyle(m.focus == FocusIterations).
			Width(itersW).Height(itersH).
			Render(m.iterationsPanel.View()),
	)

	rightCol := lipgloss.JoinVertical(lipgloss.Left,
		m.theme.PanelBorderStyle(m.focus == FocusMain).
			Width(mainW).Height(mainH).
			Render(m.mainView.View()),
		m.theme.PanelBorderStyle(m.focus == FocusSecondary).
			Width(secW).Height(secH).
			Render(m.secondary.View()),
	)

	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, rightCol)
	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

// innerDims returns the content dimensions for a panel rect accounting for
// the 1-character border on each side (2 total per dimension).
func innerDims(r Rect) (w, h int) {
	w = r.Width - 2
	if w < 1 {
		w = 1
	}
	h = r.Height - 2
	if h < 1 {
		h = 1
	}
	return
}
