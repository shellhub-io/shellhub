import axios from "axios";
import { adminApi } from "@/api/http";

const getLicense = async () => adminApi.getLicense();

const uploadLicense = async (file: File) => {
  const formData = new FormData();
  formData.append("file", file);

  const token = localStorage.getItem("cloud_token");

  const response = await axios.post(
    `${window.location.origin}/admin/api/license`,
    formData,
    {
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "multipart/form-data",
      },
    },
  );

  return response;
};

export { getLicense, uploadLicense };
