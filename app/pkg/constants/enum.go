package constants

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
