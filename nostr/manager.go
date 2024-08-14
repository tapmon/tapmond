package nostr

const MessageStartRange = 51928

const (
	KindMintedMon = MessageStartRange + iota
	KindFoundMonLevel
)

type Manager struct {
}
