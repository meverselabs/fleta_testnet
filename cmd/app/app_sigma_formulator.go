package app

import (
	"github.com/fletaio/fleta_testnet/common"
	"github.com/fletaio/fleta_testnet/core/types"
	"github.com/fletaio/fleta_testnet/process/formulator"
	"github.com/fletaio/fleta_testnet/process/vault"
)

func setupSigmaFormulator(sp *vault.Vault, ctw *types.ContextWrapper, sigmaPolicy *formulator.SigmaPolicy, alphaPolicy *formulator.AlphaPolicy) {
	addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("iUqb4PxXQ12JShdtEsb6SLipFFPHmSLW29zqHKGjvB"), common.MustParsePublicHash("4YjmYcLVvBSmtjh4Z7frRZhWgdEAYTSABCoqqzhKEJa"), common.MustParseAddress("5CyLcFhpyN"), "node1")
	addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3wnNVxsHm3gad87PQG22HkoGZxaXnkQF2bD1P3oJ65U"), common.MustParsePublicHash("EMLGsnW7RvSWTtmArG7aJuASvR7iFwg7uy59FmAwT2"), common.MustParseAddress("56NtzpE5Wh"), "node2")
	addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("324QLx4QrYrh9hE7dQb8xbmy4anyCvn6cGaE5jt3qE"), common.MustParsePublicHash("4ew8HQEwwSqeepMDCnwN9PiYg1uvoeZXyudqdQZBCb3"), common.MustParseAddress("58aNsJ3zfc"), "bluebird")
	addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3z1S6ZzWKGfSHmW519sDBgSvoWJthzcprhJziofdNHQ"), common.MustParsePublicHash("3ZdKaqaCbGSQ5xmAphzVTeEF1eGzX6iU4LLGD2ox2g9"), common.MustParseAddress("51ywFraFCw"), "THSG")
	addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3yADixVW3KxWFhf1dNHkoDFbJCsKLPArQyg5btbh6nB"), common.MustParsePublicHash("3HhrC3gPR951SjnxjnHpfhRSWH1iR3SbCSwtCHvTLuC"), common.MustParseAddress("54BR8LQAMr"), "THSJ")
	addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2Jid4fJm3Kf2GD2hvSMTyCbvW5gGCuo2p2oDWo5GhKT"), common.MustParsePublicHash("27n37VV3ebGWSNH5r9wX3ZhUwzxC2heY34UvXjizLDK"), common.MustParseAddress("4wayWtvQuB"), "hongpa")
	addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3isKGxrjaDhB5cCK85xdqJsKCBSPdSiHbXx6dssrC2S"), common.MustParsePublicHash("CV3cNk8UZxJcsLYjSgMdKuMf7VbDnbHXyqvb2rSE4y"), common.MustParseAddress("4ynTPNkL46"), "hongpa1")
}
