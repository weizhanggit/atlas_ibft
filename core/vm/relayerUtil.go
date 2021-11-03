package vm

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/mapprotocol/atlas/params"
	"golang.org/x/crypto/sha3"
	"math/big"
)

type SummayEpochInfo struct {
	EpochID     uint64
	SaCount     uint64
	DaCount     uint64
	BeginHeight uint64
	EndHeight   uint64
	AllAmount   *big.Int
}
type RegisterSummay struct {
	LastReward uint64
	Accounts   uint64
	AllAmount  *big.Int
	Infos      []*SummayEpochInfo
}

func ToJSON(ii *RegisterSummay) map[string]interface{} {
	item := make(map[string]interface{})
	item["lastRewardHeight"] = ii.LastReward
	item["AccountsCounts"] = ii.Accounts
	item["currentAllRegister"] = (*hexutil.Big)(ii.AllAmount)
	items := make([]map[string]interface{}, 0, 0)
	for _, val := range ii.Infos {
		info := make(map[string]interface{})
		info["EpochID"] = val.EpochID
		info["SaCount"] = val.SaCount
		info["DaCount"] = val.DaCount
		info["BeginHeight"] = val.BeginHeight
		info["EndHeight"] = val.EndHeight
		info["AllAmount"] = (*hexutil.Big)(val.AllAmount)
		items = append(items, info)
	}
	item["EpochInfos"] = items
	return item
}

type RewardInfo struct {
	Address  common.Address `json:"Address"`
	Amount   *big.Int       `json:"Amount"`
	Register *big.Int       `json:"register"`
}

func (e *RewardInfo) clone() *RewardInfo {
	return &RewardInfo{
		Address:  e.Address,
		Amount:   new(big.Int).Set(e.Amount),
		Register: new(big.Int).Set(e.Register),
	}
}
func (e *RewardInfo) String() string {
	return fmt.Sprintf("[Address:%v,Amount:%s\n]", e.Address.String(), e.Amount)
}
func (e *RewardInfo) ToJson() map[string]interface{} {
	item := make(map[string]interface{})
	item["Address"] = e.Address
	item["Amount"] = (*hexutil.Big)(e.Amount)
	item["Register"] = (*hexutil.Big)(e.Register)
	return item
}
func FetchOne(sas []*RewardInfos, addr common.Address) []*RewardInfo {
	items := make([]*RewardInfo, 0, 0)
	for _, val := range sas {
		if len(val.Items) > 0 {
			saAddr := val.getSaAddress()
			if bytes.Equal(saAddr.Bytes(), addr.Bytes()) {
				items = mergeRewardInfos(items, val.Items)
			}
		}
	}
	return items
}

func mergeRewardInfos(items1, itmes2 []*RewardInfo) []*RewardInfo {
	for _, v1 := range itmes2 {
		found := false
		for _, v2 := range items1 {
			if bytes.Equal(v1.Address.Bytes(), v2.Address.Bytes()) {
				found = true
				v2.Amount = new(big.Int).Add(v2.Amount, v1.Amount)
			}
		}
		if !found {
			items1 = append(items1, v1)
		}
	}
	return items1
}

type RewardInfos struct {
	Items []*RewardInfo `json:"Items"`
}

func (s *RewardInfos) clone() *RewardInfos {
	var res RewardInfos
	for _, v := range s.Items {
		res.Items = append(res.Items, v.clone())
	}
	return &res
}
func (s *RewardInfos) getSaAddress() common.Address {
	if len(s.Items) > 0 {
		return s.Items[0].Address
	}
	return common.Address{}
}

func (s *RewardInfos) String() string {
	var ss string
	for _, v := range s.Items {
		ss += v.String()
	}
	return ss
}
func (s *RewardInfos) StringToToken() map[string]interface{} {
	ss := make([]map[string]interface{}, 0, 0)
	for _, v := range s.Items {
		ss = append(ss, v.ToJson())
	}
	item := make(map[string]interface{})
	item["SaReward"] = ss
	return item
}

type TimedChainReward struct {
	St     uint64
	Number uint64
	Reward *ChainReward
}

