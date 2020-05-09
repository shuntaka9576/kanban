package git

type Repository interface {
	RepoName() string
	RepoOwner() string
}

type baseRepository struct {
	owner, name string
}

func (b *baseRepository) RepoOwner() string {
	return b.owner
}

func (b *baseRepository) RepoName() string {
	return b.name
}

func Repo(owner, name string) (*baseRepository, error) {
	return &baseRepository{
		owner: owner,
		name:  name,
	}, nil

}
