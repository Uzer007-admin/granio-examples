package app

import (
	"time"

	"github.com/Uzer007-admin/granio/v2"
)

// scheduleViaScheduler queues the notification through the in-process scheduler,
// returning a [granio.ScheduledNotification] populated with the assigned ID. Used by
// platform implementations that do not have a native scheduling API.
func (a *fyneApp) scheduleViaScheduler(n *granio.Notification, when time.Time) (*granio.ScheduledNotification, error) {
	id, err := a.scheduler.Schedule(n, when)
	if err != nil {
		return nil, err
	}
	return granio.NewScheduledNotification(id, n, when), nil
}

// cancelViaScheduler removes a pending in-process schedule by ID. Used by platform
// implementations that do not have a native scheduling API.
func (a *fyneApp) cancelViaScheduler(id string) error {
	return a.scheduler.Cancel(id)
}
