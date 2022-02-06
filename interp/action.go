package interp

import (
	"context"
	"fmt"
)

type CmdStop struct{}

func (c CmdStop) Execute(ctx context.Context, d *RuntimeData) error {
	return ErrStop
}

type CmdFileInto struct {
	Mailbox string
}

func (c CmdFileInto) Execute(ctx context.Context, d *RuntimeData) error {
	found := false
	for _, m := range d.Mailboxes {
		if m == c.Mailbox {
			found = true
		}
	}
	if found {
		return nil
	}
	d.Mailboxes = append(d.Mailboxes, c.Mailbox)
	d.ImplicitKeep = false
	return nil
}

type CmdRedirect struct {
	Addr string
}

func (c CmdRedirect) Execute(ctx context.Context, d *RuntimeData) error {
	if d.Callback.RedirectAllowed != nil {
		ok, err := d.Callback.RedirectAllowed(ctx, d, c.Addr)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
	}
	d.RedirectAddr = append(d.RedirectAddr, c.Addr)
	d.ImplicitKeep = false

	if len(d.RedirectAddr) > d.Script.opts.MaxRedirects {
		return fmt.Errorf("too many actions")
	}
	return nil
}

type CmdKeep struct{}

func (c CmdKeep) Execute(_ context.Context, d *RuntimeData) error {
	d.Keep = true
	return nil
}

type CmdDiscard struct{}

func (c CmdDiscard) Execute(_ context.Context, d *RuntimeData) error {
	d.ImplicitKeep = false
	return nil
}
