package contracts

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/core/vm"
)

// Method represents a contract's method
type Method struct {
	abi    *abi.ABI
	method string
	maxGas uint64
}

// NewMethod creates a new Method
func NewMethod(abi *abi.ABI, method string, maxGas uint64) Method {
	return Method{
		abi:    abi,
		method: method,
		maxGas: maxGas,
	}
}

/*
   用地址绑定一个合约（Api）上的方法
   resolveAddress: contractAddress
*/
func (am Method) Bind(contractAddress common.Address) *BoundMethod {
	return &BoundMethod{
		Method:         am,
		resolveAddress: noopResolver(contractAddress),
	}
}

// encodeCall will encodes the msg into []byte format for EVM consumption
func (am Method) encodeCall(args ...interface{}) ([]byte, error) {
	return am.abi.Pack(am.method, args...)
}

// decodeResult will decode the result of msg execution into the result parameter
func (am Method) decodeResult(result interface{}, output []byte) error {
	if result == nil {
		return nil
	}

	// adapter new pattern
	results, err := am.abi.Unpack(am.method, output)
	// fetch first data
	result = results[0]
	return err
}

/**
 *@notice NewBoundMethod 构造一个绑定到给定地址的新绑定方法实例。
 *@param  参数
 */
func NewBoundMethod(contractAddress common.Address, abi *abi.ABI, methodName string, maxGas uint64) *BoundMethod {
	return NewMethod(abi, methodName, maxGas).Bind(contractAddress)
}

/*
 @notice 返回&BoundMethod{}
 @param  Method 对应api方法
 @param  resolveAddress 对应合约地址
*/
func NewRegisteredContractMethod(registryId common.Hash, abi *abi.ABI, methodName string, maxGas uint64) *BoundMethod {
	return &BoundMethod{
		Method: NewMethod(abi, methodName, maxGas),
		resolveAddress: func(vmRunner vm.EVMRunner) (common.Address, error) {
			return resolveAddressForCall(vmRunner, registryId, methodName) // 内有到register查询的步骤
		},
	}
}

// BoundMethod represents a Method that is bounded to an address
// In particular, instead of address we use an address resolver to cope the fact
// that addresses need to be obtained from the Registry before making a call
type BoundMethod struct {
	Method
	resolveAddress func(vm.EVMRunner) (common.Address, error)
}

// Query executes the method with the given EVMRunner as a read only action, the returned
// value is unpacked into result.
func (bm *BoundMethod) Query(vmRunner vm.EVMRunner, result interface{}, args ...interface{}) error {
	return bm.run(vmRunner, result, true, nil, args...)
}

// Execute executes the method with the given EVMRunner and unpacks the return value into result.
// If the method does not return a value then result should be nil.
func (bm *BoundMethod) Execute(vmRunner vm.EVMRunner, result interface{}, value *big.Int, args ...interface{}) error {
	return bm.run(vmRunner, result, false, value, args...)
}

func (bm *BoundMethod) run(vmRunner vm.EVMRunner, result interface{}, readOnly bool, value *big.Int, args ...interface{}) error {
	defer meterExecutionTime(bm.method)()

	contractAddress, err := bm.resolveAddress(vmRunner)
	//fmt.Println("zhangwei contractAddress:",contractAddress)
	if err != nil {
		return err
	}

	logger := log.New("to", contractAddress, "method", bm.method)
	input, err := bm.encodeCall(args...)
	if err != nil {
		logger.Error("Error invoking evm function: can't encode method arguments", "args", args, "err", err)
		return err
	}

	var output []byte
	if readOnly {
		output, err = vmRunner.Query(contractAddress, input, bm.maxGas)
	} else {
		output, err = vmRunner.Execute(contractAddress, input, bm.maxGas, value)
	}

	if err != nil {
		message, _ := unpackError(output)
		logger.Error("Error invoking evm function: EVM call failure", "args", args, "input", hexutil.Encode(input), "maxgas", bm.maxGas, "err", err, "message", message)
		return err
	}

	if err := bm.decodeResult(result, output); err != nil {
		logger.Error("Error invoking evm function: can't unpack result", "err", err, "maxgas", bm.maxGas)
		return err
	}

	logger.Trace("EVM call successful", "input", hexutil.Encode(input), "output", hexutil.Encode(output))
	return nil
}
