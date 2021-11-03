package contracts

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/contracts/abis"
	"github.com/mapprotocol/atlas/core/vm"
	"github.com/mapprotocol/atlas/params"
)

var getAddressMethod = NewBoundMethod(params.RegistrySmartContractAddress, abis.Registry, "getAddressFor", params.MaxGasForGetAddressFor)

// TODO(kevjue) - Re-Enable caching of the retrieved registered address
// See this commit for the removed code for caching:  https://github.com/celo-org/geth/commit/43a275273c480d307a3d2b3c55ca3b3ee31ec7dd.

// GetRegisteredAddress returns the address on the registry for a given id
func GetRegisteredAddress(vmRunner vm.EVMRunner, registryId common.Hash) (common.Address, error) {
	if registryId == common.HexToHash("0x235a6f54090e9b94aa4e585a699c4375a2ff8f572c68114d138f0ed121527849") {
		//contractAddress = common.HexToAddress("0x000000000000000000000000000000000000d013")
		//err = nil
		fmt.Println("-----------------ele-------err------------------")
	}
	vmRunner.StopGasMetering()
	defer vmRunner.StartGasMetering()

	var contractAddress common.Address
	// 通过common.hash注册Register
	err := getAddressMethod.Query(vmRunner, &contractAddress, registryId)

	// TODO (mcortesi) Remove ErrEmptyArguments check after we change Proxy to fail on unset impl
	// TODO(asa): Why was this change necessary?
	if err == abi.ErrEmptyArguments || err == vm.ErrExecutionReverted {
		return common.BytesToAddress([]byte{}), ErrRegistryContractNotDeployed
	} else if err != nil {
		return common.BytesToAddress([]byte{}), err
	}

	if contractAddress == common.BytesToAddress([]byte{}) {
		return common.BytesToAddress([]byte{}), ErrSmartContractNotDeployed
	}

	return contractAddress, nil
}
