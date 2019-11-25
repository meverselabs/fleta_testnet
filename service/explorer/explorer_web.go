package explorer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fletaio/fleta/common/hash"

	"github.com/fletaio/fleta/core/chain"

	"github.com/fletaio/ecrf/process/visit"
	"github.com/fletaio/fleta/core/types"

	"github.com/fletaio/fleta/encoding"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"

	"github.com/fletaio/ecrf/service/explorer/handler"
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/process/formulator"
	"github.com/fletaio/webserver"
)

func (ex *Explorer) initWeb(e *echo.Echo, loginChecker func(next echo.HandlerFunc) echo.HandlerFunc) {
	e.GET("/", ex.HomeHandler, loginChecker)
	e.GET("/login", ex.StaticHandler("login"))
	e.GET("/logout", ex.LogoutHandler)
	e.GET("/formulators", ex.FomrulatorsHandler, loginChecker)
	e.GET("/formulator/:address", ex.FomrulatorHandler, loginChecker)
	e.GET("/blocks", ex.StaticHandler("blocks"), loginChecker)
	e.GET("/block/:hash_or_height", ex.blockDetail, loginChecker)
	e.GET("/txs", ex.StaticHandler("txs"), loginChecker)
	e.GET("/tx/:txid", ex.TransactionDetail, loginChecker)

	gAPI := e.Group("/api")
	gAPI.POST("/login", ex.APILogin())
	gAPI.GET("/datas/mainsummary", ex.MainSummaryHandler, loginChecker)
	gAPI.GET("/datas/transactions", ex.lastestTransactions, loginChecker)
	gAPI.GET("/datas/lasttxs", ex.GetLastTxsHandler, loginChecker)
	gAPI.GET("/datas/lastblocks", ex.GetLastBlocksHandler, loginChecker)
	gAPI.GET("/datas/search", ex.APIDataSearch(), loginChecker)
	gAPI.GET("/formulators", ex.FomrulatorsList, loginChecker)
}

func (ex *Explorer) RunWeb() {
	ex.path = "../../service/explorer/webfiles"
	assets := webserver.NewFileAsset(Assets, ex.path)
	ex.web = webserver.NewWebServer(assets, ex.path, func() {
		handler.LanguageInit(assets)
	})
	handler.LanguageInit(assets)
	e := ex.web.Echo

	loginChecker := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			sess, _ := session.Get("session", c)
			if err != nil {
				return err
			}
			if value, has := sess.Values["login"]; has {
				UserID := value.(string)
				c.Set("user_id", UserID)
				return next(c)
			}
			return c.Redirect(http.StatusFound, "/login")
		}
	}
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
		}
		c.Logger().Error(err, c.Request().URL.String())
		c.HTML(code, err.Error())
	}

	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))

	fs := http.FileServer(assets)
	e.GET("/assets/*", echo.WrapHandler(fs))

	ex.initWeb(e, loginChecker)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%v", ex.port)))
}

func ReturnFile(assets http.FileSystem, path string) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJavaScript)
		if f, err := assets.Open(path); err != nil {
			return err
		} else {
			bs, err := ioutil.ReadAll(f)
			if err != nil {
				return err
			}
			return c.String(http.StatusOK, string(bs))
		}
	}
}

func (ex *Explorer) StaticHandler(page string) func(c echo.Context) error {
	return func(c echo.Context) error {
		args := map[string]interface{}{
			"name": page,
		}
		if err := handler.FillArgs(c, args); err != nil {
			return err
		}
		return c.Render(http.StatusOK, page+".html", args)
	}
}

func (ex *Explorer) LogoutHandler(c echo.Context) error {
	sess, _ := session.Get("session", c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}
	sess.Values = map[interface{}]interface{}{}
	sess.Save(c.Request(), c.Response())
	return c.Redirect(http.StatusFound, "/login")
}

func (ex *Explorer) lastestTransactions(c echo.Context) error { // FIXME
	csStr := c.QueryParam("chartSize")
	cs, err := strconv.Atoi(csStr)
	if err != nil {
		cs = 60
	}

	timeList, err := ex.GetLastTxsForChart(cs)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, timeList)

}

