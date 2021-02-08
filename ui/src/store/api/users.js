import http from '@/store/helpers/http';

export const putUser = async (data) => http().put(`/users/${data.id}`, {
  username: data.username,
  email: data.email,
  currentPassword: data.currentPassword,
  newPassword: data.newPassword,
});

export const putSecurity = async (data) => http().put(`/users/security/${data.id}`, {
  session_record: data.status,
});

export const getSecurity = async () => http().get('/users/security');
