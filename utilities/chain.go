package utilities

import (
	pb "github.com/fieldkit/data-protocol"
)

type Chain struct {
}

func (c *Chain) Begin(df *DataFile) error {
	return nil
}

func (c *Chain) Process(df *DataFile, record *pb.DataRecord) error {
	return nil
}

func (c *Chain) End(df *DataFile) error {
	return nil
}

type TransformerChain struct {
	Chain []RecordTransformer
}

func (t *TransformerChain) Begin(df *DataFile, chain BeginChainFunc) error {
	if len(t.Chain) == 0 {
		return chain(df)
	}
	return t.invokeNextBegin(df, chain, 0)

}

func (t *TransformerChain) Process(df *DataFile, record *pb.DataRecord, begin BeginChainFunc, chain ProcessChainFunc, end EndChainFunc) error {
	if len(t.Chain) == 0 {
		return chain(df, record)
	}
	return t.invokeNextProcess(df, record, begin, chain, end, 0)
}

func (t *TransformerChain) End(df *DataFile, chain EndChainFunc) error {
	if len(t.Chain) == 0 {
		return chain(df)
	}
	return t.invokeNextEnd(df, chain, 0)

}

func (t *TransformerChain) invokeNextBegin(df *DataFile, last BeginChainFunc, n int) error {
	following := last

	if n < len(t.Chain)-1 {
		following = func(df *DataFile) error {
			return t.invokeNextBegin(df, last, n+1)
		}
	}

	return t.Chain[n].Begin(df, following)
}

func (t *TransformerChain) invokeNextProcess(df *DataFile, record *pb.DataRecord, lastBegin BeginChainFunc, lastProcess ProcessChainFunc, lastEnd EndChainFunc, n int) error {
	nextBegin := lastBegin
	nextProcess := lastProcess
	nextEnd := lastEnd

	if n < len(t.Chain)-1 {
		nextBegin = func(df *DataFile) error {
			return t.invokeNextBegin(df, lastBegin, n+1)
		}
		nextProcess = func(df *DataFile, record *pb.DataRecord) error {
			return t.invokeNextProcess(df, record, lastBegin, lastProcess, lastEnd, n+1)
		}
		nextEnd = func(df *DataFile) error {
			return t.invokeNextEnd(df, lastEnd, n+1)
		}
	}

	return t.Chain[n].Process(df, record, nextBegin, nextProcess, nextEnd)
}

func (t *TransformerChain) invokeNextEnd(df *DataFile, last EndChainFunc, n int) error {
	following := last

	if n < len(t.Chain)-1 {
		following = func(df *DataFile) error {
			return t.invokeNextEnd(df, last, n+1)
		}
	}

	return t.Chain[n].End(df, following)
}
