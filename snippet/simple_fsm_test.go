package snippet

import "testing"

func TestLightSwitchStateMachine(t *testing.T) {
	// Create a new instance of the light switch state machine.
	lightSwitchFsm := newLightSwitchFSM()

	// Set the initial "off" state in the state machine.
	err := lightSwitchFsm.SendEvent(SwitchOff, nil)
	if err != nil {
		t.Errorf("Couldn't set the initial state of the state machine, err: %v", err)
	}

	// Send the switch-off event again and expect the state machine to return an error.
	err = lightSwitchFsm.SendEvent(SwitchOff, nil)
	if err != ErrEventRejected {
		t.Errorf("Expected the event rejected error, got nil")
	}

	// Send the switch-on event and expect the state machine to transition to the
	// "on" state.
	err = lightSwitchFsm.SendEvent(SwitchOn, nil)
	if err != nil {
		t.Errorf("Couldn't switch the light on, err: %v", err)
	}

	// Send the switch-on event again and expect the state machine to return an error.
	err = lightSwitchFsm.SendEvent(SwitchOn, nil)
	if err != ErrEventRejected {
		t.Errorf("Expected the event rejected error, got nil")
	}

	// Send the switch-off event and expect the state machine to transition back
	// to the "off" state.
	err = lightSwitchFsm.SendEvent(SwitchOff, nil)
	if err != nil {
		t.Errorf("Couldn't switch the light off, err: %v", err)
	}
}
