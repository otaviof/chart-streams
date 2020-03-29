package chartstreams

import (
	"fmt"
	"strings"
	"time"

	git "github.com/libgit2/git2go/v28"
	log "github.com/sirupsen/logrus"
)

// CommitInfo holds together time and commit hash.
type CommitInfo struct {
	Time     *time.Time // commit time
	ID       string     // commit-id (sha)
	Revision string     // repository branch name
	Digest   string     // data digest
}

// GitRepo represents the actual Git repository, all actions taken on Git backend are here.
type GitRepo struct {
	cfg        *Config
	WorkingDir string
	r          *git.Repository
	head       *git.Oid
	branches   []string
}

// CommitIterFn function to be executed against each commit.
type CommitIterFn func(string, *git.Commit, *git.Tree, bool) error

// branchIterFn function to be executed against each branch.
type branchIterFn func(string, *git.Odb) error

// originPrefix common origin string prefix.
const originPrefix = "origin/"

// checkoutOpts common checkout options
var checkoutOpts = &git.CheckoutOpts{
	Strategy: git.CheckoutSafe | git.CheckoutForce | git.CheckoutUseTheirs,
}

func (g *GitRepo) HeadCommit() (*git.Commit, error) {
	return g.r.LookupCommit(g.head)
}

// sortBranches will sort local list of branches, skipping "master".
func (g *GitRepo) sortBranches() []string {
	sorted := []string{"master"}
	for _, branch := range g.branches {
		if branch == "master" || branch == "origin/master" {
			continue
		}
		sorted = append(sorted, branch)
	}
	return sorted
}

func (g *GitRepo) lookupBranch(branch string) (*git.Branch, error) {
	remoteBranch := fmt.Sprintf("%s%s", originPrefix, branch)
	b, err := g.r.LookupBranch(remoteBranch, git.BranchRemote)
	if err == nil && b != nil {
		log.Infof("Found remote branch '%s'", remoteBranch)
		return b, nil
	}

	log.Infof("Searching for local branch '%s'...", branch)
	return g.r.LookupBranch(branch, git.BranchLocal)
}

func (g *GitRepo) checkoutTree(oid *git.Oid) (*git.Tree, error) {
	tree, err := g.r.LookupTree(oid)
	if err != nil {
		return nil, err
	}
	opts := &git.CheckoutOpts{Strategy: git.CheckoutSafe}
	if err = g.r.CheckoutTree(tree, opts); err != nil {
		return nil, err
	}
	return tree, nil
}

func (g *GitRepo) CheckoutCommit(branch string, c *git.Commit) error {
	log.Infof("Checking out commit-id '%s/%s'", branch, c.Id().String())
	tree, err := g.checkoutTree(c.TreeId())
	if err != nil {
		return err
	}
	defer tree.Free()

	r, err := g.r.References.Create(fmt.Sprintf("refs/head/%s", branch), c.Id(), true, branch)
	if err != nil {
		return err
	}
	defer r.Free()

	head := fmt.Sprintf("refs/heads/%s", branch)
	log.Debugf("Setting head as '%s' for branch '%s'", head, branch)
	g.r.SetHead(head)
	return g.r.CheckoutHead(checkoutOpts)
}

func (g *GitRepo) checkoutBranch(branch string) error {
	log.Infof("Checking out branch '%s' HEAD...", branch)
	b, err := g.lookupBranch(branch)
	if err != nil {
		return err
	}
	defer b.Free()

	c, err := g.r.LookupCommit(b.Target())
	if err != nil {
		return err
	}
	defer c.Free()

	tree, err := g.checkoutTree(c.TreeId())
	if err != nil {
		return err
	}
	defer tree.Free()

	reference := fmt.Sprintf("refs/heads/%s", branch)
	log.Debugf("Setting reference '%s' on branch '%s'", reference, branch)
	g.r.SetHead(reference)
	_, err = g.r.References.Create(reference, c.Id(), true, branch)
	return err
}

func (g *GitRepo) branchIter(fn branchIterFn) error {
	for _, branch := range g.sortBranches() {
		if branch != "master" {
			if err := g.checkoutBranch(branch); err != nil {
				return err
			}
		}

		log.Infof("Transversing '%s' commits...", branch)
		odb, err := g.r.Odb()
		if err != nil {
			return err
		}
		defer odb.Free()

		if err = fn(branch, odb); err != nil {
			return err
		}
	}
	return nil
}

