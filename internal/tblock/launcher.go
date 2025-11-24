package tblock

import (
	"bytes"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
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
	statusText binding.String
}

func NewLauncher(cfg *launcher.Config) *Launcher {
	err := discord.Login()
	if err != nil {
		log.Println("discord login failed: ", err)
	}

	core := launcher.NewFabricLauncher(cfg)

	a := app.NewWithID("com.github.tblockmc.launcher")
	a.Settings().SetTheme(newTheme())
	w := a.NewWindow("TBlockMC Launcher")

	state := Idle
	if !core.IsFabricInstalled() {
		state = ClientNotInstalled
	}

	mainBtnText := binding.NewString()
	mainBtnText.Set(mainBtnTexts[state])

	statusText := binding.NewString()

	return &Launcher{
		state: state, w: w,
		core: core, Config: cfg,
		statusText: statusText,
	}
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

func (l *Launcher) buildUI() {
	l.mainButton = l.buildMainButton()
	usernameInput := l.buildUsernameInput()

	bottom := container.New(
		NewRatioLayout(2.0),
		usernameInput, l.mainButton,
	)

	background := canvas.NewImageFromReader(
		bytes.NewReader(static.Background), "background.png",
	)

	page := container.NewPadded(
		container.NewBorder(
			nil, container.NewVBox(
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
	l.mainButton.SetText(mainBtnTexts[l.state])
}

func (l *Launcher) buildUsernameInput() *widget.Entry {
	entry := widget.NewEntry()
	entry.OnChanged = func(data string) {
		l.Config.Username = data
	}
	entry.Text = l.Config.Username
	entry.PlaceHolder = "Enter your username"

	return entry
}

func (l *Launcher) buildMainButton() *widget.Button {
	return widget.NewButton(mainBtnTexts[l.state], func() {
		switch l.state {
		case ClientNotInstalled:
			go func() {
				fyne.Do(func() { l.setState(Downloading) })

				err := l.install()
				if err != nil {
					log.Fatal(err)
				}

				fyne.Do(func() { l.setState(Idle) })
			}()
		case Idle:
			l.setState(StartedClient)
			l.statusText.Set("")
			l.w.Hide()

			err := discord.SetPlayingActivity()
			if err != nil {
				log.Println("failed to set playing activity: ", err)
			}

			go func() {
				if err := l.core.Launch(); err != nil {
					log.Println("failed to launch minecraft: ", err)
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

	l.statusText.Set("Getting version info...")
	versionURL, err := d.GetVersionURL(l.Config.Version)
	if err != nil {
		return err
	}

	details, err := d.GetVersionDetails(versionURL)
	if err != nil {
		return err
	}

	l.statusText.Set("Downloading Minecraft jar...")
	if err := d.DownloadClient(details); err != nil {
		return err
	}

	l.statusText.Set("Downloading libraries...")
	if err := d.DownloadLibraries(details.Libraries); err != nil {
		return err
	}

	l.statusText.Set("Downloading assets...")
	if err := d.DownloadAssets(details.AssetIndex); err != nil {
		return err
	}

	l.statusText.Set("Download fabric...")
	if err := fabricInstaller.InstallFabric("1.21.4"); err != nil {
		return err
	}

	l.statusText.Set("Downloading mods...")
	if err := d.DownloadMods(mods); err != nil {
		return err
	}

	l.statusText.Set("Writing static files...")
	if err := d.WriteStaticFiles(statics); err != nil {
		return err
	}

	l.statusText.Set("Downloading Java 17...")
	if err := d.DownloadJava(); err != nil {
		return err
	}

	l.Config.JavaPath = d.GetJavaPath()

	l.statusText.Set("Successfully installed minecraft!")
	return nil
}
