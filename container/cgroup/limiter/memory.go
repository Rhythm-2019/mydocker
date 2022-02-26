package limiter

type MemoryLimitInBytesLimiter struct {
	BaseLimiter
}

func (l *MemoryLimitInBytesLimiter) Subsystem() string {
	return MEM_SUBSYSTEM
}
func (l *MemoryLimitInBytesLimiter) Item() string {
	return MEMORY_USAGE_IN_BYTES
}
