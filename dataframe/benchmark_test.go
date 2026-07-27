package dataframe

type PipeFunc func(*dataframe.DataFrame) (*dataframe.DataFrame, error)

func Func(fn func(*dataframe.DataFrame) (*dataframe.DataFrame, error)) PipeFunc {
	return fn
}
