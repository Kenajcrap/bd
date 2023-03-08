package ui

import (
	"context"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/leighmacdonald/bd/internal/model"
	"github.com/leighmacdonald/bd/internal/tr"
	"github.com/leighmacdonald/steamid/v2/steamid"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"
	"log"
	"net/url"
	"sort"
	"strings"
)

type menuButton struct {
	widget.Button
	menu *fyne.Menu
}

func (m *menuButton) Tapped(event *fyne.PointEvent) {
	widget.ShowPopUpMenuAtPosition(m.menu, fyne.CurrentApp().Driver().CanvasForObject(m), event.AbsolutePosition)
}

func newMenuButton(menu *fyne.Menu) *menuButton {
	c := &menuButton{menu: menu}
	c.ExtendBaseWidget(c)
	c.SetIcon(theme.SettingsIcon())

	return c
}

const newItemLabel = "New..."

func generateAttributeMenu(window fyne.Window, sid64 steamid.SID64, attrList binding.StringList, markFunc model.MarkFunc) *fyne.Menu {
	mkAttr := func(attrName string) func() {
		clsAttribute := attrName
		clsSteamId := sid64
		return func() {
			log.Printf("marking %d as %s", clsSteamId, clsAttribute)
			if errMark := markFunc(sid64, []string{clsAttribute}); errMark != nil {
				log.Printf("Failed to mark player: %v\n", errMark)
			}
		}
	}
	markAsMenuLabel := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "menu_markas_label", Other: "Mark As..."}})
	markAsMenu := fyne.NewMenu(markAsMenuLabel)
	knownAttributes, errGet := attrList.Get()
	if errGet != nil {
		log.Panicf("Failed to get list: %v\n", errGet)
	}
	sort.Slice(knownAttributes, func(i, j int) bool {
		return strings.ToLower(knownAttributes[i]) < strings.ToLower(knownAttributes[j])
	})
	for _, mi := range knownAttributes {
		markAsMenu.Items = append(markAsMenu.Items, fyne.NewMenuItem(mi, mkAttr(mi)))
	}
	entry := widget.NewEntry()
	entry.Validator = func(s string) error {
		if s == "" {
			msg := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "error_attribute_empty", Other: "Attribute cannot be empty"}})
			return errors.New(msg)
		}
		for _, knownAttr := range knownAttributes {
			if strings.EqualFold(knownAttr, s) {
				msg := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{ID: "error_attribute_duplicate", Other: "Duplicate attribute: {{ .Attr }} "},
					TemplateData:   map[string]any{"Attr": knownAttr}})
				return errors.New(msg)
			}
		}
		return nil
	}
	attributeLabel := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "mark_label_attr", Other: "Attribute Name"}})
	fi := widget.NewFormItem(attributeLabel, entry)
	markAsMenu.Items = append(markAsMenu.Items, fyne.NewMenuItem(newItemLabel, func() {
		title := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "mark_title", Other: "Add custom mark attribute"}})
		save := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "mark_button_save", Other: "Save"}})
		cancel := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "mark_button_cancel", Other: "Cancel"}})
		w := dialog.NewForm(title, save, cancel,
			[]*widget.FormItem{fi}, func(success bool) {
				if !success {
					return
				}
				if errMark := markFunc(sid64, []string{entry.Text}); errMark != nil {
					log.Printf("Failed to mark player: %v\n", errMark)
				}
			}, window)
		w.Show()
	}))
	markAsMenu.Refresh()
	return markAsMenu
}

func generateExternalLinksMenu(steamId steamid.SID64, links model.LinkConfigCollection, urlOpener func(url *url.URL) error) *fyne.Menu {
	lk := func(link *model.LinkConfig, sid64 steamid.SID64, urlOpener func(url *url.URL) error) func() {
		clsLinkValue := link
		clsSteamId := sid64
		return func() {
			u := clsLinkValue.URL
			switch model.SteamIdFormat(clsLinkValue.IdFormat) {
			case model.Steam:
				u = fmt.Sprintf(u, steamid.SID64ToSID(clsSteamId))
			case model.Steam3:
				u = fmt.Sprintf(u, steamid.SID64ToSID3(clsSteamId))
			case model.Steam32:
				u = fmt.Sprintf(u, steamid.SID64ToSID32(clsSteamId))
			case model.Steam64:
				u = fmt.Sprintf(u, clsSteamId.Int64())
			default:
				log.Printf("Got unhandled steamid format, trying steam64: %v", clsLinkValue.IdFormat)
			}
			ul, urlErr := url.Parse(u)
			if urlErr != nil {
				log.Printf("Failed to create link: %v", urlErr)
				return
			}
			if errOpen := urlOpener(ul); errOpen != nil {
				log.Printf("Failed to open url: %v", errOpen)
			}
		}
	}

	var items []*fyne.MenuItem
	sort.Slice(links, func(i, j int) bool {
		return strings.ToLower(links[i].Name) < strings.ToLower(links[j].Name)
	})
	for _, link := range links {
		if !link.Enabled {
			continue
		}
		items = append(items, fyne.NewMenuItem(link.Name, lk(link, steamId, urlOpener)))
	}
	return fyne.NewMenu("Sub Menu", items...)
}

