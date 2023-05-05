package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	clone "github.com/huandu/go-clone/generic"
	"github.com/leighmacdonald/bd/internal/detector"
	"github.com/leighmacdonald/bd/internal/tr"
	"github.com/leighmacdonald/bd/pkg/util"
	"github.com/leighmacdonald/steamid/v2/steamid"
	"github.com/leighmacdonald/steamweb"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"path/filepath"
	"strings"
)

func newSettingsDialog(logger *zap.Logger, parent fyne.Window) dialog.Dialog {
	const testSteamId = 76561197961279983
	origSettings := detector.Settings()
	settings := clone.Clone[*detector.UserSettings](origSettings)

	var createSelectorRow = func(label string, icon fyne.Resource, entry *widget.Entry, defaultPath string) *container.Split {
		fileInputContainer := container.NewHSplit(widget.NewButtonWithIcon("Edit", icon, func() {
			d := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
				if err != nil || uri == nil {
					return
				}
				entry.SetText(uri.Path())
			}, parent)
			d.Show()
		}), entry)
		fileInputContainer.SetOffset(0.0)
		return fileInputContainer
	}
	apiKeyOriginal := settings.GetAPIKey()
	apiKeyEntry := widget.NewPasswordEntry()
	apiKey := settings.GetAPIKey()
	apiKeyEntry.Bind(binding.BindString(&apiKey))
	apiKeyEntry.Validator = func(newApiKey string) error {
		if len(newApiKey) > 0 && len(newApiKey) != 32 {
			msg := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{ID: "error_invalid_api_key", Other: "Invalid API Key"}})
			return errors.New(msg)
		}
		// Wait until all validation is complete to keep the setting
		defer func() {
			_ = steamweb.SetKey(apiKeyOriginal)
		}()
		if newApiKey == "" {
			return nil
		}
		if errSetKey := steamweb.SetKey(newApiKey); errSetKey != nil {
			return errSetKey
		}
		res, errRes := steamweb.PlayerSummaries(steamid.Collection{testSteamId})
		if errRes != nil {
			logger.Error("Failed to fetch player summary for validation", zap.Error(errRes))
			msg := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{ID: "error_invalid_api_key", Other: "Failed to validate"}})
			return errors.New(msg)
		}
		if len(res) != 1 {
			logger.Error("Received incorrect number of results in steam api validation call")
			msg := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{ID: "error_invalid_api_invalid_response", Other: "Invalid Response"}})
			return errors.New(msg)
		}
		return nil
	}

	steamIdEntry := widget.NewEntry()
	steamIdVal := settings.GetSteamId()
	steamId := steamIdVal.String()
	steamIdEntry.Bind(binding.BindString(&steamId))
	steamIdEntry.Validator = validateSteamId
	tf2Dir := settings.GetTF2Dir()
	tf2RootEntry := widget.NewEntryWithData(binding.BindString(&tf2Dir))
	tf2RootEntry.Validator = validateSteamRoot
	steamDir := settings.GetSteamDir()
	steamDirEntry := widget.NewEntryWithData(binding.BindString(&steamDir))
	steamDirEntry.Validator = func(newRoot string) error {
		if len(newRoot) > 0 {
			if !util.Exists(newRoot) {
				msg := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "error_invalid_path", Other: "Invalid Path"}})
				return errors.New(msg)
			}
			userDataDir := filepath.Join(newRoot, "userdata")
			if !util.Exists(userDataDir) {
				msg := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{ID: "error_invalid_steam_dir_user_data", Other: "Could not find userdata folder"}})
				return errors.New(msg)
			}
			if tf2RootEntry.Text == "" {
				dp := filepath.Join(newRoot, "steamapps/common/Team Fortress 2/tf")
				if errValid := validateSteamRoot(dp); errValid == nil && util.Exists(dp) {
					tf2RootEntry.SetText(dp)
				}
			}
		}
		return nil
	}
	autoCloseOnGameExit := settings.GetAutoCloseOnGameExit()
	autoCloseOnGameExitEntry := widget.NewCheckWithData("", binding.BindBool(&autoCloseOnGameExit))

	autoLaunchGame := settings.GetAutoLaunchGame()
	autoLaunchGameEntry := widget.NewCheckWithData("", binding.BindBool(&autoLaunchGame))

	kickerEnabled := settings.GetKickerEnabled()
	kickerEnabledEntry := widget.NewCheckWithData("", binding.BindBool(&kickerEnabled))

	chatWarningsEnabled := settings.GetChatWarningsEnabled()
	chatWarningsEnabledEntry := widget.NewCheckWithData("", binding.BindBool(&chatWarningsEnabled))

	partyWarningsEnabled := settings.GetPartyWarningsEnabled()
	partyWarningsEnabledEntry := widget.NewCheckWithData("", binding.BindBool(&partyWarningsEnabled))

	discordPresenceEnabled := settings.GetDiscordPresenceEnabled()
	discordPresenceEnabledEntry := widget.NewCheckWithData("", binding.BindBool(&discordPresenceEnabled))

	voiceBansEnabled := settings.GetVoiceBansEnabled()
	voiceBanEnabledEntry := widget.NewCheckWithData("", binding.BindBool(&voiceBansEnabled))

	rconStatic := settings.GetRCONStatic()

	rconModeStaticEntry := widget.NewCheckWithData("", binding.BindBool(&rconStatic))

	debugLogEnabled := settings.GetDebugLogEnabled()
	debugLogEnabledEntry := widget.NewCheckWithData("", binding.BindBool(&debugLogEnabled))

	staticConfig := detector.NewRconConfig(true)
	boundTags := binding.NewString()
	if errSet := boundTags.Set(strings.Join(settings.GetKickTags(), ",")); errSet != nil {
		logger.Error("Failed to set tags", zap.Error(errSet))
	}
	tagsEntry := widget.NewEntryWithData(boundTags)
	tagsEntry.Validator = validateTags
	linksDialog := newLinksDialog(parent, logger, settings)
	linksButton := widget.NewButtonWithIcon("Edit Links", theme.SettingsIcon(), func() {
		linksDialog.Show()
	})
	linksButton.Alignment = widget.ButtonAlignLeading
	linksButton.Refresh()

	listsDialog := newRuleListConfigDialog(parent, logger, settings)
	listsButton := widget.NewButtonWithIcon("Edit Lists", theme.SettingsIcon(), func() {
		listsDialog.Show()
	})
	listsButton.Alignment = widget.ButtonAlignLeading
	listsButton.Refresh()

	labelLists := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_lists", Other: "Lists & Rules"}})
	labelListsHint := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_lists_hint", Other: "Configure your 3rd party player and rule lists"}})
	labelLinks := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_links", Other: "External Links"}})
	labelLinksHint := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_links_hint", Other: "Customize external links menu"}})
	labelKickerEnabled := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_kicker_enabled", Other: "Vote Kicker"}})
	labelKickerEnabledHint := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_kicker_enabled_hint", Other: "Enable vote kick functionality in-game"}})
	labelKickableTags := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_kickable_tags", Other: "Kickable Tags"}})
	labelKickableTagsHint := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_kickable_tags_hint", Other: "Attributes/Tags that when matched will trigger a in-game kick."}})
	labelChatWarnEnabled := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_chat_warn_enabled", Other: "Chat Warnings"}})
	labelChatWarnEnabledHint := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_chat_warn_enabled_hint", Other: "Show warning message using in-game chat"}})
	labelPartyWarnEnabled := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_party_warn_enabled", Other: "Party Warnings"}})
	labelPartyWarnEnabledHint := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_party_warn_enabled_hint", Other: "Show lobby only warning messages"}})
	labelDiscordPresence := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_discord_presence_enabled", Other: "Discord Presence"}})
	labelDiscordPresenceHint := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_discord_presence_enabled_hint", Other: "Enables discord rich presence if discord is running"}})
	labelAutoLaunch := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_auto_launch", Other: "Auto Launch TF2"}})
	labelAutoLaunchHint := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_auto_launch_hint", Other: "When launching bd, also automatically launch tf2"}})
	labelAutoExit := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_auto_exit", Other: "Auto Close"}})
	labelAutoExitHint := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_auto_exit_hint", Other: "When TF2 exits, close bd as well"}})
	labelSteamAPIKey := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_steam_api_key", Other: "Steam API Key"}})
	labelSteamAPIKeyHint := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_steam_api_key_hint", Other: "Steam web api key. https://steamcommunity.com/dev/apikey"}})
	labelSteamID := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_steam_id", Other: "Steam ID"}})
	labelSteamIDHint := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_steam_id_hint", Other: "Your steam id in any of the following formats: steam,steam3,steam32,steam64"}})
	labelSteamRoot := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_steam_root", Other: "Steam Root"}})
	labelSteamRootHint := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_steam_root_hint", Other: "Location of your steam install directory containing a userdata folder."}})
	labelTF2Root := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_tf2_root", Other: "TF2 Root"}})
	labelTF2RootHint := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_tf2_root_hint", Other: "Path to your steamapps/common/Team Fortress 2/tf folder"}})
	labelRCONMode := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_rcon_mode", Other: "RCON Mode"}})
	labelRCONModeHint := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_rcon_mode_hint", Other: "Static: Port: {{ .Port }}, Password: {{ .Password }}"},
		TemplateData:   map[string]interface{}{"Port": staticConfig.Port(), "Password": staticConfig.Password()}})
	labelVoiceBanEnabled := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_voice_ban_enabled", Other: "Gen. Voice Bans"}})
	labelVoiceBanEnabledHint := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_voice_ban_enabled_hint",
			Other: "WARN: This will overwrite your current ban list. Mutes the 200 most recent marked entries."}})
	labelSelect := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_select_folder",
			Other: "Select"}})
	labelDebugLogEnabled := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_debug_log_enabled",
			Other: "Debug Log"}})
	labelDebugLogEnabledHint := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: "settings_label_debug_log_enabled_hint",
			Other: "Log events are save to bd.log. Requires restart of application."}})

	settingsForm := &widget.Form{
		Items: []*widget.FormItem{
			{Text: labelLists, Widget: listsButton, HintText: labelListsHint},
			{Text: labelLinks, Widget: linksButton, HintText: labelLinksHint},
			{Text: labelKickerEnabled, Widget: kickerEnabledEntry, HintText: labelKickerEnabledHint},
			{Text: labelKickableTags, Widget: tagsEntry, HintText: labelKickableTagsHint},
			{Text: labelChatWarnEnabled, Widget: chatWarningsEnabledEntry, HintText: labelChatWarnEnabledHint},
			{Text: labelPartyWarnEnabled, Widget: partyWarningsEnabledEntry, HintText: labelPartyWarnEnabledHint},
			{Text: labelDiscordPresence, Widget: discordPresenceEnabledEntry, HintText: labelDiscordPresenceHint},
			{Text: labelAutoLaunch, Widget: autoLaunchGameEntry, HintText: labelAutoLaunchHint},
			{Text: labelAutoExit, Widget: autoCloseOnGameExitEntry, HintText: labelAutoExitHint},
			{Text: labelDebugLogEnabled, Widget: debugLogEnabledEntry, HintText: labelDebugLogEnabledHint},
			{Text: labelSteamAPIKey, Widget: apiKeyEntry, HintText: labelSteamAPIKeyHint},
			{Text: labelSteamID, Widget: steamIdEntry, HintText: labelSteamIDHint},
			{Text: labelSteamRoot,
				Widget:   createSelectorRow(labelSelect, theme.FolderIcon(), steamDirEntry, ""),
				HintText: labelSteamRootHint},
			{Text: labelTF2Root,
				Widget:   createSelectorRow(labelSelect, theme.FolderIcon(), tf2RootEntry, ""),
				HintText: labelTF2RootHint},
			{Text: labelRCONMode, Widget: rconModeStaticEntry, HintText: labelRCONModeHint},
			{Text: labelVoiceBanEnabled, Widget: voiceBanEnabledEntry, HintText: labelVoiceBanEnabledHint},
		},
	}
	onSave := func(status bool) {
		if !status {
			return
		}
		// Update it to our preferred format
		if steamIdEntry.Text != "" {
			newSid, errSid := steamid.StringToSID64(steamIdEntry.Text)
			if errSid != nil {
				// Should never happen? was validated previously.
				logger.Panic("Steamid state invalid?", zap.Error(errSid))
			}
			origSettings.SetSteamID(newSid.String())
			steamIdEntry.SetText(newSid.String())
		}
		var newTags []string
		for _, t := range strings.Split(tagsEntry.Text, ",") {
			if t == "" {
				continue
			}
			newTags = append(newTags, strings.Trim(t, " "))
		}
		origSettings.SetKickTags(newTags)
		origSettings.SetAPIKey(apiKeyEntry.Text)
		origSettings.SetSteamDir(steamDirEntry.Text)
		origSettings.SetTF2Dir(tf2RootEntry.Text)
		origSettings.SetKickerEnabled(kickerEnabledEntry.Checked)
		origSettings.SetChatWarningsEnabled(chatWarningsEnabledEntry.Checked)
		origSettings.SetPartyWarningsEnabled(partyWarningsEnabledEntry.Checked)
		origSettings.SetRconStatic(rconModeStaticEntry.Checked)
		origSettings.SetAutoCloseOnGameExit(autoCloseOnGameExitEntry.Checked)
		origSettings.SetAutoLaunchGame(autoLaunchGameEntry.Checked)
		origSettings.SetVoiceBansEnabled(voiceBanEnabledEntry.Checked)
		origSettings.SetDebugLogEnabled(debugLogEnabledEntry.Checked)
		origSettings.SetDiscordPresenceEnabled(discordPresenceEnabledEntry.Checked)
		origSettings.SetLinks(settings.GetLinks())
		origSettings.SetLists(settings.GetLists())

		if apiKeyOriginal != apiKeyEntry.Text {
			if errSetKey := steamweb.SetKey(apiKeyEntry.Text); errSetKey != nil {
				logger.Error("Failed to set new steam key", zap.Error(errSetKey))
			}
		}
		if errSave := origSettings.Save(); errSave != nil {
			logger.Error("Failed to save settings", zap.Error(errSave))
		}
	}
	titleSettings := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "settings_title", Other: "Edit UserSettings"}})
	buttonSave := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "settings_button_apply", Other: "Save"}})
	buttonCancel := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "settings_button_cancel", Other: "Cancel"}})
	settingsWindow := dialog.NewCustomConfirm(titleSettings, buttonSave, buttonCancel, container.NewVScroll(settingsForm), onSave, parent)
	settingsForm.Refresh()
	settingsWindow.Resize(fyne.NewSize(sizeDialogueWidth, sizeWindowMainHeight-200))
	return settingsWindow
}
