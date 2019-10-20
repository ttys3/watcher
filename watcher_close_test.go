package watcher

import (
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

func TestDirCreate(t *testing.T) {
	rootDirectory := "/tmp/watcher-test"
	os.RemoveAll(rootDirectory)
	os.Mkdir(rootDirectory, os.ModePerm)

	var numEvtExpected = 10
	var numEvt = 0

	var wg sync.WaitGroup
	var isClosedCorrectly = false
	log.Printf("init the watcher")
	// adds a watcher w which watches files change
	w := New()
	w.FilterOps(Rename, Move, Create, Remove, Write)

	wg.Add(1)
	// in case of file change assign new index
	go func() {
		log.Printf("start the event loop")
		defer func() {
			log.Printf("end the event loop")
			wg.Done()
		}()
		for {
			select {
			//rename sub dir (dir WRITE) = event MOVE all sub dir files
			case event := <-w.Event:
				numEvt++
				log.Printf("event: %#v", event)
			case e := <-w.Error:
				t.Fatalf("err: %#v", e)
				return
			case <-w.Closed:
				log.Printf("recv from w.Closed: watcher closed ============>")
				isClosedCorrectly = true
				return
			}
			time.Sleep(time.Millisecond * 10)
		}
	}()

	err := w.AddRecursive(rootDirectory)
	if err != nil && err != os.ErrPermission && err != os.ErrNotExist && err != os.ErrInvalid {
		if e, ok := err.(*os.PathError); ok {
			//add this path to ignore
			log.Printf("watcher: os.PathError: %#v", e)
			log.Printf("successfully added %s to the watcher, waiting for event ...", rootDirectory)
		} else {
			log.Printf("failed to add %s to the watcher, err: %s", rootDirectory, err.Error())
		}
	} else {
		log.Printf("successfully added %s to the watcher, waiting for event ...", rootDirectory)
	}

	go func() {
		w.Wait()
		time.Sleep(time.Second * 3)
		log.Printf("mkdir abc/def/ghi")
		os.MkdirAll(rootDirectory+"/abc/def/ghi", os.ModePerm)
		time.Sleep(time.Second)

		log.Printf("remove dir abc/def/ghi")
		os.Remove(rootDirectory + "/abc/def/ghi")
		time.Sleep(time.Second)

		log.Printf("remove dir abc/def")
		os.Remove(rootDirectory + "/abc/def")
		time.Sleep(time.Second)

		log.Printf("remove dir abc")
		os.Remove(rootDirectory + "/abc")

		time.Sleep(time.Second * 1)
		// exit the test
		w.Close()
		log.Printf("w.Close() called")
	}()

	// time.Millisecond * 100
	// blocks until all operation has been successfully
	err = w.Start(time.Second * 1)
	if err != nil {
		log.Printf("watcher exited with error: %s", err)
	} else {
		log.Printf("watcher exited with no error")
	}
	// wait recv from w.Closed channel
	wg.Wait()

	if !isClosedCorrectly {
		t.Fail()
	}

	if numEvtExpected != numEvt {
		t.Fail()
	}
}

func TestStartLongIntervalChRecv(t *testing.T) {
	rootDirectory := "/tmp/watcher-test-long-interval"
	os.RemoveAll(rootDirectory)
	os.Mkdir(rootDirectory, os.ModePerm)

	var numEvtExpected = 4
	var numEvt = 0

	var wg sync.WaitGroup
	var isClosedCorrectly = false
	log.Printf("init the watcher")
	// adds a watcher w which watches files change
	w := New()
	w.FilterOps(Rename, Move, Create, Remove, Write)

	wg.Add(1)
	// in case of file change assign new index
	go func() {
		log.Printf("start the event loop")
		defer func() {
			log.Printf("end the event loop")
			wg.Done()
		}()
		for {
			select {
			//rename sub dir (dir WRITE) = event MOVE all sub dir files
			case event := <-w.Event:
				numEvt++
				log.Printf("event: %#v", event)
			case e := <-w.Error:
				log.Printf("err: %#v", e)
				return
			case <-w.Closed:
				log.Printf("recv from w.Closed: watcher closed ============>")
				isClosedCorrectly = true
				return
			}
			time.Sleep(time.Millisecond * 10)
		}
	}()

	err := w.AddRecursive(rootDirectory)
	if err != nil && err != os.ErrPermission && err != os.ErrNotExist && err != os.ErrInvalid {
		if e, ok := err.(*os.PathError); ok {
			//add this path to ignore
			log.Printf("watcher: os.PathError: %#v", e)
			log.Printf("successfully added %s to the watcher, waiting for event ...", rootDirectory)
		} else {
			t.Errorf("failed to add %s to the watcher, err: %s", rootDirectory, err.Error())
		}
	} else {
		log.Printf("successfully added %s to the watcher, waiting for event ...", rootDirectory)
	}

	go func() {
		w.Wait()

		time.Sleep(time.Second * 1)
		os.MkdirAll(rootDirectory+"/abc/def/ghi", os.ModePerm)
		log.Printf("dir created")

		wt := time.Second * 12
		log.Printf("wait for %s and close the watcher", wt)
		time.Sleep(wt)
		// exit the test
		w.Close()
		log.Printf("w.Close() called")
	}()

	// use a very long duration here to test recv from w.Closed
	// blocks until all operation has been successfully
	err = w.Start(time.Second * 10)
	if err != nil {
		log.Printf("watcher exited with error: %s", err)
	} else {
		log.Printf("watcher exited with no error")
	}
	// wait recv from w.Closed channel
	wg.Wait()

	if !isClosedCorrectly {
		t.Fail()
	}

	if numEvtExpected != numEvt {
		t.Fail()
	}
}

func TestWatcherNoPermSubDir(t *testing.T) {
	// prepare
	// as root: mkdir -p /tmp/watcher-test-no-perm/noperm && chmod 700 /tmp/watcher-test-no-perm/noperm && chown root:root /tmp/watcher-test-no-perm/noperm
	rootDirectory := "/tmp/watcher-test-no-perm"
	os.Mkdir(rootDirectory, os.ModePerm)

	var wg sync.WaitGroup
	var isClosedCorrectly = false
	log.Printf("init the watcher")
	// adds a watcher w which watches files change
	w := New()
	w.FilterOps(Rename, Move, Create, Remove, Write)

	wg.Add(1)
	// in case of file change assign new index
	go func() {
		log.Printf("start the event loop")
		defer func() {
			log.Printf("end the event loop")
			wg.Done()
		}()
		for {
			select {
			//rename sub dir (dir WRITE) = event MOVE all sub dir files
			case event := <-w.Event:
				log.Printf("event: %#v", event)
			case e := <-w.Error:
				t.Errorf("err: %#v", e)
				return
			case <-w.Closed:
				log.Printf("recv from w.Closed: watcher closed ============>")
				isClosedCorrectly = true
				return
			}
			//time.Sleep(time.Millisecond * 10)
		}
		log.Printf("end the event loop")
	}()

	err := w.AddRecursive(rootDirectory)
	if err != nil {
		t.Errorf("failed to add %s to the watcher, err: %s", rootDirectory, err.Error())
		t.FailNow()
	} else {
		log.Printf("successfully added %s to the watcher, waiting for event ...", rootDirectory)
	}

	go func() {
		w.Wait()

		time.Sleep(time.Millisecond * 8000)
		log.Printf("before w.Close()")
		// exit the test
		w.Close()
		log.Printf("after w.Close()")
	}()

	// use a very long duration here to test recv from w.Closed
	// blocks until all operation has been successfully
	err = w.Start(time.Second * 2)
	if err != nil {
		log.Printf("watcher exited with error: %s", err)
		// fail if no perm error
		t.FailNow()
	} else {
		log.Printf("watcher exited with no error")
	}
	// wait recv from w.Closed channel
	wg.Wait()

	if !isClosedCorrectly {
		t.Fail()
	}
}
