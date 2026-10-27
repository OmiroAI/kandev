import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";

export function PreviewTaskDescription({
  description,
  children,
  fallback,
}: {
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
      <p className="mt-4 whitespace-pre-wrap break-words text-sm">{description}</p>
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
