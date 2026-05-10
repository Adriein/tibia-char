package constants

//Env var keys

const (
	DatabaseUser     = "DATABASE_USER"
	DatabasePassword = "DATABASE_PASSWORD"
	DatabaseName     = "DATABASE_NAME"
	ServerPort       = "SERVER_PORT"
	Env              = "ENV"
	Production       = "PRODUCTION"
	PosthogSdkApiKey = "POSTHOG_SDK_API_KEY"
)

const (
	DEV  = "dev"
	PROD = "prod"
)

const (
	AuctionModule = "AUCTION_MODULE"
)

const (
	SourceKey = "source"
)

const (
	SessionCookie = "tc_session"
)

// Tibia Char

const (
	OkResKey    = "ok"
	ErrorResKey = "error"
	DataResKey  = "data"
)

const (
	GenderMale       = 1
	GenderFemale     = 2
	VocationKnight   = 1
	VocationPaladin  = 2
	VocationSorcerer = 3
	VocationDruid    = 4
	VocationMonk     = 5
	VocationNone     = 6
)

const (
	Male     = "male"
	Female   = "female"
	Knight   = "knight"
	Paladin  = "paladin"
	Sorcerer = "sorcerer"
	Druid    = "druid"
	Monk     = "monk"
	None     = "none"
)

const (
	ProxyAddr  = "PROXY_ADDR"
	LocalProxy = "local"
)

const (
	TibiaOfficialWebsite        = "www.tibia.com"
	IncomingTimeFormat          = "20060102150405"
	TibiaCoinEuroEquivalence    = 0.04
	MaxLongTailRefreshAllowance = 50
	GRoutineID                  = "GRoutineID"
	Phase                       = "PHASE"
	ScrapPhase                  = "SCRAP"
	WatchPhase                  = "WATCH"
	ConsolidatePhase            = "CONSOLIDATE"
)

// Errors

const (
	ServerGenericError         = "SERVER_ERROR"
	NoGoodSearchParamProvided  = "NO_GOOD_SEARCH_PARAM_PROVIDED"
	NoWorldSearchParamProvided = "NO_WORLD_SEARCH_PARAM_PROVIDED"
)
