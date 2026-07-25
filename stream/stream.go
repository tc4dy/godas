package stream

import (
	"github.com/tc4dy/godas/dataframe"
	"github.com/tc4dy/godas/io"
)

type Stream struct {
	source     source
	chunkSize  int
	pipes      []PipeFunc
}

type source interface {
	collect(chunkSize int) (*dataframe.DataFrame, error)
}

type csvSource struct {
	path string
	opts []io.CSVOption
}

func (c csvSource) collect(chunkSize int) (*dataframe.DataFrame, error) {
	return io.ReadCSV(c.path, c.opts...)
}

func FromCSV(path string, opts ...io.CSVOption) *Stream {
	return &Stream{
		source:    csvSource{path: path, opts: opts},
		chunkSize: 10000,
	}
}

func (s *Stream) Chunk(size int) *Stream {
	if size > 0 {
		s.chunkSize = size
	}
	return s
}

func (s *Stream) Pipe(fn PipeFunc) *Stream {
	s.pipes = append(s.pipes, fn)
	return s
}

func (s *Stream) Collect() (*dataframe.DataFrame, error) {
	df, err := s.source.collect(s.chunkSize)
	if err != nil {
		return nil, err
	}
	for _, fn := range s.pipes {
		df, err = fn(df)
		if err != nil {
			return nil, err
		}
	}
	return df, nil
}