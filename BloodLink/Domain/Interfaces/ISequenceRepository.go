package Domain

type ISequenceRepository interface {
	GetNextSequenceValue(counterType string, year int) (int, error)
}
