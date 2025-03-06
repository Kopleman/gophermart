package log

type MockLogger struct {
	*SugaredLogger
}

func (m MockLogger) Debug(_ ...interface{}) {
}

func (m MockLogger) Debugf(_ string, _ ...interface{}) {
}

func (m MockLogger) Info(_ ...interface{}) {
}

func (m MockLogger) Infof(_ string, _ ...interface{}) {
}

func (m MockLogger) Infow(_ string, _ ...interface{}) {
}

func (m MockLogger) Infoln(_ ...interface{}) {
}

func (m MockLogger) Warn(_ ...interface{}) {
}

func (m MockLogger) Warnf(_ string, _ ...interface{}) {
}

func (m MockLogger) Error(_ ...interface{}) {
}

func (m MockLogger) Errorf(_ string, _ ...interface{}) {
}

func (m MockLogger) Fatal(_ ...interface{}) {
}

func (m MockLogger) Fatalf(_ string, _ ...interface{}) {
}

func (m MockLogger) Panic(_ ...interface{}) {
}

func (m MockLogger) Panicf(_ string, _ ...interface{}) {
}

func (m MockLogger) With(_ ...interface{}) Logger {
	return m
}

func (m MockLogger) Sync() error {
	return nil
}

func (m MockLogger) Sugar() *SugaredLogger {
	return m.SugaredLogger
}
