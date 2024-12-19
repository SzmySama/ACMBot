package helper

type ChainContext[T any] struct {
	ctx   *T
	tasks []task[T]
}

type task[T any] func(*T) error

func NewChainContext[T any](ctx *T) *ChainContext[T] {
	return &ChainContext[T]{ctx: ctx}
}

func (c *ChainContext[T]) Then(task task[T]) *ChainContext[T] {
	c.tasks = append(c.tasks, task)
	return c
}

func (c *ChainContext[T]) Execute() error {
	ctx := c.ctx
	var err error
	for _, task := range c.tasks {
		err = task(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
