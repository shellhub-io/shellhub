import { adminApi } from "./../../api/http";

const getStats = async () => adminApi.getStats();

export default getStats;
