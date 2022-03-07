package limiter

type CpuCfsPeriodLimiter struct {
	BaseLimiter
}

func (l *CpuCfsPeriodLimiter) Subsystem() string {
	return CPU_SUBSYSTEM
}
func (l *CpuCfsPeriodLimiter) Item() string {
	return CPU_PERIOD
}

type CpuCfsQuotaLimiter struct {
	BaseLimiter
}

func (l *CpuCfsQuotaLimiter) Subsystem() string {
	return CPU_SUBSYSTEM
}
func (l *CpuCfsQuotaLimiter) Item() string {
	return CPU_QUOTE
}
