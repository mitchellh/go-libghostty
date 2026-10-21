package libghostty

import (
	"testing"
)

// The benchmarks in this file measure reading the image list from a render
// state, which a renderer does after each update. They use the scenes from
// render_state_benchmark_test.go.

// readRenderStateImages fills it from rs and reads every image, the way a
// renderer rebuilds its queue of images to draw. If queue is nil, it reads
// each image with Info. Otherwise it reads them with InfoInto into the
// reused *queue, like a renderer that keeps its queue between frames. It
// returns the number of images read.
func readRenderStateImages(b *testing.B, rs *RenderState, it *RenderStateImageIterator, queue *[]RenderStateImageInfo) int {
	if err := rs.ImageIterator(it); err != nil {
		b.Fatal(err)
	}
	if queue != nil {
		*queue = (*queue)[:0]
	}
	n := 0
	for it.Next() {
		if queue == nil {
			if _, err := it.Info(); err != nil {
				b.Fatal(err)
			}
		} else {
			*queue = append(*queue, RenderStateImageInfo{})
			if err := it.InfoInto(&(*queue)[len(*queue)-1]); err != nil {
				b.Fatal(err)
			}
		}
		n++
	}
	return n
}

// BenchmarkRenderStateImages measures reading the image list:
//
//   - Read: fill the iterator and read every image with Info.
//   - ReadInto: the same, reading each image with InfoInto into a queue
//     that is reused every time.
//   - Generation: read only [RenderState.ImagePlacementsGeneration], which
//     is all a renderer needs on frames where no image changed.
//
// It reports the number of images per read as images/op.
func BenchmarkRenderStateImages(b *testing.B) {
	for _, scene := range renderStateBenchmarkScenes {
		if !scene.kitty {
			continue
		}

		for _, into := range []bool{false, true} {
			name := scene.name + "/Read"
			if into {
				name = scene.name + "/ReadInto"
			}
			b.Run(name, func(b *testing.B) {
				_, rs := newRenderStateBenchmarkTerminal(b, scene)
				it, err := NewRenderStateImageIterator()
				if err != nil {
					b.Fatal(err)
				}
				b.Cleanup(it.Close)

				var queue *[]RenderStateImageInfo
				if into {
					queue = new([]RenderStateImageInfo)
				}
				n := 0
				b.ReportAllocs()
				for b.Loop() {
					n = readRenderStateImages(b, rs, it, queue)
				}
				b.ReportMetric(float64(n), "images/op")
			})
		}

		b.Run(scene.name+"/Generation", func(b *testing.B) {
			_, rs := newRenderStateBenchmarkTerminal(b, scene)
			b.ReportAllocs()
			for b.Loop() {
				if _, err := rs.ImagePlacementsGeneration(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkRenderStateImageFrame measures a whole frame of image work after
// a program rewrote one row of text, which doesn't change any image. Images
// are read with InfoInto into a reused queue.
//
//   - Always: update, then read the image list every frame.
//   - Generation: update, then read the image list only when
//     [RenderState.ImagePlacementsGeneration] changed, as the
//     documentation recommends.
func BenchmarkRenderStateImageFrame(b *testing.B) {
	for _, scene := range renderStateBenchmarkScenes {
		if !scene.kitty {
			continue
		}

		for _, useGeneration := range []bool{false, true} {
			name := scene.name + "/Always"
			if useGeneration {
				name = scene.name + "/Generation"
			}
			b.Run(name, func(b *testing.B) {
				term, rs := newRenderStateBenchmarkTerminal(b, scene)
				it, err := NewRenderStateImageIterator()
				if err != nil {
					b.Fatal(err)
				}
				b.Cleanup(it.Close)

				var queue []RenderStateImageInfo
				var lastGen uint64
				haveList := false
				frames, reads := 0, 0
				b.ReportAllocs()
				for b.Loop() {
					frames++
					term.VTWrite(renderStateBenchmarkRow)
					if err := rs.Update(term); err != nil {
						b.Fatal(err)
					}
					if useGeneration {
						gen, err := rs.ImagePlacementsGeneration()
						if err != nil {
							b.Fatal(err)
						}
						if haveList && gen == lastGen {
							continue
						}
						lastGen, haveList = gen, true
					}
					readRenderStateImages(b, rs, it, &queue)
					reads++
				}

				// Report how many frames read the list, so a regression in
				// the generation (changing on frames where no image did)
				// shows up as reads/op near 1 instead of near 0.
				b.ReportMetric(float64(reads)/float64(frames), "reads/op")
			})
		}
	}
}