func (ex *Explorer) GetLastTxsForChart(cs int) ([]*ChartTimeData, error) {
	chartSize := uint64(cs)
	timeList := make([]*ChartTimeData, chartSize)
	var step int32 = 10
	var start int32
	end := start + step
	max := uint64(time.Now().UnixNano())
	max /= 1000000000

	for i, _ := range timeList {
		timeList[i] = &ChartTimeData{
			Time: (max - uint64(i)) * 1000000000,
		}
	}
	min := max - chartSize

	_, txs, _, err := ex.getTransactionTotalFromTail(start, end)
	for err == nil && len(txs) > 0 {
		for _, tx := range txs {
			t := tx.Timestamp() / 1000000000
			if min > t {
				return timeList, nil
			}
			if max-t < chartSize {
				timeList[max-t].Count++
			}
		}
		start = end
		end = start + step
		_, txs, _, err = ex.getTransactionTotalFromTail(start, end)
	}
	if err != nil {
		return nil, err
	}
	return timeList, nil
}
func (ex *Explorer) GetLastTxs() ([]*TransactionData, error) {
	//_txids, _txs, results, err := ex.getTransactionTotalFromTail(0, 10)
	_txids, _txs, results, err := ex.getTransactionTotalFromTail(0, 1000)
	if err != nil {
		return nil, err
	}
	txs := []*TransactionData{}
	fc := encoding.Factory("transaction")
	for i, _tx := range _txs {
		var result string
		if results[i] == 1 {
			result = "success"
		} else {
			result = "fail"
		}

		txType, err := fc.TypeOf(_tx)
		if err != nil {
			continue
		}
		txName, err := fc.TypeName(txType)
		if err != nil {
			continue
		}

		tx := &TransactionData{
			Name:      txName,
			TXID:      _txids[i],
			Status:    result,
			Timestamp: _tx.Timestamp(),
		}
		txs = append(txs, tx)
	}
	return txs, nil
}
func (ex *Explorer) GetLastBlocks() ([]*HeaderData, error) {
	blocks, err := ex.getBlocks(ex.Height(), ex.Height()-9, true)
	if err != nil {
		return nil, err
	}
	return blocks, nil
}

func (ex *Explorer) GetLastTxsHandler(c echo.Context) error {
	tx, err := ex.GetLastTxs()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, tx)
}
func (ex *Explorer) GetLastBlocksHandler(c echo.Context) error {
	bs, err := ex.GetLastBlocks()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, bs)
}

func (ex *Explorer) MainSummaryHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, ex.getMainSummary())
}

func (ex *Explorer) HomeHandler(c echo.Context) error {
	args := map[string]interface{}{
		"name": "home",
	}
	err := handler.FillArgs(c, args)
	if err != nil {
		return err
	}
	args["summary"] = ex.getMainSummary()
	args["chartTxs"], err = ex.GetLastTxsForChart(60)
	if err != nil {
		return err
	}
	args["txs"], err = ex.GetLastTxs()
	if err != nil {
		return err
	}
	args["blocks"], err = ex.GetLastBlocks()
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "index.html", args)
}

type formulatorInfos struct {
	Name       string `json:"Name"`
	Address    string `json:"Address"`
	BlockCount uint32 `json:"BlockCount"`
	IsDeleted  bool   `json:"Deleted"`
	Type       string `json:"Type"`
}

type formulatorSummary struct {
	Count   int `json:"count"`
	_Amount *amount.Amount
	Amount  string `json:"amount"`
}

func (ex *Explorer) getFormulatorInfos(getBlockCount map[string]struct{}) map[string]interface{} {
	Summary := map[string]*formulatorSummary{
		"Hyper": &formulatorSummary{_Amount: amount.NewCoinAmount(0, 0)},
		"Omega": &formulatorSummary{_Amount: amount.NewCoinAmount(0, 0)},
		"Sigma": &formulatorSummary{_Amount: amount.NewCoinAmount(0, 0)},
		"Alpha": &formulatorSummary{_Amount: amount.NewCoinAmount(0, 0)},
	}
	cs := ex.cs.Candidates()
	fls := []formulatorInfos{}
	for _, c := range cs {
		addr := c.Address.String()
		acc, err := ex.cn.NewLoaderWrapper(0).Account(c.Address)
		facc, ok := acc.(*formulator.FormulatorAccount)
		if !ok {
			continue
		}
		var Type string
		switch facc.FormulatorType {
		case formulator.HyperFormulatorType:
			Type = "Hyper"
		case formulator.OmegaFormulatorType:
			Type = "Omega"
		case formulator.SigmaFormulatorType:
			Type = "Sigma"
		case formulator.AlphaFormulatorType:
			Type = "Alpha"
		}
		fs := Summary[Type]
		fs.Count++
		if facc.Amount != nil {
			fs._Amount = fs._Amount.Add(facc.Amount)
		}
		fi := formulatorInfos{
			Address: addr,
			Name:    acc.Name(),
			Type:    Type,
		}
		if getBlockCount != nil {
			if _, has := getBlockCount[addr]; has {
				fi.BlockCount = ex.getFormulatorBlockCount(c.Address)
			}
		}
		if err != nil {
			fi.IsDeleted = true
		}
		fls = append(fls, fi)
	}
	for _, fs := range Summary {
		fs.Amount = fs._Amount.String()
		ss := strings.Split(fs.Amount, ".")
		if len(ss) > 1 {
			fs.Amount = ss[0]
		}
	}

	m := map[string]interface{}{}
	m["formulators"] = fls
	m["summary"] = Summary
	return m
}

