package painter

import (
	"image"
	"image/color"
	"image/draw"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"golang.org/x/exp/shiny/screen"
)

type MockScreen struct {
	mock.Mock
}

func (m *MockScreen) NewTexture(size image.Point) (screen.Texture, error) {
	args := m.Called(size)
	if tex := args.Get(0); tex != nil {
		return tex.(screen.Texture), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockScreen) NewWindow(opts *screen.NewWindowOptions) (screen.Window, error) {
	args := m.Called(opts)
	if win := args.Get(0); win != nil {
		return win.(screen.Window), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockScreen) NewBuffer(size image.Point) (screen.Buffer, error) {
	args := m.Called(size)
	if buf := args.Get(0); buf != nil {
		return buf.(screen.Buffer), args.Error(1)
	}
	return nil, args.Error(1)
}

type MockTexture struct {
	mock.Mock
}

func (m *MockTexture) Release()          { m.Called() }
func (m *MockTexture) Size() image.Point { args := m.Called(); return args.Get(0).(image.Point) }
func (m *MockTexture) Bounds() image.Rectangle {
	args := m.Called()
	return args.Get(0).(image.Rectangle)
}
func (m *MockTexture) Upload(dp image.Point, src screen.Buffer, sr image.Rectangle) {
	m.Called(dp, src, sr)
}
func (m *MockTexture) Fill(dr image.Rectangle, src color.Color, op draw.Op) {
	m.Called(dr, src, op)
}

type MockReceiver struct {
	mock.Mock
	LastTexture screen.Texture
}

func (m *MockReceiver) Update(t screen.Texture) {
	m.Called(t)
}

func TestLoop_Initialization(t *testing.T) {
	mockScreen := new(MockScreen)
	mockTexture := new(MockTexture)
	size := image.Point{X: 800, Y: 800}
	initialBgColor := color.RGBA{G: 0xff, A: 0xff}
	initialFigureColor := color.RGBA{R: 0xff, G: 0xff, B: 0x00, A: 0xff}

	mockTexture.On("Release").Maybe()
	mockTexture.On("Bounds").Return(image.Rectangle{Max: size})
	mockTexture.On("Fill", mock.AnythingOfType("image.Rectangle"), initialBgColor, draw.Src).Return().Once()
	mockTexture.On("Fill", mock.AnythingOfType("image.Rectangle"), initialFigureColor, draw.Src).Return().Times(2)
	mockScreen.On("NewTexture", size).Return(mockTexture, nil).Once()

	l := NewLoop(mockScreen)

	require.NotNil(t, l)
	require.NotNil(t, l.State)
	assert.NotNil(t, l.State.Texture)
	assert.Equal(t, initialBgColor, l.State.Background)
	assert.Len(t, l.State.Figures, 1)
	if len(l.State.Figures) == 1 {
		assert.Equal(t, 400, l.State.Figures[0].X)
		assert.Equal(t, 400, l.State.Figures[0].Y)
	}
	assert.Nil(t, l.State.BgRect)
	assert.Equal(t, size, l.State.WindowSize)

	mockScreen.AssertExpectations(t)
	mockTexture.AssertExpectations(t)
}

func TestLoop_PostAndProcess(t *testing.T) {
	mockScreen := new(MockScreen)
	mockReceiver := new(MockReceiver)
	mockTexture := new(MockTexture)
	size := image.Point{X: 800, Y: 800}

	mockTexture.On("Release").Maybe()
	mockTexture.On("Bounds").Return(image.Rectangle{Max: size})
	mockTexture.On("Fill", mock.AnythingOfType("image.Rectangle"), mock.Anything, draw.Src).Return().Maybe()

	mockScreen.On("NewTexture", size).Return(mockTexture, nil).Once()
	mockReceiver.On("Update", mockTexture).Return().Times(2)

	l := NewLoop(mockScreen)
	l.SetReceiver(mockReceiver)
	go l.Start()
	time.Sleep(50 * time.Millisecond)

	l.Post(GreenOperation{})
	l.Post(FigureOperation{X: 0.5, Y: 0.5})
	l.Post(UpdateOperation{})
	time.Sleep(100 * time.Millisecond)

	expectedBg := color.RGBA{G: 0xff, A: 0xff}
	assert.Equal(t, expectedBg, l.State.Background)

	require.Len(t, l.State.Figures, 2)
	assert.Equal(t, 400, l.State.Figures[1].X)
	assert.Equal(t, 400, l.State.Figures[1].Y)

	mockReceiver.AssertExpectations(t)

	l.Stop()
	mockScreen.AssertExpectations(t)
	mockTexture.AssertCalled(t, "Release")
}

func TestLoop_ResetOperation(t *testing.T) {
	mockScreen := new(MockScreen)
	mockReceiver := new(MockReceiver)
	mockTexture := new(MockTexture)
	size := image.Point{X: 800, Y: 800}

	mockTexture.On("Release").Maybe()
	mockTexture.On("Bounds").Return(image.Rectangle{Max: size})
	mockTexture.On("Fill", mock.AnythingOfType("image.Rectangle"), mock.Anything, draw.Src).Return().Maybe()

	mockScreen.On("NewTexture", size).Return(mockTexture, nil).Once()
	mockReceiver.On("Update", mockTexture).Return().Times(2)

	l := NewLoop(mockScreen)
	l.SetReceiver(mockReceiver)
	go l.Start()
	time.Sleep(50 * time.Millisecond)

	l.Post(GreenOperation{})
	l.Post(FigureOperation{X: 0.2, Y: 0.2})
	l.Post(BgRectOperation{X1: 0.1, Y1: 0.1, X2: 0.9, Y2: 0.9})
	time.Sleep(50 * time.Millisecond)

	require.NotEqual(t, color.Black, l.State.Background)
	require.NotEmpty(t, l.State.Figures)
	require.NotNil(t, l.State.BgRect)

	l.Post(ResetOperation{})
	l.Post(UpdateOperation{})
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, color.Black, l.State.Background)
	assert.Empty(t, l.State.Figures)
	assert.Nil(t, l.State.BgRect)

	mockReceiver.AssertExpectations(t)

	l.Stop()
	mockScreen.AssertExpectations(t)
	mockTexture.AssertCalled(t, "Release")
}
