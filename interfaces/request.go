package interfaces

type RequestInterface interface {
	ThisIsRequest()
	Equals(...interface{}) bool
}
