package study

import "github.com/fletaio/fleta_testnet/core/types"

type Form struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Priority uint64                 `json:"priority"`
	Extra    *types.StringStringMap `json:"extra"`
	Groups   []*Group               `json:"groups"`
}

type Group struct {
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Type  string                 `json:"type"`
	Extra *types.StringStringMap `json:"extra"`
	Items []*Item                `json:"items"`
}

type Item struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	TypeArgv string                 `json:"type_argv"`
	Extra    *types.StringStringMap `json:"extra"`
	Codes    []*Code                `json:"codes"`
}

type Code struct {
	ID    string                 `json:"id"`
	Value string                 `json:"value"`
	Name  string                 `json:"name"`
	Extra *types.StringStringMap `json:"extra"`
}
