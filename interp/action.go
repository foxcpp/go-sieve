package interp

import (
	"context"
	"fmt"
)

type CmdStop struct{}

func (c CmdStop) Execute(_ context.Context, _ *RuntimeData) error {
	return ErrStop
}

type CmdFileInto struct {
	Mailbox string
	Flags   Flags
}

func (c CmdFileInto) Execute(_ context.Context, d *RuntimeData) error {
	mailbox := expandVars(d, c.Mailbox)
	found := false
	for _, m := range d.Mailboxes {
		if m == mailbox {
			found = true
		}
	}
	if found {
		return nil
	}
	d.Mailboxes = append(d.Mailboxes, mailbox)
	d.ImplicitKeep = false
	if c.Flags != nil {
		d.Flags = canonicalFlags(expandVarsList(d, c.Flags), nil, d.FlagAliases)
	}
	return nil
}

type CmdRedirect struct {
	Addr string
}

func (c CmdRedirect) Execute(ctx context.Context, d *RuntimeData) error {
	addr := expandVars(d, c.Addr)

	ok, err := d.Policy.RedirectAllowed(ctx, d, addr)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	d.RedirectAddr = append(d.RedirectAddr, addr)
	d.ImplicitKeep = false

	if len(d.RedirectAddr) > d.Script.opts.MaxRedirects {
		return fmt.Errorf("too many actions")
	}
	return nil
}

type CmdKeep struct {
	Flags Flags
}

func (c CmdKeep) Execute(_ context.Context, d *RuntimeData) error {
	d.Keep = true
	if c.Flags != nil {
		d.Flags = canonicalFlags(expandVarsList(d, c.Flags), nil, d.FlagAliases)
	}
	return nil
}

type CmdDiscard struct{}

func (c CmdDiscard) Execute(_ context.Context, d *RuntimeData) error {
	d.ImplicitKeep = false
	d.Flags = make([]string, 0)
	return nil
}

type CmdSetFlag struct {
	Flags Flags
}

func (c CmdSetFlag) Execute(_ context.Context, d *RuntimeData) error {
	if c.Flags != nil {
		d.Flags = canonicalFlags(expandVarsList(d, c.Flags), nil, d.FlagAliases)
	}
	return nil
}

type CmdAddFlag struct {
	Flags Flags
}

func (c CmdAddFlag) Execute(_ context.Context, d *RuntimeData) error {
	if c.Flags != nil {
		flags := expandVarsList(d, c.Flags)

		if d.Flags == nil {
			d.Flags = make([]string, len(flags))
			copy(d.Flags, flags)
		} else {
			// Use canonicalFlags to remove duplicates
			d.Flags = canonicalFlags(append(d.Flags, flags...), nil, d.FlagAliases)
		}
	}
	return nil
}

type CmdRemoveFlag struct {
	Flags Flags
}

func (c CmdRemoveFlag) Execute(_ context.Context, d *RuntimeData) error {
	if c.Flags != nil {
		// Use canonicalFlags to remove duplicates
		d.Flags = canonicalFlags(d.Flags, expandVarsList(d, c.Flags), d.FlagAliases)
	}
	return nil
}
