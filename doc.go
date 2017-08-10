/*
Package testerator is TEST execution accelERATOR for GoogleAppEngine/Go starndard environment.

When you call the `aetest.NewInstance`, dev_appserver.py will spinup every time.
This is very high cost operation.
Testerator compress the time of dev_appserver.py operation.

following code launch dev server. It's slow (~4s).

	opt := &aetest.Options{AppID: "unittest", StronglyConsistentDatastore: true}
	inst, err := aetest.NewInstance(opt)

testerator wrapped devserver spinup.
for example.

	func TestMain(m *testing.M) {
		_, _, err := testerator.SpinUp()
		if err != nil {
			fmt.Printf(err.Error())
			os.Exit(1)
		}

		status := m.Run()

		err = testerator.SpinDown()
		if err != nil {
			fmt.Printf(err.Error())
			os.Exit(1)
		}

		os.Exit(status)
	}

	func TestFooBar(t *testing.T) {
		_, c, err := testerator.SpinUp()
		if err != nil {
			t.Fatal(err.Error())
		}
		defer testerator.SpinDown()

		// write some test!
	}

If you want to clean up Datastore or Search API or Memcache, You should import above packages.

	import (
		// do testerator feature setup
		_ "github.com/favclip/testerator/datastore"
		_ "github.com/favclip/testerator/search"
		_ "github.com/favclip/testerator/memcache"
	)

*/
package testerator
