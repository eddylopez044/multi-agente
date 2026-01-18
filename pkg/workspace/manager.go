package workspace

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/nanochip/multi-agent/pkg/types"
)

// Manager gestiona el workspace git
type Manager struct {
	repoPath      string
	repo          *git.Repository
	baseBranch    string
	currentBranch string
	tmpDir        string
}

// NewManager crea un nuevo workspace manager
func NewManager(repoPath string) (*Manager, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repo: %w", err)
	}

	tmpDir := filepath.Join(repoPath, ".multi-agent", "branches")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create tmp dir: %w", err)
	}

	return &Manager{
		repoPath:   repoPath,
		repo:       repo,
		baseBranch: "main",
		tmpDir:     tmpDir,
	}, nil
}

// CheckoutBranch crea y cambia a una nueva rama
func (m *Manager) CheckoutBranch(branchName string) error {
	worktree, err := m.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Verificar si la rama ya existe
	branchRef := plumbing.NewBranchReferenceName(branchName)
	_, err = m.repo.Reference(branchRef, false)

	if err == nil {
		// La rama existe, hacer checkout
		if err := worktree.Checkout(&git.CheckoutOptions{
			Branch: branchRef,
			Create: false,
		}); err != nil {
			return fmt.Errorf("failed to checkout branch: %w", err)
		}
	} else {
		// Crear nueva rama
		headRef, err := m.repo.Head()
		if err != nil {
			return fmt.Errorf("failed to get HEAD: %w", err)
		}

		newRef := plumbing.NewHashReference(branchRef, headRef.Hash())
		if err := m.repo.Storer.SetReference(newRef); err != nil {
			return fmt.Errorf("failed to create branch: %w", err)
		}

		if err := worktree.Checkout(&git.CheckoutOptions{
			Branch: branchRef,
			Create: false,
		}); err != nil {
			return fmt.Errorf("failed to checkout new branch: %w", err)
		}
	}

	m.currentBranch = branchName
	return nil
}

// GetCurrentBranch retorna la rama actual
func (m *Manager) GetCurrentBranch() string {
	return m.currentBranch
}

// ApplyPatch aplica un patch al workspace
func (m *Manager) ApplyPatch(patch types.Evidence) error {
	// Implementación de aplicación de patch
	// Por ahora, solo loguear
	return nil
}

// GetDiff retorna el diff del workspace actual
func (m *Manager) GetDiff() (string, error) {
	worktree, err := m.repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return "", fmt.Errorf("failed to get status: %w", err)
	}

	diffOutput := ""
	for file, fileStatus := range status {
		diffOutput += fmt.Sprintf("%s %s\n", fileStatus.Staging, file)
	}

	return diffOutput, nil
}

// Commit crea un commit con los cambios actuales
func (m *Manager) Commit(message string) error {
	worktree, err := m.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Agregar todos los cambios
	if _, err := worktree.Add("."); err != nil {
		return fmt.Errorf("failed to add changes: %w", err)
	}

	// Crear commit
	_, err = worktree.Commit(message, &git.CommitOptions{
		Author: &git.Signature{
			Name:  "Multi-Agent System",
			Email: "agent@nanochip.dev",
			When:  time.Now(),
		},
	})

	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

// RunCommand ejecuta un comando en el workspace
func (m *Manager) RunCommand(cmd string, args ...string) (string, error) {
	command := exec.Command(cmd, args...)
	command.Dir = m.repoPath

	output, err := command.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}

// Cleanup limpia recursos temporales
func (m *Manager) Cleanup() error {
	// Volver a la rama base
	worktree, err := m.repo.Worktree()
	if err != nil {
		return err
	}

	if err := worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(m.baseBranch),
	}); err != nil {
		return err
	}

	return nil
}

// GetRepoPath retorna la ruta del repositorio
func (m *Manager) GetRepoPath() string {
	return m.repoPath
}
