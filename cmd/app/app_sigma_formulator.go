package app

import (
	"github.com/fletaio/fleta_testnet/common"
	"github.com/fletaio/fleta_testnet/core/types"
	"github.com/fletaio/fleta_testnet/process/formulator"
	"github.com/fletaio/fleta_testnet/process/vault"
)

func setupSigmaFormulator(sp *vault.Vault, ctw *types.ContextWrapper, sigmaPolicy *formulator.SigmaPolicy, alphaPolicy *formulator.AlphaPolicy) {
	addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("iUqb4PxXQ12JShdtEsb6SLipFFPHmSLW29zqHKGjvB"), common.MustParsePublicHash("4YjmYcLVvBSmtjh4Z7frRZhWgdEAYTSABCoqqzhKEJa"), common.MustParseAddress("5CyLcFhpyN"), "node1")
	addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2Jid4fJm3Kf2GD2hvSMTyCbvW5gGCuo2p2oDWo5GhKT"), common.MustParsePublicHash("27n37VV3ebGWSNH5r9wX3ZhUwzxC2heY34UvXjizLDK"), common.MustParseAddress("4wayWtvQuB"), "hongpa")
	addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4oV8S1dEuTKQrsac7CS81jZdQQpiG31CgoUd66eHXsk"), common.MustParsePublicHash("4GzTnuP7Hky1Dye1AJMLzEXTX2a5kEka5h9AJVvZyTD"), common.MustParseAddress("4sC1mwGabR"), "hongpa2")
	addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("324QLx4QrYrh9hE7dQb8xbmy4anyCvn6cGaE5jt3qE"), common.MustParsePublicHash("4ew8HQEwwSqeepMDCnwN9PiYg1uvoeZXyudqdQZBCb3"), common.MustParseAddress("58aNsJ3zfc"), "bluebird")
	addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("mVssPMvS4RnSK6LmpYrWbXVxxhhE5AAyRbuU8Br74r"), common.MustParsePublicHash("VbMwA5AwSfn93ks8HMv7vvSx4THuzfeefTWVoANEha"), common.MustParseAddress("4gCcRY8zq4"), "zutenbe1")
	addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2Egzma6KP4yERrhEAeBdFiBEhCQHFyDaaJ1vGR1DYKf"), common.MustParsePublicHash("8eDJ3h8DLW8RSovYUjxmcDi1QNvo7UW64MQxGZ9dnS"), common.MustParseAddress("5mwYfT6aH5"), "shin1")
	addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3z1S6ZzWKGfSHmW519sDBgSvoWJthzcprhJziofdNHQ"), common.MustParsePublicHash("3ZdKaqaCbGSQ5xmAphzVTeEF1eGzX6iU4LLGD2ox2g9"), common.MustParseAddress("51ywFraFCw"), "THSG")
	addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4X8Fbz4HurLjpbdBsmhmqNbd8an7aPmCrRPRDLGkqVe"), common.MustParsePublicHash("3UHQyJwSSHHCw29fB5xiGk9W7GNf1DjGC284WhW6jpD"), common.MustParseAddress("4uPVeR6VkL"), "hongpa3")
	addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("49DaZWMvaiJU5DuZGwTJn99sMn4UuTEVU1CUKhHSPSi"), common.MustParsePublicHash("v3GwqbQehcqNVYbRzDk3TDJ7yJ19DgwoamZnMJZuVg"), common.MustParseAddress("4no42yckHf"), "hongpa4")
	addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3yADixVW3KxWFhf1dNHkoDFbJCsKLPArQyg5btbh6nB"), common.MustParsePublicHash("3HhrC3gPR951SjnxjnHpfhRSWH1iR3SbCSwtCHvTLuC"), common.MustParseAddress("54BR8LQAMr"), "THSJ")
}
