package reward

import (
	"fmt"
	abi "github.com/chenjianmei111/go-state-types/abi"
	big "github.com/chenjianmei111/go-state-types/big"
	"github.com/chenjianmei111/go-state-types/network"
	"github.com/chenjianmei111/specs-actors/actors/builtin"
	"github.com/chenjianmei111/specs-actors/actors/runtime"
	"github.com/chenjianmei111/specs-actors/actors/util/math"
)

const (
	// This number is not exported because it's not suitable for
	// calculations outside reward calculations. Importantly, there are more
	// than 365 days in a year so this number cannot be used to calculate
	// sector lifetimes, etc.
	daysInYear   = 365
	epochsInYear = daysInYear * builtin.EpochsInDay
)

// Baseline function = BaselineInitialValue * (BaselineExponent) ^(t), t in epochs
// Note: we compute exponential iteratively using recurrence e(n) = e * e(n-1).
// Caller of baseline power function is responsible for keeping track of intermediate,
// state e(n-1), the baseline power function just does the next multiplication

// Floor(e^(ln[1 + 200%] / epochsInYear) * 2^128
// Q.128 formatted number such that f(epoch) = baseExponent^epoch grows 200% in one year of epochs
// Calculation here: https://www.wolframalpha.com/input/?i=IntegerPart%5BExp%5BLog%5B1%2B200%25%5D%2F%28%28365+days%29%2F%2830+seconds%29%29%5D*2%5E128%5D
var BaselineExponentV0 = big.MustFromString("340282722551251692435795578557183609728") // Q.128
// Floor(e^(ln[1 + 100%] / epochsInYear) * 2^128
// Q.128 formatted number such that f(epoch) = baseExponent^epoch grows 100% in one year of epochs
// Calculation here: https://www.wolframalpha.com/input/?i=IntegerPart%5BExp%5BLog%5B1%2B100%25%5D%2F%28%28365+days%29%2F%2830+seconds%29%29%5D*2%5E128%5D
var BaselineExponentV3 = big.MustFromString("340282591298641078465964189926313473653") // Q.128

// 1EiB
var BaselineInitialValueV0 = big.Lsh(big.NewInt(1), 60) // Q.0
// 2.5057116798121726 EiB
var BaselineInitialValueV3 = big.NewInt(2_880_000) // Q.0

// Initialize baseline power for epoch -1 so that baseline power at epoch 0 is
// BaselineInitialValue.
func InitBaselinePower() abi.StoragePower {
	baselineInitialValue256 := big.Lsh(BaselineInitialValueV0, 2*math.Precision) // Q.0 => Q.256
	baselineAtMinusOne := big.Div(baselineInitialValue256, BaselineExponentV0)   // Q.256 / Q.128 => Q.128
	return big.Rsh(baselineAtMinusOne, math.Precision)                           // Q.128 => Q.0
}

// Compute BaselinePower(t) from BaselinePower(t-1) with an additional multiplication
// of the base exponent.
func BaselinePowerFromPrev(prevEpochBaselinePower abi.StoragePower, nv network.Version) abi.StoragePower {
	exponent := BaselineExponentV0
	if nv >= network.Version3 {
		exponent = BaselineExponentV3
	}
	thisEpochBaselinePower := big.Mul(prevEpochBaselinePower, exponent) // Q.0 * Q.128 => Q.128
	return big.Rsh(thisEpochBaselinePower, math.Precision)              // Q.128 => Q.0
}

// These numbers are placeholders, but should be in units of attoFIL, 10^-18 FIL
var SimpleTotal = big.Mul(big.NewInt(170999990), big.NewInt(1e18)) // 330M for testnet, PARAM_FINISH
var BaselineTotal = big.Mul(big.NewInt(10), big.NewInt(1e18))      // 770M for testnet, PARAM_FINISH

