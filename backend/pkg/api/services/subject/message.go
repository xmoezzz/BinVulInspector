package subject

type Message interface {
	Payload() ([]byte, error)
	Decode([]byte) error
}
