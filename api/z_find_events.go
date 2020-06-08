package api

import (
	"encoding/json"
	eventsdb "github.com/MinterTeam/events-db"
	"github.com/MinterTeam/minter-go-node/core/types"
	"github.com/tidwall/gjson"
)

type FindEventsResponse struct {
	Events eventsdb.Events `json:"events"`
}

func FindEvents(height uint64, find []types.Address) (*FindEventsResponse, error) {
	var result FindEventsResponse

	events := blockchain.GetEventsDB().LoadEvents(uint32(height))

	for _, event := range events {
		marshalEvent, _ := json.Marshal(event)
		address := gjson.GetBytes(marshalEvent, "address")
		for _, addr := range find {
			if address.String() == addr.String() {
				result.Events = append(result.Events, event)
			}
		}
	}

	return &result, nil
}
