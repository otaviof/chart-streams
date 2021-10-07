package chartstreams

import (
	"fmt"
	"os"
	"strings"
	"time"

	git "github.com/libgit2/git2go/v31"
	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/chart/loader"
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

// checkoutOpts common checkout options, to force and keep a clean tree.
var checkoutOpts = &git.CheckoutOptions{Strategy: git.CheckoutForce | git.CheckoutRemoveUntracked}

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

// lookupBranch search for a remote branch, and if not found, a local branch instead.
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

// checkoutTree execute tree look up and checkout.
func (g *GitRepo) checkoutTree(oid *git.Oid) (*git.Tree, error) {
	tree, err := g.r.LookupTree(oid)
	if err != nil {
		return nil, fmt.Errorf("looking up tree %q: %w", oid, err)
	}
	if err = g.r.CheckoutTree(tree, checkoutOpts); err != nil {
		return nil, fmt.Errorf("checking out tree %q: %w", oid, err)
	}
	return tree, nil
}

// GetFilesFromCommit returns a list of files inside the path for a given
// commit; this list of files is meant to be consumed by Helm's
// `loader.LoadFiles` function.
func (g *GitRepo) GetFilesFromCommit(
	commit *git.Commit,
	path string,
) ([]*loader.BufferedFile, error) {
	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("obtaining commit tree: %w", err)
	}

	// files contains the contents to be returned, ready to be used by
	// `loader.LoadFiles`.
	files := []*loader.BufferedFile{}

	tree.Walk(func(curpath string, te *git.TreeEntry) int {
		// don't even bother looking at something other than blobs.
		if te.Filemode != git.FilemodeBlob {
			return 0
		}

		// in the case `path` and `curpath` are equal the entry should be
		// skipped.
		path := strings.TrimPrefix(curpath, path+"/")
		if path == curpath {
			return 0
		}

		// lookup the entry blob to obtain its contents.
		blob, lookupErr := g.r.LookupBlob(te.Id)
		if lookupErr != nil {
			err = fmt.Errorf("looking up blob for '%s': %w", te.Id, lookupErr)
			return -1 // interrupts tree.Walk
		}

		files = append(
			files,
			&loader.BufferedFile{
				Name: path + te.Name,
				Data: blob.Contents(),
			})

		return 0
	})

	// `err` might be populated if a blob can't be looked up while walking the
	// commit tree.
	return files, err
}

// CheckoutCommit based in branch and commit id, execute tree checkout.
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
	err = g.r.SetHead(head)
	if err != nil {
		return err
	}
	return g.r.CheckoutHead(checkoutOpts)
}

// checkoutBranch look up branch, and look up head commit, with this information it can checkout the
// branch tree and finally, set repository information about new head.
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
	err = g.r.SetHead(reference)
	if err != nil {
		return err
	}
	_, err = g.r.References.Create(reference, c.Id(), true, branch)
	return err
}

// branchIter execute the informed function against each branch in repository.
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

// ModifiedFiles for a given commit and tree, check what are the files that have changed, return them
// as string slice.
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
			return nil, fmt.Errorf("looking up parent commit-id %q: %w", parentID.String(), err)
		}
		defer parent.Free()

		parentTree, err := parent.Tree()
		if err != nil {
			return nil, fmt.Errorf("looking up parent's commit-id %q tree: %w", parentID.String(), err)
		}
		defer parentTree.Free()

		diff, err := g.r.DiffTreeToTree(parentTree, tree, opts)
		if err != nil {
			return nil, fmt.Errorf(
				"creating diff between parent's commit-id %q and %q: %w",
				parentID.String(), c.Id().String(), err)
		}
		defer func() {
			// there isn't anything useful to do.
			_ = diff.Free()
		}()

		_ = diff.ForEach(func(f git.DiffDelta, p float64) (git.DiffForEachHunkCallback, error) {
			modified = append(modified, f.OldFile.Path)
			return nil, nil
		}, git.DiffDetailFiles)
	}
	return modified, nil
}

// CommitIter executed informed function on each branch commit.
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

