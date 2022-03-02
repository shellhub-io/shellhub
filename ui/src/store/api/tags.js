import http from '@/store/helpers/http';

export const updateTag = async (data) => http().put(`tags/${data.oldTag}`, { name: data.newTag });

export const removeTag = async (name) => http().delete(`tags/${name}`);

export const getTags = async () => http().get('/tags');
