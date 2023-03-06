package ui

import (
	"context"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/leighmacdonald/bd/internal/model"
	"github.com/leighmacdonald/bd/internal/translations"
	"github.com/pkg/errors"
	"sync"
	"time"
)

type searchWindow struct {
	fyne.Window
	ctx         context.Context
	app         fyne.App
	list        *widget.Table
	boundList   binding.ExternalUntypedList
	queryString binding.String
	objectMu    *sync.RWMutex
	boundListMu *sync.RWMutex
	resultCount binding.Int
	avatarCache *avatarCache
	queryEntry  *widget.Entry
	cb          callBacks
}

func (screen *searchWindow) Reload(results model.PlayerCollection) error {
	bl := results.AsAny()
	screen.boundListMu.Lock()
	if errSet := screen.boundList.Set(bl); errSet != nil {
		return errors.Wrapf(errSet, "failed to set player results")
	}
	if errReload := screen.boundList.Reload(); errReload != nil {
		return errors.Wrap(errReload, "Failed to reload results")
	}
	if errSet := screen.resultCount.Set(len(bl)); errSet != nil {
		return errors.Wrap(errSet, "Failed to set result count")
	}
	screen.boundListMu.Unlock()
	screen.Content().Refresh()
	return nil
}

func newSearchWindow(ctx context.Context, app fyne.App, cb callBacks, attrs binding.StringList, settings *model.Settings, cache *avatarCache) *searchWindow {
	window := app.NewWindow(translations.One(translations.WindowPlayerSearch))
	window.Canvas().AddShortcut(
		&desktop.CustomShortcut{KeyName: fyne.KeyW, Modifier: fyne.KeyModifierControl},
		func(shortcut fyne.Shortcut) {
			window.Hide()
		})
	window.SetCloseIntercept(func() {
		window.Hide()
	})

	sw := searchWindow{
		Window:      window,
		ctx:         ctx,
		app:         app,
		list:        nil,
		boundList:   binding.BindUntypedList(&[]interface{}{}),
		objectMu:    &sync.RWMutex{},
		boundListMu: &sync.RWMutex{},
		avatarCache: cache,
		queryString: binding.NewString(),
		cb:          cb,
		resultCount: binding.NewInt(),
	}

	sw.list = widget.NewTable(func() (int, int) {
		return sw.boundList.Length() + 1, 4
	}, func() fyne.CanvasObject {
		return container.NewMax(
			widget.NewLabel(""),
			widget.NewIcon(theme.ContentClearIcon()),
			widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			newContextMenuIcon())
	}, func(i widget.TableCellID, o fyne.CanvasObject) {
		sw.objectMu.Lock()
		defer sw.objectMu.Unlock()
		label := o.(*fyne.Container).Objects[0].(*widget.Label)
		icon := o.(*fyne.Container).Objects[1].(*widget.Icon)
		ctxMenu := o.(*fyne.Container).Objects[3].(*contextMenuIcon)
		if i.Row == 0 {
			switch i.Col {
			case 0:
				label.Show()
				icon.Hide()
				ctxMenu.Hide()
				label.TextStyle.Bold = true
				label.SetText("Last Seen")
			case 1:
				label.Hide()
				icon.Hide()
				ctxMenu.Hide()
			case 2:
				icon.Hide()
				ctxMenu.Hide()
				label.TextStyle.Bold = true
				label.SetText("Profile Name")
				label.Show()
			case 3:
				label.Hide()
				icon.Hide()
				ctxMenu.Hide()
			}
			return
		}
		value, valueErr := sw.boundList.GetValue(i.Row - 1)
		if valueErr != nil {
			return
		}
		ps := value.(*model.Player)
		label.Hide()
		icon.Hide()
		ctxMenu.Hide()
		switch i.Col {
		case 0:
			update := ps.UpdatedOn.Format(time.RFC822)
			label.Bind(binding.BindString(&update))
			label.Show()
		case 1:
			icon.Show()
			icon.SetResource(sw.avatarCache.GetAvatar(ps.SteamId))
		case 2:
			label.Bind(binding.BindString(&ps.Name))
			label.Show()
		case 3:
			ctxMenu.menu = generateUserMenu(sw.ctx, app, window, ps.SteamId, ps.UserId, cb, attrs, settings.GetLinks())
			ctxMenu.Show()
			//labelIcon.Refresh()
		}
	})

	sw.list.SetColumnWidth(0, 150)
	sw.list.SetColumnWidth(1, 40)
	sw.list.SetColumnWidth(2, 400)
	sw.list.SetColumnWidth(3, 40)
	//sw.list.SetColumnWidth(2, 200)
	sw.queryEntry = widget.NewEntryWithData(sw.queryString)
	sw.queryEntry.PlaceHolder = "SteamID or Name"
	sw.queryEntry.OnSubmitted = func(s string) {
		results, errSearch := cb.searchPlayer(sw.ctx, model.SearchOpts{Query: s})
		if errSearch != nil {
			showUserError(errSearch, window)
			return
		}
		if errReload := sw.Reload(results); errReload != nil {
			showUserError(errReload, sw.Window)
		}

	}
	//sw.list.Resize(fyne.NewSize(600, 500))
	sw.SetContent(container.NewBorder(
		container.NewBorder(
			nil,
			nil,
			nil,
			widget.NewLabelWithData(binding.IntToStringWithFormat(
				sw.resultCount,
				fmt.Sprintf("%s%%d", translations.One(translations.LabelResultCount)))),
			container.NewMax(sw.queryEntry),
		),
		nil, nil, nil,
		container.NewMax(sw.list)))
	sw.Window.Resize(fyne.NewSize(650, 700))

	return &sw
}