func (ex *Explorer) FomrulatorsList(c echo.Context) error {
	addrs := map[string]struct{}{}
	for i := 1; ; i++ {
		q := c.Param(fmt.Sprintf("top%n", i))
		if q == "" {
			break
		}
		addrs[q] = struct{}{}
	}

	return c.JSON(http.StatusOK, ex.getFormulatorInfos(addrs))

}

func (ex *Explorer) FomrulatorsHandler(c echo.Context) error {
	args := map[string]interface{}{
		"name": "formulators",
	}
	if err := handler.FillArgs(c, args); err != nil {
		return err
	}

	args["finfo"] = ex.getFormulatorInfos(nil)

	return c.Render(http.StatusOK, "formulators.html", args)
}

func (ex *Explorer) FomrulatorHandler(c echo.Context) error {
	args := map[string]interface{}{
		"name": "formulators",
	}
	if err := handler.FillArgs(c, args); err != nil {
		return err
	}

	Address := c.Param("address")
	addr, err := common.ParseAddress(Address)
	if err != nil {
		return err
	}

	ctw := ex.cn.NewLoaderWrapper(1)
	acc, err := ctw.Account(addr)
	if err != nil {
		return err
	}

	facc, is := acc.(*formulator.FormulatorAccount)
	if !is {
		return ErrIsNotFormulator
	}

	fd := &FormulatorData{
		Addr:           Address,
		Name:           facc.Name(),
		CreateBlock:    facc.UpdatedHeight,
		StakingAmount:  facc.Amount.String(),
		GenerateBlocks: ex.getFormulatorBlockCount(addr),
	}
	switch facc.FormulatorType {
	case formulator.AlphaFormulatorType:
		fd.FormulatorType = "Alpha"
	case formulator.SigmaFormulatorType:
		fd.FormulatorType = "Sigma"
	case formulator.OmegaFormulatorType:
		fd.FormulatorType = "Omega"
	case formulator.HyperFormulatorType:
		fd.FormulatorType = "Hyper"
	}

	args["fd"] = fd

	return c.Render(http.StatusOK, "formulator.html", args)
}

func (ex *Explorer) blockDetail(c echo.Context) error {
	args := map[string]interface{}{
		"name": "block",
	}
	if err := handler.FillArgs(c, args); err != nil {
		return err
	}
	var height uint32
	h := c.Param("hash_or_height")
	if len(h) == 64 {
		bh, err := hash.ParseHash(h)
		if err != nil {
			return err
		}
		v, err := ex.BlockHeightByHash(bh)
		if err != nil {
			return err
		}
		height = v
	} else {
		v, err := strconv.Atoi(h)
		if err != nil {
			return err
		}
		height = uint32(v)
	}
	b, err := ex.cn.Block(uint32(height))
	if err != nil {
		return err
	}

	fAddr := b.Header.Generator
	args["header"] = NewBlockHeaderData(b.Header)
	args["formulator"], _, err = ex.getAccountInfo(fAddr.String())
	if err != nil {
		return err
	}

	txs := []*TransactionDetailData{}
	for i, _ := range b.Transactions {
		m := ex.txInfo(b, uint16(i))
		txs = append(txs, &TransactionDetailData{
			err:            m["err"].(string),
			Type:           m["Type"].(string),
			Result:         m["Result"].(string),
			BlockHash:      m["Block Hash"].(string),
			BlockTimestamp: m["Block Timestamp"].(int64),
			TxHash:         m["Tx Hash"].(string),
			TxTimeStamp:    m["Tx TimeStamp"].(int64),
		})
	}

	args["txs"] = txs

	return c.Render(http.StatusOK, "block.html", args)
}

