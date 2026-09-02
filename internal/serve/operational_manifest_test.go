package serve

import (
	"reflect"
	"testing"

	"github.com/practable/book/internal/serve/models"
)

func TestOperationalTemplatesRoundTripThroughAPIModel(t *testing.T) {
	workflow, timeout := "video-health", "4s"
	attempts := int64(3)
	actionType, actionLabel := "contact_support", "Contact the lab"
	stageName, recoveryStageName, cleanupStageName, templateName := "check-video", "reset-video", "stop-video", "standard-video-check"
	recoveryAttempts := int64(1)
	pipelineName := "standard-video-activation"
	manifest := models.Manifest{
		OperationalJobTemplates: map[string]models.OperationalJobTemplate{
			templateName: {
				Workflow: &workflow, Timeout: &timeout,
				Parameters: map[string]string{"mode": "video"}, AllowedOverrides: map[string]string{"stream": "string"},
				Retry:            &models.OperationalRetryPolicy{Attempts: &attempts, InitialDelay: "250ms", Backoff: 2, MaximumDelay: "1s", TotalTimeout: "4s", RetryableCodes: []string{"not_ready"}},
				ProgressMessages: &models.OperationalProgressMessages{Initial: "Starting video", Retry: "Video is not ready; retrying"},
				FailureGuidance:  &models.OperationalFailureGuidance{Title: "Video unavailable", Message: "Choose another experiment or contact the lab.", Actions: []*models.OperationalFailureAction{{Type: &actionType, Label: &actionLabel, URL: "https://support.example.test"}}},
			},
		},
		OperationalPipelineTemplates: map[string]models.OperationalPipelineTemplate{
			pipelineName: {
				Stages:           []*models.OperationalPipelineStage{{Name: &stageName, JobTemplate: &templateName, WaitAfter: "500ms"}},
				Recovery:         []*models.OperationalPipelineStage{{Name: &recoveryStageName, JobTemplate: &templateName, WaitAfter: "250ms"}},
				RecoveryAttempts: &recoveryAttempts,
				Cleanup:          []*models.OperationalPipelineStage{{Name: &cleanupStageName, JobTemplate: &templateName}},
			},
		},
	}
	stored, err := convertModelsManifestToStore(manifest)
	if err != nil {
		t.Fatal(err)
	}
	exportedJobs, exportedPipelines := storeOperationalTemplatesToModel(stored)
	if !reflect.DeepEqual(exportedJobs, manifest.OperationalJobTemplates) {
		t.Fatalf("job templates changed during round trip\n got: %#v\nwant: %#v", exportedJobs, manifest.OperationalJobTemplates)
	}
	if !reflect.DeepEqual(exportedPipelines, manifest.OperationalPipelineTemplates) {
		t.Fatalf("pipeline templates changed during round trip\n got: %#v\nwant: %#v", exportedPipelines, manifest.OperationalPipelineTemplates)
	}
}

func TestStreamOperationBindingRoundTripThroughAPIModel(t *testing.T) {
	pipeline := "standard-video-activation"
	input := map[string]models.OperationalStreamBinding{"video": {ActivationPipeline: &pipeline, Parameters: map[string]models.OperationalParameterBinding{"stream": {From: "resource.properties.video_stream"}}}}
	stored := modelStreamOperationsToStore(input)
	exported := storeStreamOperationsToModel(stored)
	if !reflect.DeepEqual(exported, input) {
		t.Fatalf("stream bindings changed during round trip\n got: %#v\nwant: %#v", exported, input)
	}
}

func TestOperatorOnlyManifestStreamConvertsToStore(t *testing.T) {
	connection, purpose, topic, endpoint := "session", "control", "prototype", "https://relay.example.test"
	manifest := models.Manifest{Streams: map[string]models.ManifestStream{"control": {
		ConnectionType: &connection, For: &purpose, Scopes: []string{"read", "write"}, Topic: &topic, URL: &endpoint, OperatorOnly: true,
	}}}
	stored, err := convertModelsManifestToStore(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if !stored.Streams["control"].OperatorOnly {
		t.Fatal("operator-only stream flag was discarded")
	}
}
