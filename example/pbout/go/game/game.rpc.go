package game

import (
	context "context"
	grpc "google.golang.org/grpc"
	proto "google.golang.org/protobuf/proto"
)

type GameServer interface {
	GameAddActivity(context.Context, *CGameAddActivity) (*SGameAddActivity, error)
	GameUpdateActivity(context.Context, *CGameUpdateActivity) (*SGameUpdateActivity, error)
	GameUpdateActivityStatus(context.Context, *CGameUpdateActivityStatus) (*SGameUpdateActivityStatus, error)
	GetActivityList(context.Context, *CGetActivityList) (*SGetActivityList, error)
	LoginGame(context.Context, *CLoginGame) (*SLoginGame, error)
}

func _Game_GetActivityList_Handler(srv any, ctx context.Context, in proto.Message) (any, error) {
	req := in.(*CGetActivityList)
	return srv.(GameServer).GetActivityList(ctx, req)
}

func _Game_LoginGame_Handler(srv any, ctx context.Context, in proto.Message) (any, error) {
	req := in.(*CLoginGame)
	return srv.(GameServer).LoginGame(ctx, req)
}

func _Game_GameAddActivity_Handler(srv any, ctx context.Context, in proto.Message) (any, error) {
	req := in.(*CGameAddActivity)
	return srv.(GameServer).GameAddActivity(ctx, req)
}

func _Game_GameUpdateActivity_Handler(srv any, ctx context.Context, in proto.Message) (any, error) {
	req := in.(*CGameUpdateActivity)
	return srv.(GameServer).GameUpdateActivity(ctx, req)
}

func _Game_GameUpdateActivityStatus_Handler(srv any, ctx context.Context, in proto.Message) (any, error) {
	req := in.(*CGameUpdateActivityStatus)
	return srv.(GameServer).GameUpdateActivityStatus(ctx, req)
}

var Game_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Game",
	HandlerType: (*GameServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetActivityList",
			Handler:    _Game_GetActivityList_Handler,
		},
		{
			MethodName: "LoginGame",
			Handler:    _Game_LoginGame_Handler,
		},
		{
			MethodName: "GameAddActivity",
			Handler:    _Game_GameAddActivity_Handler,
		},
		{
			MethodName: "GameUpdateActivity",
			Handler:    _Game_GameUpdateActivity_Handler,
		},
		{
			MethodName: "GameUpdateActivityStatus",
			Handler:    _Game_GameUpdateActivityStatus_Handler,
		},
	},
}
