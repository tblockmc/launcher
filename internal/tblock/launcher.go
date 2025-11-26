package tblock

import (
	"bytes"
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/havrydotdev/tblock-launcher/internal/discord"
	"github.com/havrydotdev/tblock-launcher/internal/static"
	"github.com/havrydotdev/tblock-launcher/pkg/downloader"
	"github.com/havrydotdev/tblock-launcher/pkg/launcher"
	"github.com/havrydotdev/tblock-launcher/pkg/utils"
	"github.com/mouuff/go-rocket-update/pkg/provider"
	"github.com/mouuff/go-rocket-update/pkg/updater"
)

type LauncherState int

const (
	Downloading LauncherState = iota
	StartedClient
	CanUpdate
	ClientNotInstalled
	Idle
)

var (
	statics = []downloader.StaticAsset{
		{Path: "options.txt", Data: static.OptionsTXT},
		{Path: "servers.dat", Data: static.ServersDAT},
	}

	mods = []downloader.ModData{
		{Name: "fabric-api", URL: "https://cdn.modrinth.com/data/P7dR8mSH/versions/g58ofrov/fabric-api-0.136.1%2B1.21.8.jar"},
		{Name: "sodium", URL: "https://cdn.modrinth.com/data/AANobbMI/versions/7pwil2dy/sodium-fabric-0.7.3%2Bmc1.21.8.jar"},
		{Name: "simple-voice-chat", URL: "https://cdn.modrinth.com/data/9eGKb6K1/versions/2Z1g1v36/voicechat-fabric-1.21.8-2.6.6.jar"},
		{Name: "modmenu", URL: "https://cdn.modrinth.com/data/mOgUt4GM/versions/am1Siv7F/modmenu-15.0.0.jar"},
		{Name: "placeholder-api", URL: "https://cdn.modrinth.com/data/eXts2L7r/versions/1S1kjZ9W/placeholder-api-2.7.2%2B1.21.8.jar"},
		{Name: "zoomify", URL: "https://cdn.modrinth.com/data/w7ThoJFB/versions/qMqviL3t/Zoomify-2.14.6%2B1.21.6.jar"},
		{Name: "iris-shaders", URL: "https://cdn.modrinth.com/data/YL57xq9U/versions/Rhzf61g1/iris-fabric-1.9.6%2Bmc1.21.8.jar"},
		{Name: "emotecraft", URL: "https://cdn.modrinth.com/data/pZ2wrerK/versions/VeMVR6lp/emotecraft-fabric-for-MC1.21.7-3.0.0-b.build.127.jar"},
		{Name: "player-animation-lib", URL: "https://cdn.modrinth.com/data/ha1mEyJS/versions/xbjrgVCf/PlayerAnimationLibFabric-1.0.13%2Bmc.1.21.8.jar"},
		{Name: "bendable-cuboids", URL: "https://cdn.modrinth.com/data/OI3FlFon/versions/mqKPHO6f/BendableCuboids-1.0.5%2Bmc1.21.7.jar"},
		{Name: "yet-another-configlib", URL: "https://cdn.modrinth.com/data/1eAoo2KR/versions/WxYlHLu6/yet_another_config_lib_v3-3.7.1%2B1.21.6-fabric.jar"},
		{Name: "fabric-language-kotlin", URL: "https://cdn.modrinth.com/data/Ha28R6CL/versions/LcgnDDmT/fabric-language-kotlin-1.13.7%2Bkotlin.2.2.21.jar"},
	}

	mainBtnTexts = map[LauncherState]string{
		Idle:               "Play!",
		ClientNotInstalled: "Download",
		StartedClient:      "Running...",
		Downloading:        "Downloading...",
		CanUpdate:          "Update",
	}
)

type Launcher struct {
	Config  *launcher.Config
	version string
	core    *launcher.FabricLauncher
	state   LauncherState

	w          fyne.Window
	a          fyne.App
	u          *updater.Updater
	mainButton *widget.Button
	settings   *dialog.CustomDialog
	statusText binding.String
}

