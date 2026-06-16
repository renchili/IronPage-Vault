package core

import "testing"

func TestNotificationTrimCount(t *testing.T) {
	cases := []struct {
		name   string
		unread int
		limit  int
		want   int
	}{
		{"empty", 0, MaxUnreadNotifications, 0},
		{"below limit", MaxUnreadNotifications - 1, MaxUnreadNotifications, 0},
		{"at limit", MaxUnreadNotifications, MaxUnreadNotifications, 1},
		{"over limit", MaxUnreadNotifications + 1, MaxUnreadNotifications, 2},
		{"invalid limit", 10, 0, 0},
	}

	for _, tc := range cases {
		if got := NotificationTrimCount(tc.unread, tc.limit); got != tc.want {
			t.Fatalf("%s: NotificationTrimCount(%d,%d)=%d want %d", tc.name, tc.unread, tc.limit, got, tc.want)
		}
	}
}
