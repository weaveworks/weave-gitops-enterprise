package query

type FluxObject struct{}

type Opts struct {
	Limit  int64
	Offset int64
}

type Key struct {
	Kind        string
	ClusterName string
	Namespace   string
	Name        string
}

type StoreWriter interface {
	AddObject(k Key, obj FluxObject) error
}

type StoreClient interface {
	StoreWriter
	List(opts Opts) ([]FluxObject, error)
}

type QueryService interface {
	List(opts Opts) ([]FluxObject, error)
}

type qs struct {
	c StoreClient
}

func (q *qs) List(opts Opts) ([]FluxObject, error) {
	return []FluxObject{}, nil
}

func NewQueryService(cl StoreClient) QueryService {
	return &qs{c: cl}
}
