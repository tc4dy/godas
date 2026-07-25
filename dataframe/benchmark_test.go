package stream

import "github.com/tc4dy/godas/dataframe"

type PipeFunc func(*dataframe.DataFrame) (*dataframe.DataFrame, error)

func Func(fn func(*dataframe.DataFrame) (*dataframe.DataFrame, error)) PipeFunc {
	return fn
}