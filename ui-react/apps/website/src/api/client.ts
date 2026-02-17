import axios from "axios";

const apiClient = axios.create({
  baseURL: `${window.location.protocol}//${window.location.host}`,
});

export default apiClient;
