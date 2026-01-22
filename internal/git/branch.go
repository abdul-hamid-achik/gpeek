package git

import (
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type Branch struct {
	Name      string `json:"name"`
	Hash      string `json:"hash"`
	IsRemote  bool   `json:"is_remote"`
	IsCurrent bool   `json:"is_current"`
	Upstream  string `json:"upstream,omitempty"`
}

func (r *Repository) CurrentBranch() string {
	head, err := r.repo.Head()
	if err != nil {
		return "HEAD"
	}

	if head.Name().IsBranch() {
		return head.Name().Short()
	}

	return head.Hash().String()[:7]
}

func (r *Repository) ListBranches() ([]Branch, error) {
	var branches []Branch

	current := r.CurrentBranch()

	refs, err := r.repo.Branches()
	if err != nil {
		return nil, err
	}

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		name := ref.Name().Short()
		branches = append(branches, Branch{
			Name:      name,
			Hash:      ref.Hash().String(),
			IsRemote:  false,
			IsCurrent: name == current,
		})
		return nil
	})

	if err != nil {
		return nil, err
	}

	return branches, nil
}

func (r *Repository) ListRemoteBranches() ([]Branch, error) {
	var branches []Branch

	refs, err := r.repo.References()
	if err != nil {
		return nil, err
	}

	defaultRemote := r.DefaultRemote()
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsRemote() {
			name := ref.Name().Short()
			name = strings.TrimPrefix(name, defaultRemote+"/")
			branches = append(branches, Branch{
				Name:     name,
				Hash:     ref.Hash().String(),
				IsRemote: true,
			})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return branches, nil
}

func (r *Repository) Checkout(name string) error {
	wt, err := r.repo.Worktree()
	if err != nil {
		return err
	}

	return wt.Checkout(&gogit.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(name),
	})
}

func (r *Repository) CreateBranch(name string) error {
	head, err := r.repo.Head()
	if err != nil {
		return err
	}

	ref := plumbing.NewHashReference(plumbing.NewBranchReferenceName(name), head.Hash())
	return r.repo.Storer.SetReference(ref)
}

func (r *Repository) DeleteBranch(name string) error {
	return r.repo.Storer.RemoveReference(plumbing.NewBranchReferenceName(name))
}

func (r *Repository) IsDetachedHead() bool {
	head, err := r.repo.Head()
	if err != nil {
		return false
	}
	return !head.Name().IsBranch()
}
