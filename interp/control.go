package interp

import (
	"context"
)

type CmdIf struct {
	Test  Test
	Block []Cmd
}

func (c CmdIf) Execute(ctx context.Context, d *RuntimeData) error {
	res, err := c.Test.Check(ctx, d)
	if err != nil {
		return err
	}
	if res {
		for _, c := range c.Block {
			if err := c.Execute(ctx, d); err != nil {
				return err
			}
		}
	}
	d.ifResult = res
	return nil
}

type CmdElsif struct {
	Test  Test
	Block []Cmd
}

func (c CmdElsif) Execute(ctx context.Context, d *RuntimeData) error {
	if d.ifResult {
		return nil
	}
	res, err := c.Test.Check(ctx, d)
	if err != nil {
		return err
	}
	if res {
		for _, c := range c.Block {
			if err := c.Execute(ctx, d); err != nil {
				return err
			}
		}
	}
	d.ifResult = res
	return nil
}

type CmdElse struct {
	Block []Cmd
}

func (c CmdElse) Execute(ctx context.Context, d *RuntimeData) error {
	if d.ifResult {
		return nil
	}
	for _, c := range c.Block {
		if err := c.Execute(ctx, d); err != nil {
			return err
		}
	}
	return nil
}
