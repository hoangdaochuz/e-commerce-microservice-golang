package repo

type IDBConnection interface {
	Connect(string) error
}
