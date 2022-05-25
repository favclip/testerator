emulator:
	gcloud --project=unittest beta emulators datastore start --host-port=localhost:8081 --no-store-on-disk --consistency=1.0 

test:
	DATASTORE_EMULATOR_HOST=localhost:8081 DATASTORE_PROJECT_ID=unittest go test ./...
