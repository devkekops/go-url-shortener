package storage

type LinkRepository interface {
	FindByID(id int64) (string, error)
	Save(link string) (int64, error)
}
