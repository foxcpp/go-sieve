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
	Copy    bool
}

func (ActionFileInto) testActionName() string            { return "fileinto" }
func (a ActionFileInto) cancelsImplicitKeep() bool { return !a.Copy }

type ActionRedirect struct {
	Address string
	Copy    bool
}

func (ActionRedirect) testActionName() string              { return "redirect" }
func (a ActionRedirect) cancelsImplicitKeep() bool { return !a.Copy }

type ActionReject struct {
	Reason string
}

func (ActionReject) testActionName() string    { return "reject" }
func (ActionReject) cancelsImplicitKeep() bool { return true }

type ActionEReject struct {
	Reason string
}

func (ActionEReject) testActionName() string    { return "ereject" }
func (ActionEReject) cancelsImplicitKeep() bool { return true }
