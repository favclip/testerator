# testerator

appengine unit test accelerator.

## Description

appengine unit test is very slow.
testerator improve testing speed.

following code launch devserver. It's slow (~4s).

```
opt := &aetest.Options{AppID: "unittest", StronglyConsistentDatastore: true}
inst, err := aetest.NewInstance(opt)
```

testerator wrapped devserver spinup.
for example.

```
testerator.SpinUp() // spin up!
testerator.SpinUp() // no effect
testerator.SpinUp() // no effect

testerator.SpinDown() // clear environment! Datastore, Memcache and Search API
testerator.SpinDown() // clear environment! Datastore, Memcache and Search API
testerator.SpinDown() // spin down
```
