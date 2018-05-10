package mutualexclusion

import (
	"sync"
)

var wg sync.WaitGroup

func start(processNumber, takenTimes int) {
	wg.Add(processNumber)

	wg.Wait()
}
