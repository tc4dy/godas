package dataframe

type PipeFunc func(*DataFrame) (*DataFrame, error)

func Func(fn func(*DataFrame) (*DataFrame, error)) PipeFunc {
	return fn
}
