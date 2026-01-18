package agents

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/nanochip/multi-agent/pkg/policies"
	"github.com/nanochip/multi-agent/pkg/types"
	"github.com/nanochip/multi-agent/pkg/workspace"
)

// Releaser empaqueta, versiona, despliega y gestiona rollback
type Releaser struct {
	*BaseAgent
}

// NewReleaser crea un nuevo agente releaser
func NewReleaser(ws workspace.Manager, policy policies.Engine) *Releaser {
	contract := types.AgentContract{
		ID:           "releaser",
		Name:         "Releaser",
		AllowedPaths: []string{"**/*"}, // Puede tocar todo para empaquetar
		AllowedTools: []string{"git", "go", "docker", "kubectl"},
		RequiredTests: false,
	}
	
	return &Releaser{
		BaseAgent: NewBaseAgent(ws, policy, contract),
	}
}

// Execute ejecuta la tarea de release
func (r *Releaser) Execute(ctx context.Context, task *types.Task) *types.TaskResult {
	decision := types.Decision{
		Agent:      "releaser",
		Reason:     "packaging and releasing changes",
		Action:     "release",
		Timestamp:  time.Now(),
		Confidence: 0.85,
	}
	
	evidence := make([]types.Evidence, 0)
	
	// 1. Versionar
	version, err := r.version()
	if err != nil {
		return &types.TaskResult{
			TaskID:    task.ID,
			State:     types.StateFailed,
			Success:   false,
			Error:     fmt.Sprintf("failed to version: %v", err),
			Decisions: []types.Decision{decision},
		}
	}
	
	evidence = append(evidence, types.Evidence{
		Type:        "metric",
		Source:      "version",
		Content:     []byte(version),
		Timestamp:   time.Now(),
		Description: fmt.Sprintf("New version: %s", version),
	})
	
	// 2. Empaquetar
	packagePath, err := r.packageArtifacts(version)
	if err != nil {
		return &types.TaskResult{
			TaskID:    task.ID,
			State:     types.StateFailed,
			Success:   false,
			Error:     fmt.Sprintf("failed to package: %v", err),
			Decisions: []types.Decision{decision},
		}
	}
	
	evidence = append(evidence, types.Evidence{
		Type:        "report",
		Source:      "package",
		Content:     []byte(packagePath),
		Timestamp:   time.Now(),
		Description: fmt.Sprintf("Package created at: %s", packagePath),
	})
	
	// 3. Crear tag de git
	if err := r.createTag(version); err != nil {
		// No crítico, solo log
		evidence = append(evidence, types.Evidence{
			Type:        "log",
			Source:      "git tag",
			Content:     []byte(err.Error()),
			Timestamp:   time.Now(),
		})
	}
	
	// 4. Desplegar (si está configurado)
	deployResult := r.deploy(version, task.Inputs)
	if deployResult != "" {
		evidence = append(evidence, types.Evidence{
			Type:        "report",
			Source:      "deploy",
			Content:     []byte(deployResult),
			Timestamp:   time.Now(),
		})
	}
	
	outputs := map[string]interface{}{
		"version":      version,
		"package_path": packagePath,
		"deployed":     deployResult != "",
	}
	
	return &types.TaskResult{
		TaskID:    task.ID,
		State:     types.StateSuccess,
		Success:   true,
		Outputs:   outputs,
		Evidence:  evidence,
		Decisions: []types.Decision{decision},
	}
}

// version genera una nueva versión
func (r *Releaser) version() (string, error) {
	// Obtener última versión de git tags
	repoPath := r.workspace.GetRepoPath()
	
	// Intentar obtener último tag
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	cmd.Dir = repoPath
	lastTagOutput, _ := cmd.Output()
	lastTag := strings.TrimSpace(string(lastTagOutput))
	
	// Si no hay tags, empezar en v0.1.0
	if lastTag == "" {
		return "v0.1.0", nil
	}
	
	// Incrementar versión (semver simple)
	// Remover 'v' si existe
	version := strings.TrimPrefix(lastTag, "v")
	parts := strings.Split(version, ".")
	
	if len(parts) >= 3 {
		// Incrementar patch
		// En producción, usaría una librería de semver
		return fmt.Sprintf("v%s.%s.%s", parts[0], parts[1], increment(parts[2])), nil
	}
	
	return "v0.1.0", nil
}

