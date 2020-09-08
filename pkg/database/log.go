package database

import "k8s.io/klog"

type dbLogger struct {
}

func (l dbLogger) Print(v ...interface{}) {
	klog.Info(v)
}
