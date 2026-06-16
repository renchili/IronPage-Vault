package core

const MaxUnreadNotifications = 500

func NotificationTrimCount(unread int, limit int) int {
	if limit < 1 {
		return 0
	}
	if unread < limit {
		return 0
	}
	return unread - limit + 1
}
