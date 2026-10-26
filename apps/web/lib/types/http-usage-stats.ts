export type UsageSource = "all" | "native" | "external";

export type UsageTotalsDTO = {
  known_tokens_total: number;
  tokens_total: number | null;
  run_count: number;
  native_event_count: number;
  incomplete_run_count: number;
  incomplete_native_event_count: number;
  review_iterations: number;
  qa_runs: number;
  worker_duration_ms: number;
};

export type UsageStatsDTO = {
  totals: UsageTotalsDTO;
  daily: (UsageTotalsDTO & { date: string })[];
  tasks: (UsageTotalsDTO & { task_id: string; title: string; state: string })[];
  providers: (UsageTotalsDTO & {
    source: Exclude<UsageSource, "all">;
    provider: string;
    model: string;
  })[];
  next_task_cursor: string;
  next_provider_cursor: string;
  daily_from: string;
  daily_until: string;
};
