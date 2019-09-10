package btc_indexer

type AddressWatcher interface {
	GetAddresses() []string
}
