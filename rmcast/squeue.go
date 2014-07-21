package rmcast
// squeue, sender queue which queue unacked pkg
import (
	"container/list"
	//"log"
	"fmt"
)

type Squeue struct {
	lst *list.List
}

func NewSqueue () *Rqueue {
	fmt.Println ("create Squeue")
	return &Rqueue {lst: list.New ()}
}

/*
 * return a list, which contains all the pkg in sequence
 * they should be dispose by upper application.
 */
func (pqueue *Squeue) Enque (pkg *PKG) {
	// new pkg, append it
	pqueue.lst.PushBack (pkg)
	return
}

// deque the first item
func (pqueue *Squeue) Deque () *PKG {
	first := pqueue.lst.Front ()
	return pqueue.lst.Remove (first).(*PKG)
}

func (pqueue *Squeue) First () *PKG {
	if pqueue.lst.Len () == 0 {
		return nil
	}

	return pqueue.lst.Front ().Value.(*PKG)
}

// test use 
func (pqueue *Squeue) SeqInQueue (query_seq uint32) bool {
	for e := pqueue.lst.Front (); e != nil; e = e.Next () {
		seq := e.Value.(*PKG).GetSeq ()
		if query_seq > seq {
			return false
		}

		if query_seq == seq {
			return true
		}
	}

	return false
}
