package serve

import (
	"fmt"
	"time"

	"github.com/practable/book/internal/serve/models"
	"github.com/practable/book/internal/store"
)

func parseOptionalDuration(value, field string) (time.Duration, error) {
	if value == "" {
		return 0, nil
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("error parsing %s: %w", field, err)
	}
	return d, nil
}

func modelOperationalTemplatesToStore(mm models.Manifest) (map[string]store.OperationalJobTemplate, map[string]store.OperationalPipelineTemplate, error) {
	jobs := make(map[string]store.OperationalJobTemplate, len(mm.OperationalJobTemplates))
	for name, value := range mm.OperationalJobTemplates {
		if value.Workflow == nil || value.Timeout == nil {
			return nil, nil, fmt.Errorf("operational job template %s is missing workflow or timeout", name)
		}
		timeout, err := time.ParseDuration(*value.Timeout)
		if err != nil {
			return nil, nil, fmt.Errorf("error parsing timeout in operational job template %s: %w", name, err)
		}
		item := store.OperationalJobTemplate{Workflow: *value.Workflow, Timeout: timeout, Parameters: value.Parameters, AllowedOverrides: value.AllowedOverrides}
		if value.Retry != nil {
			initial, err := parseOptionalDuration(value.Retry.InitialDelay, "initial_delay in operational job template "+name)
			if err != nil {
				return nil, nil, err
			}
			maximum, err := parseOptionalDuration(value.Retry.MaximumDelay, "maximum_delay in operational job template "+name)
			if err != nil {
				return nil, nil, err
			}
			total, err := parseOptionalDuration(value.Retry.TotalTimeout, "total_timeout in operational job template "+name)
			if err != nil {
				return nil, nil, err
			}
			attempts := 0
			if value.Retry.Attempts != nil {
				attempts = int(*value.Retry.Attempts)
			}
			item.Retry = store.OperationalRetryPolicy{Attempts: attempts, InitialDelay: initial, Backoff: value.Retry.Backoff, MaximumDelay: maximum, TotalTimeout: total, RetryableCodes: value.Retry.RetryableCodes}
		}
		if value.ProgressMessages != nil {
			item.ProgressMessages = store.OperationalProgressMessages{Initial: value.ProgressMessages.Initial, Retry: value.ProgressMessages.Retry}
		}
		if value.FailureGuidance != nil {
			item.FailureGuidance.Title, item.FailureGuidance.Message = value.FailureGuidance.Title, value.FailureGuidance.Message
			for _, action := range value.FailureGuidance.Actions {
				if action == nil || action.Type == nil || action.Label == nil {
					continue
				}
				item.FailureGuidance.Actions = append(item.FailureGuidance.Actions, store.OperationalFailureAction{Type: *action.Type, Label: *action.Label, URL: action.URL})
			}
		}
		jobs[name] = item
	}
	pipelines := make(map[string]store.OperationalPipelineTemplate, len(mm.OperationalPipelineTemplates))
	convertStages := func(values []*models.OperationalPipelineStage, field string) ([]store.OperationalPipelineStage, error) {
		if values == nil {
			return nil, nil
		}
		result := make([]store.OperationalPipelineStage, 0, len(values))
		for _, value := range values {
			if value == nil || value.Name == nil || value.JobTemplate == nil {
				return nil, fmt.Errorf("%s contains an incomplete stage", field)
			}
			wait, err := parseOptionalDuration(value.WaitAfter, "wait_after in "+field)
			if err != nil {
				return nil, err
			}
			result = append(result, store.OperationalPipelineStage{Name: *value.Name, JobTemplate: *value.JobTemplate, WaitAfter: wait})
		}
		return result, nil
	}
	for name, value := range mm.OperationalPipelineTemplates {
		stages, err := convertStages(value.Stages, "operational pipeline "+name)
		if err != nil {
			return nil, nil, err
		}
		recovery, err := convertStages(value.Recovery, "operational pipeline "+name+" recovery")
		if err != nil {
			return nil, nil, err
		}
		cleanup, err := convertStages(value.Cleanup, "operational pipeline "+name+" cleanup")
		if err != nil {
			return nil, nil, err
		}
		recoveryAttempts := 0
		if value.RecoveryAttempts != nil {
			recoveryAttempts = int(*value.RecoveryAttempts)
		}
		pipelines[name] = store.OperationalPipelineTemplate{Stages: stages, Recovery: recovery, RecoveryAttempts: recoveryAttempts, Cleanup: cleanup}
	}
	return jobs, pipelines, nil
}

