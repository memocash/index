package queue

type Item struct {
	Topic string
	Uid   []byte
	Data  []byte
}
