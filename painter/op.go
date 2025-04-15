package painter

import (
	"image/color"
	"log"
	"strconv"
)

type WhiteOperation struct{}

func (o WhiteOperation) Do(state *LoopState) bool {
	state.Background = color.White
	log.Println("Background set to white")
	return false
}

type GreenOperation struct{}

func (o GreenOperation) Do(state *LoopState) bool {
	state.Background = color.RGBA{G: 0xff, A: 0xff}
	log.Println("Background set to green")
	return false
}

type UpdateOperation struct{}

func (o UpdateOperation) Do(_ *LoopState) bool {
	log.Println("Update operation triggered: Requesting screen update")
	return true
}

type BgRectOperation struct {
	X1, Y1, X2, Y2 float64
}

func (o BgRectOperation) Do(state *LoopState) bool {
	state.BgRect = &RelativeRectangle{
		X1: o.X1, Y1: o.Y1, X2: o.X2, Y2: o.Y2,
	}
	log.Printf("Background rectangle set to: [%.2f, %.2f] -> [%.2f, %.2f]", o.X1, o.Y1, o.X2, o.Y2)
	return false
}

type FigureOperation struct {
	X, Y float64
}

func (o FigureOperation) Do(state *LoopState) bool {
	if state.Texture == nil {
		log.Println("Error: Cannot add figure, texture is nil")
		return false
	}
	bounds := state.Texture.Bounds()
	pxX := int(o.X * float64(bounds.Dx()))
	pxY := int(o.Y * float64(bounds.Dy()))
	state.Figures = append(state.Figures, &Figure{X: pxX, Y: pxY})
	log.Printf("Figure added at relative: %.2f, %.2f (pixels: %d, %d)", o.X, o.Y, pxX, pxY)
	return false
}

type MoveOperation struct {
	X, Y float64
}

func (o MoveOperation) Do(state *LoopState) bool {
	if state.Texture == nil {
		log.Println("Error: Cannot move figures, texture is nil")
		return false
	}
	bounds := state.Texture.Bounds()
	newX := int(o.X * float64(bounds.Dx()))
	newY := int(o.Y * float64(bounds.Dy()))
	log.Printf("Moving all %d figures to relative: %.2f, %.2f (pixels: %d, %d)", len(state.Figures), o.X, o.Y, newX, newY)
	for _, f := range state.Figures {
		f.X = newX
		f.Y = newY
	}
	return false
}

type ResetOperation struct{}

func (o ResetOperation) Do(state *LoopState) bool {
	log.Println("Resetting state to default (black background, no figures)")
	state.Background = color.Black
	state.BgRect = nil
	state.Figures = make([]*Figure, 0)
	return false
}

func ParseCoords(args []string, count int) ([]float64, bool) {
	if len(args) != count {
		log.Printf("Error: Expected %d coordinate arguments, got %d: %v", count, len(args), args)
		return nil, false
	}
	coords := make([]float64, count)
	var err error
	for i, arg := range args {
		coords[i], err = strconv.ParseFloat(arg, 64)
		if err != nil {
			log.Printf("Error: Could not parse coordinate '%s': %v", arg, err)
			return nil, false
		}
		if coords[i] < 0 || coords[i] > 1 {
			if coords[i] < 0 {
				coords[i] = 0
			}
			if coords[i] > 1 {
				coords[i] = 1
			}
			log.Printf("Warning: Coordinate clamped to %.3f", coords[i])
		}
	}
	return coords, true
}
