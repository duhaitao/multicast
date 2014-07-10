package rmcast
// rqueue, receiver queue which queue unordered pkg
// queue is FIFO data struct
import (
	"container/list"
)

type Rqueue struct {
	lst *list.List
}

func NewRqueue () *Rqueue {
	return &Rqueue {list.New ()}
}

func (pqueue *Rqueue) Enque (pkg *PKG) {
	insert_seq := pkg.GetSeq ()
	var seq uint32
	// find the proper place to enque it
	for e := lst.Front (); e != nil; e = e.Next () {
		seq = e.Value.(*PKG).GetSeq ()
		if seq == insert_seq { // duplicate 
			return
		}

		if insert_seq < seq {
			pqueue.lst.InsertBefore (pkg, e)
			return
		}
	}
	pqueue.lst.PushBack (pkg)
}

// deque the last item
func (pqueue *Rqueue) Deque () *PKG {
	last := pqueue.lst.Back ()
	return pqueue.lst.Remove (last).(*PKG)
}

func (pqueue *RQueue) Walk (action func ()) {
}
