package locktest

import "testing"

/*
*
if we delete the code about lock, we can find the counter not equals as 2 * 10000
but once we add these code back, it's always equal to 2 * 10000
*/
func TestLockPractice(t *testing.T) {
	LockPractice()
}
