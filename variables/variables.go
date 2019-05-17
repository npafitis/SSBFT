package variables


var (
	//Number of processors
	N int

	F int
	//This processor's id.
	Id int
	// This processor's view.
	View int
	// This processor's known current primary
	Prim int

	T    int
	// Size of Clients Set
	K int
)

func Initialise(id int, n int, t int, k int) {
	N = n
	Id = id
	F = (N - 1) / 5
	T = t
	K = k
}
