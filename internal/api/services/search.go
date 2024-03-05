package services

import "sync"

type search struct {
	index sync.Map
}

func NewSearch() *search {
	return &search{}
}

func (s *search) Add(account string, txID uint64) {
	if ids, ok := s.index.Load(account); ok {
		s.index.Store(account, append(ids.([]uint64), txID))
	} else {
		s.index.Store(account, []uint64{txID})
	}
}

func (s *search) Get(account string) []uint64 {
	if ids, ok := s.index.Load(account); ok {
		return ids.([]uint64)
	}
	return nil
}
