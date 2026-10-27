package requiredstores

import "testing"

func TestTaskRequiresExternalActivityLedger(t *testing.T) {
	for _, descriptor := range Catalog() {
		if descriptor.ID != "task" {
			continue
		}
		for _, table := range descriptor.RequiredTables {
			if table == "task_external_activity" {
				return
			}
		}
	}
	t.Fatal("task readiness omits external activity ledger")
}
