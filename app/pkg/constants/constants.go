package constants

//Env var keys

const (
	DatabaseUser     = "DATABASE_USER"
	DatabasePassword = "DATABASE_PASSWORD"
	DatabaseName     = "DATABASE_NAME"
	ServerPort       = "SERVER_PORT"
	Env              = "ENV"
	Production       = "PRODUCTION"
)

// Tibia Char

const (
	OkResKey    = "ok"
	ErrorResKey = "error"
	DataResKey  = "data"
)

const (
	VolatileMarketStatus     = "Volatile"
	StableMarketStatus       = "Stable"
	RiskyMarketStatus        = "Risky"
	BullMarketType           = "Bull"
	BearMarketType           = "Bear"
	SidewaysMarketType       = "Sideways"
	BullExhaustionMarketType = "Bull Exhaustion"
	PullbackMarketType       = "Pullback"
	UnclearMarketType        = "Unclear"
)

const (
	EventDataIngestion            = "DATA_INGESTION"
	EventDataIngestionDescription = "Ingested market data"
)

const (
	GenderMale       = 1
	GenderFemale     = 2
	VocationKnight   = 1
	VocationPaladin  = 2
	VocationSorcerer = 3
	VocationDruid    = 4
	VocationMonk     = 5
)

const (
	Male     = "male"
	Female   = "female"
	Knight   = "knight"
	Paladin  = "paladin"
	Sorcerer = "sorcerer"
	Druid    = "druid"
	Monk     = "monk"
)

// Errors

const (
	ServerGenericError         = "SERVER_ERROR"
	NoGoodSearchParamProvided  = "NO_GOOD_SEARCH_PARAM_PROVIDED"
	NoWorldSearchParamProvided = "NO_WORLD_SEARCH_PARAM_PROVIDED"
)

const (
	IncomingTimeFormat = "20060102150405"
)

const (
	TibiaOfficialWebsite = "www.tibia.com"
)
