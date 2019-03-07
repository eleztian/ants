// MIT License

// Copyright (c) 2018 Andy Pan

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package ants

import (
	"github.com/eleztian/gid"
	"log"
	"sync"
	"time"
)

// Worker is the actual executor who runs the tasks,
// it starts a goroutine that accepts tasks and
// performs function calls.
type Worker struct {
	// pool who owns this worker.
	pool *Pool

	// gid is this goroutine id
	gid int64

	// task is a job should be done.
	task chan func()

	// recycleTime will be update when putting a worker back into queue.
	recycleTime time.Time
}

// run starts a goroutine to repeat the process
// that performs the function calls.
func (w *Worker) run(wg *sync.WaitGroup) {
	w.pool.incRunning()
	go func() {
		defer func() {
			if p := recover(); p != nil {
				w.pool.decRunning()
				if w.pool.PanicHandler != nil {
					w.pool.PanicHandler(w.gid, p)
				} else {
					log.Printf("GO %d exits from a panic: %v", w.gid, p)
				}
				wg.Done()
			}
		}()

		for f := range w.task {
			if f == nil {
				w.pool.decRunning()
				w.pool.workerCache.Put(w)
				return
			}
			gid.WithGoID(&w.gid, f)
			wg.Done()
			w.pool.revertWorker(w)

			if DEBUG {
				log.Printf("GO %d finished\n", w.gid)
			}
		}
	}()
}
