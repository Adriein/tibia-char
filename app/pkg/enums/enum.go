package enums

import "github.com/rotisserie/eris"

type AuctionRecordableStatus int

const (
	RecordableActive AuctionRecordableStatus = iota + 1
	RecordableArchived
	RecordableDeleted
)

var auctionRecordableStatusName = map[AuctionRecordableStatus]string{
	RecordableActive:   "active",
	RecordableArchived: "archived",
	RecordableDeleted:  "deleted",
}

var auctionRecordableStatusValue = map[string]AuctionRecordableStatus{
	"active":   RecordableActive,
	"archived": RecordableArchived,
	"deleted":  RecordableDeleted,
}

func (ars AuctionRecordableStatus) String() string {
	return auctionRecordableStatusName[ars]
}

func GetAuctionRecordableStatusFromString(status string) (AuctionRecordableStatus, error) {
	if val, ok := auctionRecordableStatusValue[status]; ok {
		return val, nil
	}

	return 0, eris.Errorf("Unknown AuctionRecordableStatus: %s", status)
}

type BattleEye int

const (
	BattleEyeYellow BattleEye = iota + 1
	BattleEyeGreen
)

var battlEyeName = map[BattleEye]string{
	BattleEyeYellow: "yellow",
	BattleEyeGreen:  "green",
}

var battleEyeValue = map[string]BattleEye{
	"yellow": BattleEyeYellow,
	"green":  BattleEyeGreen,
}

func (b BattleEye) String() string {
	return battlEyeName[b]
}

func GetBattleEyeFromString(battleEye string) (BattleEye, error) {
	if val, ok := battleEyeValue[battleEye]; ok {
		return val, nil
	}

	return 0, eris.Errorf("Unknown BattleEye: %s", battleEye)
}

type WorldTransferAllowance int

const (
	WorldTransferImmediately WorldTransferAllowance = iota + 1
	WorldTransferForbidden
)

var worldTransferAllowanceName = map[WorldTransferAllowance]string{
	WorldTransferImmediately: "immediately",
	WorldTransferForbidden:   "forbidden",
}

var worldTransferAllowanceValue = map[string]WorldTransferAllowance{
	"immediately": WorldTransferImmediately,
	"forbidden":   WorldTransferForbidden,
}

func (w WorldTransferAllowance) String() string {
	return worldTransferAllowanceName[w]
}

func GetWorldTransferAllowanceFromString(allowance string) (WorldTransferAllowance, error) {
	if val, ok := worldTransferAllowanceValue[allowance]; ok {
		return val, nil
	}

	return 0, eris.Errorf("Unknown WorldTransferAllowance: %s", allowance)
}
