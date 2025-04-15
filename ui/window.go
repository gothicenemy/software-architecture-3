package ui

import (
	"github.com/gothicenemy/software-architecture-3/painter"
	"image"
	"image/draw"
	"log"

	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
)

type Window struct {
	Title            string
	Debug            bool
	window           screen.Window
	events           chan interface{}
	tx               chan screen.Texture
	closeReq         chan struct{}
	closed           chan struct{}
	windowSize       size.Event
	figureX, figureY int
	painterLoop      *painter.Loop
}

const (
	WindowWidth  = 800
	WindowHeight = 800
)

func NewWindow(s screen.Screen, p *painter.Loop) *Window {
	log.Println("Creating UI Window...")
	win, err := s.NewWindow(&screen.NewWindowOptions{
		Title:  "Painter Final",
		Width:  WindowWidth,
		Height: WindowHeight,
	})
	if err != nil {
		log.Fatalf("!!! CRITICAL: Failed to create shiny window in ui.NewWindow: %v", err)
		return nil
	}
	log.Println("UI Shiny window created.")

	w := &Window{
		Title:       "Painter Final",
		Debug:       true,
		window:      win,
		events:      make(chan interface{}),
		tx:          make(chan screen.Texture),
		closeReq:    make(chan struct{}),
		closed:      make(chan struct{}),
		figureX:     WindowWidth / 2,
		figureY:     WindowHeight / 2,
		windowSize:  size.Event{WidthPx: WindowWidth, HeightPx: WindowHeight},
		painterLoop: p,
	}

	if w.painterLoop != nil {
		w.painterLoop.SetReceiver(w)
	} else {
		log.Println("Warning: ui.NewWindow received nil painterLoop")
	}

	go w.eventReader()

	return w
}

func (w *Window) eventReader() {
	defer close(w.closed)
	for {
		e := w.window.NextEvent()
		if w.events == nil {
			log.Println("Event channel is nil, stopping reader.")
			return
		}
		select {
		case w.events <- e:
		default:
			if _, ok := e.(paint.Event); !ok {
				log.Printf("Event queue full or closed, dropping event: %T", e)
			}
		}
		if lcEvent, ok := e.(lifecycle.Event); ok && lcEvent.To == lifecycle.StageDead {
			log.Println("Lifecycle dead in eventReader, stopping.")
			return
		}
		select {
		case <-w.closeReq:
			log.Println("Close request received in eventReader, stopping.")
			return
		default:
		}
	}
}

func (w *Window) Loop() {
	if w.window == nil {
		log.Println("Error: Window loop started with nil window.")
		close(w.closed)
		return
	}
	log.Println("Window event loop started.")

	if w.painterLoop != nil {
		w.painterLoop.Post(painter.UpdateOperation{})
	}

	for {
		select {
		case e, ok := <-w.events:
			if !ok {
				log.Println("Event channel closed, exiting Window loop.")
				return
			}
			if w.handleEvent(e) {
				return
			}
		case t, ok := <-w.tx:
			if !ok {
				log.Println("Texture channel closed.")
				continue
			}
			if w.window != nil {
				windowBounds := image.Rect(0, 0, w.windowSize.WidthPx, w.windowSize.HeightPx)
				w.window.Scale(windowBounds, t, t.Bounds(), draw.Src, nil)
				w.window.Publish()
			} else {
				t.Release()
			}
		case <-w.closeReq:
			log.Println("Close requested, exiting Window loop.")
			return
		}
	}
}

func (w *Window) handleEvent(e interface{}) bool {
	if w.Debug {
		log.Printf("event: %T", e)
	}
	switch ev := e.(type) {
	case lifecycle.Event:
		if ev.To == lifecycle.StageDead {
			log.Println("Lifecycle dead received in handleEvent")
			return true
		}
	case key.Event:
		if ev.Code == key.CodeEscape {
			log.Println("Escape pressed received in handleEvent")
			return true
		}
	case mouse.Event:
		if ev.Button == mouse.ButtonLeft && ev.Direction == mouse.DirPress {
			w.figureX = int(ev.X)
			w.figureY = int(ev.Y)
			if w.painterLoop != nil && w.windowSize.WidthPx > 0 && w.windowSize.HeightPx > 0 {
				relX := float64(ev.X) / float64(w.windowSize.WidthPx)
				relY := float64(ev.Y) / float64(w.windowSize.HeightPx)
				if relX < 0 {
					relX = 0
				}
				if relX > 1 {
					relX = 1
				}
				if relY < 0 {
					relY = 0
				}
				if relY > 1 {
					relY = 1
				}
				cmd := painter.MoveOperation{X: relX, Y: relY}
				w.painterLoop.Post(cmd)
				w.painterLoop.Post(painter.UpdateOperation{})
				log.Printf("Sent 'move %.2f %.2f' and 'update' on click", relX, relY)
			}
		}
	case size.Event:
		w.windowSize = ev
		log.Printf("Resized to: %dx%d", ev.WidthPx, ev.HeightPx)
	case paint.Event:
		if w.Debug {
			log.Println("Paint event")
		}
	case error:
		log.Printf("System error event: %v", e)
	}
	return false
}

func (w *Window) Update(t screen.Texture) {
	select {
	case w.tx <- t:
	default:
		log.Println("Warning: UI loop busy, skipping texture frame.")
	}
}

func (w *Window) Stop() {
	log.Println("Stop requested for Window")
	select {
	case w.closeReq <- struct{}{}:
	default:
		log.Println("Close request already pending.")
	}
}

func (w *Window) Closed() <-chan struct{} { return w.closed }
