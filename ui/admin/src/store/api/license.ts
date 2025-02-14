import axios from "axios";
import { adminApi } from "./../../api/http";

const getLicense = async () => adminApi.getLicense();

const uploadLicense = async (file: File) => {
  const form = new FormData();
  form.append("file", file);
  const postLicense = await axios.post(`${window.location.protocol}//${window.location.host}/admin/api/license`, form, {
    headers: {
      Authorization: `Bearer ${localStorage.getItem("cloud_token")}`,
    },
  });
  return postLicense;
};

export { getLicense, uploadLicense };
