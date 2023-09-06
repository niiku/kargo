package stages

import (
	"context"
	"fmt"

	kargoapi "github.com/akuity/kargo/api/v1alpha1"
	libArgoCD "github.com/akuity/kargo/internal/argocd"
)

func (r *reconciler) checkHealth(
	ctx context.Context,
	currentFreight kargoapi.Freight,
	argoCDAppUpdates []kargoapi.ArgoCDAppUpdate,
) kargoapi.Health {
	if len(argoCDAppUpdates) == 0 {
		return kargoapi.Health{
			Status: kargoapi.HealthStateUnknown,
			Issues: []string{
				"no spec.promotionMechanisms.argoCDAppUpdates are defined",
			},
		}
	}

	health := kargoapi.Health{
		// We'll start healthy and degrade as we find issues
		Status: kargoapi.HealthStateHealthy,
		Issues: []string{},
	}

	for _, check := range argoCDAppUpdates {
		app, err :=
			r.getArgoCDAppFn(ctx, r.argoClient, check.AppNamespace, check.AppName)
		if err != nil {
			if health.Status != kargoapi.HealthStateUnhealthy {
				health.Status = kargoapi.HealthStateUnknown
			}
			health.Issues = append(
				health.Issues,
				fmt.Sprintf(
					"error finding Argo CD Application %q in namespace %q: %s",
					check.AppName,
					check.AppNamespace,
					err,
				),
			)
		} else if app == nil {
			if health.Status != kargoapi.HealthStateUnhealthy {
				health.Status = kargoapi.HealthStateUnknown
			}
			health.Issues = append(
				health.Issues,
				fmt.Sprintf(
					"unable to find Argo CD Application %q in namespace %q",
					check.AppName,
					check.AppNamespace,
				),
			)
		} else if len(app.Spec.Sources) > 0 {
			if health.Status != kargoapi.HealthStateUnhealthy {
				health.Status = kargoapi.HealthStateUnknown
			}
			health.Issues = append(
				health.Issues,
				fmt.Sprintf(
					"bugs in Argo CD currently prevent a comprehensive assessment of "+
						"the health of multi-source Application %q in namespace %q",
					check.AppName,
					check.AppNamespace,
				),
			)
		} else {
			var desiredRevision string
			for _, commit := range currentFreight.Commits {
				if commit.RepoURL == app.Spec.Source.RepoURL {
					if commit.HealthCheckCommit != "" {
						desiredRevision = commit.HealthCheckCommit
					} else {
						desiredRevision = commit.ID
					}
				}
			}
			if desiredRevision == "" {
				for _, chart := range currentFreight.Charts {
					if chart.RegistryURL == app.Spec.Source.RepoURL &&
						chart.Name == app.Spec.Source.Chart {
						desiredRevision = chart.Version
					}
				}
			}
			if healthy, reason := libArgoCD.IsApplicationHealthyAndSynced(
				app,
				desiredRevision,
			); !healthy {
				health.Status = kargoapi.HealthStateUnhealthy
				health.Issues = append(health.Issues, reason)
			}
		}
	}

	return health
}
