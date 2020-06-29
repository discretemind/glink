package glink

type clusterManager struct {
}

func ClusterManager(url string, token string) (res *clusterManager){
	res = &clusterManager{}
	return
}

func (m *clusterManager) Error(err error) {

}