type ChainReward struct {
	Height   uint64
	St       uint64
	CoinBase *RewardInfo `json:"blockminer"`
}

func (s *ChainReward) CoinRewardInfo() map[string]interface{} {
	feild := map[string]interface{}{
		"blockminer": s.CoinBase.ToJson(),
	}
	return feild
}

func (s *ChainReward) RelayerRewardInfo() map[string]interface{} {
	infos := make([]map[string]interface{}, 0, 0)
	feild := map[string]interface{}{
		"RelayerReward": infos,
	}
	return feild
}

func CloneChainReward(reward *ChainReward) *ChainReward {
	var res ChainReward
	res.Height, res.St = reward.Height, reward.St
	res.CoinBase = reward.CoinBase.clone()
	return &res
}

type BalanceInfo struct {
	Address common.Address `json:"address"`
	Valid   *big.Int       `json:"valid"`
	Lock    *big.Int       `json:"lock"`
}

type BlockBalance struct {
	Balance []*BalanceInfo `json:"addrWithBalance"       gencodec:"required"`
}

func (s *BlockBalance) ToMap() map[common.Address]*BalanceInfo {
	infos := make(map[common.Address]*BalanceInfo)
	for _, v := range s.Balance {
		infos[v.Address] = v
	}
	return infos
}

func ToBalanceInfos(items map[common.Address]*BalanceInfo) []*BalanceInfo {
	infos := make([]*BalanceInfo, 0, 0)
	for k, v := range items {
		infos = append(infos, &BalanceInfo{
			Address: k,
			Valid:   new(big.Int).Set(v.Valid),
			Lock:    new(big.Int).Set(v.Lock),
		})
	}
	return infos
}

func NewChainReward(height, tt uint64, coin *RewardInfo, fruits []*RewardInfo, relayer []*RewardInfos) *ChainReward {
	return &ChainReward{
		Height:   height,
		St:       tt,
		CoinBase: coin,
	}
}
func ToRewardInfos1(items map[common.Address]*big.Int) []*RewardInfo {
	infos := make([]*RewardInfo, 0, 0)
	for k, v := range items {
		infos = append(infos, &RewardInfo{
			Address: k,
			Amount:  new(big.Int).Set(v),
		})
	}
	return infos
}
func ToRewardInfos2(items map[common.Address]*big.Int) []*RewardInfos {
	infos := make([]*RewardInfos, 0, 0)
	for k, v := range items {
		items := []*RewardInfo{&RewardInfo{
			Address: k,
			Amount:  new(big.Int).Set(v),
		}}

		infos = append(infos, &RewardInfos{
			Items: items,
		})
	}
	return infos
}
func MergeReward(map1, map2 map[common.Address]*big.Int) map[common.Address]*big.Int {
	for k, v := range map2 {
		if vv, ok := map1[k]; ok {
			map1[k] = new(big.Int).Add(vv, v)
		} else {
			map1[k] = v
		}
	}
	return map1
}

type EpochIDInfo struct {
	EpochID     uint64
	BeginHeight uint64
	EndHeight   uint64
}

func (e *EpochIDInfo) isValid() bool {
	if e.EpochID < 0 {
		return false
	}
	if e.EpochID == 0 && params.PowForkPoint+1 != e.BeginHeight {
		return false
	}
	if e.BeginHeight < 0 || e.EndHeight <= 0 || e.EndHeight <= e.BeginHeight {
		return false
	}
	return true
}
func (e *EpochIDInfo) String() string {
	return fmt.Sprintf("[id:%v,begin:%v,end:%v]", e.EpochID, e.BeginHeight, e.EndHeight)
}

// the key is epochid if RelayerValue as a locked asset,otherwise key is block height if RelayerValue as a register asset
type RelayerValue struct {
	Value map[uint64]*big.Int
}

type LockedItem struct {
	Amount *big.Int
	Locked bool
}

//// LockedValue,the key of Value is epochid
//type LockedValue struct {
//	Value map[uint64]*LockedItem
//}

func (s *RelayerValue) ToUnlockedValue(height uint64) *big.Int {
	value := big.NewInt(0)
	for k, v := range s.Value {
		if IsUnlocked(k, height) {
			value.Add(value, v)
		}
	}
	return value
}

