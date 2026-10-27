import type { ComponentProps } from "react";
import type { TaskOverview } from "@/lib/state/slices/task-overview-types";
import { usePreviewTaskDetails } from "@/hooks/domains/task/use-preview-task-details";
import { PreviewSessionTabs } from "./preview-session-tabs";
import { PreviewTaskDetailState } from "./preview-task-description";

type Props = Omit<ComponentProps<typeof PreviewSessionTabs>, "taskId" | "description"> & {
  task: Pick<TaskOverview, "id" | "description" | "primarySessionId">;
};

export function PreviewTaskDetails({ task, sessionId, ...props }: Props) {
  const details = usePreviewTaskDetails(task.id, !sessionId && !task.primarySessionId);
  return (
    <PreviewTaskDetailState loading={details.loading} error={details.error} onRetry={details.retry}>
      <PreviewSessionTabs
        {...props}
        taskId={task.id}
        sessionId={sessionId}
        description={details.description ?? task.description}
      />
    </PreviewTaskDetailState>
  );
}
