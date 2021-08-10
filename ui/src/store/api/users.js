import http from '@/store/helpers/http';

export const signUp = async (data) => http().post('/register', {
  name: data.name,
  email: data.email,
  username: data.username,
  password: data.password,
});

export const postResendEmail = async (email) => http().post('/user/resend_email', {
  email,
});

export const postRecoverPassword = async (email) => http().post('/user/recover_password', {
  email,
});

export const patchUserData = async (data) => http().patch(`/users/${data.id}/data`, {
  name: data.name,
  username: data.username,
  email: data.email,
});

export const patchUserPassword = async (data) => http().patch(`/users/${data.id}/password`, {
  currentPassword: data.currentPassword,
  newPassword: data.newPassword,
});

export const putSecurity = async (data) => http().put(`/users/security/${data.id}`, {
  session_record: data.status,
});

export const getSecurity = async () => http().get('/users/security');
