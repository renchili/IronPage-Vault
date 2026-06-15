package app

import "testing"

func TestNotificationTrimCount(t *testing.T) {
    cases := []struct{
        unread int
        limit int
        want int
    }{
        {0, 500, 0},
        {499, 500, 0},
        {500, 500, 1},
        {501, 500, 2},
        {10, 0, 0},
    }
    for _, tc := range cases {
        if got := notificationTrimCount(tc.unread, tc.limit); got != tc.want {
            t.Fatalf("notificationTrimCount(%d,%d)=%d want %d", tc.unread, tc.limit, got, tc.want)
        }
    }
}
