package app

import "time"

func makeIdentifier(prefix string) string {
    return prefix + "_" + time.Now().UTC().Format("20060102150405.000000000")
}
