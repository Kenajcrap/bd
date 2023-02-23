package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/leighmacdonald/bd/model"
	"log"
	"sync"
	"time"
)

type userMessageList struct {
	list        *widget.List
	boundList   binding.ExternalUntypedList
	content     fyne.CanvasObject
	objectMu    sync.RWMutex
	boundListMu sync.RWMutex
}

func (chatList *userMessageList) Reload(rr []model.UserMessage) error {
	bl := make([]interface{}, len(rr))
	for i, r := range rr {
		bl[i] = r
	}
	chatList.boundListMu.Lock()
	defer chatList.boundListMu.Unlock()
	if errSet := chatList.boundList.Set(bl); errSet != nil {
		log.Printf("failed to set player list: %v\n", errSet)
	}
	if errReload := chatList.boundList.Reload(); errReload != nil {
		return errReload
	}
	chatList.list.ScrollToBottom()
	return nil
}

func (chatList *userMessageList) Append(msg model.UserMessage) error {
	chatList.boundListMu.Lock()
	defer chatList.boundListMu.Unlock()
	if errSet := chatList.boundList.Append(msg); errSet != nil {
		log.Printf("failed to append message: %v\n", errSet)
	}
	if errReload := chatList.boundList.Reload(); errReload != nil {
		log.Printf("Failed to update chat list: %v\n", errReload)
	}
	chatList.list.ScrollToBottom()
	return nil
}

// Widget returns the actual select list widget.
func (chatList *userMessageList) Widget() *widget.List {
	return chatList.list
}

func (ui *Ui) createGameChatMessageList() *userMessageList {
	uml := &userMessageList{}
	boundList := binding.BindUntypedList(&[]interface{}{})
	userMessageListWidget := widget.NewListWithData(
		boundList,
		func() fyne.CanvasObject {
			return container.NewBorder(
				nil,
				nil,
				container.NewHBox(widget.NewLabel(""), newContextMenuRichText(nil)),
				nil,
				widget.NewRichTextWithText(""))
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			value := i.(binding.Untyped)
			obj, _ := value.Get()
			um := obj.(model.UserMessage)
			uml.objectMu.Lock()
			rootContainer := o.(*fyne.Container)
			timeAndProfileContainer := rootContainer.Objects[1].(*fyne.Container)
			timeStamp := timeAndProfileContainer.Objects[0].(*widget.Label)
			profileButton := timeAndProfileContainer.Objects[1].(*contextMenuRichText)
			messageRichText := rootContainer.Objects[0].(*widget.RichText)

			timeStamp.SetText(um.Created.Format(time.Kitchen))
			profileButton.SetText(um.Player)
			sz := profileButton.Size()
			sz.Width = 200
			profileButton.Resize(sz)
			profileButton.menu = ui.generateUserMenu(um.PlayerSID, um.UserId)
			profileButton.menu.Refresh()
			profileButton.Refresh()

			messageRichText.Segments[0] = &widget.TextSegment{
				Style: widget.RichTextStyleInline,
				Text:  um.Message,
			}
			messageRichText.Refresh()

			uml.objectMu.Unlock()
		})
	uml.list = userMessageListWidget
	uml.boundList = boundList
	uml.content = container.NewVScroll(userMessageListWidget)
	return uml
}

func (ui *Ui) createUserHistoryMessageList() *userMessageList {
	uml := &userMessageList{}
	boundList := binding.BindUntypedList(&[]interface{}{})
	userMessageListWidget := widget.NewListWithData(
		boundList,
		func() fyne.CanvasObject {
			return container.NewBorder(
				nil,
				nil,
				widget.NewLabel(""),
				nil,
				widget.NewRichTextWithText(""))
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			value := i.(binding.Untyped)
			obj, _ := value.Get()
			um := obj.(model.UserMessage)
			uml.objectMu.Lock()
			rootContainer := o.(*fyne.Container)
			timeStamp := rootContainer.Objects[1].(*widget.Label)
			timeStamp.SetText(um.Created.Format(time.RFC822))
			messageRichText := rootContainer.Objects[0].(*widget.RichText)
			messageRichText.Segments[0] = &widget.TextSegment{
				Style: widget.RichTextStyleInline,
				Text:  um.Message,
			}
			messageRichText.Refresh()
			uml.objectMu.Unlock()
		})
	uml.list = userMessageListWidget
	uml.boundList = boundList
	uml.content = container.NewVScroll(userMessageListWidget)
	return uml
}

func (ui *Ui) createChatWidget(msgList *userMessageList) fyne.Window {
	chatWindow := ui.application.NewWindow("Chat History")
	chatWindow.SetIcon(resourceIconPng)
	chatWindow.SetContent(msgList.Widget())
	chatWindow.Resize(fyne.NewSize(1000, 500))
	chatWindow.SetCloseIntercept(func() {
		chatWindow.Hide()
	})

	return chatWindow
}
