package rmcast
// rqueue, receiver queue which queue unordered pkg
// queue is FIFO data struct and sorted by seq
import (
	"container/list"
	"log"
	"fmt"
)

type LostSeqInfo struct {
	Begin, Count uint32
}

type Rqueue struct {
	lst *list.List
	lostSeq []LostSeqInfo
}

func NewRqueue () *Rqueue {
	return &Rqueue {lst: list.New (), lostSeq: make([]LostSeqInfo, 160)}
}

/*
 * return a list, which contains all the pkg in sequence
 * they should be dispose by upper application.
 */
func (pqueue *Rqueue) Enque (pkg *PKG) {
	insert_seq := pkg.GetSeq ()
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

func (pqueue *Rqueue) GetLostSeqInfo (last_rcv_seq uint32) []LostSeqInfo {
	var i int
	var total_lost_count uint32
	begin_seq := last_rcv_seq + 1
	// fmt.Println ("=================================")
	for e := pqueue.lst.Front (); e != nil; e = e.Next () {
		seq := e.Value.(*PKG).GetSeq ()

		//fmt.Println ("queue seq: ", seq, ", begin_seq: ", begin_seq)

		if seq <= last_rcv_seq {
			log.Fatal ("unorder seq: ", seq, ", last_rcv_seq: ", last_rcv_seq)
		}

		if seq == last_rcv_seq + 1 {
			log.Fatal ("unorder seq: ", seq, ", last_rcv_seq: ", last_rcv_seq)
		}

		if begin_seq == seq {
			begin_seq++
			continue
		}

		if i == 0 {
			pqueue.lostSeq[i].Begin = last_rcv_seq + 1
			pqueue.lostSeq[i].Count = seq - last_rcv_seq - 1
		} else {
			pqueue.lostSeq[i].Begin = begin_seq
			pqueue.lostSeq[i].Count = seq - begin_seq - 1
		}

		total_lost_count += pqueue.lostSeq[i].Count
		fmt.Println ("begin: ", begin_seq, ", count: ", seq - begin_seq - 1)
		begin_seq = seq + 1
		i++
		if i > 160 {
			break
		}

		if total_lost_count > 1000 {
			break
		}
	}

	return pqueue.lostSeq[:i]
}

// test use 
func (pqueue *Rqueue) SeqInQueue (query_seq uint32) bool {
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
