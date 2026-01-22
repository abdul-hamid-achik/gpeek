package git

import (
	"sort"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Tag struct {
	Name        string    `json:"name"`
	Hash        string    `json:"hash"`
	Message     string    `json:"message,omitempty"`
	Tagger      string    `json:"tagger,omitempty"`
	TaggerTime  time.Time `json:"tagger_time,omitempty"`
	IsAnnotated bool      `json:"is_annotated"`
}

func (r *Repository) ListTags() ([]Tag, error) {
	var tags []Tag

	tagRefs, err := r.repo.Tags()
	if err != nil {
		return nil, err
	}

	err = tagRefs.ForEach(func(ref *plumbing.Reference) error {
		tag := Tag{
			Name: ref.Name().Short(),
			Hash: ref.Hash().String(),
		}

		// Try to get annotated tag info
		tagObj, err := r.repo.TagObject(ref.Hash())
		if err == nil {
			// This is an annotated tag
			tag.IsAnnotated = true
			tag.Message = tagObj.Message
			tag.Tagger = tagObj.Tagger.Name
			tag.TaggerTime = tagObj.Tagger.When
		} else {
			// Lightweight tag - get commit info
			commit, err := r.repo.CommitObject(ref.Hash())
			if err == nil {
				tag.TaggerTime = commit.Author.When
			}
		}

		tags = append(tags, tag)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort by time descending (newest first)
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].TaggerTime.After(tags[j].TaggerTime)
	})

	return tags, nil
}

func (r *Repository) CreateTag(name, message string) error {
	head, err := r.repo.Head()
	if err != nil {
		return err
	}

	commit, err := r.repo.CommitObject(head.Hash())
	if err != nil {
		return err
	}

	if message == "" {
		// Create lightweight tag
		_, err = r.repo.CreateTag(name, head.Hash(), nil)
	} else {
		// Create annotated tag
		_, err = r.repo.CreateTag(name, head.Hash(), &git.CreateTagOptions{
			Message: message,
			Tagger: &object.Signature{
				Name:  commit.Author.Name,
				Email: commit.Author.Email,
				When:  time.Now(),
			},
		})
	}

	return err
}

func (r *Repository) DeleteTag(name string) error {
	return r.repo.DeleteTag(name)
}

func (r *Repository) CheckoutTag(name string) error {
	tagRef, err := r.repo.Tag(name)
	if err != nil {
		return err
	}

	wt, err := r.repo.Worktree()
	if err != nil {
		return err
	}

	// Resolve to commit hash (handle annotated vs lightweight)
	hash := tagRef.Hash()
	tagObj, err := r.repo.TagObject(hash)
	if err == nil {
		// Annotated tag - get the target commit
		hash = tagObj.Target
	}

	return wt.Checkout(&git.CheckoutOptions{
		Hash: hash,
	})
}
