package transactor

import (
	"crypto/ecdsa"
	"errors"
	"math/big"

	"github.com/DIMO-Network/trips-api/internal/config"
	"github.com/DIMO-Network/trips-api/internal/tripNft"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/net/context"
)

type Transactor struct {
	Client     *ethclient.Client
	TripNft    *tripNft.TripNft
	privateKey *ecdsa.PrivateKey
	addr       common.Address
	chainId    int64
}

func New(settings *config.Settings) (*Transactor, error) {
	client, err := ethclient.Dial(settings.RPCUrl)
	if err != nil {
		return nil, err
	}

	privateKey, err := crypto.HexToECDSA(settings.DeployedContractPrivateKey)
	if err != nil {
		return nil, err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	addr := crypto.PubkeyToAddress(*publicKeyECDSA)

	instance, err := tripNft.NewTripNft(common.HexToAddress(settings.DeployedContractAddress), client)
	if err != nil {
		return nil, err
	}

	return &Transactor{
		Client:     client,
		TripNft:    instance,
		privateKey: privateKey,
		addr:       addr,
		chainId:    settings.ChainID,
	}, nil
}

func (t *Transactor) MintSegment(ctx context.Context, owner common.Address, vehicleNode *big.Int, startTime uint64, endTime uint64, startHex uint64, endHex uint64, bundlrId string) (*types.Transaction, error) {
	nonce, err := t.Client.PendingNonceAt(ctx, t.addr)
	if err != nil {
		return nil, err
	}

	gasPrice, err := t.Client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(t.privateKey, big.NewInt(t.chainId))
	if err != nil {
		return nil, err
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(300000)
	auth.GasPrice = gasPrice

	return t.TripNft.Mint(auth, owner, vehicleNode, startTime, endTime, startHex, endHex, bundlrId)
}
