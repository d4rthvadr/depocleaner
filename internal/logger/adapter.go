package logger

type Adapter struct {
	logger *Logger
}

func NewAdapter(logger *Logger) *Adapter {
	return &Adapter{logger: logger}
}

func (a *Adapter) Debug(msg string, fields ...interface{}) {
	a.logger.Debugw(msg, fields...)
}

func (a *Adapter) Info(msg string, fields ...interface{}) {
	a.logger.Infow(msg, fields...)
}

func (a *Adapter) Warn(msg string, fields ...interface{}) {
	a.logger.Warnw(msg, fields...)
}

func (a *Adapter) Error(msg string, fields ...interface{}) {
	a.logger.Errorw(msg, fields...)
}
