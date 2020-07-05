package glink

import (
	"context"
	"github.com/discretemind/glink/rdp"
	"go.uber.org/zap"
)

type clusterManager struct {
	url   string
	token string
}

func ClusterManager(url string, token string) (res *clusterManager) {
	res = &clusterManager{
		url: url,
	}
	return
}

func (m *clusterManager) Connect(ctx context.Context, modelId string) {
	l, _ := zap.NewProduction()
	client := rdp.Client(rdp.FromString("1.0.0"), l)
	if err := client.Connect(context.Background(), m.url, modelId); err != nil {
		l.Error("connection error", zap.Error(err))
	}
}

func (m *clusterManager) Error(err error) {

}
