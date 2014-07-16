package rmcast

import (
	//"container/list"
	"sync"
)
// pkgcache store *pkg, pkg has fixed size 4K

type PkgCache struct {
	pool sync.Pool
}

func NewPkgCache (initsize int) *PkgCache {
	pkgcache := new (PkgCache)
	if initsize < 512 {
		initsize = 512
	}

	for i := 0; i < initsize; i++ {
		newpkg := NewPKG ()
		pkgcache.pool.Put (newpkg)
	}

	pkgcache.pool.New = func () interface {} {
		return NewPKG ()
	}
	return pkgcache
}

func (cache *PkgCache) Get () *PKG {
	var pkg *PKG
	pkg = cache.pool.Get ().(*PKG)
	if pkg == nil {
		pkg = NewPKG ()
	}

	return pkg
}

func (cache *PkgCache) Put (pkg *PKG) {
	cache.pool.Put (pkg)
}

/*
type PkgCache struct {
	lst *list.List
	freelst *list.List
	initsize int
	mutex sync.Mutex
}

func NewPkgCache (initsize int) *PkgCache {
	cache := new (PkgCache)
	cache.lst = list.New ()
	cache.freelst = list.New ()

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
		if cache.freelst.Len () > 0 {
			cache.mutex.Lock ()
			cache.lst, cache.freelst = cache.freelst, cache.lst
			cache.mutex.Unlock ()

			elem := cache.lst.Front ()
			pkg = elem.Value.(*PKG)
			cache.lst.Remove (elem)
		} else {
			pkg = NewPKG ()
		}
	}

	return pkg
}

func (cache *PkgCache) Put (pkg *PKG) {
	cache.freelst.PushBack (pkg)
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
*/
