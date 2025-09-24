package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	for _, stage := range stages {
		in = stage(in)
		out := make(Bi)
		go func(in In) {
			defer func() {
				close(out)
				for v := range in {
					_ = v
				}
			}()
			for {
				select {
				case <-done:
					return
				case v, ok := <-in:
					if !ok {
						return
					}
					select {
					case <-done:
						return
					case out <- v:
					}
				}
			}
		}(in)
		in = out
	}
	return in
}
