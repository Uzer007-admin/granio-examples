//go:build !tamago && !noos

package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Uzer007-admin/granio/v2"
)

// AssertNotificationSent allows an app developer to assert that a notification was sent.
// After the content of f has executed this utility will check that the specified notification was sent.
func AssertNotificationSent(t *testing.T, n *granio.Notification, f func()) {
	require.NotNil(t, f, "function has to be specified")
	require.IsType(t, &app{}, granio.CurrentApp())
	a, _ := granio.CurrentApp().(*app)
	a.lastNotification = nil

	f()
	if n == nil {
		assert.Nil(t, a.lastNotification)
		return
	} else if a.lastNotification == nil {
		t.Error("No notification sent")
		return
	}

	assert.Equal(t, n.Title, a.lastNotification.Title)
	assert.Equal(t, n.Content, a.lastNotification.Content)
}

// AssertNotificationScheduled checks that a notification was scheduled for delivery
// after the supplied function has run. The reported [granio.ScheduledNotification.Title]
// and [granio.ScheduledNotification.Content] must match the supplied notification.
//
// Since: 2.8
func AssertNotificationScheduled(t *testing.T, n *granio.Notification, f func()) {
	require.NotNil(t, f, "function has to be specified")
	require.IsType(t, &app{}, granio.CurrentApp())
	a, _ := granio.CurrentApp().(*app)
	a.lastScheduledNotification = nil

	f()
	if n == nil {
		assert.Nil(t, a.lastScheduledNotification)
		return
	} else if a.lastScheduledNotification == nil {
		t.Error("No notification scheduled")
		return
	}

	assert.Equal(t, n.Title, a.lastScheduledNotification.Title)
	assert.Equal(t, n.Content, a.lastScheduledNotification.Content)
}
