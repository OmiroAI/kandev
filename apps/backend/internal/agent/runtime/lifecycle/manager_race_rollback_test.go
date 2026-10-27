package lifecycle

import (
	"context"
	"testing"

	"github.com/kandev/kandev/internal/agent/executor"
	agentruntime "github.com/kandev/kandev/internal/agentruntime"
	"github.com/stretchr/testify/require"
)

type rollbackSSHRecorder struct {
	MockExecutor
	stopped *ExecutorInstance
	forced  bool
}

func (r *rollbackSSHRecorder) StopInstance(_ context.Context, instance *ExecutorInstance, force bool) error {
	r.stopped = instance
	r.forced = force
	return nil
}

func TestRacedSSHExecutionPreservesWinningRemoteProcess(t *testing.T) {
	for _, scenario := range []string{"shared", "different-pid", "missing-host", "no-winner"} {
		t.Run(scenario, func(t *testing.T) {
			mgr, _ := newEnvironmentExecutionTestManager(t, &mockWorkspaceInfoProvider{})
			metadata := map[string]interface{}{
				MetadataKeySSHHost: "example.invalid", MetadataKeySSHPort: "22", MetadataKeySSHUser: "worker",
				MetadataKeySSHRemoteAgentctlPID: "123", MetadataKeySSHRemoteAgentctlPort: "41001",
				MetadataKeySSHRemoteSessionDir: "/tmp/example-session", MetadataKeyReuseExistingProcess: true,
			}
			winner := &AgentExecution{ID: "winner", TaskID: "task", SessionID: "session", RuntimeName: agentruntime.RuntimeSSH}
			winner.SetMetadataForTesting(metadata)
			if scenario != "no-winner" {
				require.NoError(t, mgr.executionStore.Add(winner))
			}
			instance := &ExecutorInstance{InstanceID: "loser", TaskID: "task", SessionID: "session", RuntimeName: agentruntime.RuntimeSSH, Metadata: cloneSSHMetadata(metadata)}
			if scenario == "different-pid" {
				instance.Metadata[MetadataKeySSHRemoteAgentctlPID] = "456"
			}
			if scenario == "missing-host" {
				delete(instance.Metadata, MetadataKeySSHHost)
			}
			recorder := &rollbackSSHRecorder{MockExecutor: MockExecutor{name: executor.NameSSH}}
			loser := &AgentExecution{ID: "loser", TaskID: "task", SessionID: "session", RuntimeName: agentruntime.RuntimeSSH}
			mgr.rollbackRacedExecution(context.Background(), recorder, instance, loser)
			require.NotNil(t, recorder.stopped)
			require.Equal(t, "loser", recorder.stopped.InstanceID)
			if scenario == "shared" {
				require.False(t, recorder.forced, "rollback must detach its own transport without killing the winner")
				require.Equal(t, StopReasonBackendShutdown, recorder.stopped.StopReason)
				require.False(t, sshShouldStopRemoteAgentctl(recorder.stopped, recorder.forced))
				require.Empty(t, instance.StopReason, "do not mutate the caller's runtime identity")
			} else {
				require.True(t, recorder.forced, "unowned runtime processes still require rollback cleanup")
			}
			if scenario != "no-winner" {
				current, ok := mgr.executionStore.GetBySessionID("session")
				require.True(t, ok)
				require.Same(t, winner, current)
			}
		})
	}
}
