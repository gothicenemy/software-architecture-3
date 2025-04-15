package painter

import (
	"image"
	"image/color"
	"log"

	"golang.org/x/exp/shiny/screen"
)

type Receiver interface {
	Update(t screen.Texture)
}

type Operation interface {
	Do(state *LoopState) (requestUpdate bool)
}

type LoopState struct {
	Texture    screen.Texture
	Background color.Color
	BgRect     *RelativeRectangle
	Figures    []*Figure
	Screen     screen.Screen
	WindowSize image.Point
}

type Figure struct {
	X, Y int
}

type RelativeRectangle struct {
	X1, Y1, X2, Y2 float64
}

type Loop struct {
	Receiver Receiver
	State    *LoopState
	MsgQueue chan Operation
	stop     chan struct{}
	stopped  chan struct{}
}

func NewLoop(s screen.Screen) *Loop {
	l := &Loop{
		MsgQueue: make(chan Operation, 100),
		stop:     make(chan struct{}),
		stopped:  make(chan struct{}),
	}

	initialSize := image.Point{X: 800, Y: 800}

	l.State = &LoopState{
		Screen:     s,
		Background: color.RGBA{G: 0xff, A: 0xff},
		Figures: []*Figure{
			{X: initialSize.X / 2, Y: initialSize.Y / 2},
		},
		BgRect:     nil,
		WindowSize: initialSize,
	}

	l.resetTexture()
	return l
}

func (l *Loop) resetTexture() {
	if l.State.Texture != nil {
		l.State.Texture.Release()
	}
	var err error
	size := l.State.WindowSize
	if size.X == 0 || size.Y == 0 {
		size = image.Point{X: 800, Y: 800}
		l.State.WindowSize = size
	}
	l.State.Texture, err = l.State.Screen.NewTexture(size)
	if err != nil {
		log.Fatalf("Failed to create texture: %v", err)
	}
	log.Printf("Texture reset/created with size %dx%d", size.X, size.Y)
	l.drawCurrentState()
}

func (l *Loop) resetState() {
	l.State.Background = color.Black
	l.State.BgRect = nil
	l.State.Figures = make([]*Figure, 0)
	l.drawCurrentState()
}

func (l *Loop) drawCurrentState() {
	if l.State.Texture == nil {
		log.Println("Error: Loop.drawCurrentState: Texture is nil")
		return
	}
	state := l.State
	state.Texture.Fill(state.Texture.Bounds(), state.Background, screen.Src)
	if state.BgRect != nil {
		bounds := state.Texture.Bounds()
		width, height := float64(bounds.Dx()), float64(bounds.Dy())
		pxRect := image.Rect(
			int(state.BgRect.X1*width), int(state.BgRect.Y1*height),
			int(state.BgRect.X2*width), int(state.BgRect.Y2*height),
		)
		state.Texture.Fill(pxRect, color.Black, screen.Src)
	}
	for _, f := range state.Figures {
		l.drawFigure(state.Texture, f.X, f.Y)
	}
}

func (l *Loop) drawFigure(texture screen.Texture, x, y int) {
	bounds := texture.Bounds()

	figureWidth := bounds.Dx() / 2
	figureHeight := bounds.Dy() / 2
	if figureWidth < 20 {
		figureWidth = 20
	}
	if figureHeight < 20 {
		figureHeight = 20
	}

	lineWidth := figureHeight / 8
	if lineWidth < 2 {
		lineWidth = 2
	}

	vRect := image.Rect(
		x-figureWidth/2,
		y-figureHeight/2,
		x-figureWidth/2+lineWidth,
		y+figureHeight/2,
	)
	hRect := image.Rect(
		x-figureWidth/2+lineWidth,
		y-lineWidth/2,
		x+figureWidth/2,
		y+lineWidth/2,
	)

	figureColor := color.RGBA{R: 0xff, G: 0xff, B: 0x00, A: 0xff}

	texture.Fill(vRect, figureColor, screen.Src)
	texture.Fill(hRect, figureColor, screen.Src)
}

func (l *Loop) Start() {
	log.Println("Painter loop started")
	defer close(l.stopped)

	for {
		select {
		case <-l.stop:
			log.Println("Painter loop stopping...")
			if l.State.Texture != nil {
				l.State.Texture.Release()
				l.State.Texture = nil
			}
			return
		case op := <-l.MsgQueue:
			if l.State.Texture == nil && l.State.Screen != nil {
				log.Println("Warning: Texture was nil in loop, attempting recreate.")
				l.resetTexture()
				if l.State.Texture == nil {
					log.Println("Error: Failed to recreate texture, skipping operation.")
					continue
				}
			}

			updateRequested := op.Do(l.State)

			if _, isUpdate := op.(UpdateOperation); !isUpdate {
				l.drawCurrentState()
			}

			if updateRequested && l.Receiver != nil && l.State.Texture != nil {
				l.Receiver.Update(l.State.Texture)
			}
		}
	}
}

func (l *Loop) Stop() {
	log.Println("Requesting painter loop stop...")
	close(l.stop)
	<-l.stopped
	log.Println("Painter loop stopped.")
}

func (l *Loop) Post(op Operation) {
	select {
	case l.MsgQueue <- op:
	default:
		log.Println("Warning: Painter message queue full. Operation dropped.")
	}
}

func (l *Loop) SetReceiver(r Receiver) {
	l.Receiver = r
	if l.Receiver != nil {
		l.Post(UpdateOperation{})
	}
}