func (s *RelayerValue) ToUnlockingValue(height uint64) *big.Int {
	value := big.NewInt(0)
	for k, v := range s.Value {
		if !IsUnlocked(k, height) {
			value.Add(value, v)
		}
	}
	return value
}

func toReward(val *big.Float) *big.Int {
	val = val.Mul(val, params.FbaseUnit)
	ii, _ := val.Int64()
	return big.NewInt(ii)
}

//func FromBlock(block *SnailBlock) (begin, end uint64) {
//	begin, end = 0, 0
//	l := len(block.Fruits())
//	if l > 0 {
//		begin, end = block.Fruits()[0].FastNumber().Uint64(), block.Fruits()[l-1].FastNumber().Uint64()
//	}
//	return
//}
func GetFirstEpoch() *EpochIDInfo {
	return &EpochIDInfo{
		EpochID:     params.FirstNewEpochID,
		BeginHeight: params.PowForkPoint + 1,
		EndHeight:   params.PowForkPoint + params.NewEpochLength,
	}
}
func GetPreFirstEpoch() *EpochIDInfo {
	return &EpochIDInfo{
		EpochID:     params.FirstNewEpochID,
		BeginHeight: 0,
		EndHeight:   params.PowForkPoint,
	}
}
func GetEpochFromHeight(hh uint64) *EpochIDInfo {
	if hh <= params.PowForkPoint {
		return GetPreFirstEpoch()
	}
	first := GetFirstEpoch()
	if hh <= first.EndHeight {
		return first
	}
	var eid uint64
	if (hh-first.EndHeight)%params.NewEpochLength == 0 {
		eid = (hh-first.EndHeight)/params.NewEpochLength + first.EpochID
	} else {
		eid = (hh-first.EndHeight)/params.NewEpochLength + first.EpochID + 1
	}
	return GetEpochFromID(eid)
}
func GetEpochFromID(eid uint64) *EpochIDInfo {

	first := GetFirstEpoch()
	if first.EpochID >= eid {
		return first
	}
	return &EpochIDInfo{
		EpochID:     eid,
		BeginHeight: (eid-first.EpochID)*params.NewEpochLength + 1,
		EndHeight:   (eid - first.EpochID + 1) * params.NewEpochLength,
	}
}
func GetEpochFromRange(begin, end uint64) []*EpochIDInfo {
	if end == 0 || begin > end || (begin < params.PowForkPoint && end < params.PowForkPoint) {
		return nil
	}
	var ids []*EpochIDInfo
	e1 := GetEpochFromHeight(begin)
	e := uint64(0)

	if e1 != nil {
		ids = append(ids, e1)
		e = e1.EndHeight
	} else {
		e = params.PowForkPoint
	}
	for e < end {
		e2 := GetEpochFromHeight(e + 1)
		if e1.EpochID != e2.EpochID {
			ids = append(ids, e2)
		}
		e = e2.EndHeight
	}

	if len(ids) == 0 {
		return nil
	}
	return ids
}
func CopyVotePk(pk []byte) []byte {
	cc := make([]byte, len(pk))
	copy(cc, pk)
	return cc
}
func ValidPk(pk []byte) error {
	_, err := crypto.UnmarshalPubkey(pk)
	return err
}
func MinCalcRedeemHeight(eid uint64) uint64 {
	e := GetEpochFromID(eid + 1)
	return e.BeginHeight + params.MaxRedeemHeight + 1
}
func ForbidAddress(addr common.Address) error {
	if bytes.Equal(addr[:], params.RelayerAddress[:]) {
		return errors.New(fmt.Sprint("addr error:", addr, params.ErrForbidAddress))
	}
	return nil
}
func IsUnlocked(eid, height uint64) bool {
	e := GetEpochFromID(eid + 1)
	return height > e.BeginHeight+params.MaxRedeemHeight
}

func RlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewLegacyKeccak256()
	if e := rlp.Encode(hw, x); e != nil {
		log.Warn("RlpHash", "error", e.Error())
	}
	hw.Sum(h[:0])
	return h
}