func (ex *Explorer) TransactionDetail(c echo.Context) error {
	args := map[string]interface{}{
		"name": "Transaction",
	}
	if err := handler.FillArgs(c, args); err != nil {
		return err
	}

	txid := c.Param("txid")
	args["txid"] = txid
	t, result, err := ex.getTransaction(txid)
	if err != nil {
		return err
	}
	args["tx"] = txid
	if result == 0 {
		args["result"] = "fail"
	} else {
		args["result"] = "success"
	}
	if at, is := t.(chain.AccountTransaction); is {
		args["from"], _, err = ex.getAccountInfo(at.From().String())
		if err != nil {
			return err
		}
	}

	switch tx := t.(type) {
	case *visit.UpdateVisitData:
		height, index, err := types.ParseTransactionID(txid)
		if err != nil {
			return err
		}
		b, err := ex.cn.Block(height)
		if err != nil {
			return err
		}
		m := ex.txInfo(b, index)
		bs, err := ex.txDetailMap(tx)
		if err != nil {
			return err
		}
		args["txSummary"] = m
		var dm map[string]interface{}
		json.Unmarshal(bs, &dm)
		args["txDetail"] = dm

		html, err := ex.ecrf.VisitDataForm(tx.VisitID, tx.Data, true)
		if err != nil {
			return err
		}
		args["html"] = html

		TXIDs, err := ex.ecrf.VisitTXIDs(tx.VisitID)
		if err != nil {
			return err
		}

		idx := 0
		for i, TXID := range TXIDs {
			if TXID == txid {
				idx = i
				break
			}
		}

		forms, err := ex.ecrf.StudyMeta()
		if err != nil {
			return err
		}

		itemMap := map[string]string{}
		for _, f := range forms {
			for _, g := range f.Groups {
				for _, item := range g.Items {
					itemMap[item.ID] = item.Name
				}
			}
		}

		codeMap := map[string]string{}
		for _, f := range forms {
			for _, g := range f.Groups {
				for _, item := range g.Items {
					for _, c := range item.Codes {
						codeMap[c.ID] = c.Name
					}
				}
			}
		}

		diffs := [][]string{}
		isFirst := true
		if idx > 0 {
			isFirst = false
			prevHeight, prevIndex, err := types.ParseTransactionID(TXIDs[idx-1])
			if err != nil {
				return err
			}
			prevBlock, err := ex.cn.Block(prevHeight)
			if err != nil {
				return err
			}
			usedKey := map[string]bool{}
			prevTx := prevBlock.Transactions[prevIndex].(*visit.UpdateVisitData)
			tx.Data.EachAll(func(key string, value []byte) bool {
				usedKey[key] = true

				prevValue, _ := prevTx.Data.Get(key)
				if bytes.Compare(value, prevValue) != 0 {
					strs := strings.Split(key, "_")
					ITEMID := strs[0]
					if name, has := itemMap[ITEMID]; has {
						pvs := strings.Split(string(prevValue), "@")
						vs := strings.Split(string(value), "@")
						for i, v := range pvs {
							if cv, has := codeMap[v]; has {
								pvs[i] = cv
							}
						}
						for i, v := range vs {
							if cv, has := codeMap[v]; has {
								vs[i] = cv
							}
						}
						diffs = append(diffs, []string{key, name, strings.Join(pvs, "<hr/>"), strings.Join(vs, "<hr/>")})
					}
				}
				return true
			})
			prevTx.Data.EachAll(func(key string, prevValue []byte) bool {
				if !usedKey[key] {
					value, _ := tx.Data.Get(key)
					if bytes.Compare(value, prevValue) != 0 {
						strs := strings.Split(key, "_")
						ITEMID := strs[0]
						if name, has := itemMap[ITEMID]; has {
							pvs := strings.Split(string(prevValue), "@")
							vs := strings.Split(string(value), "@")
							for i, v := range pvs {
								if cv, has := codeMap[v]; has {
									pvs[i] = cv
								}
							}
							for i, v := range vs {
								if cv, has := codeMap[v]; has {
									vs[i] = cv
								}
							}
							diffs = append(diffs, []string{key, name, strings.Join(pvs, "<hr/>"), strings.Join(vs, "<hr/>")})
						}
					}
				}
				return true
			})
		}
		args["diffs"] = diffs
		args["is_first"] = isFirst

		return c.Render(http.StatusOK, "tx_update_visit_data.html", args)
	default:
		height, index, err := types.ParseTransactionID(txid)
		if err != nil {
			return err
		}
		b, err := ex.cn.Block(height)
		if err != nil {
			return err
		}
		m := ex.txInfo(b, index)
		bs, err := ex.txDetailMap(t)
		if err != nil {
			return err
		}
		args["txSummary"] = m
		var dm map[string]interface{}
		json.Unmarshal(bs, &dm)
		args["txDetail"] = dm

		return c.Render(http.StatusOK, "tx.html", args)
	}
}

type TransactionData struct {
	Name      string `json:"name"`
	Amount    string `json:"amount"`
	TXID      string `json:"txid"`
	Status    string `json:"status"`
	Timestamp uint64 `json:"timestamp"`
}

type HeaderData struct {
	Height    uint32 `json:"height"`
	Hash      string `json:"hash"`
	Timestamp uint64 `json:"timestamp"`
	Generator string `json:"generator"`
}

type ChartTimeData struct {
	Time  uint64 `json:"time"`
	Count uint32 `json:"count"`
}
