package sites

type PostInterface interface {
	GetAuthor() string
	GetContext() string
	GetURL() string
	GetImages() []string
	GetData() uint64
}
