package api

import (
	"fmt"
	"github.com/MinterTeam/minter-go-node/core/commissions"
	"github.com/MinterTeam/minter-go-node/core/transaction"
	"github.com/MinterTeam/minter-go-node/core/types"
	"github.com/MinterTeam/minter-go-node/formula"
	"github.com/MinterTeam/minter-go-node/rpc/lib/types"
	"math/big"
)

type UseMaxResponse struct {
	GasCoin          string `json:"gascoin"`
	StartValue       string `json:"startvalue"`
	TXComissionValue string `json:"txvalue"`
	EndValue         string `json:"endvalue"`
}

func CalcTxCommission(gascoin string, height int, txtype string, payload []byte, mtxs int64) (string, error) {
	var commissionInBaseCoin *big.Int
	switch txtype {
	case "SendTx":
		commissionInBaseCoin = big.NewInt(commissions.SendTx)
	case "ConvertTx":
		commissionInBaseCoin = big.NewInt(commissions.ConvertTx)
	case "DeclareCandidacyTx":
		commissionInBaseCoin = big.NewInt(commissions.DeclareCandidacyTx)
	case "DelegateTx":
		commissionInBaseCoin = big.NewInt(commissions.DelegateTx)
	case "UnbondTx":
		commissionInBaseCoin = big.NewInt(commissions.UnbondTx)
	case "ToggleCandidateStatus":
		commissionInBaseCoin = big.NewInt(commissions.ToggleCandidateStatus)
	case "EditCandidate":
		commissionInBaseCoin = big.NewInt(commissions.EditCandidate)
	case "RedeemCheckTx":
		commissionInBaseCoin = big.NewInt(commissions.RedeemCheckTx)
	case "CreateMultisig":
		commissionInBaseCoin = big.NewInt(commissions.CreateMultisig)
	case "MultiSend":
		if mtxs <= 0 {
			return "", rpctypes.RPCError{Code: 400, Message: "Set number of txs for multisend (mtxs)"}
		}
		commissionInBaseCoin = big.NewInt(commissions.MultisendDelta*(mtxs-1) + 10)
	default:
		return "", rpctypes.RPCError{Code: 401, Message: "Set correct txtype for tx"}
	}

	if len(payload) > 1024 {
		return "", rpctypes.RPCError{Code: 401, Message: fmt.Sprintf("TX payload length is over %d bytes", 1024)}
	}

	totalCommissionInBaseCoin := new(big.Int).Mul(big.NewInt(0).Add(commissionInBaseCoin, big.NewInt(int64(len(payload)))), transaction.CommissionMultiplier)

	cState, err := GetStateForHeight(height)
	if err != nil {
		return "", err
	}

	cState.RLock()
	defer cState.RUnlock()

	if gascoin == "BIP" {
		return totalCommissionInBaseCoin.String(), nil
	}

	coin := cState.Coins.GetCoin(types.StrToCoinSymbol(gascoin))

	if coin == nil {
		return "", rpctypes.RPCError{Code: 404, Message: "Gas Coin not found"}
	}

	if totalCommissionInBaseCoin.Cmp(coin.Reserve()) == 1 {
		return "", rpctypes.RPCError{Code: 400, Message: "Not enough coin reserve for pay comission"}
	}

	return formula.CalculateSaleAmount(coin.Volume(), coin.Reserve(), coin.Crr(), totalCommissionInBaseCoin).String(), nil

}

func CalcFreeCoinForTx(gascoin string, gascoinamount string, height int, txtype string, payload []byte, mtxs int64) (*UseMaxResponse, error) {

	comission, err := CalcTxCommission(gascoin, height, txtype, payload, mtxs)

	if err != nil {
		return new(UseMaxResponse), err
	}

	commissionBig, _ := big.NewInt(0).SetString(comission, 10)
	gascoinamountBig, _ := big.NewInt(0).SetString(gascoinamount, 10)

	if gascoinamountBig.Cmp(commissionBig) == -1 {
		return new(UseMaxResponse), rpctypes.RPCError{Code: 400, Message: "Not enough coin bipvalue for pay comission"}
	}

	return &UseMaxResponse{
		GasCoin:          gascoin,
		StartValue:       gascoinamountBig.String(),
		TXComissionValue: comission,
		EndValue:         big.NewInt(0).Sub(gascoinamountBig, commissionBig).String(),
	}, nil
}
