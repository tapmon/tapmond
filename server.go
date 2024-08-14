package tapmond

type Tapmond struct {
	rpcServer *TapmonRpcServer
}

func InitTapmond() (*Tapmond, error) {
	return &Tapmond{}, nil
}
