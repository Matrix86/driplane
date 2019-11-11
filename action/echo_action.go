package action

import (
	"fmt"
	"github.com/Matrix86/driplane/com"
)

type EchoAction struct {
	ActionBase
}

func NewEchoAction(conf []string) (Action, error) {
	h := &EchoAction{}

	return h, nil
}

func (h *EchoAction) Name() string {
	return "echoaction"
}

func (h *EchoAction) DoAction(msg com.DataMessage) {
	fmt.Println("%v", msg)
}

func init() {
	register("echoaction", NewEchoAction)
}
