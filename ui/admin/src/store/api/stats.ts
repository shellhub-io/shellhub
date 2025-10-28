import { adminApi } from "@/api/http";

const getAdminStats = async () => adminApi.getStats();

export default getAdminStats;
