package helper

type IfEntity interface {
	GetName() string
	ToJson() (string, error)
	FromJson([]byte, bool)
	AfterLoad()
}