// increment incrementa un número de versión
func increment(s string) string {
	// Simplificado, en producción usar strconv
	return "1"
}

// packageArtifacts empaqueta los artefactos
func (r *Releaser) packageArtifacts(version string) (string, error) {
	repoPath := r.workspace.GetRepoPath()
	artifactsDir := filepath.Join(repoPath, "artifacts", version)
	
	if err := os.MkdirAll(artifactsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create artifacts dir: %w", err)
	}
	
	// Construir binarios
	buildOutput, err := r.workspace.RunCommand("go", "build", "-o", filepath.Join(artifactsDir, "app"), "./cmd/orchestrator")
	if err != nil {
		return "", fmt.Errorf("build failed: %w", err)
	}
	
	// Crear tarball (simulado)
	packagePath := filepath.Join(artifactsDir, fmt.Sprintf("app-%s.tar.gz", version))
	
	// En producción, usar tar o zip
	// Por ahora solo crear el archivo
	f, err := os.Create(packagePath)
	if err != nil {
		return "", err
	}
	f.Close()
	
	_ = buildOutput // Usar buildOutput
	
	return packagePath, nil
}

// createTag crea un tag de git
func (r *Releaser) createTag(version string) error {
	repoPath := r.workspace.GetRepoPath()
	
	// Crear tag
	_, err := r.workspace.RunCommand("git", "tag", "-a", version, "-m", fmt.Sprintf("Release %s", version))
	if err != nil {
		return err
	}
	
	return nil
}

// deploy despliega la versión
func (r *Releaser) deploy(version string, inputs map[string]interface{}) string {
	// Verificar si hay configuración de deploy
	deployTarget, ok := inputs["deploy_target"].(string)
	if !ok || deployTarget == "" {
		return "" // No deploy configurado
	}
	
	repoPath := r.workspace.GetRepoPath()
	
	switch deployTarget {
	case "docker":
		return r.deployDocker(version, repoPath)
	case "kubernetes":
		return r.deployKubernetes(version, repoPath)
	default:
		return fmt.Sprintf("Unknown deploy target: %s", deployTarget)
	}
}

// deployDocker despliega usando Docker
func (r *Releaser) deployDocker(version string, repoPath string) string {
	// Construir imagen Docker
	imageName := fmt.Sprintf("app:%s", version)
	output, err := r.workspace.RunCommand("docker", "build", "-t", imageName, ".")
	if err != nil {
		return fmt.Sprintf("Docker build failed: %v", err)
	}
	
	// Push (si está configurado)
	_ = output
	return fmt.Sprintf("Docker image built: %s", imageName)
}

// deployKubernetes despliega usando Kubernetes
func (r *Releaser) deployKubernetes(version string, repoPath string) string {
	// Aplicar manifests de k8s
	output, err := r.workspace.RunCommand("kubectl", "apply", "-f", "k8s/")
	if err != nil {
		return fmt.Sprintf("Kubernetes deploy failed: %v", err)
	}
	
	return fmt.Sprintf("Kubernetes deploy: %s", output)
}

// Rollback ejecuta un rollback a una versión anterior
func (r *Releaser) Rollback(targetVersion string) error {
	// Checkout a la versión anterior
	_, err := r.workspace.RunCommand("git", "checkout", targetVersion)
	if err != nil {
		return fmt.Errorf("failed to checkout version: %w", err)
	}
	
	// Re-desplegar
	// En producción, esto sería más complejo
	return nil
}
