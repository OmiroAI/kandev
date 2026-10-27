package lifecycle

import agentruntime "github.com/kandev/kandev/internal/agentruntime"

func (m *Manager) racedExecutionSharesSSHProcess(instance *ExecutorInstance, execution *AgentExecution) bool {
	if instance.RuntimeName != agentruntime.RuntimeSSH {
		return false
	}
	winner, ok := m.executionStore.GetBySessionID(execution.SessionID)
	if !ok || winner.ID == execution.ID || winner.TaskID != execution.TaskID || winner.RuntimeName != agentruntime.RuntimeSSH {
		return false
	}
	metadata := winner.MetadataSnapshot()
	for _, key := range []string{MetadataKeySSHHost, MetadataKeySSHPort, MetadataKeySSHUser,
		MetadataKeySSHRemoteAgentctlPID, MetadataKeySSHRemoteAgentctlPort, MetadataKeySSHRemoteSessionDir} {
		value := getMetadataString(instance.Metadata, key)
		if value == "" || value != getMetadataString(metadata, key) {
			return false
		}
	}
	return true
}
