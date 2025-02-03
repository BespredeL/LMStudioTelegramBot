package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"log"
	"os"
	"strconv"
	"time"
)

func botTabContent() *fyne.Container {
	botControlButton := widget.NewButton(t("Launch Telegram bot"), nil)
	statusLabel := widget.NewLabel(t("The bot is not launched"))
	var pollingStopChan chan struct{}
	botRunning := false

	startBot := func() bool {
		if selectedModel == "" {
			statusLabel.SetText(t("Choose a model!"))
			return false
		}

		statusLabel.SetText(t("The bot is running!"))
		if config.UpdateMethod == "polling" {
			pollingStopChan = make(chan struct{})
			go startLongPolling(pollingStopChan)
		} else if config.UpdateMethod == "webhook" {
			go startWebhookServer()
		}

		return true
	}

	stopBot := func() {
		if config.UpdateMethod == "polling" && pollingStopChan != nil {
			close(pollingStopChan)
		}
		botRunning = false
		statusLabel.SetText(t("The bot is stopped"))
	}

	botControlButton.OnTapped = func() {
		if !botRunning {
			if startBot() {
				botRunning = true
				botControlButton.SetText(t("Stop Telegram bot"))
			}
		} else {
			stopBot()
			botControlButton.SetText("Launch Telegram bot")
		}
	}

	logEntry := widget.NewMultiLineEntry()
	logEntry.Wrapping = fyne.TextWrapWord

	logScroll := container.NewVScroll(logEntry)
	logScroll.SetMinSize(fyne.NewSize(0, 500)) // Minimum height

	// Log renewal function
	updateLog := func() {
		data, err := os.ReadFile("requests.log")
		if err == nil {
			logEntry.SetText(string(data))
			logEntry.Refresh() // We update UI manually
		}
	}

	// Run the gorutin, which updates the log every 2 seconds
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			updateLog()
		}
	}()

	botControlContainer := container.NewBorder(
		container.NewVBox(botControlButton, statusLabel),
		nil, nil, nil,
		logScroll,
	)

	return botControlContainer
}

