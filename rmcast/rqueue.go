package rmcast
// rqueue, receiver queue which queue unordered pkg
// queue is FIFO data struct and sorted by seq
import (
	"container/list"
	"log"
)

type Rqueue struct {
	lst *list.List
}

func NewRqueue () *Rqueue {
/*
	rqueue := new (Rqueue)
	lst := list.New ()
	rqueue.lst = lst
	return rqueue
*/
	return &Rqueue {lst: list.New ()}
}

/*
 * return a list, which contains all the pkg in sequence
 * they should be dispose by upper application.
 */
func (pqueue *Rqueue) Enque (pkg *PKG) {
	insert_seq := pkg.GetSeq ()
	log.Println ("insert seq: ", insert_seq)
	var seq uint32
	// find the proper place to enque it
	for e := pqueue.lst.Front (); e != nil; e = e.Next () {
		seq = e.Value.(*PKG).GetSeq ()
		if seq == insert_seq { // duplicate 
			return
		}

		if insert_seq < seq {
			pqueue.lst.InsertBefore (pkg, e)
			return
		}
	}
	// new pkg, append it
	pqueue.lst.PushBack (pkg)
	return
}

// deque the first item
func (pqueue *Rqueue) Deque () *PKG {
	first := pqueue.lst.Front ()
	return pqueue.lst.Remove (first).(*PKG)
}

func (pqueue *Rqueue) First () *PKG {
	if pqueue.lst.Len () == 0 {
		return nil
	}

	return pqueue.lst.Front ().Value.(*PKG)
}
