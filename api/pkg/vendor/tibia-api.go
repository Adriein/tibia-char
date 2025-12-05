package vendor

import (
	"encoding/json"
	"net/http"

	"github.com/rotisserie/eris"
)

type TibiaApi struct{}

type TibiaApiWorlds struct {
	Worlds struct {
		RegularWorlds []struct {
			Name string `json:"name"`
		} `json:"regular_worlds"`
	} `json:"worlds"`
}

func NewTibiaApi() TibiaApi {
	return TibiaApi{}
}

func (t *TibiaApi) GetWorlds() ([]string, error) {
	httpRes, err := http.Get("https://api.tibiadata.com/v4/worlds")

	if err != nil {
		return nil, eris.Errorf("Error fetching worlds with TibiaApi: %s", err.Error())
	}

	if httpRes.StatusCode != http.StatusOK {
		return nil, eris.Errorf("Tibia Api responded with http code %d fetching worlds", httpRes.StatusCode)
	}

	defer httpRes.Body.Close()

	var apiWorldsResponse TibiaApiWorlds

	err = json.NewDecoder(httpRes.Body).Decode(&apiWorldsResponse)

	if err != nil {
		return nil, eris.Errorf("Error parsing json: %s", err.Error())
	}

	regularWorlds := apiWorldsResponse.Worlds.RegularWorlds

	results := make([]string, 0, len(regularWorlds))

	for _, world := range regularWorlds {
		results = append(results, world.Name)
	}

	return results, nil
}
