package user

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"
	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/pixa/src/app/proxy/common"
)

type actionType string

type JobDescription struct {
	Scheme      string
	Profile     string
	Source      string
	Destination string
	emoji       string
	err         error
}

type Summary struct {
	Measurement string
	Err         error
}

type model struct {
	inputs     *common.ShrinkCommandInputs
	executable string
	status     string
	arity      uint
	level      int32
	latest     JobDescription
	di         common.DriverTraverseInfo
	ui         walker
	delay      time.Duration
	spinner    spinner.Model
	result     *nav.TraverseResult
	err        error
	program    *tea.Program
}

func (m *model) attach(ui walker) {
	// The reason why we have attach and detach methods is the model
	// and the ui both hold a reference to each which would mean that
	// they could never be garbage collected resulting in a memory
	// leak. However, pixa is a short lived program and the memory leak
	// is of no significance. The attach/detach process, prevents this
	// potential of a leak and acknowledges this cyclic relationship.
	//
	m.ui = ui
}

func (m *model) detach() {
	m.ui = nil
}

func discover(ti common.ClientTraverseInfo, ui walker) tea.Cmd {
	return func() tea.Msg {
		var (
			result *nav.TraverseResult
			err    error
		)

		result, err = ui.navigate(ti) // peek

		return &common.DiscoveredMsg{
			Result: result,
			Err:    err,
		}
	}
}

func principal(ti common.ClientTraverseInfo, ui walker) tea.Cmd {
	return func() tea.Msg {
		var (
			result *nav.TraverseResult
			err    error
		)

		result, err = ui.navigate(ti) // poke

		return &common.FinishedMsg{
			Result: result,
			Err:    err,
		}
	}
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		discover(m.di, m.ui),
	)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case *common.DiscoveredMsg:
		m.result = msg.Result
		m.err = msg.Err
		m.status = "🎭 discovered"

		if msg.Err != nil {
			return m, tea.Quit
		}

		m.di.Next()

		return m, principal(m.di, m.ui)

	case *common.ProgressMsg:
		atomic.AddInt32(&m.level, 1)
		m.latest.Source = msg.Source
		m.latest.Destination = msg.Destination
		m.latest.Scheme = msg.Scheme
		m.latest.Profile = msg.Profile
		m.latest.emoji = randemoji()
		m.latest.err = msg.Err
		m.status = "🚀 progressing"

	case *common.FinishedMsg:
		m.result = msg.Result
		m.err = msg.Err
		m.status = summary(msg.Result, msg.Err)
		m.detach()

		return m, tea.Quit

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)

		return m, cmd

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	}

	return m, nil
}

type bodyContent struct {
	source      string
	destination string
	emoji       string
}

func (bc *bodyContent) view() string {
	from, source := filepath.Split(bc.source)
	enriched := fmt.Sprintf("%v %v", bc.emoji, source)
	to, destination := filepath.Split(bc.destination)

	content := fmt.Sprintf(
		`		-->      source: %v
		-->        from: %v
		--> destination: %v
		-->          to: %v`,
		enriched,
		from,
		destination,
		to,
	)

	return content
}

func (m *model) View() string {
	executable := fmt.Sprintf("%v %v", common.Definitions.Pixa.Emoji, m.executable)
	scheme := lo.Ternary(m.latest.Scheme != "", m.latest.Scheme, "[NONE]")
	profile := lo.Ternary(m.latest.Profile != "", m.latest.Profile, "[NONE]")
	executing := lo.TernaryF(m.di.IsDryRun(),
		func() string {
			return "⛔ dry with"
		},
		func() string {
			return "🌀 go using"
		},
	)
	wpf := m.inputs.Root.WorkerPoolFam.Native
	cpus := lo.TernaryF(wpf.NoWorkers > 1 || wpf.CPU,
		func() string {
			if wpf.CPU {
				return fmt.Sprintf("NumCPUs %v", runtime.NumCPU())
			}
			return fmt.Sprintf("%v workers", fmt.Sprintf("%v", wpf.NoWorkers))
		},
		func() string {
			return "single CPU"
		},
	)

	action := fmt.Sprintf("%v %v", executing, cpus)
	info := fmt.Sprintf("action: '%v', scheme: '%v', profile: '%v'", action, scheme, profile)
	bc := bodyContent{
		source:      m.latest.Source,
		destination: m.latest.Destination,
		emoji:       m.latest.emoji,
	}

	e := lo.Ternary(m.latest.err != nil,
		fmt.Sprintf("💥 %v", m.latest.err),
		"💫 ok",
	)
	latestView := fmt.Sprintf(
		`
	- %v status(%v): %v
		-->        info: %v
%v
		-->       error: %v
		-->    progress: %v
	`,
		m.spinner.View(), executable, m.status,
		info,
		bc.view(),
		e,
		m.level,
	)

	// the view will be a series of 'lanes', each one representing a job
	// currently in progress. There will be an overall progress bar, that
	// will tick every time a lane reports a completion. Each lane will
	// record the time each job takes, so that a batch completion time
	// can be estimated based upon the current moving average.
	// We can also represent each lane with an emoji, just for the fun of it
	//
	return latestView
}