func startGUI() {
	// Create a new application and a window with a given heading
	application := app.NewWithID("telegram.lmstudio.bot")
	//application.Settings().SetTheme(theme.LightTheme())
	window := application.NewWindow("Telegram Bot + LM Studio")
	window.Resize(fyne.NewSize(800, 650))

	// --------------------------
	// The Configuration tab
	// --------------------------
	apiAddressEntry := widget.NewEntry()
	apiAddressEntry.SetText(config.APIAddress)
	apiAddressEntry.SetPlaceHolder("http://localhost:1234")

	tokenLimitEntry := widget.NewEntry()
	tokenLimitEntry.SetText(strconv.Itoa(config.TokenLimit))
	tokenLimitEntry.SetPlaceHolder("2048")

	timeoutEntry := widget.NewEntry()
	timeoutEntry.SetText(strconv.Itoa(config.PollingTimeout))
	timeoutEntry.SetPlaceHolder("60")

	botTokenEntry := widget.NewPasswordEntry()
	botTokenEntry.SetText(config.BotToken)
	botTokenEntry.SetPlaceHolder(t("Enter the bot token"))

	updateMethodSelect := widget.NewSelect([]string{"polling", "webhook"}, func(val string) {
		config.UpdateMethod = val
	})
	updateMethodSelect.SetSelected(config.UpdateMethod)

	webhookDomainEntry := widget.NewEntry()
	webhookDomainEntry.SetText(config.WebhookDomain)
	webhookDomainEntry.SetPlaceHolder("mybot.example.com")

	webhookPortEntry := widget.NewEntry()
	webhookPortEntry.SetText(config.WebhookPort)
	webhookPortEntry.SetPlaceHolder("")

	certFileEntry := widget.NewEntry()
	certFileEntry.SetText(config.CertFile)
	certFileEntry.SetPlaceHolder("cert.pem")

	keyFileEntry := widget.NewEntry()
	keyFileEntry.SetText(config.KeyFile)
	keyFileEntry.SetPlaceHolder("key.pem")

	systemRoleEntry := widget.NewMultiLineEntry()
	systemRoleEntry.Wrapping = fyne.TextWrapWord
	systemRoleEntry.SetText(config.SystemRole)
	systemRoleEntry.SetPlaceHolder(t("System message (role)..."))

	lmModeSelect := widget.NewSelect([]string{"stream", "full"}, func(val string) {
		config.LMStudioMode = val
	})
	lmModeSelect.SetSelected(config.LMStudioMode)
	lmModeSelect.PlaceHolder = t("Select the LM Studio mode")

	languageSelect := widget.NewSelect([]string{"en", "ru"}, func(val string) {
		config.Language = val
	})
	languageSelect.SetSelected(config.Language)
	languageSelect.PlaceHolder = t("Select a language")

	saveConfigButton := widget.NewButtonWithIcon(t("Save the configuration"), theme.DocumentSaveIcon(), func() {
		config.APIAddress = apiAddressEntry.Text
		if n, err := fmt.Sscanf(tokenLimitEntry.Text, "%d", &config.TokenLimit); n != 1 || err != nil {
			dialog.ShowError(fmt.Errorf("the wrong value of the maximum number of tokens"), window)
			return
		}
		if n, err := fmt.Sscanf(timeoutEntry.Text, "%d", &config.PollingTimeout); n != 1 || err != nil {
			dialog.ShowError(fmt.Errorf("the wrong meaning of the timeout"), window)
			return
		}
		config.BotToken = botTokenEntry.Text
		config.UpdateMethod = updateMethodSelect.Selected
		config.WebhookDomain = webhookDomainEntry.Text
		config.WebhookPort = webhookPortEntry.Text
		config.CertFile = certFileEntry.Text
		config.KeyFile = keyFileEntry.Text
		config.SystemRole = systemRoleEntry.Text
		config.LMStudioMode = lmModeSelect.Selected

		if err := saveConfig(); err != nil {
			dialog.ShowError(fmt.Errorf("configuration conservation error: %v", err), window)
			return
		}
		dialog.ShowInformation(t("Success"), t("The configuration is saved!"), window)
	})

	configForm := container.NewVBox(
		widget.NewLabelWithStyle(t("Configuration"), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewForm(
			widget.NewFormItem(t("LM Studio API address"), apiAddressEntry),
			widget.NewFormItem(t("Max. tokens"), tokenLimitEntry),
			widget.NewFormItem(t("Timeout (sec)"), timeoutEntry),
			widget.NewFormItem(t("Bot token"), botTokenEntry),
			widget.NewFormItem(t("Update method"), updateMethodSelect),
			widget.NewFormItem(t("Webhook domain"), webhookDomainEntry),
			widget.NewFormItem(t("Webhook port"), webhookPortEntry),
			widget.NewFormItem(t("The path to Cert.pem"), certFileEntry),
			widget.NewFormItem(t("The path to Key.pem"), keyFileEntry),
			widget.NewFormItem(t("System message"), systemRoleEntry),
			widget.NewFormItem(t("LM Studio mode"), lmModeSelect),
			widget.NewFormItem(t("Language"), languageSelect),
		),
		saveConfigButton,
	)

	// --------------------------
	// Tab "LM Studio Models"
	// --------------------------
	modelSelect := widget.NewSelect([]string{}, func(val string) {
		selectedModel = val
		log.Printf("Selected model: %s", selectedModel)
	})
	modelSelect.PlaceHolder = t("Select a model")

	refreshModelsButton := widget.NewButtonWithIcon(t("Update models"), theme.ViewRefreshIcon(), func() {
		models, err := fetchModels()
		if err != nil {
			dialog.ShowError(fmt.Errorf("error getting models: %v", err), window)
			modelSelect.Options = []string{}
			modelSelect.PlaceHolder = t("Error getting models")
			modelSelect.Refresh()
			return
		}
		modelSelect.Options = models
		modelSelect.Refresh()
		dialog.ShowInformation(t("Models"), t("Models found: %d", len(models)), window)
	})

	modelsContainer := container.NewVBox(
		widget.NewLabelWithStyle(t("LM Studio Models"), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		modelSelect,
		refreshModelsButton,
	)

	// --------------------------
	// Tab "Users"
	// --------------------------
	usersTableContainer := container.NewVBox()
	refreshUsersTable := func() {
		if err := loadUsers(); err != nil {
			dialog.ShowError(fmt.Errorf("error loading users: %v", err), window)
			return
		}

		rows := []fyne.CanvasObject{
			container.NewHBox(
				widget.NewLabelWithStyle(t("ID"), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
				widget.NewLabelWithStyle(t("Username"), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
				widget.NewLabelWithStyle(t("Allowed"), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			),
		}

		for _, u := range getSortedUsers() {
			uid := u.ID // Local copy to close
			allowedCheck := widget.NewCheck("", func(val bool) {
				usersMutex.Lock()
				if user, ok := users[uid]; ok {
					user.Allowed = val
				}
				usersMutex.Unlock()
				if err := saveUsers(); err != nil {
					dialog.ShowError(fmt.Errorf("error saving users: %v", err), window)
				}
			})

			allowedCheck.SetChecked(u.Allowed)
			row := container.NewHBox(
				widget.NewLabel(fmt.Sprintf("%d", u.ID)),
				widget.NewLabel(u.Username),
				allowedCheck,
			)

			rows = append(rows, row)
		}
		usersTableContainer.Objects = rows
		usersTableContainer.Refresh()
	}

	refreshUsersButton := widget.NewButtonWithIcon(t("Refresh list"), theme.ViewRefreshIcon(), func() {
		refreshUsersTable()
	})

	usersScroll := container.NewVScroll(usersTableContainer)
	usersScroll.SetMinSize(fyne.NewSize(0, 400))

	usersContainer := container.NewBorder(
		nil,
		refreshUsersButton,
		nil,
		nil,
		usersScroll,
	)

	refreshUsersTable()

	// --------------------------
	// Basic layout via TabContainer
	// --------------------------
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon(t("Configuration"), theme.SettingsIcon(), container.NewVScroll(configForm)),
		container.NewTabItemWithIcon(t("Models"), theme.ComputerIcon(), container.NewVScroll(modelsContainer)),
		container.NewTabItemWithIcon(t("Users"), theme.AccountIcon(), container.NewVScroll(usersContainer)),
		container.NewTabItemWithIcon(t("Bot"), theme.MediaPlayIcon(), botTabContent()),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	window.SetContent(tabs)
	window.ShowAndRun()
}
