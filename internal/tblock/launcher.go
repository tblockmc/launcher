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
)

type LauncherState int

const (
	Downloading LauncherState = iota
	StartedClient
	ClientNotInstalled
	Idle
)

var (
	statics = []downloader.StaticAsset{
		{Path: "options.txt", Data: static.OptionsTXT},
		{Path: "servers.dat", Data: static.ServersDAT},
	}

	mods = []downloader.ModData{
		{Name: "fabric-api", URL: "https://cdn.modrinth.com/data/P7dR8mSH/versions/p96k10UR/fabric-api-0.119.4%2B1.21.4.jar"},
		{Name: "sodium", URL: "https://cdn.modrinth.com/data/AANobbMI/versions/c3YkZvne/sodium-fabric-0.6.13%2Bmc1.21.4.jar"},
		{Name: "simple-voice-chat", URL: "https://cdn.modrinth.com/data/9eGKb6K1/versions/5ODvTv8E/voicechat-fabric-1.21.4-2.6.6.jar"},
		{Name: "modmenu", URL: "https://cdn.modrinth.com/data/mOgUt4GM/versions/7iGb2ltH/modmenu-13.0.3.jar"},
		{Name: "zoomify", URL: "https://cdn.modrinth.com/data/w7ThoJFB/versions/RKRjd2h1/Zoomify-2.14.2%2B1.21.3.jar"},
		{Name: "iris-shaders", URL: "https://cdn.modrinth.com/data/YL57xq9U/versions/Ca054sTe/iris-fabric-1.8.8%2Bmc1.21.4.jar"},
		// limbs dont bend on 1.21.4 with sodium, update when fixed
		{Name: "emotecraft", URL: "https://cdn.modrinth.com/data/pZ2wrerK/versions/njhJbosE/emotecraft-fabric-for-MC1.21.4-2.5.8.jar"},
		{Name: "yet-another-configlib", URL: "https://cdn.modrinth.com/data/1eAoo2KR/versions/F67XvT8M/yet_another_config_lib_v3-3.8.0%2B1.21.4-fabric.jar"},
		{Name: "fabric-language-kotlin", URL: "https://cdn.modrinth.com/data/Ha28R6CL/versions/LcgnDDmT/fabric-language-kotlin-1.13.7%2Bkotlin.2.2.21.jar"},
	}

	mainBtnTexts = map[LauncherState]string{
		Idle:               "Play!",
		ClientNotInstalled: "Download",
		StartedClient:      "Running...",
		Downloading:        "Downloading...",
	}
)

type Launcher struct {
	Config *launcher.Config
	core   *launcher.FabricLauncher
	state  LauncherState

	w          fyne.Window
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

	// hardcode ukrainian for now
	err = lang.AddTranslationsForLocale(uk, lang.SystemLocale())
	if err != nil {
		return nil, err
	}

	a.Settings().SetTheme(newTheme())
	w := a.NewWindow(lang.L("TBlockMC"))

	state := Idle
	if !core.IsFabricInstalled() {
		state = ClientNotInstalled
	}

	mainBtnText := binding.NewString()
	mainBtnText.Set(lang.L(mainBtnTexts[state]))

	statusText := binding.NewString()

	return &Launcher{
		state: state, w: w,
		core: core, Config: cfg,
		statusText: statusText,
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
	if err := fabricInstaller.InstallFabric("1.21.4"); err != nil {
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
