package utils

import "sync"

func FanIn[T any](chs ...chan T) chan T {
	var wg sync.WaitGroup
	outCh := make(chan T, len(chs))

	// определяем функцию output для каждого канала в chs.
	// функция output копирует значения из канала в канал outCh, пока с не будет закрыт.
	output := func(c chan T) {
		for n := range c {
			outCh <- n
		}
		wg.Done()
	}

	// добавляем в группу столько горутин, сколько каналов пришло в fanIn.
	wg.Add(len(chs))
	// перебираем все каналы, которые пришли и отправляем каждый в отдельную горутину.
	for _, c := range chs {
		go output(c)
	}

	// запускаем горутину для закрытия outCh после того, как все горутины отработают
	go func() {
		wg.Wait()
		close(outCh)
	}()

	// возвращаем общий канал
	return outCh
}