func (g *GitRepo) ModifiedFiles(c *git.Commit, tree *git.Tree) ([]string, error) {
	opts := &git.DiffOptions{}
	modified := []string{}
	parentCount := c.ParentCount()
	for i := uint(0); i <= parentCount; i++ {
		parentID := c.ParentId(i)
		if parentID == nil {
			continue
		}

		log.Debugf("Looking up parent commit-id '%s'", parentID.String())
		parent, err := g.r.LookupCommit(parentID)
		if err != nil {
			return nil, err
		}
		defer parent.Free()

		parentTree, err := parent.Tree()
		if err != nil {
			return nil, err
		}
		defer parentTree.Free()

		diff, err := g.r.DiffTreeToTree(parentTree, tree, opts)
		if err != nil {
			return nil, err
		}
		defer diff.Free()

		_ = diff.ForEach(func(f git.DiffDelta, p float64) (git.DiffForEachHunkCallback, error) {
			modified = append(modified, f.OldFile.Path)
			return nil, nil
		}, git.DiffDetailFiles)
	}
	return modified, nil
}

func (g *GitRepo) CommitIter(fn CommitIterFn) error {
	return g.branchIter(func(branch string, odb *git.Odb) error {
		head, err := g.r.Head()
		if err != nil {
			return err
		}
		defer head.Free()

		c, err := g.r.LookupCommit(head.Target())
		if err != nil {
			return err
		}
		defer c.Free()

		tree, err := c.Tree()
		if err != nil {
			return err
		}
		defer tree.Free()

		if err = fn(branch, c, tree, true); err != nil {
			return err
		}

		var counter = 1
		return odb.ForEach(func(oid *git.Oid) error {
			if g.cfg.CloneDepth > 0 && counter >= g.cfg.CloneDepth {
				return nil
			}

			obj, err := g.r.Lookup(oid)
			if err != nil {
				return err
			}
			if obj.Type() != git.ObjectCommit {
				return nil
			}

			c, err := obj.AsCommit()
			if err != nil {
				return err
			}
			defer c.Free()
			if err = g.CheckoutCommit(branch, c); err != nil {
				return err
			}
			tree, err := c.Tree()
			if err != nil {
				return err
			}
			defer tree.Free()

			counter++
			return fn(branch, c, tree, false)
		})
	})
}

func (g *GitRepo) LookupCommit(id string) (*git.Commit, error) {
	oid, err := git.NewOid(id)
	if err != nil {
		return nil, err
	}
	return g.r.LookupCommit(oid)
}

func extractBranches(r *git.Repository) ([]string, error) {
	iter, err := r.NewBranchIterator(git.BranchAll)
	if err != nil {
		return nil, err
	}
	branchRef := []string{}
	err = iter.ForEach(func(branch *git.Branch, branchType git.BranchType) error {
		name, err := branch.Name()
		if err != nil {
			return err
		}
		name = strings.TrimPrefix(name, originPrefix)
		branchRef = append(branchRef, name)
		return nil
	})
	return branchRef, err
}

func NewGitRepo(cfg *Config, workdingDir string) (*GitRepo, error) {
	log.Infof("Working directory at '%s'", workdingDir)
	opts := &git.CloneOptions{
		FetchOptions: &git.FetchOptions{
			DownloadTags: git.DownloadTagsAll,
		},
		CheckoutOpts: checkoutOpts,
	}
	r, err := git.Clone(cfg.RepoURL, workdingDir, opts)
	if err != nil {
		return nil, err
	}

	head, err := r.Head()
	if err != nil {
		return nil, err
	}
	defer head.Free()

	branches, err := extractBranches(r)
	if err != nil {
		return nil, err
	}

	log.Infof("Repository branches '%v'", branches)
	if !ContainsStringSlice(branches, "master") {
		return nil, fmt.Errorf("can't find 'master' branch in [%v]", branches)
	}

	return &GitRepo{
		cfg:        cfg,
		WorkingDir: workdingDir,
		r:          r,
		head:       head.Target(),
		branches:   branches,
	}, nil
}
