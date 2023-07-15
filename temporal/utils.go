package temporal

import (
	"context"
	"errors"
	"fmt"
	"github.com/danthegoodman1/GoAPITemplate/utils"
	"strings"
	"time"

	"go.temporal.io/sdk/workflow"
)

func execActivityIO[Tin any, Tout any](ctx workflow.Context, activity func(ctx context.Context, params Tin) (res Tout, err error), input Tin, scheduleToClose time.Duration) (Tout, error) {
	if scheduleToClose != 0 {
		ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			ScheduleToCloseTimeout: scheduleToClose,
		})
	}
	f := workflow.ExecuteActivity(ctx, activity, input)
	var res Tout
	err := f.Get(ctx, &res)
	if err != nil {
		return res, fmt.Errorf("error in activity '%s': %w", utils.FuncName(activity), err)
	}
	return res, nil
}

func execActivity[Tin any](ctx workflow.Context, activity func(ctx context.Context, params Tin) (err error), input Tin, scheduleToClose time.Duration) error {
	if scheduleToClose != 0 {
		ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			ScheduleToCloseTimeout: scheduleToClose,
		})
	}
	f := workflow.ExecuteActivity(ctx, activity, input)
	err := f.Get(ctx, nil)
	if err != nil {
		return fmt.Errorf("error in activity '%s': %w", utils.FuncName(activity), err)
	}
	return nil
}

func continueAsNew[Tin any](ctx workflow.Context, wfFunc func(ctx workflow.Context, input Tin) error, input Tin) error {
	return workflow.NewContinueAsNewError(ctx, wfFunc, input)
}

func getEventWFUserID(ctx workflow.Context) (string, error) {
	wfInfo := workflow.GetInfo(ctx)
	parts := strings.SplitN(wfInfo.WorkflowExecution.ID, "__", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("workflow execution ID %s did not have proper split to find userID, got parts %+v: %w", wfInfo.WorkflowExecution.ID, parts, ErrInvalidWorkflowID)
	}
	userID := parts[1]
	return userID, nil
}

func IsWorkflowAlreadyFinishedError(err error) bool {
	if err == nil {
		return false
	}
	if strings.Contains(err.Error(), "workflow execution already completed") || strings.Contains(err.Error(), "Workflow execution already finished successfully") {
		return true
	}
	return false
}

func IsWorkflowNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "unknown external workflow execution")
}

func convertTemporalError(err error) error {
	if err == nil {
		return nil
	}
	// temporal is weird - will open an issue
	if context.DeadlineExceeded.Error() == err.Error() {
		return errors.Join(err, context.DeadlineExceeded)
	}
	if context.Canceled.Error() == err.Error() {
		return errors.Join(err, context.Canceled)
	}
	return err
}
