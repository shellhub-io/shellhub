import { adminApi } from "@admin/api/http";

const getAdminStats = async () => adminApi.getStats();

export default getAdminStats;