// Computes RewardTheta which is is precise fractional value of effectiveNetworkTime.
// The effectiveNetworkTime is defined by CumsumBaselinePower(theta) == CumsumRealizedPower
// As baseline power is defined over integers and the RewardTheta is required to be fractional,
// we perform linear interpolation between CumsumBaseline(⌊theta⌋) and CumsumBaseline(⌈theta⌉).
// The effectiveNetworkTime argument is ceiling of theta.
// The result is a fractional effectiveNetworkTime (theta) in Q.128 format.
func computeRTheta(effectiveNetworkTime abi.ChainEpoch, baselinePowerAtEffectiveNetworkTime, cumsumRealized, cumsumBaseline big.Int) big.Int {
	var rewardTheta big.Int
	if effectiveNetworkTime != 0 {
		rewardTheta = big.NewInt(int64(effectiveNetworkTime)) // Q.0
		rewardTheta = big.Lsh(rewardTheta, math.Precision)    // Q.0 => Q.128
		diff := big.Sub(cumsumBaseline, cumsumRealized)
		diff = big.Lsh(diff, math.Precision)                      // Q.0 => Q.128
		diff = big.Div(diff, baselinePowerAtEffectiveNetworkTime) // Q.128 / Q.0 => Q.128
		rewardTheta = big.Sub(rewardTheta, diff)                  // Q.128
	} else {
		// special case for initialization
		rewardTheta = big.Zero()
	}
	return rewardTheta
}

var (
	// lambda = ln(2) / (6 * epochsInYear)
	// for Q.128: int(lambda * 2^128)
	// Calculation here: https://www.wolframalpha.com/input/?i=IntegerPart%5BLog%5B2%5D+%2F+%286+*+%281+year+%2F+30+seconds%29%29+*+2%5E128%5D
	lambda = big.MustFromString("37396271439864487274534522888786")
	// expLamSubOne = e^lambda - 1
	// for Q.128: int(expLamSubOne * 2^128)
	// Calculation here: https://www.wolframalpha.com/input/?i=IntegerPart%5B%5BExp%5BLog%5B2%5D+%2F+%286+*+%281+year+%2F+30+seconds%29%29%5D+-+1%5D+*+2%5E128%5D
	expLamSubOne = big.MustFromString("37396273494747879394193016954629")
)

