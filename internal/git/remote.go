package git

import (
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func (r *Repository) Fetch() error {
	err := r.repo.Fetch(&gogit.FetchOptions{
		RemoteName: r.DefaultRemote(),
	})
	if err == gogit.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

func (r *Repository) Pull() error {
	wt, err := r.repo.Worktree()
	if err != nil {
		return err
	}

	err = wt.Pull(&gogit.PullOptions{
		RemoteName: r.DefaultRemote(),
	})
	if err == gogit.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

func (r *Repository) Push() error {
	return r.repo.Push(&gogit.PushOptions{
		RemoteName: r.DefaultRemote(),
	})
}

func (r *Repository) GetRemotes() ([]string, error) {
	remotes, err := r.repo.Remotes()
	if err != nil {
		return nil, err
	}

	var names []string
	for _, remote := range remotes {
		names = append(names, remote.Config().Name)
	}

	return names, nil
}

func (r *Repository) AddRemote(name, url string) error {
	_, err := r.repo.CreateRemote(&config.RemoteConfig{
		Name: name,
		URLs: []string{url},
	})
	return err
}

func (r *Repository) RemoveRemote(name string) error {
	return r.repo.DeleteRemote(name)
}

func (r *Repository) GetAheadBehind() (ahead, behind int, err error) {
	head, err := r.repo.Head()
	if err != nil {
		return 0, 0, err
	}

	if !head.Name().IsBranch() {
		return 0, 0, nil
	}

	branchName := head.Name().Short()

	remoteBranch, err := r.repo.Reference(plumbing.ReferenceName("refs/remotes/"+r.DefaultRemote()+"/"+branchName), true)
	if err != nil {
		return 0, 0, nil
	}

	localCommit, err := r.repo.CommitObject(head.Hash())
	if err != nil {
		return 0, 0, err
	}

	remoteCommit, err := r.repo.CommitObject(remoteBranch.Hash())
	if err != nil {
		return 0, 0, err
	}

	localIter, err := r.repo.Log(&gogit.LogOptions{From: localCommit.Hash})
	if err != nil {
		return 0, 0, err
	}

	localCommits := make(map[string]bool)
	_ = localIter.ForEach(func(c *object.Commit) error {
		localCommits[c.Hash.String()] = true
		return nil
	})

	remoteIter, err := r.repo.Log(&gogit.LogOptions{From: remoteCommit.Hash})
	if err != nil {
		return 0, 0, err
	}

	remoteCommits := make(map[string]bool)
	_ = remoteIter.ForEach(func(c *object.Commit) error {
		remoteCommits[c.Hash.String()] = true
		return nil
	})

	for hash := range localCommits {
		if !remoteCommits[hash] {
			ahead++
		}
	}

	for hash := range remoteCommits {
		if !localCommits[hash] {
			behind++
		}
	}

	return ahead, behind, nil
}
