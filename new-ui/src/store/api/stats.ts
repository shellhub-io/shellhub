import { devicesApi } from "../../api/http";

const getStats = async () => devicesApi.getStatusDevices();

export default getStats;
