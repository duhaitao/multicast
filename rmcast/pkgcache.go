package rmcast

import (
	"container/list"
)
// pkgcache store *pkg, pkg has fixed size 4K

type PkgCache struct {
	lst *list.List
	initsize int
}

func NewPkgCache (initsize int) *PkgCache {
	cache := new (PkgCache)
	cache.lst = list.New ()

	if initsize < 512 {
		initsize = 512
	}

	cache.initsize = initsize

	for i := 0; i < initsize; i++ {
		newpkg := NewPKG ()
		cache.lst.PushBack (newpkg)
	}

	return cache
}

func (cache *PkgCache) Get () *PKG {
	var pkg *PKG
	if cache.lst.Len () > 0 {
		elem := cache.lst.Front ()
		pkg = elem.Value.(*PKG)
		cache.lst.Remove (elem)
	} else {
		pkg = NewPKG ()
	}

	return pkg
}

func (cache *PkgCache) Put (pkg *PKG) {
	cache.lst.PushBack (pkg)
}

// PkgCache will increase when peak time, normally, it shouldn't
// occpy memory in peak time, Shink will release some PKG, and shink
// to initsize
func (cache *PkgCache) Shink () {
	listsiz := cache.lst.Len ()
	shinksiz := listsiz / 2

	if listsiz < cache.initsize { // impossible
		return
	}

	if shinksiz < cache.initsize {
		shinksiz = cache.initsize
	}

	for cache.lst.Len () > shinksiz {
		elem := cache.lst.Front ()
		cache.lst.Remove (elem)
		// gc will free PKG
	}
}
