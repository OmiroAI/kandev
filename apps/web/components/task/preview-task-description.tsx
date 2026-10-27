import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { TaskDescriptionEditable } from "./simple/components/task-description-editable";

export function PreviewTaskDescription({
  taskId,
  description,
  children,
  fallback,
}: {
  taskId: string;
  description?: string;
  children?: ReactNode;
  fallback: ReactNode;
}) {
  if (!description?.trim()) return fallback;
  return (
    <div
      className="flex min-h-0 flex-1 flex-col overflow-y-auto px-4 pb-4"
      data-testid="preview-task-description"
    >
      <TaskDescriptionEditable taskId={taskId} description={description} sessions={[]} />
      {children}
    </div>
  );
}

export function PreviewEmptyState() {
  const { t } = useTranslation();
  return (
    <div className="flex h-full flex-col">
      <div
        className="flex flex-1 items-center justify-center text-sm text-muted-foreground"
        data-testid="preview-empty-state"
      >
        {t("task:noAgentsYet2")}
      </div>
    </div>
  );
}
