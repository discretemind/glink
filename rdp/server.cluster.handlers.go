package rdp

import (
	"go.uber.org/zap"
	"log"
	"reflect"
)

type handlerType reflect.Value

func (c *cluster) registerHandler(value interface{}) {
	v := reflect.ValueOf(value)
	t := v.Type()
	if t.NumIn() != 2 {
		log.Fatal("Invalid handler func")
	}
	cmdType := t.In(1)

	id, ok := ProtectedCommands.GetIdByType(cmdType)
	if !ok {
		log.Fatal("Invalid handler func. Command not found")
	}
	c.handlers[id] = handlerType(v)
}

//func (c *cluster) execHandler(id uint16, p *Peer, cmd reflect.Value) error {
//	f, ok := c.handlers[id]
//	if !ok {
//		return fmt.Errorf("can't execute. handler not found %d", id)
//	}
//	v := reflect.Value(f)
//	res := v.Call([]reflect.Value{reflect.ValueOf(p), reflect.ValueOf(cmd)})
//	fmt.Println(res)
//	return nil
//}

func (c *cluster) handleMetrics(p *Peer, command *MetricsCmd) error {
	c.logger.Info("handleMetrics: ", zap.String("peer", p.addr.String()), zap.Any("cmd", command))
	return p.UpdateMetrics(command)
}
