package agents

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nanochip/multi-agent/pkg/policies"
	"github.com/nanochip/multi-agent/pkg/types"
	"github.com/nanochip/multi-agent/pkg/workspace"
)

// Release gestiona empaquetado, versionado, despliegue y rollback
type Release struct {
	*BaseAgent
}

// NewRelease crea un nuevo agente de release
func NewRelease(ws workspace.Manager, policy policies.Engine) *Release {
	contract := types.AgentContract{
		ID:           "release",
		Name:         "Release/SRE",
		AllowedPaths: []string{"**/Dockerfile", "**/Makefile", "**/*.yaml", "**/*.yml", "**/go.mod"},
		ForbiddenPaths: []string{"**/*.go"}, // No modifica código
		AllowedTools:   []string{"git", "go", "docker", "kubectl"},
		RequiredTests:  false,
	}
	
	return &Release{
		BaseAgent: NewBaseAgent(ws, policy, contract),
	}
}

// Execute ejecuta la tarea de release
func (r *Release) Execute(ctx context.Context, task *types.Task) *types.TaskResult {
	decision := types.Decision{
		Agent:      "release",
		Reason:     fmt.Sprintf("executing release: %s", task.Objective),
		Action:     "release",
		Timestamp:  time.Now(),
		Confidence: 0.85,
	}
	
	startTime := time.Now()
	
	// Determinar acción basada en el objetivo
	action := r.determineAction(task.Objective)
	
	outputs := make(map[string]interface{})
	evidence := make([]types.Evidence, 0)
	
	switch action {
	case "package":
		result := r.packageArtifacts(task)
		outputs["package"] = result
		evidence = append(evidence, types.Evidence{
			Type:        "report",
			Source:      "package",
			Description: "Build artifacts",
			Timestamp:   time.Now(),
		})
		
	case "version":
		version, err := r.versionArtifacts(task)
		if err != nil {
			return &types.TaskResult{
				TaskID:    task.ID,
				State:     types.StateFailed,
				Success:   false,
				Error:     err.Error(),
				Decisions: []types.Decision{decision},
				Duration:  time.Since(startTime),
			}
		}
		outputs["version"] = version
		evidence = append(evidence, types.Evidence{
			Type:        "report",
			Source:      "version",
			Description: fmt.Sprintf("Versioned as %s", version),
			Timestamp:   time.Now(),
		})
		
	case "deploy":
		result := r.deployArtifacts(task)
		outputs["deploy"] = result
		evidence = append(evidence, types.Evidence{
			Type:        "report",
			Source:      "deploy",
			Description: "Deployment result",
			Timestamp:   time.Now(),
		})
		
	case "rollback":
		result := r.rollbackDeployment(task)
		outputs["rollback"] = result
		evidence = append(evidence, types.Evidence{
			Type:        "report",
			Source:      "rollback",
			Description: "Rollback result",
			Timestamp:   time.Now(),
		})
		
	default:
		// Release completo: package → version → deploy
		packageResult := r.packageArtifacts(task)
		outputs["package"] = packageResult
		
		version, err := r.versionArtifacts(task)
		if err == nil {
			outputs["version"] = version
		}
		
		deployResult := r.deployArtifacts(task)
		outputs["deploy"] = deployResult
	}
	
	success := true
	if errMsg, ok := outputs["error"].(string); ok && errMsg != "" {
		success = false
	}
	
	return &types.TaskResult{
		TaskID:    task.ID,
		State:     mapState(success),
		Success:   success,
		Outputs:   outputs,
		Evidence:  evidence,
		Decisions: []types.Decision{decision},
		Duration:  time.Since(startTime),
	}
}

// determineAction determina qué acción realizar
func (r *Release) determineAction(objective string) string {
	objective = strings.ToLower(objective)
	
	if strings.Contains(objective, "package") || strings.Contains(objective, "build") {
		return "package"
	}
	if strings.Contains(objective, "version") {
		return "version"
	}
	if strings.Contains(objective, "deploy") {
		return "deploy"
	}
	if strings.Contains(objective, "rollback") {
		return "rollback"
	}
	return "full" // Release completo
}

// packageArtifacts empaqueta los artefactos
func (r *Release) packageArtifacts(task *types.Task) map[string]interface{} {
	result := make(map[string]interface{})
	
	repoPath := r.workspace.GetRepoPath()
	
	// Build Go binary
	output, err := r.workspace.RunCommand("go", "build", "-o", "bin/app", "./cmd/...")
	if err != nil {
		result["error"] = fmt.Sprintf("build failed: %v", err)
		return result
	}
	
	result["build_output"] = output
	result["binary_path"] = filepath.Join(repoPath, "bin/app")
	
	// Si hay Dockerfile, construir imagen
	dockerfilePath := filepath.Join(repoPath, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); err == nil {
		imageName := r.getImageName()
		dockerOutput, err := r.workspace.RunCommand("docker", "build", "-t", imageName, ".")
		if err != nil {
			result["docker_error"] = fmt.Sprintf("docker build failed: %v", err)
		} else {
			result["docker_image"] = imageName
			result["docker_output"] = dockerOutput
		}
	}
	
	return result
}

