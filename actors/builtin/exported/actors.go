package exported

import (
	"github.com/chenjianmei111/specs-actors/v2/actors/builtin/account"
	"github.com/chenjianmei111/specs-actors/v2/actors/builtin/cron"
	init_ "github.com/chenjianmei111/specs-actors/v2/actors/builtin/init"
	"github.com/chenjianmei111/specs-actors/v2/actors/builtin/market"
	"github.com/chenjianmei111/specs-actors/v2/actors/builtin/miner"
	"github.com/chenjianmei111/specs-actors/v2/actors/builtin/multisig"
	"github.com/chenjianmei111/specs-actors/v2/actors/builtin/paych"
	"github.com/chenjianmei111/specs-actors/v2/actors/builtin/power"
	"github.com/chenjianmei111/specs-actors/v2/actors/builtin/reward"
	"github.com/chenjianmei111/specs-actors/v2/actors/builtin/system"
	"github.com/chenjianmei111/specs-actors/v2/actors/builtin/verifreg"
	"github.com/chenjianmei111/specs-actors/v2/actors/runtime"
)

func BuiltinActors() []runtime.VMActor {
	return []runtime.VMActor{
		account.Actor{},
		cron.Actor{},
		init_.Actor{},
		market.Actor{},
		miner.Actor{},
		multisig.Actor{},
		paych.Actor{},
		power.Actor{},
		reward.Actor{},
		system.Actor{},
		verifreg.Actor{},
	}
}
