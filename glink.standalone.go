package glink

//import "go.uber.org/zap"

type standaloneManager struct {
	//logger zap.Logger
}

func StandaloneManager() (res *standaloneManager) {
	res = &standaloneManager{}
	return
}

func (m *standaloneManager) Error(err error) {
	//m.logger.Error(err.Error())
}
