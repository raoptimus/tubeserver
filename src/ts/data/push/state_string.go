// generated by stringer -type=State; DO NOT EDIT

package push

import "fmt"

const _State_name = "StateWaitStateInProgressStateErrorStateFinish"

var _State_index = [...]uint8{0, 9, 24, 34, 45}

func (i State) String() string {
	if i < 0 || i+1 >= State(len(_State_index)) {
		return fmt.Sprintf("State(%d)", i)
	}
	return _State_name[_State_index[i]:_State_index[i+1]]
}
