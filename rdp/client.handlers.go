package rdp

import (
	"go.uber.org/zap"
	"log"
	"reflect"
)

func (c *client) registerHandler(value interface{}) {
	v := reflect.ValueOf(value)
	t := v.Type()
	if t.NumIn() != 1 {
		log.Fatal("invalid handler func")
	}
	cmdType := t.In(0)

	id, ok := ProtectedCommands.GetIdByType(cmdType)
	if !ok {
		log.Fatal("invalid handler func. Command not found")
	}
	c.handlers[id] = handlerType(v)
}

//func (c *client) execHandler(id uint16, cmd reflect.Value) error {
//	f, ok := c.handlers[id]
//	if !ok {
//		return fmt.Errorf("can't execute. handler not found %d", id)
//	}
//	v := reflect.Value(f)
//	//res := v.Call([]reflect.Value{reflect.ValueOf(cmd)})
//
//
//	return nil
//}

func (c *client) startHandler(cmd *startCmd) {
	c.logger.Info("startHandler", zap.String("cmd", string(cmd.Config)))
}

func (c *client) OnStart(f func(cfg string)) (err error) {
	//return c.subscribe(func(cmd *startCmd) {
	//	f(string(cmd.Config))
	//})
	return nil
}

func (c *client) OnStop(f func() error) {

}
