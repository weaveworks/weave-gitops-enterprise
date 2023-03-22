package git

type Repository struct {
	Domain string
	Org    string
	Name   string
}

type TreeEntry struct {
	Name string
	Path string
	Type string
	Size int
	SHA  string
	Link string
}
