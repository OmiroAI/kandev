import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@kandev/ui/button";
import { PanelLoadingState } from "@/components/panel-loading-state";

export function PreviewTaskDetailState({
  loading,
  error,
  onRetry,
  children,
}: {
  loading: boolean;
  error: boolean;
  onRetry: () => void;
  children: ReactNode;
}) {
  const { t } = useTranslation();
  if (loading)
    return (
      <PanelLoadingState testId="preview-task-details-loading" label={t("common:loadingTask")} />
    );
  if (error)
    return (
      <div role="alert" className="flex flex-col items-center gap-3 p-4">
        <span>{t("common:requestFailed")}</span>
        <Button
          variant="outline"
          size="sm"
          onClick={onRetry}
          className="cursor-pointer [@media(pointer:coarse)]:min-h-11 [@media(pointer:coarse)]:min-w-11"
        >
          {t("common:retryPreview")}
        </Button>
      </div>
    );
  return children;
}

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
