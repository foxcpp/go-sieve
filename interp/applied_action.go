package interp

type AppliedAction interface {
	testActionName() string
	cancelsImplicitKeep() bool
}

type ActionKeep struct {
	Implicit bool
	Flags    Flags
}

func (ActionKeep) testActionName() string    { return "keep" }
func (ActionKeep) cancelsImplicitKeep() bool { return true }

type ActionDiscard struct{}

func (ActionDiscard) testActionName() string    { return "discard" }
func (ActionDiscard) cancelsImplicitKeep() bool { return true }

type ActionFileInto struct {
	Mailbox string
	Flags   Flags
}

func (ActionFileInto) testActionName() string    { return "fileinto" }
func (ActionFileInto) cancelsImplicitKeep() bool { return true }

type ActionRedirect struct {
	Address string
}

func (ActionRedirect) testActionName() string    { return "redirect" }
func (ActionRedirect) cancelsImplicitKeep() bool { return true }
