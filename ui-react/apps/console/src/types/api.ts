export interface PaginatedResponse<T> {
  data: T[];
  totalCount: number;
}
