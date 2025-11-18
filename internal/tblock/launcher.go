package tblock

import (
	"bytes"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/havrydotdev/tblock-launcher/pkg/launcher"
	"github.com/havrydotdev/tblock-launcher/pkg/mc"
)

type LauncherState int

const (
	Downloading LauncherState = iota
	StartedClient
	ClientNotInstalled
	Idle
)

type Launcher struct {
	Config *launcher.Config

	state      LauncherState
	core       *launcher.GameLauncher
	background []byte

	a          fyne.App
	w          fyne.Window
	logs       *widget.Label
	mainButton *widget.Button
}

func NewLauncher(cfg *launcher.Config, background, options, servers []byte) *Launcher {
	logs := widget.NewLabel("Your logs will be there.")

	labelWriter := &LabelWriter{logs}
	vm := mc.NewVersionManager(cfg.GameDir, labelWriter, options, servers)
	core := launcher.NewGameLauncher(cfg, vm)

	a := app.NewWithID("wtf.tblock.launcher")
	w := a.NewWindow("tblock")

	state := ClientNotInstalled
	if core.IsInstalled() {
		state = Idle
	}

	return &Launcher{
		logs: logs, a: a, w: w,
		core: core, background: background,
		state: state, Config: cfg,
	}
}

func (l *Launcher) Run() {
	l.buildUI()
	l.w.Resize(fyne.NewSize(640, 480))
	l.w.SetFixedSize(true)
	l.w.ShowAndRun()
}

func (l *Launcher) buildUI() {
	l.mainButton = l.buildMainButton(l.state)

	usernameInput := widget.NewEntry()
	usernameInput.OnChanged = func(data string) {
		l.Config.Username = data
	}
	usernameInput.Text = l.Config.Username

	bottom := container.NewHSplit(usernameInput, l.mainButton)
	bottom.Offset = 0.66

	btns := container.NewBorder(
		widget.NewLabel("TBlockMC Launcher"), bottom, l.logs, nil,
	)

	l.w.SetContent(container.New(layout.NewStackLayout(),
		canvas.NewImageFromReader(bytes.NewReader(l.background), "background.png"), btns))
}

func (l *Launcher) setState(state LauncherState) {
	l.state = state
	newBtn := l.buildMainButton(l.state)
	l.mainButton.OnTapped = newBtn.OnTapped
	l.mainButton.SetText(newBtn.Text)
}

func (l *Launcher) buildMainButton(state LauncherState) *widget.Button {
	switch state {
	case StartedClient:
		return widget.NewButton("Client started", func() {})
	case Downloading:
		return widget.NewButton("Downloading...", func() {})
	case ClientNotInstalled:
		return widget.NewButton("Download", func() {
			go func() {
				fyne.Do(func() { l.setState(Downloading) })
				if err := l.core.Install(); err != nil {
					l.logs.Text = err.Error()
				}
				fyne.Do(func() { l.setState(Idle) })
			}()
		})
	default:
		return widget.NewButton("Play", func() {
			go func() {
				fyne.Do(func() {
					l.setState(StartedClient)
					l.w.Hide()
				})

				if err := l.core.Launch(l.Config.Username); err != nil {
					l.logs.Text = err.Error()
				}

				fyne.Do(func() {
					l.setState(Idle)
					l.w.Show()
				})
			}()
		})
	}
}