// LookupCommit search for informed commit id.
func (g *GitRepo) LookupCommit(id string) (*git.Commit, error) {
	oid, err := git.NewOid(id)
	if err != nil {
		return nil, fmt.Errorf("creating Oid for %q: %w", id, err)
	}

	c, err := g.r.LookupCommit(oid)
	if err != nil {
		err = fmt.Errorf("looking up commit %q: %w", id, err)
	}
	return c, err
}

// extractBranches given a repository, inspect branches and return as a string slice.
func extractBranches(r *git.Repository) ([]string, error) {
	iter, err := r.NewBranchIterator(git.BranchAll)
	if err != nil {
		return nil, fmt.Errorf("creating branch iterator: %w", err)
	}
	branchRef := []string{}
	err = iter.ForEach(func(branch *git.Branch, branchType git.BranchType) error {
		name, err := branch.Name()
		if err != nil {
			return fmt.Errorf("obtaining branch name: %w", err)
		}
		name = strings.TrimPrefix(name, originPrefix)
		if name == "HEAD" {
			return nil
		}
		branchRef = append(branchRef, name)
		return nil
	})
	return branchRef, err
}

func (g *GitRepo) FetchBranch(branchName string) error {
	remoteName := "origin"

	log.Infof("Fetching branch '%s' from remote '%s'...", branchName, remoteName)

	remote, err := g.r.Remotes.Lookup(remoteName)
	if err != nil {
		return fmt.Errorf("looking up remote '%s': %w", remoteName, err)
	}

	if err := remote.Fetch([]string{}, nil, ""); err != nil {
		return fmt.Errorf("fetching updates from '%s': %w", remoteName, err)
	}

	return nil
}

// NewGitRepo instantiate git repository by cloning, and extract repository information.
func NewGitRepo(cfg *Config, workdingDir string) (*GitRepo, error) {
	log.Infof("Working directory at '%s'", workdingDir)
	opts := &git.CloneOptions{
		Bare: true,
		FetchOptions: &git.FetchOptions{
			DownloadTags: git.DownloadTagsAll,
		},
	}

	if cfg.ForceClone {
		log.Infof("Removing working-dir '%s' per user request", workdingDir)
		if err := os.RemoveAll(workdingDir); err != nil {
			return nil, fmt.Errorf("removing working-dir '%s': %w", workdingDir, err)
		}
	}

	log.Infof("Cloning repository '%s' on '%s'", cfg.RepoURL, workdingDir)
	repo, err := git.Clone(cfg.RepoURL, workdingDir, opts)
	if err != nil {
		log.Infof("Error cloning repository, will try to recover: %s", err)

		// naively assume the directory contains a git repository and is the
		// right repository
		if repo, err = git.OpenRepository(workdingDir); err != nil {
			return nil, fmt.Errorf("cloning repository '%s' on '%s': %w", cfg.RepoURL, workdingDir, err)
		}

		if err := assertRemoteUrl(repo, "origin", cfg.RepoURL); err != nil {
			log.Errorf("asserting remote '%s' url: %s", cfg.RepoURL, err)
		}

		log.Infof("Repository found in '%s' opened successfully", workdingDir)
	}

	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("obtaining repository's head: %w", err)
	}
	defer head.Free()

	err = repo.CheckoutHead(&git.CheckoutOptions{Strategy: git.CheckoutForce | git.CheckoutRecreateMissing})
	if err != nil {
		return nil, err
	}

	branches, err := extractBranches(repo)
	if err != nil {
		return nil, fmt.Errorf("extracting branches: %w", err)
	}
	log.Infof("Repository branches '%v'", branches)
	// TODO: make the main branch configurable instead of "master"
	if !ContainsStringSlice(branches, "master") {
		return nil, fmt.Errorf("can't find 'master' branch in [%v]", branches)
	}

	return &GitRepo{
		cfg:        cfg,
		WorkingDir: workdingDir,
		r:          repo,
		head:       head.Target(),
		branches:   branches,
	}, nil
}

func assertRemoteUrl(repo *git.Repository, name string, expected string) error {
	r, err := repo.Remotes.Lookup(name)
	if err != nil {
		return fmt.Errorf("looking up remote '%s': %w", name, err)
	}
	if r.Url() != expected {
		return fmt.Errorf("expected repository url '%s', got '%s'", expected, r.Url())
	}
	return nil
}
