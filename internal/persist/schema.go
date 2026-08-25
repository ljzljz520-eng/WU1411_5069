package persist

var bucketNames = struct {
	Shares  []byte
	Audits  []byte
	Configs []byte
	Windows []byte
}{
	Shares:  []byte("shares"),
	Audits:  []byte("audits"),
	Configs: []byte("configs"),
	Windows: []byte("windows"),
}
