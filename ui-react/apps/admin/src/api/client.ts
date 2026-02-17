import axios from "axios";
import { setupInterceptors } from "./interceptors";

const apiClient = axios.create({
  baseURL: `${window.location.protocol}//${window.location.host}`,
});

setupInterceptors(apiClient);

export default apiClient;