func NewLauncher(cfg *launcher.Config) (*Launcher, error) {
	err := discord.Login()
	if err != nil {
		log.Println("discord login failed: ", err)
	}

	core := launcher.NewFabricLauncher(cfg)

	a := app.NewWithID("com.github.tblockmc.launcher")
	uk, err := static.Translations.ReadFile("translations/uk.json")
	if err != nil {
		return nil, err
	}

	a.Settings().SetTheme(newTheme())

	// hardcode ukrainian for now
	err = lang.AddTranslationsForLocale(uk, lang.SystemLocale())
	if err != nil {
		return nil, err
	}

	version := a.Metadata().Version
	if version == "" {
		version = "dev-build"
	}

	u := &updater.Updater{
		Provider: &provider.Github{
			RepositoryURL: "github.com/tblockmc/launcher",
			ArchiveName:   getReleaseArchive(),
		},
		ExecutableName: getBinaryName(),
		Version:        fmt.Sprintf("v%s", version),
	}

	w := a.NewWindow(fmt.Sprintf("%s %s", lang.L("TBlockMC"), version))

	state := Idle
	if !core.IsFabricInstalled() {
		state = ClientNotInstalled
	}

	canUpdate, err := u.CanUpdate()
	if err != nil {
		w.SetTitle(err.Error())
	}

	if canUpdate {
		state = CanUpdate
	}

	mainBtnText := binding.NewString()
	mainBtnText.Set(lang.L(mainBtnTexts[state]))

	statusText := binding.NewString()

	return &Launcher{
		state: state, w: w, u: u, a: a,
		core: core, Config: cfg,
		statusText: statusText,
		version:    version,
	}, nil
}

func (l *Launcher) Run() {
	err := discord.SetIdleActivity()
	if err != nil {
		log.Println("failed to set discord idle activity: ", err)
	}

	l.buildUI()

	l.w.Resize(fyne.NewSize(640, 480))
	l.w.SetFixedSize(true)
	l.w.ShowAndRun()
}

func (l *Launcher) openSettings() {
	l.settings.Resize(fyne.NewSize(500, 200))
	l.settings.Show()
}

func (l *Launcher) buildSettingsDialog() *dialog.CustomDialog {
	javaPathInputLabel := widget.NewLabel(lang.L("Java path"))
	javaPathInput := widget.NewEntry()
	javaPathInput.SetText(l.Config.JavaPath)
	javaPathInput.OnChanged = func(javaPath string) {
		l.Config.JavaPath = javaPath
	}

	memoryInputLabel := widget.NewLabel(lang.L("Minecraft memory"))
	memoryInput := widget.NewEntry()
	memoryInput.SetText(l.Config.Memory)
	memoryInput.OnChanged = func(memory string) {
		l.Config.Memory = memory
	}

	return dialog.NewCustom(lang.L("Settings"), lang.L("Close"),
		container.NewVBox(
			layout.NewSpacer(),
			container.New(layout.NewFormLayout(),
				javaPathInputLabel, javaPathInput,
				memoryInputLabel, memoryInput,
			),
			layout.NewSpacer(),
		), l.w,
	)
}

func (l *Launcher) buildUI() {
	l.mainButton = l.buildMainButton()
	l.settings = l.buildSettingsDialog()

	usernameInput := l.buildUsernameInput()

	bottom := container.New(
		NewRatioLayout(2.0),
		usernameInput, l.mainButton,
	)

	background := canvas.NewImageFromReader(
		bytes.NewReader(static.Background), "background.png",
	)

	topMenu := container.NewBorder(
		nil, nil, nil,
		widget.NewButtonWithIcon("", theme.Icon(theme.IconNameSettings), l.openSettings),
	)

	page := container.NewPadded(
		container.NewBorder(
			topMenu, container.NewVBox(
				widget.NewLabelWithData(l.statusText), bottom,
			), nil, nil,
		),
	)

	l.w.SetContent(
		container.NewStack(background, page),
	)
	l.w.SetPadded(false)
}

