package api

import (
	"encoding/json"
	eventsdb "github.com/MinterTeam/events-db"
	"github.com/tidwall/gjson"
)

type FindEventsResponse struct {
	Events eventsdb.Events `json:"events"`
}

func FindEvents(height uint64, find []string) (*FindEventsResponse, error) {
	var result FindEventsResponse
	//todo check find
	events := blockchain.GetEventsDB().LoadEvents(uint32(height))

	for _, event := range events {
		marshalEvent, _ := json.Marshal(event) //todo add interface to Event
		address := gjson.GetManyBytes(marshalEvent, "address", "validator_pub_key")
		for _, addr := range find {
			if address[0].String() == addr || address[1].String() == addr {
				result.Events = append(result.Events, event)
			}
		}
	}

	return &result, nil
}
