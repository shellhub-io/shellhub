import http from '@/store/helpers/http';

export const putUser = async (data) => http().put('/user', {
  username: data.username,
  email: data.email,
  currentPassword: data.currentPassword,
  newPassword: data.newPassword,
});

export const putSecurity = async (status) => http().put('/user/security', {
  sessionRecord: status,
});

export const getSecurity = async () => http().get('/user/security');
