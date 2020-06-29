package main

import (
	"context"
	"github.com/discretemind/glink/rdp"
	"go.uber.org/zap"
	"time"
)

func main() {
	l, _ := zap.NewDevelopment()
	srv := rdp.Server("test server", l)

	cluster := srv.CreateCluster(1, "cluster 1", 2)
	l.Info("Cluster Added", zap.String("id", cluster.ID().String()))

	go runClient(cluster)
	if err := srv.Listen(context.Background(), 5555); err != nil {
		l.Error("Listening error", zap.Error(err))
	}
	l.Info("Exit")
}

func runClient(cluster rdp.ICluster) {
	time.Sleep(1 * time.Second)

	l, _ := zap.NewDevelopment()
	c := rdp.Client(rdp.FromString("1.1.1"), l)
	if err := c.OnStart(func(cfg string) {
		l.Info("Start ", zap.String("cfg", cfg))
	}); err != nil {
		l.Info("Can't subscribe on start")
	}

	go func() {
		time.Sleep(3 * time.Second)
		l.Info("Start peer")
		if err := cluster.Start(c.ID().String(), struct {
			Brokers []string
		}{
			Brokers: []string{"0.0.0.0:1111","0.0.0.0:2222"},
		}); err != nil {
			l.Error("Cant accept ", zap.Error(err))
		}
	}()

	err := c.Connect(context.Background(), "0.0.0.0:5555", cluster.ID().String())
	if err != nil {
		l.Error("Connection error", zap.Error(err))
		//logger.With("client ")

	}

}
