import Axios from "axios";
// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore //* use while route mapping is not in TypeScript
export default () => {
  const axios = Axios.create({
    baseURL: `${window.location.protocol}//${window.location.host}/admin`,
    headers: {
      Authorization: `Bearer ${localStorage.getItem("cloud_token")}`,
    },
  });

  return axios;
};
