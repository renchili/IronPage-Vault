package core

// MaxUnreadNotifications is the per-user unread notification ceiling.
const MaxUnreadNotifications = 500

// NotificationTrimCount returns how many oldest unread notifications to trim.
//
// The returned count includes room for the notification about to be inserted, so
// callers can enforce the unread ceiling before adding the new record.
func NotificationTrimCount(unread int, limit int) int {
	if limit < 1 {
		return 0
	}
	if unread < limit {
		return 0
	}
	return unread - limit + 1
}
