package lang_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gothicenemy/software-architecture-3/painter"
	"github.com/gothicenemy/software-architecture-3/painter/lang"
)

func TestParser_Parse_ValidCommands(t *testing.T) {
	p := &lang.Parser{}
	input := `
# This is a comment
white
green

bgrect 0.1 0.2 0.8 0.9
figure 0.5 0.5
move 0.1 0.1
update
reset
figure 0.0 1.0 # Coordinates within range
bgrect 0 0 1 1
`
	reader := strings.NewReader(input)
	ops, err := p.Parse(reader)

	require.NoError(t, err)
	require.Len(t, ops, 9)

	assert.IsType(t, painter.WhiteOperation{}, ops[0])
	assert.IsType(t, painter.GreenOperation{}, ops[1])
	assert.IsType(t, painter.BgRectOperation{}, ops[2])
	if bgrectOp, ok := ops[2].(painter.BgRectOperation); ok {
		assert.InDelta(t, 0.1, bgrectOp.X1, 0.001)
		assert.InDelta(t, 0.2, bgrectOp.Y1, 0.001)
		assert.InDelta(t, 0.8, bgrectOp.X2, 0.001)
		assert.InDelta(t, 0.9, bgrectOp.Y2, 0.001)
	}
	assert.IsType(t, painter.FigureOperation{}, ops[3])
	if figureOp, ok := ops[3].(painter.FigureOperation); ok {
		assert.InDelta(t, 0.5, figureOp.X, 0.001)
		assert.InDelta(t, 0.5, figureOp.Y, 0.001)
	}
	assert.IsType(t, painter.MoveOperation{}, ops[4])
	if moveOp, ok := ops[4].(painter.MoveOperation); ok {
		assert.InDelta(t, 0.1, moveOp.X, 0.001)
		assert.InDelta(t, 0.1, moveOp.Y, 0.001)
	}
	assert.IsType(t, painter.UpdateOperation{}, ops[5])
	assert.IsType(t, painter.ResetOperation{}, ops[6])
	assert.IsType(t, painter.FigureOperation{}, ops[7])
	if figureOp, ok := ops[7].(painter.FigureOperation); ok {
		assert.InDelta(t, 0.0, figureOp.X, 0.001)
		assert.InDelta(t, 1.0, figureOp.Y, 0.001)
	}
	assert.IsType(t, painter.BgRectOperation{}, ops[8])
	if bgrectOp, ok := ops[8].(painter.BgRectOperation); ok {
		assert.InDelta(t, 0.0, bgrectOp.X1, 0.001)
		assert.InDelta(t, 0.0, bgrectOp.Y1, 0.001)
		assert.InDelta(t, 1.0, bgrectOp.X2, 0.001)
		assert.InDelta(t, 1.0, bgrectOp.Y2, 0.001)
	}
}

func TestParser_Parse_InvalidCommands(t *testing.T) {
	p := &lang.Parser{}
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"Empty", "", 0}, {"Only Comment", "# white", 0}, {"Unknown Command", "red", 0},
		{"White With Arg", "white 0.5", 0}, {"Green With Arg", "green 1", 0},
		{"Update With Arg", "update status", 0}, {"Reset With Arg", "reset now", 0},
		{"BgRect Wrong Arg Count", "bgrect 0.1 0.2 0.3", 0}, {"BgRect Wrong Arg Type", "bgrect 0.1 0.2 text 0.4", 0},
		{"BgRect Out Of Range", "bgrect -0.1 0.2 0.8 1.1", 1},
		{"Figure Wrong Arg Count", "figure 0.5", 0}, {"Figure Wrong Arg Type", "figure text 0.5", 0},
		{"Figure Out Of Range", "figure 1.2 0.5", 1},
		{"Move Wrong Arg Count", "move 0.5", 0}, {"Move Wrong Arg Type", "move 0.5 text", 0},
		{"Move Out Of Range", "move 0.5 -0.2", 1},
		{"Mixed Valid Invalid", "white\nfigure 0.1\nupdate", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			ops, err := p.Parse(reader)
			require.NoError(t, err)
			assert.Len(t, ops, tt.want)
		})
	}
}
