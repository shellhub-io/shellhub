import http from '@/store/helpers/http';

export const putUser = async (data) => http().put(`/users/${data.id}`, {
  username: data.username,
  email: data.email,
  currentPassword: data.currentPassword,
  newPassword: data.newPassword,
});

export const putSecurity = async (status) => http().put('/user/security', {
  session_record: status,
});

export const getSecurity = async () => http().get('/user/security');
