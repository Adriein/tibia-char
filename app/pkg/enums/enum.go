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

type AuctionStage string

const (
	StageInitial AuctionStage = "initial"
	StageCurrent AuctionStage = "current"
	StageWinning AuctionStage = "winning"
)

func (s AuctionStage) Valid() bool {
	switch s {
	case StageInitial, StageCurrent, StageWinning:
		return true
	}

	return false
}