// Computes a reward for all expected leaders when effective network time changes from prevTheta to currTheta
// Inputs are in Q.128 format
func computeReward(rt runtime.Runtime, epoch abi.ChainEpoch, prevTheta, currTheta big.Int) abi.TokenAmount {
	//1 epoch = 20s, 1576800 epoch 1year
	if epoch < 6307200 || epoch == 6307200 {

		if rt == nil {
			return big.Mul(big.NewInt(11415525114155), big.NewInt(1e6))
		}

		actualreward := big.Sub(SimpleTotal, rt.CurrentBalance())
		expectreward := big.Mul(big.NewInt(int64(epoch)), big.Mul(big.NewInt(11415525114155), big.NewInt(1e6)))

		if -1 == big.Cmp(actualreward, expectreward) {
			return big.Mul(big.NewInt(12915525114155), big.NewInt(1e6))
		} else {
			return big.Mul(big.NewInt(10415525114155), big.NewInt(1e6))
		}
	}
	if 6307200 < epoch && epoch < 12614400 {

		if rt == nil {
			return big.Mul(big.NewInt(57077625570776), big.NewInt(1e5))
		}

		actualreward := big.Sub(SimpleTotal, rt.CurrentBalance())
		firststagereward := big.Mul(big.NewInt(72000000), big.NewInt(1e18))
		secondexpectreward := big.Mul(big.NewInt(int64(epoch-6307200)), big.Mul(big.NewInt(57077625570776), big.NewInt(1e5)))
		totalexpectreward := big.Add(secondexpectreward, firststagereward)

		if -1 == big.Cmp(actualreward, totalexpectreward) {
			return big.Mul(big.NewInt(67077625570776), big.NewInt(1e5))
		} else {
			return big.Mul(big.NewInt(52077625570776), big.NewInt(1e5))
		}
	}
	if 12614400 < epoch && epoch < 18921600 {

		if rt == nil {
			return big.Mul(big.NewInt(28538812785388), big.NewInt(1e5))
		}

		actualreward := big.Sub(SimpleTotal, rt.CurrentBalance())
		frontstagereward := big.Add(big.Mul(big.NewInt(72000000), big.NewInt(1e18)), big.Mul(big.NewInt(36000000), big.NewInt(1e18)))
		currentexpectreward := big.Mul(big.NewInt(int64(epoch-12614400)), big.Mul(big.NewInt(28538812785388), big.NewInt(1e5)))
		totalexpectreward := big.Add(frontstagereward, currentexpectreward)

		if -1 == big.Cmp(actualreward, totalexpectreward) {
			return big.Mul(big.NewInt(33538812785388), big.NewInt(1e5))
		} else {
			return big.Mul(big.NewInt(23538812785388), big.NewInt(1e5))
		}
	}
	if 18921600 < epoch && epoch < 50457600 {

		if rt == nil {
			return big.Mul(big.NewInt(14269406392694), big.NewInt(1e5))
		}

		actualreward := big.Sub(SimpleTotal, rt.CurrentBalance())
		frontstagereward := big.Add(big.Mul(big.NewInt(72000000), big.NewInt(1e18)), big.Mul(big.NewInt(54000000), big.NewInt(1e18)))
		currentexpectreward := big.Mul(big.NewInt(int64(epoch-18921600)), big.Mul(big.NewInt(14269406392694), big.NewInt(1e5)))
		totalexpectreward := big.Add(frontstagereward, currentexpectreward)

		if -1 == big.Cmp(actualreward, totalexpectreward) {
			return big.Mul(big.NewInt(19269406392694), big.NewInt(1e5))
		} else {
			return big.Mul(big.NewInt(9269406392694), big.NewInt(1e5))
		}
	}
	if epoch > 50457600 {
		simpleReward := big.Mul(SimpleTotal, expLamSubOne)    //Q.0 * Q.128 =>  Q.128
		epochLam := big.Mul(big.NewInt(int64(epoch)), lambda) // Q.0 * Q.128 => Q.128

		simpleReward = big.Mul(simpleReward, big.Int{Int: expneg(epochLam.Int)}) // Q.128 * Q.128 => Q.256
		simpleReward = big.Rsh(simpleReward, math.Precision)                     // Q.256 >> 128 => Q.128

		baselineReward := big.Sub(computeBaselineSupply(currTheta), computeBaselineSupply(prevTheta)) // Q.128

		reward := big.Add(simpleReward, baselineReward) // Q.128
		fmt.Printf("reward = %d", reward)

		//return big.Rsh(reward, math.Precision) // Q.128 => Q.0
		return big.Mul(big.NewInt(1), big.NewInt(1e8))
	}
	return big.Mul(big.NewInt(1), big.NewInt(1e8))
}

// Computes baseline supply based on theta in Q.128 format.
// Return is in Q.128 format
func computeBaselineSupply(theta big.Int) big.Int {
	thetaLam := big.Mul(theta, lambda)           // Q.128 * Q.128 => Q.256
	thetaLam = big.Rsh(thetaLam, math.Precision) // Q.256 >> 128 => Q.128

	eTL := big.Int{Int: expneg(thetaLam.Int)} // Q.128

	one := big.NewInt(1)
	one = big.Lsh(one, math.Precision) // Q.0 => Q.128
	oneSub := big.Sub(one, eTL)        // Q.128

	return big.Mul(BaselineTotal, oneSub) // Q.0 * Q.128 => Q.128
}

// SlowConvenientBaselineForEpoch computes baseline power for use in epoch t
// by calculating the value of ThisEpochBaselinePower that shows up in block at t - 1
// It multiplies ~t times so it should not be used in actor code directly.  It is exported as
// convenience for consuming node.
func SlowConvenientBaselineForEpoch(targetEpoch abi.ChainEpoch, nv network.Version) abi.StoragePower {
	baseline := InitBaselinePower()
	baseline = BaselinePowerFromPrev(baseline, nv) // value in genesis block (for epoch 1)
	for i := abi.ChainEpoch(1); i < targetEpoch; i++ {
		baseline = BaselinePowerFromPrev(baseline, nv) // value in block i (for epoch i+1)
	}
	return baseline
}
