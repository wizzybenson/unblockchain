package database

type Account string

func NewAccount(name string) Account {
	return Account(name)
}