func modelStreamOperationsToStore(values map[string]models.OperationalStreamBinding) map[string]store.OperationalStreamBinding {
	result := make(map[string]store.OperationalStreamBinding, len(values))
	for name, value := range values {
		pipeline := ""
		if value.ActivationPipeline != nil {
			pipeline = *value.ActivationPipeline
		}
		parameters := make(map[string]store.OperationalParameterBinding, len(value.Parameters))
		for key, binding := range value.Parameters {
			parameters[key] = store.OperationalParameterBinding{Value: binding.Value, From: binding.From}
		}
		result[name] = store.OperationalStreamBinding{ActivationPipeline: pipeline, Parameters: parameters}
	}
	return result
}

func storeOperationalTemplatesToModel(sm store.Manifest) (map[string]models.OperationalJobTemplate, map[string]models.OperationalPipelineTemplate) {
	jobs := make(map[string]models.OperationalJobTemplate, len(sm.OperationalJobTemplates))
	for name, value := range sm.OperationalJobTemplates {
		workflow, timeout := value.Workflow, value.Timeout.String()
		item := models.OperationalJobTemplate{Workflow: &workflow, Timeout: &timeout, Parameters: value.Parameters, AllowedOverrides: value.AllowedOverrides}
		if value.Retry.Attempts != 0 || value.Retry.InitialDelay != 0 || value.Retry.Backoff != 0 || value.Retry.MaximumDelay != 0 || value.Retry.TotalTimeout != 0 || len(value.Retry.RetryableCodes) > 0 {
			attempts := int64(value.Retry.Attempts)
			item.Retry = &models.OperationalRetryPolicy{Attempts: &attempts, Backoff: value.Retry.Backoff, RetryableCodes: value.Retry.RetryableCodes}
			if value.Retry.InitialDelay != 0 {
				item.Retry.InitialDelay = value.Retry.InitialDelay.String()
			}
			if value.Retry.MaximumDelay != 0 {
				item.Retry.MaximumDelay = value.Retry.MaximumDelay.String()
			}
			if value.Retry.TotalTimeout != 0 {
				item.Retry.TotalTimeout = value.Retry.TotalTimeout.String()
			}
		}
		if value.ProgressMessages != (store.OperationalProgressMessages{}) {
			item.ProgressMessages = &models.OperationalProgressMessages{Initial: value.ProgressMessages.Initial, Retry: value.ProgressMessages.Retry}
		}
		if value.FailureGuidance.Title != "" || value.FailureGuidance.Message != "" || len(value.FailureGuidance.Actions) > 0 {
			guidance := &models.OperationalFailureGuidance{Title: value.FailureGuidance.Title, Message: value.FailureGuidance.Message}
			for _, action := range value.FailureGuidance.Actions {
				actionType, label := action.Type, action.Label
				guidance.Actions = append(guidance.Actions, &models.OperationalFailureAction{Type: &actionType, Label: &label, URL: action.URL})
			}
			item.FailureGuidance = guidance
		}
		jobs[name] = item
	}
	pipelines := make(map[string]models.OperationalPipelineTemplate, len(sm.OperationalPipelineTemplates))
	convertStages := func(values []store.OperationalPipelineStage) []*models.OperationalPipelineStage {
		if values == nil {
			return nil
		}
		result := make([]*models.OperationalPipelineStage, 0, len(values))
		for _, value := range values {
			name, template := value.Name, value.JobTemplate
			item := &models.OperationalPipelineStage{Name: &name, JobTemplate: &template}
			if value.WaitAfter != 0 {
				item.WaitAfter = value.WaitAfter.String()
			}
			result = append(result, item)
		}
		return result
	}
	for name, value := range sm.OperationalPipelineTemplates {
		item := models.OperationalPipelineTemplate{Stages: convertStages(value.Stages), Recovery: convertStages(value.Recovery), Cleanup: convertStages(value.Cleanup)}
		if value.RecoveryAttempts != 0 {
			recoveryAttempts := int64(value.RecoveryAttempts)
			item.RecoveryAttempts = &recoveryAttempts
		}
		pipelines[name] = item
	}
	return jobs, pipelines
}

func storeStreamOperationsToModel(values map[string]store.OperationalStreamBinding) map[string]models.OperationalStreamBinding {
	result := make(map[string]models.OperationalStreamBinding, len(values))
	for name, value := range values {
		pipeline := value.ActivationPipeline
		parameters := make(map[string]models.OperationalParameterBinding, len(value.Parameters))
		for key, binding := range value.Parameters {
			parameters[key] = models.OperationalParameterBinding{Value: binding.Value, From: binding.From}
		}
		result[name] = models.OperationalStreamBinding{ActivationPipeline: &pipeline, Parameters: parameters}
	}
	return result
}
