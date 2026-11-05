export type Resource = {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  taskId: number;
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