func generateSteamIdMenu(window fyne.Window, steamId steamid.SID64) *fyne.Menu {
	title := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "menu_title_steam_id", Other: "Copy SteamID..."}})
	m := fyne.NewMenu(title,
		fyne.NewMenuItem(fmt.Sprintf("%d", steamId), func() {
			window.Clipboard().SetContent(fmt.Sprintf("%d", steamId))
		}),
		fyne.NewMenuItem(string(steamid.SID64ToSID(steamId)), func() {
			window.Clipboard().SetContent(string(steamid.SID64ToSID(steamId)))
		}),
		fyne.NewMenuItem(string(steamid.SID64ToSID3(steamId)), func() {
			window.Clipboard().SetContent(string(steamid.SID64ToSID3(steamId)))
		}),
		fyne.NewMenuItem(fmt.Sprintf("%d", steamid.SID64ToSID32(steamId)), func() {
			window.Clipboard().SetContent(fmt.Sprintf("%d", steamid.SID64ToSID32(steamId)))
		}),
	)
	return m
}

func generateKickMenu(ctx context.Context, userId int64, kickFunc model.KickFunc) *fyne.Menu {
	fn := func(reason model.KickReason) func() {
		return func() {
			log.Printf("Calling vote: %d %v", userId, reason)
			if errKick := kickFunc(ctx, userId, reason); errKick != nil {
				log.Printf("Error trying to call kick: %v\n", errKick)
			}
		}
	}
	title := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "menu_title_call_vote", Other: "Call Vote..."}})
	labelCheating := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "menu_call_vote_cheating", Other: "Cheating"}})
	labelIdle := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "menu_call_vote_idle", Other: "Idle"}})
	labelScamming := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "menu_call_vote_scamming", Other: "Scamming"}})
	labelOther := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "menu_call_vote_other", Other: "Other"}})
	return fyne.NewMenu(title,
		&fyne.MenuItem{Label: labelCheating, Action: fn(model.KickReasonCheating)},
		&fyne.MenuItem{Label: labelIdle, Action: fn(model.KickReasonIdle)},
		&fyne.MenuItem{Label: labelScamming, Action: fn(model.KickReasonScamming)},
		&fyne.MenuItem{Label: labelOther, Action: fn(model.KickReasonOther)},
	)
}

func generateUserMenu(ctx context.Context, app fyne.App, window fyne.Window, steamId steamid.SID64, userId int64, cb callBacks,
	knownAttributes binding.StringList, links model.LinkConfigCollection) *fyne.Menu {
	kickTitle := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "user_menu_call_vote", Other: "Call Vote..."}})
	markTitle := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "user_menu_mark", Other: "Mark As..."}})
	externalTitle := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "user_menu_external", Other: "Open External..."}})
	steamIdTitle := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "user_menu_steam_id", Other: "Copy SteamID..."}})
	chatHistoryTitle := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "user_menu_chat_hist", Other: "View Chat History"}})
	nameHistoryTitle := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "user_menu_name_hist", Other: "View Name History"}})
	whitelistTitle := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "user_menu_whitelist", Other: "Whitelist Player"}})
	notesTitle := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "user_menu_notes", Other: "Edit Notes"}})
	var items []*fyne.MenuItem
	if userId > 0 {
		items = append(items, &fyne.MenuItem{
			Icon:      theme.CheckButtonCheckedIcon(),
			ChildMenu: generateKickMenu(ctx, userId, cb.kickFunc),
			Label:     kickTitle})
	}
	items = append(items, []*fyne.MenuItem{
		{
			Icon:      theme.ZoomFitIcon(),
			ChildMenu: generateAttributeMenu(window, steamId, knownAttributes, cb.markFn),
			Label:     markTitle},
		{
			Icon:      theme.SearchIcon(),
			ChildMenu: generateExternalLinksMenu(steamId, links, app.OpenURL),
			Label:     externalTitle},
		{
			Icon:      theme.ContentCopyIcon(),
			ChildMenu: generateSteamIdMenu(window, steamId),
			Label:     steamIdTitle},
		{
			Icon: theme.ListIcon(),
			Action: func() {
				cb.createUserChat(steamId)
			},
			Label: chatHistoryTitle},
		{
			Icon: theme.VisibilityIcon(),
			Action: func() {
				cb.createNameHistory(steamId)
			},
			Label: nameHistoryTitle},
		{
			Icon: theme.VisibilityOffIcon(),
			Action: func() {
				if err := cb.whitelistFn(steamId); err != nil {
					showUserError(err, window)
				}
			},
			Label: whitelistTitle},
		{
			Icon: theme.DocumentCreateIcon(),
			Action: func() {
				offline := false
				player := cb.getPlayer(steamId)
				if player == nil {
					player = model.NewPlayer(steamId, "")
					if errOffline := cb.getPlayerOffline(ctx, steamId, player); errOffline != nil {
						showUserError(errors.Errorf("Unknown player: %v", errOffline), window)
						return
					}
					offline = true
				}
				entry := widget.NewMultiLineEntry()
				entry.SetMinRowsVisible(30)
				player.RLock()
				entry.SetText(player.Notes)
				player.RUnlock()
				item := widget.NewFormItem("", entry)
				sz := item.Widget.Size()
				sz.Height = sizeDialogueHeight
				item.Widget.Resize(sz)

				editNoteTitle := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "edit_note_title", Other: "Edit Player Notes"}})
				editNoteSave := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "edit_note_button_save", Other: "Save"}})
				editNoteCancel := tr.Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "edit_note_button_cancel", Other: "Cancel"}})

				d := dialog.NewForm(editNoteTitle, editNoteSave, editNoteCancel, []*widget.FormItem{item}, func(b bool) {
					if !b {
						return
					}
					player.Lock()
					player.Notes = entry.Text
					player.Touch()
					player.Unlock()
					if offline {
						if errSave := cb.savePlayer(ctx, player); errSave != nil {
							log.Printf("Failed to save: %v\n", errSave)
						}
					}

				}, window)
				d.Resize(window.Canvas().Size())
				d.Show()
			},
			Label: notesTitle},
	}...)
	menu := fyne.NewMenu("User Actions", items...)
	return menu
}
