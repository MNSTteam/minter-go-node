package api

import (
	"github.com/MinterTeam/minter-go-node/rpc/lib/types"
	"time"
)

func DurationTime(height1 int64, count int64) time.Duration {
	if count == 0 {
		count = 100
	}
	block1, _ := client.Block(&height1)
	time1 := block1.Block.Time

	height2 := height1 - count
	block2, _ := client.Block(&height2)
	duration := time1.Sub(block2.Block.Time)

	return time.Duration(duration.Nanoseconds() / count)
}

func HeightByTime(query string, height int64) (int64, error) {

	h := int64(blockchain.Height())
	if height > 0 && height > int64(blockchain.Height()) {
		return 0, rpctypes.RPCError{Code: 404, Message: "Inputed block higher actual block"}
	}
	if height == 0 {
		height = h
	}

	duration := DurationTime(height, 100)

	block, _ := client.Block(&height)

	var difference int64

	switch query {

	case "day":
		difference = int64(time.Hour * 24 / duration)
	case "week":
		difference = int64(time.Hour * 24 * 7 / duration)
	case "":
		return height, nil
	default:
		targettime, err := time.Parse(time.RFC3339, query)
		if err != nil {
			return 0, rpctypes.RPCError{Code: 404, Message: "Incorrect query time", Data: err.Error()}
		}
		difference = int64(block.Block.Time.Sub(targettime) / duration)
	}

	for {
		height -= difference
		block2, _ := client.Block(&height)

		result := block.Block.Time.Sub(block2.Block.Time.Add(time.Duration(-difference))).Nanoseconds()
		difference = result / duration.Nanoseconds()
		if difference == 0 {
			if result > 0 {
				return height, nil
			}
			return height, nil
		}
	}
}
