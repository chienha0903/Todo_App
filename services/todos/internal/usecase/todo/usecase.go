package todo

import (
	"context"
	"time"

	"github.com/chienha0903/Todo_App/services/todos/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/todos/internal/domain/gateway"
	vo "github.com/chienha0903/Todo_App/services/todos/internal/domain/valueobject"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/input"
	"github.com/chienha0903/Todo_App/services/todos/internal/usecase/todo/output"
	"github.com/google/wire"
)

// TodoUsecase defines business operations for Todo.
type TodoUsecase interface {
	CreateTodo(ctx context.Context, in input.CreateTodoInput) (*output.TodoOutput, error)
}

// ProviderSet registers this usecase for Wire.
var ProviderSet = wire.NewSet(NewTodoUsecase)

type todoUsecase struct {
	cmdGW gateway.TodoCommandGateway
	qryGW gateway.TodoQueryGateway
}

func NewTodoUsecase(cmdGW gateway.TodoCommandGateway, qryGW gateway.TodoQueryGateway) TodoUsecase {
	return &todoUsecase{cmdGW: cmdGW, qryGW: qryGW}
}

func (uc *todoUsecase) CreateTodo(ctx context.Context, in input.CreateTodoInput) (*output.TodoOutput, error) {
	title, err := vo.NewTodoTitle(in.Title)
	if err != nil {
		return nil, err
	}

	desc, err := vo.NewTodoDescription(in.Description)
	if err != nil {
		return nil, err
	}

	priority, err := vo.NewTodoPriority(in.Priority)
	if err != nil {
		return nil, err
	}

	var dueDate *vo.TodoDueDate
	if in.DueDate != nil {
		dd, err := vo.NewTodoDueDate(*in.DueDate)
		if err != nil {
			return nil, err
		}
		dueDate = &dd
	}

	now := time.Now()
	todo := &entity.Todo{
		UserID:      entity.UserID(in.UserID),
		Title:       title,
		Description: desc,
		Status:      vo.TODO_STATUS_PENDING,
		Priority:    priority,
		DueDate:     dueDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.cmdGW.CreateTodo(ctx, todo); err != nil {
		return nil, err
	}

	return toOutput(todo), nil
}

func toOutput(t *entity.Todo) *output.TodoOutput {
	out := &output.TodoOutput{
		ID:          int64(t.ID),
		UserID:      int64(t.UserID),
		Title:       t.Title.Value(),
		Description: t.Description.Value(),
		Status:      t.Status.String(),
		Priority:    t.Priority.String(),
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
	if t.DueDate != nil {
		v := t.DueDate.Value()
		out.DueDate = &v
	}
	return out
}
