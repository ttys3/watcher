package watcher

import (
	"os"
	"testing"
	"time"
)

func TestWatcher(t *testing.T) {
	rootDirectory := "/tmp/watcher-test"
	os.RemoveAll(rootDirectory)
	os.Mkdir(rootDirectory, os.ModePerm)

	var isClosedCorrectly = false
	t.Logf("init the watcher")
	// adds a watcher w which watches files change
	w := New()
	w.FilterOps(Rename, Move, Create, Remove, Write)

	// in case of file change assign new index
	go func() {
		t.Logf("start the event loop")
		for {
			select {
			//rename sub dir (dir WRITE) = event MOVE all sub dir files
			case event := <-w.Event:
				t.Logf("event: %#v", event)
			case e := <-w.Error:
				t.Fatalf("err: %#v", e)
				return
			case <-w.Closed:
				t.Logf("Recieved from w.Closed: watcher closed ============>")
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
			t.Logf("watcher: os.PathError: %#v", e)
			t.Logf("successfully added %s to the watcher, waiting for event ...", rootDirectory)
		} else {
			t.Logf("failed to add %s to the watcher, err: %s", rootDirectory, err.Error())
		}
	} else {
		t.Logf("successfully added %s to the watcher, waiting for event ...", rootDirectory)
	}

	go func() {
		w.Wait()
		time.Sleep(time.Second * 3)
		os.MkdirAll(rootDirectory+"/abc/def/ghi", os.ModePerm)
		time.Sleep(time.Second)
		os.Remove(rootDirectory + "/abc/def/ghi")
		time.Sleep(time.Second)
		os.Remove(rootDirectory + "/abc/def")
		time.Sleep(time.Second)
		os.Remove(rootDirectory + "/abc")

		time.Sleep(time.Second * 1)
		t.Logf("before w.Close()")
		// exit the test
		w.Close()
		t.Logf("after w.Close()")
	}()

	// time.Millisecond * 100
	// blocks until all operation has been successfully
	err = w.Start(time.Second * 3)
	if err != nil {
		t.Logf("watcher exited with error: %s", err)
	} else {
		t.Logf("watcher exited with no error")
	}
	// wait recv from w.Closed channel
	time.Sleep(time.Millisecond * 200)

	if !isClosedCorrectly {
		t.Fail()
	}
}

func TestWatcher_StartLongIntervalChRecv(t *testing.T) {
	rootDirectory := "/tmp/watcher-test-long-interval"
	os.RemoveAll(rootDirectory)
	os.Mkdir(rootDirectory, os.ModePerm)

	var isClosedCorrectly = false
	t.Logf("init the watcher")
	// adds a watcher w which watches files change
	w := New()
	w.FilterOps(Rename, Move, Create, Remove, Write)

	// in case of file change assign new index
	go func() {
		t.Logf("start the event loop")
		for {
			select {
			//rename sub dir (dir WRITE) = event MOVE all sub dir files
			case event := <-w.Event:
				t.Logf("event: %#v", event)
			case e := <-w.Error:
				t.Errorf("err: %#v", e)
				return
			case <-w.Closed:
				t.Logf("Recieved from w.Closed: watcher closed ============>")
				isClosedCorrectly = true
				return
			}
			//time.Sleep(time.Millisecond * 10)
		}
		t.Logf("end the event loop")
	}()

	err := w.AddRecursive(rootDirectory)
	if err != nil && err != os.ErrPermission && err != os.ErrNotExist && err != os.ErrInvalid {
		if e, ok := err.(*os.PathError); ok {
			//add this path to ignore
			t.Errorf("watcher: os.PathError: %#v", e)
			t.Logf("successfully added %s to the watcher, waiting for event ...", rootDirectory)
		} else {
			t.Errorf("failed to add %s to the watcher, err: %s", rootDirectory, err.Error())
		}
	} else {
		t.Logf("successfully added %s to the watcher, waiting for event ...", rootDirectory)
	}

	go func() {
		w.Wait()

		time.Sleep(time.Millisecond * 3000)
		t.Logf("before w.Close()")
		// exit the test
		w.Close()
		t.Logf("after w.Close()")
	}()

	// use a very long duration here to test recv from w.Closed
	// blocks until all operation has been successfully
	err = w.Start(time.Second * 3600)
	if err != nil {
		t.Logf("watcher exited with error: %s", err)
	} else {
		t.Logf("watcher exited with no error")
	}
	// wait recv from w.Closed channel
	time.Sleep(time.Millisecond * 200)

	if !isClosedCorrectly {
		t.Fail()
	}
}

func TestWatcherNoPermSubDir(t *testing.T) {
	// prepare
	// as root: mkdir -p /tmp/watcher-test-no-perm/noperm && chmod 700 /tmp/watcher-test-no-perm/noperm && chown root:root /tmp/watcher-test-no-perm/noperm
	rootDirectory := "/tmp/watcher-test-no-perm"
	os.Mkdir(rootDirectory, os.ModePerm)

	var isClosedCorrectly = false
	t.Logf("init the watcher")
	// adds a watcher w which watches files change
	w := New()
	w.FilterOps(Rename, Move, Create, Remove, Write)

	// in case of file change assign new index
	go func() {
		t.Logf("start the event loop")
		for {
			select {
			//rename sub dir (dir WRITE) = event MOVE all sub dir files
			case event := <-w.Event:
				t.Logf("event: %#v", event)
			case e := <-w.Error:
				t.Errorf("err: %#v", e)
				return
			case <-w.Closed:
				t.Logf("Recieved from w.Closed: watcher closed ============>")
				isClosedCorrectly = true
				return
			}
			//time.Sleep(time.Millisecond * 10)
		}
		t.Logf("end the event loop")
	}()

	err := w.AddRecursive(rootDirectory)
	if err != nil {
		t.Errorf("failed to add %s to the watcher, err: %s", rootDirectory, err.Error())
		t.FailNow()
	} else {
		t.Logf("successfully added %s to the watcher, waiting for event ...", rootDirectory)
	}

	go func() {
		w.Wait()

		time.Sleep(time.Millisecond * 8000)
		t.Logf("before w.Close()")
		// exit the test
		w.Close()
		t.Logf("after w.Close()")
	}()

	// use a very long duration here to test recv from w.Closed
	// blocks until all operation has been successfully
	err = w.Start(time.Second * 2)
	if err != nil {
		t.Logf("watcher exited with error: %s", err)
	} else {
		t.Logf("watcher exited with no error")
	}
	// wait recv from w.Closed channel
	time.Sleep(time.Millisecond * 200)

	if !isClosedCorrectly {
		t.Fail()
	}
}
