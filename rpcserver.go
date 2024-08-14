package tapmond

import (
	"context"

	"github.com/tapmon/tapmond/mons"
	"github.com/tapmon/tapmond/tapmonrpc"
)

type TapmonRpcServer struct {
	tapmonManager *mons.Manager

	tapmonrpc.UnimplementedTapmonServer
}

func NewTapmonRpcServer(manager *mons.Manager) *TapmonRpcServer {
	return &TapmonRpcServer{
		tapmonManager: manager,
	}
}

func (t *TapmonRpcServer) GetMon(ctx context.Context,
	req *tapmonrpc.GetMonRequest) (*tapmonrpc.GetMonResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (t *TapmonRpcServer) ListOwnedMons(ctx context.Context,
	req *tapmonrpc.ListOwnedMonsRequest) (*tapmonrpc.ListOwnedMonsResponse,
	error) {

	panic("not implemented") // TODO: Implement
}

func (t *TapmonRpcServer) ListAllMons(ctx context.Context,
	req *tapmonrpc.ListAllMonsRequest) (*tapmonrpc.ListAllMonsResponse, error) {

	panic("not implemented") // TODO: Implement
}

func (t *TapmonRpcServer) MintMon(ctx context.Context,
	req *tapmonrpc.MintMonRequest) (*tapmonrpc.MintMonResponse, error) {

	mon, err := t.tapmonManager.MintMon(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	return &tapmonrpc.MintMonResponse{
		Mon: monToRpc(mon),
	}, nil

}

func (t *TapmonRpcServer) LevelMon(ctx context.Context,
	req *tapmonrpc.LevelMonRequest) (*tapmonrpc.LevelMonResponse, error) {

	panic("not implemented") // TODO: Implement
}

func monToRpc(mon *mons.Mon) *tapmonrpc.Mon {
	attributes := make([]int32, 0, len(mon.Scores))
	for _, rarity := range mon.Scores {
		attributes = append(attributes, int32(rarity))
	}

	return &tapmonrpc.Mon{
		Id:          mon.Id,
		RarityScore: mon.CalculateRarityScore(0),
		Attributes:  attributes,
	}
}
