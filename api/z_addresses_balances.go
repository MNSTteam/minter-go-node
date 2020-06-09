package api

import (
	"github.com/MinterTeam/minter-go-node/core/state"
	"github.com/MinterTeam/minter-go-node/core/types"
	"github.com/MinterTeam/minter-go-node/formula"
	"math/big"
)

type CoinBalance struct {
	Coin     string `json:"coin"`
	Value    string `json:"value"`
	BipValue string `json:"bip_value"`
}

type AddressBalanceResponse struct {
	Freecoins []*CoinBalance `json:"freecoins"`
	Delegated []*CoinBalance `json:"delegated"`

	//todo: unbound (замороженные)

	Total            []*CoinBalance `json:"total"`
	TransactionCount uint64         `json:"transaction_count"`
	Bipvalue         string         `json:"bipvalue"`
}

type AddressesBalancesResponse struct {
	Address types.Address           `json:"address"`
	Balance *AddressBalanceResponse `json:"balance"`
}

type UserStake struct {
	Value    *big.Int
	BipValue *big.Int
}

func CustomCoinBipBalance(coinToSellString string, valueToSell *big.Int, cState *state.State) *big.Int {
	coinToSell := types.StrToCoinSymbol(coinToSellString)
	coinToBuy := types.StrToCoinSymbol("BIP")

	if coinToSell == coinToBuy {
		return valueToSell
	}

	if coinToSell == types.GetBaseCoin() {
		coin := cState.Coins.GetCoin(coinToBuy)
		return formula.CalculatePurchaseReturn(coin.Volume(), coin.Reserve(), coin.Crr(), valueToSell)
	}

	if coinToBuy == types.GetBaseCoin() {
		coin := cState.Coins.GetCoin(coinToSell)
		return formula.CalculateSaleReturn(coin.Volume(), coin.Reserve(), coin.Crr(), valueToSell)
	}

	coinFrom := cState.Coins.GetCoin(coinToSell)
	coinTo := cState.Coins.GetCoin(coinToBuy)
	basecoinValue := formula.CalculateSaleReturn(coinFrom.Volume(), coinFrom.Reserve(), coinFrom.Crr(), valueToSell)
	return formula.CalculatePurchaseReturn(coinTo.Volume(), coinTo.Reserve(), coinTo.Crr(), basecoinValue)

}

func MakeAddressBalance(address types.Address, height int) (*AddressBalanceResponse, error) {
	cState, err := GetStateForHeight(height)
	if err != nil {
		return nil, err
	}

	if height != 0 {
		cState.Lock()
		cState.Candidates.LoadCandidates()
		cState.Candidates.LoadStakes()
		cState.Unlock()
	}

	cState.RLock()
	defer cState.RUnlock()

	balances := cState.Accounts.GetBalances(address)
	var response AddressBalanceResponse

	totalStakesGroupByCoin := map[types.CoinSymbol]*UserStake{}
	freecoinStakesGroupByCoin := map[types.CoinSymbol]*UserStake{}

	response.Freecoins = make([]*CoinBalance, 0, len(balances))
	for coin, value := range balances {
		result := CustomCoinBipBalance(coin.String(), value, cState)
		freecoinStakesGroupByCoin[coin] = &UserStake{
			Value:    value,
			BipValue: result,
		}
		totalStakesGroupByCoin[coin] = &UserStake{
			Value:    value,
			BipValue: result,
		}
		response.Freecoins = append(response.Freecoins, &CoinBalance{
			Coin:     coin.String(),
			Value:    value.String(),
			BipValue: result.String(),
		})
	}

	var userDelegatedStakesGroupByCoin = map[types.CoinSymbol]*UserStake{}
	allCandidates := cState.Candidates.GetCandidates()
	for _, candidate := range allCandidates {
		userStakes := UserStakes(cState, candidate.PubKey, address)
		for coin, userStake := range userStakes {
			stake, ok := userDelegatedStakesGroupByCoin[coin]
			if !ok {
				stake = &UserStake{
					Value:    big.NewInt(0),
					BipValue: big.NewInt(0),
				}
			}
			stake.Value.Add(stake.Value, userStake.Value)
			stake.BipValue.Add(stake.BipValue, userStake.BipValue)
			userDelegatedStakesGroupByCoin[coin] = stake
		}
	}

	response.Delegated = make([]*CoinBalance, 0, len(userDelegatedStakesGroupByCoin))
	for coin, delegatedStake := range userDelegatedStakesGroupByCoin {
		response.Delegated = append(response.Delegated, &CoinBalance{
			Coin:     coin.String(),
			Value:    delegatedStake.Value.String(),
			BipValue: delegatedStake.BipValue.String(),
		})

		totalStake, ok := totalStakesGroupByCoin[coin]
		if !ok {
			totalStake = &UserStake{
				Value:    big.NewInt(0),
				BipValue: big.NewInt(0),
			}
		}
		totalStake.Value.Add(totalStake.Value, delegatedStake.Value)
		totalStake.BipValue.Add(totalStake.BipValue, delegatedStake.BipValue)
		totalStakesGroupByCoin[coin] = totalStake
	}

	coinsBipValue := big.NewInt(0)
	response.Total = make([]*CoinBalance, 0, len(totalStakesGroupByCoin))
	for coin, stake := range totalStakesGroupByCoin {
		response.Total = append(response.Total, &CoinBalance{
			Coin:     coin.String(),
			Value:    stake.Value.String(),
			BipValue: stake.BipValue.String(),
		})
		coinsBipValue.Add(coinsBipValue, stake.BipValue)
	}

	response.TransactionCount = cState.Accounts.GetNonce(address)
	response.Bipvalue = coinsBipValue.String()

	return &response, nil
}

func UserStakes(state *state.State, c types.Pubkey, address types.Address) map[types.CoinSymbol]*UserStake {
	var userStakes = map[types.CoinSymbol]*UserStake{}

	stakes := state.Candidates.GetStakes(c)

	for _, stake := range stakes {
		if stake.Owner != address {
			continue
		}
		userStakes[stake.Coin] = &UserStake{
			Value:    stake.Value,
			BipValue: stake.BipValue,
		}
	}

	return userStakes
}