// versionArtifacts versiona los artefactos
func (r *Release) versionArtifacts(task *types.Task) (string, error) {
	// Leer versión actual de go.mod o VERSION file
	repoPath := r.workspace.GetRepoPath()
	versionFile := filepath.Join(repoPath, "VERSION")
	
	var currentVersion string
	if data, err := os.ReadFile(versionFile); err == nil {
		currentVersion = strings.TrimSpace(string(data))
	} else {
		// Intentar leer de go.mod
		goModPath := filepath.Join(repoPath, "go.mod")
		if data, err := os.ReadFile(goModPath); err == nil {
			re := regexp.MustCompile(`module\s+\S+\s+v?(\d+\.\d+\.\d+)`)
			matches := re.FindStringSubmatch(string(data))
			if len(matches) > 1 {
				currentVersion = matches[1]
			}
		}
	}
	
	// Incrementar versión (patch por defecto)
	newVersion := r.incrementVersion(currentVersion, "patch")
	
	// Escribir nueva versión
	if err := os.WriteFile(versionFile, []byte(newVersion), 0644); err != nil {
		return "", fmt.Errorf("failed to write version: %w", err)
	}
	
	// Crear tag git
	tagName := fmt.Sprintf("v%s", newVersion)
	if _, err := r.workspace.RunCommand("git", "tag", tagName); err != nil {
		return newVersion, fmt.Errorf("failed to create tag: %w", err)
	}
	
	return newVersion, nil
}

// incrementVersion incrementa una versión semántica
func (r *Release) incrementVersion(version, level string) string {
	if version == "" {
		return "0.1.0"
	}
	
	re := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)
	matches := re.FindStringSubmatch(version)
	if len(matches) != 4 {
		return "0.1.0"
	}
	
	major := matches[1]
	minor := matches[2]
	patch := matches[3]
	
	switch level {
	case "major":
		major = fmt.Sprintf("%d", parseInt(major)+1)
		minor = "0"
		patch = "0"
	case "minor":
		minor = fmt.Sprintf("%d", parseInt(minor)+1)
		patch = "0"
	default: // patch
		patch = fmt.Sprintf("%d", parseInt(patch)+1)
	}
	
	return fmt.Sprintf("%s.%s.%s", major, minor, patch)
}

func parseInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

// deployArtifacts despliega los artefactos
func (r *Release) deployArtifacts(task *types.Task) map[string]interface{} {
	result := make(map[string]interface{})
	
	// Verificar si hay configuración de Kubernetes
	repoPath := r.workspace.GetRepoPath()
	k8sPath := filepath.Join(repoPath, "k8s")
	
	if _, err := os.Stat(k8sPath); err == nil {
		// Desplegar con kubectl
		output, err := r.workspace.RunCommand("kubectl", "apply", "-f", k8sPath)
		if err != nil {
			result["error"] = fmt.Sprintf("kubectl apply failed: %v", err)
			result["output"] = output
			return result
		}
		result["k8s_output"] = output
		result["deployment_method"] = "kubernetes"
	} else {
		// Otros métodos de despliegue (Docker, etc.)
		result["deployment_method"] = "manual"
		result["message"] = "No deployment configuration found, manual deployment required"
	}
	
	return result
}

// rollbackDeployment hace rollback del despliegue
func (r *Release) rollbackDeployment(task *types.Task) map[string]interface{} {
	result := make(map[string]interface{})
	
	// Obtener versión anterior del tag usando shell
	cmd := exec.Command("sh", "-c", "git tag --sort=-version:refname | head -2")
	cmd.Dir = r.workspace.GetRepoPath()
	output, err := cmd.CombinedOutput()
	if err != nil {
		result["error"] = fmt.Sprintf("failed to get previous version: %v", err)
		return result
	}
	
	// Parsear tags y hacer checkout a la versión anterior
	// Por ahora simplificado
	result["rollback_version"] = "previous"
	result["message"] = "Rollback initiated"
	
	// Si hay k8s, hacer rollback
	repoPath := r.workspace.GetRepoPath()
	k8sPath := filepath.Join(repoPath, "k8s")
	if _, err := os.Stat(k8sPath); err == nil {
		k8sOutput, err := r.workspace.RunCommand("kubectl", "rollout", "undo", "deployment/app")
		if err != nil {
			result["k8s_error"] = fmt.Sprintf("kubectl rollout failed: %v", err)
		} else {
			result["k8s_output"] = k8sOutput
		}
	}
	
	return result
}

// getImageName genera un nombre de imagen Docker
func (r *Release) getImageName() string {
	// Por defecto, usar nombre del repo
	repoPath := r.workspace.GetRepoPath()
	repoName := filepath.Base(repoPath)
	return fmt.Sprintf("%s:latest", repoName)
}

// mapState convierte un bool a TaskState
func mapState(success bool) types.TaskState {
	if success {
		return types.StateSuccess
	}
	return types.StateFailed
}
