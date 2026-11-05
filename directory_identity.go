package main

// directoryIdentity uses the same identity fields as os.SameFile.
type directoryIdentity struct {
	device uint64
	file   uint64
}
