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
	addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2tqyyKDje5iiTD8Wvm6VyRSagN1QRzGDwrevhq1kmaJ"), common.MustParsePublicHash("4Ei1HSF3KtDfGrdzHCWfRf4NSTZ2oYCT1CNGFkjV1WB"), common.MustParseAddress("5p92XvvVRz"), "shin2")
	addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4aDsBd3UCd74BsMYpfeZmxeUWeRph9WnDi14HHMFCKT"), common.MustParsePublicHash("3u6v76WAknSq1j86Pfb6p31FsBAJztPdVmY1kkw4k66"), common.MustParseAddress("5hYavVSjyK"), "shin3")
	addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3isKGxrjaDhB5cCK85xdqJsKCBSPdSiHbXx6dssrC2S"), common.MustParsePublicHash("CV3cNk8UZxJcsLYjSgMdKuMf7VbDnbHXyqvb2rSE4y"), common.MustParseAddress("4ynTPNkL46"), "hongpa1")
	addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4nY6bpa94VVybCfEUUHFjgEX1j9wSv5hyoUNBDzmtJd"), common.MustParsePublicHash("38qmoMNCuBht1ihjCKVV5nTWvfiDU7NBNeeHWhB7eT7"), common.MustParseAddress("4pzXuTSfSa"), "hongpa5")
	addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2V6b9no2QhRkapd1hKErJ5DrcQf6UXKvERVHbjXYsRY"), common.MustParsePublicHash("25waPgmJrY3Wy3zoB8yiPA6YtJdGR9ci5mbj5vfwBTN"), common.MustParseAddress("4iQ6J1xuyu"), "hongpa6")
	addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3wnNVxsHm3gad87PQG22HkoGZxaXnkQF2bD1P3oJ65U"), common.MustParsePublicHash("EMLGsnW7RvSWTtmArG7aJuASvR7iFwg7uy59FmAwT2"), common.MustParseAddress("56NtzpE5Wh"), "node2")
	addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3vTRVy1dUXbwzhz6YomUpUWfFT14VCVpjs1S3ySjkCS"), common.MustParsePublicHash("3Uo6d6w1Xrebq1j42Nm2TguHn42R5MgZTMHBwP4HfrX"), common.MustParseAddress("4kbaAVnq8p"), "hongpa7")
	addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("aA8TYVA8Vh3XWaAuy48y8DSFbgwco9oR3uuhTKFRP1"), common.MustParsePublicHash("3EqB9DUVdx6Z9HW8RvbWdg5ybxSaRsdLzc5zT2d3rKE"), common.MustParseAddress("4e18Z4K5g9"), "zutenbe")
}
