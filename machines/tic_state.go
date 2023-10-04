package machines

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"

	"github.com/creativenucleus/bytejammer/embed"
)

type TicState struct {
	Code      []byte
	IsRunning bool
	CursorX   int
	CursorY   int
}

func MakeTicStateRunning(code []byte) TicState {
	return TicState{
		Code:      code,
		IsRunning: true,
	}
}

func MakeTicStateEditor(code []byte, cursorX int, cursorY int) TicState {
	return TicState{
		Code:    code,
		CursorX: cursorX,
		CursorY: cursorY,
	}
}

func MakeTicStateFromExportData(data []byte) (*TicState, error) {
	ts := TicState{}

	r := regexp.MustCompile(`(?s)^-- pos: (\d+),(\d+)\n(.*)$`)
	matches := r.FindStringSubmatch(string(data))

	ts.IsRunning = matches[1] == "0" && matches[1] == "0"
	if !ts.IsRunning {
		var err error
		ts.CursorX, err = strconv.Atoi(matches[1])
		if err != nil {
			return nil, err
		}
		ts.CursorY, err = strconv.Atoi(matches[2])
		if err != nil {
			return nil, err
		}
	}
	ts.Code = []byte(matches[3])

	return &ts, nil
}

func (ts TicState) GetCode() []byte {
	return ts.Code
}

func (ts *TicState) SetCode(code []byte) {
	ts.Code = code
}

func (ts TicState) GetIsRunning() bool {
	return ts.IsRunning
}

func (ts TicState) GetCursorX() int {
	return ts.CursorX
}

func (ts TicState) GetCursorY() int {
	return ts.CursorY
}

// Adds the control string
// This is --pos: 0,0 if running, otherwise --pos: X,Y (the cursor position)
func (ts TicState) MakeDataToImport() ([]byte, error) {
	controlString := "-- pos: 0,0\n" // Running
	if !ts.IsRunning {
		controlString = fmt.Sprintf("-- pos: %d,%d\n", ts.CursorX, ts.CursorY)
	}

	return append([]byte(controlString), ts.Code...), nil
}

func (ts1 TicState) IsEqual(ts2 TicState) bool {
	return bytes.Equal(ts1.Code, ts2.Code) &&
		ts1.IsRunning == ts2.IsRunning && ts1.CursorX == ts2.CursorX && ts1.CursorY == ts2.CursorY
}

func CodeAddAuthorShim(code []byte, author string) []byte {
	shim := CodeReplace(embed.LuaAuthorShim, map[string]string{"DISPLAY_NAME": author})
	return append(code, shim...)
}
