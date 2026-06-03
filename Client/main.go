package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/coder/websocket" // UPDATED: Coder package path handles deprecation notice
	//"github.com/coder/websocket/wsjson" // UPDATED: Coder JSON sub-package path
)

//go:embed Icon.png
var WordleIconRawBytes []byte
var mainWindow fyne.Window
var trueLayout *fyne.Container

var socket *websocket.Conn
var serverCallsChannel = make(chan []byte, 25)
var ctx = context.Background()

type Button struct {
	widget.Button
	CustomSize fyne.Size
}

func (b *Button) MinSize() fyne.Size {
	return b.CustomSize
}

func NewCustomButton(label string, width, height float32, tapped func()) *Button {
	var btn = &Button{CustomSize: fyne.NewSize(width, height)}
	btn.Text = label
	btn.OnTapped = tapped
	btn.ExtendBaseWidget(btn) // Critical for custom widgets in Fyne
	return btn
}

func padding(sizeX float32, sizeY float32) fyne.CanvasObject {
	var block = canvas.NewRectangle(color.Transparent)
	block.SetMinSize(fyne.NewSize(sizeX, sizeY))
	return block
}

func SignalServer(actionName string, params map[string]any) {
	if socket == nil {
		println("Cannot signal server. Connection not yet established!")
		return
	}
	if params == nil {
		params = make(map[string]any)
	}

	params["ActionName"] = actionName
	var payload, err = json.Marshal(params)
	if err != nil {
		println("Error marshaling JSON:", err.Error())
		return
	}

	select {
	case serverCallsChannel <- payload:
	default:
		println("Outbound channel buffer full!")
	}
	time.Sleep(1 * time.Millisecond)
}

func InitSocketWriting() {
	for payload := range serverCallsChannel {
		// FIXED: coder/websocket Write method takes (ctx, messageType, payload)
		var err = socket.Write(ctx, websocket.MessageText, payload)
		if err != nil {
			println("Error sending message to server:", err.Error())
		}
		time.Sleep(6 * time.Millisecond)
	}
}

func InitSocketListening() {
	defer socket.Close(websocket.StatusNormalClosure, "Closing")

	for {
		var _, incomingData, err = socket.Read(ctx)
		if err != nil {
			println("Socket read error or player disconnected:", err.Error())
			break
		}

		var payload = string(incomingData)
		text := canvas.NewText(payload, color.White)

		fyne.Do(func() {
			trueLayout.Add(text)
			trueLayout.Refresh()
		})

		time.Sleep(100 * time.Millisecond)
	}
}

func main() {
	wordleApp := app.New()

	mainWindow = wordleApp.NewWindow("Wordle PvP Client")
	mainWindow.Resize(fyne.NewSize(450, 250))

	// DISPLAYED LOGO
	var logoResource = fyne.NewStaticResource("Icon.png", WordleIconRawBytes)
	var logoImage = canvas.NewImageFromResource(logoResource)
	logoImage.SetMinSize(fyne.NewSize(130, 100))
	logoImage.FillMode = canvas.ImageFillContain

	// TITLE BELOW
	var title = canvas.NewText("Wordle PvP", theme.Color(theme.ColorNameForeground))
	title.TextSize = 24
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	var createLobbyBtn = NewCustomButton("Create Lobby", 150, 35, func() {
		SignalServer("Hi", nil)
	})

	var joinLobbyBtn = NewCustomButton("Join Lobby", 150, 35, func() {

	})

	trueLayout = container.NewVBox(
		container.NewBorder(logoImage, nil, nil, nil, title),
		padding(0, 20),
		container.NewCenter(container.NewVBox(
			createLobbyBtn,
			padding(0, 5),
			joinLobbyBtn,
		)),
	)
	mainWindow.SetContent(container.NewCenter(trueLayout))

	go func() {
		var err error
		println("Connecting to WebSocket server...")
		//socket, _, err = websocket.Dial(ctx, "ws://localhost:8080/ws", nil)
		socket, _, err = websocket.Dial(ctx, "wss://echo.websocket.org", nil)
		if err != nil {
			println("WebSocket establishment failed, error:", err.Error())
			return
		}
		println("WebSocket connected successfully!")

		go InitSocketListening()
		go InitSocketWriting()
	}()

	mainWindow.ShowAndRun()
}
