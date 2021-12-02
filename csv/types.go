package csv

// D is an ordered representation of a document. This type should be used when the order of the elements matter.
type D []E

// E represents a element for a D. It is usually used inside a D.
type E struct {
	Key   string
	Value interface{}
}
