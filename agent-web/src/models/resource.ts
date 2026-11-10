export type Resource = {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  taskId: number | null;
  name: string;
  size: number;
  mimeType: string;
  sha256: string;
  state: string;
};

export type ResourceSource = {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  resourceId: number;
  mcpServerCode: string;
  externalId: string;
  sourceName: string;
  sourceReference: string;
  metadata: string;
};

export type ResourceDetail = Resource & { sources: ResourceSource[] };

export type AgentResourceTask = {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  candidateId: number;
  mcpServerCode: string;
  externalTaskId: string;
  state: "queued" | "running" | "completed" | "failed" | "cancelled";
  stage: string;
  progressPercent: number;
  progressDone: number;
  progressTotal: number;
  progressUnit: string;
  errorCode: string;
  errorMessage: string;
  startedAt: string | null;
  completedAt: string | null;
};
