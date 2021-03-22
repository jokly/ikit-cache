package service

type CacheService struct{}

func (cs *CacheService) Set(key, values string) error {
	return nil
}

func (cs *CacheService) Get(key string) string {
	return key
}
