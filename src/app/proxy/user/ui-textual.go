package user

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/pixa/src/app/proxy/common"
)

type textualUI struct {
	interaction
	m  *model
	po *common.PresentationOptions
}

// Decorate allows the interaction to provide a wrapper around the callback.
// If the interaction does not need it, then it just returns the target. Only
// the Principal callback is decorated.
func (ui *textualUI) Decorate(target *nav.LabelledTraverseCallback) *nav.LabelledTraverseCallback {
	return &nav.LabelledTraverseCallback{
		Label: "💝💝 Principal Textual Shrink Callback",
		Fn: func(item *nav.TraverseItem) error {
			return target.Fn(item)
		},
	}
}

// Performs the full traversal which consists of a discovery phase followed
// by the principal phase.
func (ui *textualUI) Traverse(di common.DriverTraverseInfo,
) (*nav.TraverseResult, error) {
	ui.m = &model{
		inputs:     ui.inputs,
		executable: ui.inputs.Root.Configs.Advanced.Executable().Symbol(),
		status:     "🔎 discovering ...",
		arity:      ui.arity,
		latest: JobDescription{
			Source:      "waiting ...",
			Destination: "waiting ...",
		},
		di:      di,
		delay:   ui.inputs.Root.Configs.Interaction.TuiConfig().PerItemDelay(),
		spinner: spinner.New(),
	}

	const (
		hotPink = "201"
	)

	ui.m.spinner.Spinner = spinner.Dot
	ui.m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(hotPink))

	ui.m.attach(ui)

	options := []tea.ProgramOption{}
	if ui.po.WithoutRenderer {
		options = []tea.ProgramOption{tea.WithoutRenderer()}
	}

	ui.m.program = tea.NewProgram(ui.m, options...)

	if _, err := ui.m.program.Run(); err != nil {
		ui.logger.Error(fmt.Sprintf("could not start: '%v'", err)) // make this i18n error

		return nil, err
	}

	return ui.m.result, ui.m.err
}

// Tick allows the model to be updated, as activity occurs during
// the traversal.
func (ui *textualUI) Tick(msg *common.ProgressMsg) {
	if ui.m.delay > 0 {
		time.Sleep(ui.m.delay)
	}

	ui.m.program.Send(msg)
}