func (l *Launcher) setState(state LauncherState) {
	l.state = state
	l.mainButton.SetText(lang.L(mainBtnTexts[l.state]))
}

func (l *Launcher) buildUsernameInput() *widget.Entry {
	entry := widget.NewEntry()
	entry.OnChanged = func(data string) {
		l.Config.Username = data
	}
	entry.Text = l.Config.Username
	entry.PlaceHolder = lang.L("Enter your username")

	return entry
}

func (l *Launcher) showError(err error) {
	if err == nil {
		return
	}

	d := dialog.NewError(err, l.w)
	d.Resize(fyne.NewSize(350, 150))
	d.Show()
}

func (l *Launcher) buildMainButton() *widget.Button {
	return widget.NewButton(lang.L(mainBtnTexts[l.state]), func() {
		switch l.state {
		case CanUpdate:
			l.setState(Downloading)

			go func() {
				if l.version != "dev-build" {
					if _, err := l.u.Update(); err != nil {
						l.showError(err)
					} else {
						l.a.Quit()
					}
				}

				l.setState(Idle)
			}()
		case ClientNotInstalled:
			go func() {
				fyne.Do(func() { l.setState(Downloading) })

				err := l.install()
				if err != nil {
					l.showError(err)
				}

				fyne.Do(func() { l.setState(Idle) })
			}()
		case Idle:
			l.setState(StartedClient)
			l.statusText.Set("")

			err := discord.SetPlayingActivity()
			if err != nil {
				log.Println("failed to set playing activity: ", err)
			}

			go func() {
				if err := l.core.Launch(); err != nil {
					l.showError(err)
				}

				fyne.Do(func() {
					l.setState(Idle)
					l.w.Show()
				})
			}()
		}
	})
}

func (l *Launcher) install() error {
	d := downloader.New(l.Config.GameDir)
	fabricInstaller := downloader.NewFabricInstaller(l.Config.GameDir)

	l.statusText.Set(lang.L("Getting version info..."))
	versionURL, err := d.GetVersionURL(l.Config.Version)
	if err != nil {
		return fmt.Errorf("failed to get version url: %s", err)
	}

	details, err := d.GetVersionDetails(versionURL)
	if err != nil {
		return fmt.Errorf("failed to get version details: %s", err)
	}

	l.statusText.Set(lang.L("Downloading Minecraft jar..."))
	if err := d.DownloadClient(details); err != nil {
		return fmt.Errorf("failed to download minecraft jar: %s", err)
	}

	l.statusText.Set(lang.L("Downloading libraries..."))
	if err := d.DownloadLibraries(details.Libraries); err != nil {
		return fmt.Errorf("failed to download minecraft libraries: %s", err)
	}

	l.statusText.Set(lang.L("Downloading assets..."))
	if err := d.DownloadAssets(details.AssetIndex); err != nil {
		return fmt.Errorf("failed to download minecraft assets: %s", err)
	}

	l.statusText.Set(lang.L("Download fabric..."))
	if err := fabricInstaller.InstallFabric(utils.McVersion); err != nil {
		return fmt.Errorf("failed to download fabric: %s", err)
	}

	l.statusText.Set(lang.L("Downloading mods..."))
	if err := d.DownloadMods(mods); err != nil {
		return fmt.Errorf("failed to download mods: %s", err)
	}

	l.statusText.Set(lang.L("Writing static files..."))
	if err := d.WriteStaticFiles(statics); err != nil {
		return fmt.Errorf("failed to write static files: %s", err)
	}

	l.statusText.Set(lang.L("Downloading Java..."))
	if err := d.DownloadJava(); err != nil {
		return fmt.Errorf("failed to download java: %s", err)
	}

	l.Config.JavaPath = d.GetJavaPath()
	l.statusText.Set(lang.L("Successfully installed minecraft!"))
	return nil
}
