package api

import (
	"github.com/MinterTeam/minter-go-node/core/state"
	"github.com/MinterTeam/minter-go-node/core/state/candidates"
	"github.com/MinterTeam/minter-go-node/core/types"
	"github.com/MinterTeam/minter-go-node/rpc/lib/types"
	"math/big"
)

type CandidateResponseAlt struct {
	PubKey     string `json:"pub_key"`
	TotalStake string `json:"total_stake"`
	Commission uint   `json:"commission"`
	UsedSlots  int    `json:"used_slots"`
	UniqUsers  int    `json:"uniq_users"`
	MinStake   string `json:"minstake"`
	Status     byte   `json:"status"`
}

const (
	// CandidateOff = 0x01
	// CandidateOn	= 0x02
	ValidatorOn        = 0x03
	ValidatorsMaxSlots = 1000
)

func ResponseCandidateAlt(state *state.State, c *candidates.Candidate) *CandidateResponseAlt {
	stakes := state.Candidates.GetStakes(c.PubKey)
	candidate := &CandidateResponseAlt{
		TotalStake: state.Candidates.GetTotalStake(c.PubKey).String(),
		PubKey:     c.PubKey.String(),
		Commission: c.Commission,
		Status:     c.Status,
		UsedSlots:  len(stakes),
	}
	addresses := map[string]struct{}{}

	for _, validator := range state.Validators.GetValidators() {
		if validator.PubKey.String() != candidate.PubKey {
			continue
		}
		candidate.Status = ValidatorOn
		break
	}

	minStake := big.NewInt(0)
	for i, stake := range stakes {
		if candidate.UsedSlots >= ValidatorsMaxSlots {
			if i == 0 {
				minStake = stake.BipValue
				continue
			}
			for minStake.Cmp(stake.BipValue) == 1 {
				minStake = stake.BipValue
			}
		}

		if _, ok := addresses[stake.Owner.String()]; !ok {
			addresses[stake.Owner.String()] = struct{}{}
		}
	}

	candidate.UniqUsers = len(addresses)
	candidate.MinStake = minStake.String()

	return candidate
}

func CandidateAlt(pubkey types.Pubkey, height int) (*CandidateResponseAlt, error) {
	cState, err := GetStateForHeight(height)
	if err != nil {
		return nil, err
	}

	if height != 0 {
		cState.Lock()
		cState.Candidates.LoadCandidates()
		cState.Candidates.LoadStakes()
		cState.Validators.LoadValidators()
		cState.Unlock()
	}

	cState.RLock()
	defer cState.RUnlock()

	candidate := cState.Candidates.GetCandidate(pubkey)
	if candidate == nil {
		return nil, rpctypes.RPCError{Code: 404, Message: "Candidate not found"}
	}

	return ResponseCandidateAlt(cState, candidate), nil
}

func CandidatesAlt(height int, status int) ([]*CandidateResponseAlt, error) {
	cState, err := GetStateForHeight(height)
	if err != nil {
		return nil, err
	}

	if height != 0 {
		cState.Lock()
		cState.Candidates.LoadCandidates()
		cState.Validators.LoadValidators()
		cState.Candidates.LoadStakes()
		cState.Unlock()
	}

	cState.RLock()
	defer cState.RUnlock()

	var result []*CandidateResponseAlt

	allCandidates := cState.Candidates.GetCandidates()
	for _, candidate := range allCandidates {
		candadateInfo := ResponseCandidateAlt(cState, candidate)
		if status == 0 {
			result = append(result, candadateInfo)
			continue
		}
		if candadateInfo.Status == byte(status) {
			result = append(result, candadateInfo)
		}
	}

	return result, nil
}
