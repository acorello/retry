package retry

type Spec struct {
	// Retries indicates how many times to re-attempt after the first try
	Retries uint
	Stop    func(error) bool
}

// Fn executes fn once. If it returns an error and `Spec.Stop` returns 'false' then it retries for the specified `Spec.Retries`.
func Fn[T any](fn func() (T, error), retry Spec) (result T, err error) {
	maxExecutions := 1 + retry.Retries
	for range maxExecutions {
		result, err = fn()
		if err == nil || retry.Stop != nil && retry.Stop(err) {
			break
		}
	}
	return result, err
}
