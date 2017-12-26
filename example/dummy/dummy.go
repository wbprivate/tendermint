package dummy

import (
	"bytes"
	"fmt"

	"github.com/tendermint/abci/example/code"
	"github.com/tendermint/abci/types"
	"github.com/tendermint/iavl"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
)

var _ types.Application = (*DummyApplication)(nil)

type DummyApplication struct {
	types.BaseApplication

	state *iavl.VersionedTree
}

func NewDummyApplication() *DummyApplication {
	state := iavl.NewVersionedTree(dbm.NewMemDB(), 0)
	return &DummyApplication{state: state}
}

func (app *DummyApplication) Info(req types.RequestInfo) (resInfo types.ResponseInfo) {
	return types.ResponseInfo{Data: fmt.Sprintf("{\"size\":%v}", app.state.Size())}
}

// tx is either "key=value" or just arbitrary bytes
func (app *DummyApplication) DeliverTx(tx []byte) types.ResponseDeliverTx {
	var key, value []byte
	parts := bytes.Split(tx, []byte("="))
	if len(parts) == 2 {
		key, value = parts[0], parts[1]
	} else {
		key, value = tx, tx
	}
	app.state.Set(key, value)

	tags := []cmn.KVPair{
		{[]byte("app.creator"), []byte("jae")},
		{[]byte("app.key"), key},
	}
	return types.ResponseDeliverTx{Code: code.CodeTypeOK, Tags: tags}
}

func (app *DummyApplication) CheckTx(tx []byte) types.ResponseCheckTx {
	return types.ResponseCheckTx{Code: code.CodeTypeOK}
}

func (app *DummyApplication) Commit() types.ResponseCommit {
	// Save a new version
	var hash []byte
	var err error

	if app.state.Size() > 0 {
		hash, _, err = app.state.SaveVersion()
		if err != nil {
			// if this wasn't a dummy app, we'd do something smarter
			panic(err)
		}
	}

	return types.ResponseCommit{Data: hash}
}

func (app *DummyApplication) Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery) {
	if reqQuery.Prove {
		value, proof, err := app.state.GetWithProof(reqQuery.Data)
		// if this wasn't a dummy app, we'd do something smarter
		if err != nil {
			panic(err)
		}
		resQuery.Index = -1 // TODO make Proof return index
		resQuery.Key = reqQuery.Data
		resQuery.Value = value
		resQuery.Proof = proof.Bytes()
		if value != nil {
			resQuery.Log = "exists"
		} else {
			resQuery.Log = "does not exist"
		}
		return
	} else {
		index, value := app.state.Get(reqQuery.Data)
		resQuery.Index = int64(index)
		resQuery.Value = value
		if value != nil {
			resQuery.Log = "exists"
		} else {
			resQuery.Log = "does not exist"
		}
		return
	}
}
