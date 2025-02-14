/* eslint-disable import/prefer-default-export */
export const envVariables = {
  announcementsEnable: process.env.SHELLHUB_ANNOUNCEMENTS === "true",
  tinyMceKey: process.env.TINYMCE_API_KEY,
};
