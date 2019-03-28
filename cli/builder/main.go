package builder

type ComposeBuilder interface {
	New(rootPath string)
	Build() error
	Delete() error
}
